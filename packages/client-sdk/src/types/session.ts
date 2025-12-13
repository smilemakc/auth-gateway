/**
 * Session management types
 */

/** Session entity */
export interface Session {
  id: string;
  ipAddress: string;
  userAgent: string;
  createdAt: string;
  lastActivity: string;
  isCurrent?: boolean;
}

/** Session list response */
export interface SessionListResponse {
  sessions: Session[];
  total: number;
  page: number;
}

/** Session statistics (admin) */
export interface SessionStats {
  totalActive: number;
  totalToday: number;
  averageSessionDuration: number;
  topLocations: SessionLocationStats[];
}

/** Session location statistics */
export interface SessionLocationStats {
  country: string;
  city: string;
  count: number;
}
