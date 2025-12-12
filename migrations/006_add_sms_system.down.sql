-- Drop SMS logs table and its indexes
DROP INDEX IF EXISTS idx_sms_logs_phone_type;
DROP INDEX IF EXISTS idx_sms_logs_phone_created;
DROP INDEX IF EXISTS idx_sms_logs_user_id;
DROP INDEX IF EXISTS idx_sms_logs_created_at;
DROP INDEX IF EXISTS idx_sms_logs_status;
DROP INDEX IF EXISTS idx_sms_logs_type;
DROP INDEX IF EXISTS idx_sms_logs_phone;
DROP TABLE IF EXISTS sms_logs;

-- Drop SMS settings table and its indexes
DROP INDEX IF EXISTS idx_sms_settings_enabled;
DROP TABLE IF EXISTS sms_settings;

-- Drop new otps indexes
DROP INDEX IF EXISTS idx_otps_phone_type_unique;
DROP INDEX IF EXISTS idx_otps_email_type_unique;
DROP INDEX IF EXISTS idx_otps_phone_type;
DROP INDEX IF EXISTS idx_otps_phone;

-- Remove phone column from otps table
ALTER TABLE otps DROP COLUMN IF EXISTS phone;

-- Recreate original unique constraint for email
CREATE UNIQUE INDEX IF NOT EXISTS otps_email_type_idx ON otps(email, type, used, expires_at);
