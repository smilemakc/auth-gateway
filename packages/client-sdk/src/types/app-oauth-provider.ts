/**
 * Application OAuth Provider types
 */

export interface ApplicationOAuthProvider {
  id: string;
  application_id: string;
  provider: string;
  client_id: string;
  client_secret?: string;
  callback_url: string;
  scopes: string[];
  auth_url: string;
  token_url: string;
  user_info_url: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateAppOAuthProviderRequest {
  provider: string;
  client_id: string;
  client_secret: string;
  callback_url: string;
  scopes?: string[];
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
  is_active?: boolean;
}

export interface UpdateAppOAuthProviderRequest {
  client_id?: string;
  client_secret?: string;
  callback_url?: string;
  scopes?: string[];
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
  is_active?: boolean;
}
