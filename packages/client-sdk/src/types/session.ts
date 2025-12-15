/**
 * Session management types
 */

/** Session entity */
export interface Session {
  id: string;
  user_id: string;
  device_type: string;
  os: string;
  browser: string;
  session_name: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  last_active_at: string;
  is_current?: boolean;
}

/** Session list response */
export interface SessionListResponse {
  sessions: Session[];
  total: number;
  page: number;
  per_page?: number;
}

/** Session statistics (admin) */
export interface SessionStats {
  total_active: number;
  total_today: number;
  average_session_duration: number;
  top_locations: SessionLocationStats[];
}

/** Session location statistics */
export interface SessionLocationStats {
  country: string;
  city: string;
  count: number;
}
