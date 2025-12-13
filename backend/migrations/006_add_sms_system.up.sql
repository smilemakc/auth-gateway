-- Add phone column to otps table
ALTER TABLE otps ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- Create index on phone for faster lookups
CREATE INDEX IF NOT EXISTS idx_otps_phone ON otps(phone);

-- Create composite index for phone + type queries
CREATE INDEX IF NOT EXISTS idx_otps_phone_type ON otps(phone, type) WHERE used = FALSE;

-- Modify the existing unique constraint to handle both email and phone
-- Drop the old constraint if it exists
DROP INDEX IF EXISTS otps_email_type_idx;

-- Create new partial unique constraints
CREATE UNIQUE INDEX IF NOT EXISTS idx_otps_email_type_unique ON otps(email, type) WHERE email IS NOT NULL AND used = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_otps_phone_type_unique ON otps(phone, type) WHERE phone IS NOT NULL AND used = FALSE;

-- Create SMS settings table
CREATE TABLE IF NOT EXISTS sms_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('twilio', 'aws_sns', 'vonage', 'mock')),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    account_sid VARCHAR(255),
    auth_token VARCHAR(255),
    from_number VARCHAR(20),
    aws_region VARCHAR(50),
    aws_access_key_id VARCHAR(255),
    aws_secret_access_key VARCHAR(255),
    aws_sender_id VARCHAR(50),
    max_per_hour INTEGER NOT NULL DEFAULT 10,
    max_per_day INTEGER NOT NULL DEFAULT 50,
    max_per_number INTEGER NOT NULL DEFAULT 5,
    created_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Create index on enabled settings
CREATE INDEX IF NOT EXISTS idx_sms_settings_enabled ON sms_settings(enabled);

-- Create SMS logs table
CREATE TABLE IF NOT EXISTS sms_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('verification', 'password_reset', '2fa', 'login')),
    provider VARCHAR(50) NOT NULL,
    message_id VARCHAR(255),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'sent', 'failed', 'delivered')),
    error_message TEXT,
    sent_at TIMESTAMP,
    user_id UUID,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes on sms_logs
CREATE INDEX IF NOT EXISTS idx_sms_logs_phone ON sms_logs(phone);
CREATE INDEX IF NOT EXISTS idx_sms_logs_type ON sms_logs(type);
CREATE INDEX IF NOT EXISTS idx_sms_logs_status ON sms_logs(status);
CREATE INDEX IF NOT EXISTS idx_sms_logs_created_at ON sms_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sms_logs_user_id ON sms_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_sms_logs_phone_created ON sms_logs(phone, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sms_logs_phone_type ON sms_logs(phone, type, created_at DESC);
