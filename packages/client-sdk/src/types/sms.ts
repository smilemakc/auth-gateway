/**
 * SMS-related types
 */

import type { OTPType } from './common';
import type { User } from './user';

/** Send SMS OTP request */
export interface SendSMSRequest {
  phone: string;
  type: OTPType;
}

/** Send SMS response */
export interface SendSMSResponse {
  success: boolean;
  message_id: string;
  expires_at: string;
}

/** Verify SMS OTP request */
export interface VerifySMSRequest {
  phone: string;
  code: string;
}

/** Verify SMS OTP response */
export interface VerifySMSResponse {
  valid: boolean;
  access_token?: string;
  refresh_token?: string;
  user?: User;
}

/** SMS statistics (admin only) */
export interface SMSStatsResponse {
  total_sent: number;
  sent_today: number;
  delivery_rate: number;
  providers: SMSProviderStats[];
}

/** SMS provider statistics */
export interface SMSProviderStats {
  name: string;
  sent: number;
  delivered: number;
  failed: number;
}

/** SMS provider type */
export type SMSProvider = 'twilio' | 'aws_sns' | 'vonage' | 'mock';

/** SMS settings (admin) */
export interface SMSSettings {
  id: string;
  provider: SMSProvider;
  is_active: boolean;
  config: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

/** Create/Update SMS settings request */
export interface SMSSettingsRequest {
  provider: SMSProvider;
  is_active?: boolean;
  config: Record<string, unknown>;
}
