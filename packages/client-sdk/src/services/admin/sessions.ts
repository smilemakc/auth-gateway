/**
 * Admin Sessions service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type { Session, SessionListResponse, SessionStats } from '../../types/session';
import { BaseService } from '../base';

/** Admin Sessions service for session management */
export class AdminSessionsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all sessions (admin)
   * @param page Page number
   * @param perPage Items per page
   * @returns List of all sessions
   */
  async list(page = 1, perPage = 50): Promise<SessionListResponse> {
    const response = await this.http.get<Session[] | SessionListResponse>(
      '/api/admin/sessions',
      { query: { page, per_page: perPage } }
    );

    // Backend may return array directly
    if (Array.isArray(response.data)) {
      return {
        sessions: response.data,
        total: response.data.length,
        page,
        per_page: perPage,
      };
    }

    return response.data;
  }

  /**
   * Get session statistics
   * @returns Session statistics
   */
  async getStats(): Promise<SessionStats> {
    const response = await this.http.get<SessionStats>('/api/admin/sessions/stats');
    return response.data;
  }

  /**
   * List sessions for a specific user
   * @param userId User ID
   * @param page Page number
   * @param perPage Items per page
   * @returns List of user's sessions
   */
  async listUserSessions(
    userId: string,
    page = 1,
    perPage = 20
  ): Promise<SessionListResponse> {
    const response = await this.http.get<SessionListResponse>(
      `/admin/users/${userId}/sessions`,
      { query: { page, per_page: perPage } }
    );
    return response.data;
  }

  /**
   * Revoke a specific session
   * @param sessionId Session ID
   * @returns Success message
   */
  async revokeSession(sessionId: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/sessions/${sessionId}`
    );
    return response.data;
  }

  /**
   * Revoke all sessions for a user
   * @param userId User ID
   * @returns Success message
   */
  async revokeUserSessions(userId: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/admin/users/${userId}/sessions/revoke-all`
    );
    return response.data;
  }

  /**
   * Alias for revokeSession
   */
  async revoke(sessionId: string): Promise<MessageResponse> {
    return this.revokeSession(sessionId);
  }

  /**
   * Alias for revokeUserSessions
   */
  async revokeAllForUser(userId: string): Promise<MessageResponse> {
    return this.revokeUserSessions(userId);
  }

  /**
   * Get a session by ID
   */
  async get(sessionId: string): Promise<Session> {
    const response = await this.http.get<Session>(`/admin/sessions/${sessionId}`);
    return response.data;
  }

  /**
   * Get all active sessions for a user
   * @param userId User ID
   * @returns All user sessions
   */
  async getAllUserSessions(userId: string): Promise<Session[]> {
    const sessions: Session[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.listUserSessions(userId, page, 100);
      sessions.push(...response.sessions);
      hasMore = sessions.length < response.total;
      page++;
    }

    return sessions;
  }
}
