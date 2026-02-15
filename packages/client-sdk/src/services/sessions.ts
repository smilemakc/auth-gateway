/**
 * Sessions service for managing user sessions
 */

import type { HttpClient } from '../core/http';
import type { MessageResponse } from '../types/common';
import type { Session, SessionListResponse } from '../types/session';
import { BaseService } from './base';

/** Sessions service for managing user sessions */
export class SessionsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all sessions for the current user
   * @param page Page number
   * @param perPage Items per page
   * @returns List of sessions
   */
  async list(page = 1, perPage = 20): Promise<SessionListResponse> {
    const response = await this.http.get<SessionListResponse>('/api/sessions', {
      query: { page, page_size: perPage },
    });
    return response.data;
  }

  /**
   * Get all active sessions
   * @returns All sessions for the user
   */
  async getAll(): Promise<Session[]> {
    const sessions: Session[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.list(page, 100);
      sessions.push(...response.sessions);
      hasMore = sessions.length < response.total;
      page++;
    }

    return sessions;
  }

  /**
   * Revoke a specific session
   * @param id Session ID
   * @returns Success message
   */
  async revoke(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/api/sessions/${id}`);
    return response.data;
  }

  /**
   * Revoke all sessions except current
   * @returns Success message
   */
  async revokeAll(): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/api/sessions/revoke-all'
    );
    return response.data;
  }

  /**
   * Get the current session
   * @returns Current session or undefined if not found
   */
  async getCurrent(): Promise<Session | undefined> {
    const { sessions } = await this.list(1, 100);
    return sessions.find((s) => s.is_current);
  }

  /**
   * Get session count
   * @returns Total number of active sessions
   */
  async getCount(): Promise<number> {
    const { total } = await this.list(1, 1);
    return total;
  }
}
