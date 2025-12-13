/**
 * WebSocket client for real-time events
 */

import type { TokenStorage } from '../config/types';

/** WebSocket event types */
export type WebSocketEventType =
  | 'session_revoked'
  | 'password_changed'
  | 'user_updated'
  | 'token_refresh_required'
  | 'maintenance_mode'
  | 'notification'
  | 'custom';

/** WebSocket message structure */
export interface WebSocketMessage<T = unknown> {
  type: WebSocketEventType;
  payload: T;
  timestamp: string;
}

/** WebSocket connection state */
export type WebSocketState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

/** WebSocket configuration */
export interface WebSocketConfig {
  /** WebSocket URL (e.g., 'wss://api.example.com/ws') */
  url: string;
  /** Token storage for authentication */
  tokenStorage: TokenStorage;
  /** Reconnection attempts (default: 5) */
  maxReconnectAttempts?: number;
  /** Initial reconnection delay in ms (default: 1000) */
  reconnectDelayMs?: number;
  /** Maximum reconnection delay in ms (default: 30000) */
  maxReconnectDelayMs?: number;
  /** Heartbeat interval in ms (default: 30000) */
  heartbeatIntervalMs?: number;
  /** Enable debug logging */
  debug?: boolean;
}

type EventCallback<T = unknown> = (message: WebSocketMessage<T>) => void;
type StateCallback = (state: WebSocketState) => void;
type ErrorCallback = (error: Error) => void;

