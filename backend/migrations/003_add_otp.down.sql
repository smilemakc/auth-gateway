-- Drop OTP cleanup trigger
DROP TRIGGER IF EXISTS trigger_cleanup_expired_otps ON otps;
DROP FUNCTION IF EXISTS cleanup_expired_otps();

-- Remove email verification fields from users
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;

-- Drop OTP table
DROP TABLE IF EXISTS otps;
