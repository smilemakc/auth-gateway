-- Remove backup codes table
DROP TABLE IF EXISTS backup_codes;

-- Remove 2FA fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;
ALTER TABLE users DROP COLUMN IF EXISTS totp_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS totp_enabled_at;