/** WebSocket client with auto-reconnect and heartbeat */
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketConfig>;
  private state: WebSocketState = 'disconnected';
  private reconnectAttempt = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  // Event handlers
  private messageHandlers = new Map<WebSocketEventType | '*', Set<EventCallback>>();
  private stateHandlers = new Set<StateCallback>();
  private errorHandlers = new Set<ErrorCallback>();

  constructor(config: WebSocketConfig) {
    this.config = {
      maxReconnectAttempts: 5,
      reconnectDelayMs: 1000,
      maxReconnectDelayMs: 30000,
      heartbeatIntervalMs: 30000,
      debug: false,
      ...config,
    };
  }

  /** Get current connection state */
  getState(): WebSocketState {
    return this.state;
  }

  /** Check if connected */
  isConnected(): boolean {
    return this.state === 'connected' && this.ws?.readyState === WebSocket.OPEN;
  }

  /** Log debug message */
  private log(...args: unknown[]): void {
    if (this.config.debug) {
      console.log('[WebSocketClient]', ...args);
    }
  }

  /** Update and notify state change */
  private setState(state: WebSocketState): void {
    if (this.state !== state) {
      this.state = state;
      this.log(`State changed to: ${state}`);
      this.stateHandlers.forEach((handler) => handler(state));
    }
  }

  /** Connect to WebSocket server */
  async connect(): Promise<void> {
    if (this.state === 'connected' || this.state === 'connecting') {
      this.log('Already connected or connecting');
      return;
    }

    this.setState('connecting');

    try {
      // Get access token
      const token = await this.config.tokenStorage.getAccessToken();
      if (!token) {
        throw new Error('No access token available');
      }

      // Build URL with auth
      const url = new URL(this.config.url);
      url.searchParams.set('token', token);

      // Create WebSocket
      this.ws = new WebSocket(url.toString());

      // Set up event handlers
      this.ws.onopen = this.handleOpen.bind(this);
      this.ws.onclose = this.handleClose.bind(this);
      this.ws.onerror = this.handleError.bind(this);
      this.ws.onmessage = this.handleMessage.bind(this);

      // Wait for connection
      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('Connection timeout'));
        }, 10000);

        const checkConnection = (): void => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            clearTimeout(timeout);
            resolve();
          } else if (
            this.ws?.readyState === WebSocket.CLOSED ||
            this.ws?.readyState === WebSocket.CLOSING
          ) {
            clearTimeout(timeout);
            reject(new Error('Connection failed'));
          } else {
            setTimeout(checkConnection, 100);
          }
        };

        checkConnection();
      });
    } catch (error) {
      this.setState('disconnected');
      throw error;
    }
  }

  /** Disconnect from WebSocket server */
  disconnect(): void {
    this.log('Disconnecting');
    this.cleanup();
    this.ws?.close(1000, 'Client disconnect');
    this.ws = null;
    this.setState('disconnected');
  }

  /** Send message to server */
  send<T>(type: string, payload: T): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket is not connected');
    }

    const message = {
      type,
      payload,
      timestamp: new Date().toISOString(),
    };

    this.ws!.send(JSON.stringify(message));
    this.log('Sent message:', message);
  }

  /** Subscribe to a specific event type */
  on<T = unknown>(
    eventType: WebSocketEventType | '*',
    callback: EventCallback<T>
  ): () => void {
    if (!this.messageHandlers.has(eventType)) {
      this.messageHandlers.set(eventType, new Set());
    }

    this.messageHandlers.get(eventType)!.add(callback as EventCallback);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.get(eventType)?.delete(callback as EventCallback);
    };
  }

  /** Subscribe to state changes */
  onStateChange(callback: StateCallback): () => void {
    this.stateHandlers.add(callback);
    return () => {
      this.stateHandlers.delete(callback);
    };
  }

  /** Subscribe to errors */
  onError(callback: ErrorCallback): () => void {
    this.errorHandlers.add(callback);
    return () => {
      this.errorHandlers.delete(callback);
    };
  }

  /** Handle WebSocket open */
  private handleOpen(): void {
    this.log('Connection opened');
    this.setState('connected');
    this.reconnectAttempt = 0;
    this.startHeartbeat();
  }

  /** Handle WebSocket close */
  private handleClose(event: CloseEvent): void {
    this.log(`Connection closed: ${event.code} ${event.reason}`);
    this.cleanup();

    // Don't reconnect if closed normally
    if (event.code === 1000) {
      this.setState('disconnected');
      return;
    }

    // Attempt reconnection
    this.attemptReconnect();
  }

  /** Handle WebSocket error */
  private handleError(event: Event): void {
    const error = new Error('WebSocket error');
    this.log('Connection error:', event);
    this.errorHandlers.forEach((handler) => handler(error));
  }

  /** Handle incoming message */
  private handleMessage(event: MessageEvent): void {
    try {
      const message = JSON.parse(event.data) as WebSocketMessage;
      this.log('Received message:', message);

      // Handle pong (heartbeat response)
      if (message.type === ('pong' as WebSocketEventType)) {
        return;
      }

      // Notify specific handlers
      this.messageHandlers
        .get(message.type)
        ?.forEach((handler) => handler(message));

      // Notify wildcard handlers
      this.messageHandlers.get('*')?.forEach((handler) => handler(message));
    } catch (error) {
      this.log('Failed to parse message:', error);
    }
  }

  /** Start heartbeat */
  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected()) {
        this.send('ping', {});
      }
    }, this.config.heartbeatIntervalMs);
  }

  /** Attempt reconnection */
  private attemptReconnect(): void {
    if (this.reconnectAttempt >= this.config.maxReconnectAttempts) {
      this.log('Max reconnection attempts reached');
      this.setState('disconnected');
      return;
    }

    this.setState('reconnecting');
    this.reconnectAttempt++;

    // Calculate delay with exponential backoff
    const delay = Math.min(
      this.config.reconnectDelayMs * Math.pow(2, this.reconnectAttempt - 1),
      this.config.maxReconnectDelayMs
    );

    this.log(
      `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempt}/${this.config.maxReconnectAttempts})`
    );

    this.reconnectTimer = setTimeout(async () => {
      try {
        await this.connect();
      } catch {
        this.attemptReconnect();
      }
    }, delay);
  }

  /** Clean up timers */
  private cleanup(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}

/** Create WebSocket client factory */
export function createWebSocketClient(
  config: WebSocketConfig
): WebSocketClient {
  return new WebSocketClient(config);
}
