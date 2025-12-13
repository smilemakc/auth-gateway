-- ============================================================
-- Rollback Migration: Advanced Features
-- ============================================================

-- Drop views
DROP VIEW IF EXISTS login_geo_distribution;
DROP VIEW IF EXISTS webhook_stats;
DROP VIEW IF EXISTS active_sessions;
DROP VIEW IF EXISTS user_role_permissions;

-- Drop health metrics table
DROP TABLE IF EXISTS health_metrics CASCADE;

-- Drop system settings table
DROP TABLE IF EXISTS system_settings CASCADE;

-- Drop login locations table
DROP TABLE IF EXISTS login_locations CASCADE;

-- Remove geo-location fields from audit_logs
ALTER TABLE audit_logs DROP COLUMN IF EXISTS country_code;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS country_name;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS city;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS latitude;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS longitude;

-- Drop branding settings table
DROP TABLE IF EXISTS branding_settings CASCADE;

-- Drop email template tables
DROP TABLE IF EXISTS email_template_versions CASCADE;
DROP TABLE IF EXISTS email_templates CASCADE;

-- Remove account_type from users table
ALTER TABLE users DROP COLUMN IF EXISTS account_type;

-- Drop webhook tables
DROP TABLE IF EXISTS webhook_deliveries CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;

-- Drop IP filters table
DROP TABLE IF EXISTS ip_filters CASCADE;

-- Remove session tracking fields from refresh_tokens
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS device_type;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS os;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS browser;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS ip_address;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS user_agent;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS last_active_at;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS session_name;

-- Remove role_id from users table
ALTER TABLE users DROP COLUMN IF EXISTS role_id;

-- Drop RBAC tables
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
