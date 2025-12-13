-- Create OTPs table for email verification and password reset
CREATE TABLE IF NOT EXISTS otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL, -- Hashed OTP code (bcrypt)
    type VARCHAR(50) NOT NULL CHECK (type IN ('verification', 'password_reset', '2fa', 'login')),
    used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Indexes for performance
    CONSTRAINT otps_email_type_idx UNIQUE (email, type, used, expires_at)
);

-- Index for cleanup of expired OTPs
CREATE INDEX idx_otps_expires_at ON otps(expires_at) WHERE used = FALSE;

-- Index for quick lookups
CREATE INDEX idx_otps_email_type ON otps(email, type) WHERE used = FALSE;

-- Add email_verified field to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP;

-- Create trigger to clean up old OTPs (optional, can be done via cron job)
CREATE OR REPLACE FUNCTION cleanup_expired_otps() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM otps WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_cleanup_expired_otps
    AFTER INSERT ON otps
    EXECUTE FUNCTION cleanup_expired_otps();

COMMENT ON TABLE otps IS 'One-time passwords for email verification, password reset, and 2FA';
COMMENT ON COLUMN otps.code IS 'Bcrypt hashed OTP code';
COMMENT ON COLUMN otps.type IS 'Type of OTP: verification, password_reset, 2fa, login';
