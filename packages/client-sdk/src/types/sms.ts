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
  messageId: string;
  expiresAt: string;
}

/** Verify SMS OTP request */
export interface VerifySMSRequest {
  phone: string;
  code: string;
}

/** Verify SMS OTP response */
export interface VerifySMSResponse {
  valid: boolean;
  accessToken?: string;
  refreshToken?: string;
  user?: User;
}

/** SMS statistics (admin only) */
export interface SMSStatsResponse {
  totalSent: number;
  sentToday: number;
  deliveryRate: number;
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
  isActive: boolean;
  config: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

/** Create/Update SMS settings request */
export interface SMSSettingsRequest {
  provider: SMSProvider;
  isActive?: boolean;
  config: Record<string, unknown>;
}
