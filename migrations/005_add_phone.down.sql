-- Remove indexes
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_phone_unique;

-- Remove phone fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS phone_verified;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
