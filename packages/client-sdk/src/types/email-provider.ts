/**
 * Email provider and profile types
 */

import type { TimestampedEntity } from './common';

/** Email provider type */
export type EmailProviderType = 'smtp' | 'sendgrid' | 'ses' | 'mailgun';

/** Email provider entity */
export interface EmailProvider extends TimestampedEntity {
  name: string;
  type: EmailProviderType;
  is_active: boolean;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_use_tls?: boolean;
  ses_region?: string;
  ses_access_key_id?: string;
  mailgun_domain?: string;
  has_smtp_password?: boolean;
  has_sendgrid_api_key?: boolean;
  has_ses_secret_access_key?: boolean;
  has_mailgun_api_key?: boolean;
}

/** Create email provider request */
export interface CreateEmailProviderRequest {
  name: string;
  type: EmailProviderType;
  is_active?: boolean;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_password?: string;
  smtp_use_tls?: boolean;
  sendgrid_api_key?: string;
  ses_region?: string;
  ses_access_key_id?: string;
  ses_secret_access_key?: string;
  mailgun_domain?: string;
  mailgun_api_key?: string;
}

/** Update email provider request */
export interface UpdateEmailProviderRequest extends Partial<CreateEmailProviderRequest> {}

/** Email profile entity */
export interface EmailProfile extends TimestampedEntity {
  name: string;
  provider_id: string;
  from_email: string;
  from_name: string;
  reply_to?: string;
  is_default: boolean;
  is_active: boolean;
  provider?: EmailProvider;
}

/** Create email profile request */
export interface CreateEmailProfileRequest {
  name: string;
  provider_id: string;
  from_email: string;
  from_name: string;
  reply_to?: string;
  is_default?: boolean;
  is_active?: boolean;
}

/** Update email profile request */
export interface UpdateEmailProfileRequest extends Partial<CreateEmailProfileRequest> {}
