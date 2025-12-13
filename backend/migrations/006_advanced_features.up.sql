-- ============================================================
-- Migration: Advanced Features
-- Description: RBAC, Session Management, IP Filtering, Webhooks,
--              Service Accounts, Email Templates, Branding, Geo-tracking
-- ============================================================

-- ============================================================
-- 1. RBAC (Role-Based Access Control) Tables
-- ============================================================

-- Permissions table: defines all available permissions in the system
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE, -- e.g., "users.delete", "api_keys.view"
    resource VARCHAR(50) NOT NULL,      -- e.g., "users", "api_keys", "webhooks"
    action VARCHAR(50) NOT NULL,        -- e.g., "create", "read", "update", "delete"
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource, action)
);

-- Create index on resource and action for faster lookups
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- Roles table: defines roles (both system and custom roles)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,  -- e.g., "admin", "moderator", "custom_role_1"
    display_name VARCHAR(100) NOT NULL, -- e.g., "Administrator", "Content Moderator"
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE, -- System roles cannot be deleted
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Role-Permission mapping table
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- Create indexes for faster role-permission lookups
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Add role_id column to users table (will migrate existing role strings later)
ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles(id);

-- Create index on user role_id for faster lookups
CREATE INDEX idx_users_role_id ON users(role_id);

-- ============================================================
-- 2. Enhanced Session Management
-- ============================================================

-- Add session tracking fields to refresh_tokens table
ALTER TABLE refresh_tokens ADD COLUMN device_type VARCHAR(50);     -- "mobile", "desktop", "tablet"
ALTER TABLE refresh_tokens ADD COLUMN os VARCHAR(100);              -- "iOS 17.2", "Windows 11", "Ubuntu 22.04"
ALTER TABLE refresh_tokens ADD COLUMN browser VARCHAR(100);         -- "Chrome 120", "Safari 17", "Firefox 121"
ALTER TABLE refresh_tokens ADD COLUMN ip_address VARCHAR(45);       -- IPv4 or IPv6
ALTER TABLE refresh_tokens ADD COLUMN user_agent TEXT;              -- Full user agent string
ALTER TABLE refresh_tokens ADD COLUMN last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE refresh_tokens ADD COLUMN session_name VARCHAR(100);    -- User-defined session name

-- Create index for active session queries
CREATE INDEX idx_refresh_tokens_user_active ON refresh_tokens(user_id, revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_last_active ON refresh_tokens(last_active_at DESC);

-- ============================================================
-- 3. IP Whitelisting/Blacklisting
-- ============================================================

CREATE TABLE ip_filters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_cidr VARCHAR(100) NOT NULL UNIQUE,  -- IP address or CIDR range (e.g., "192.168.1.0/24")
    filter_type VARCHAR(20) NOT NULL CHECK (filter_type IN ('whitelist', 'blacklist')),
    reason TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP,                   -- NULL = never expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for IP filter lookups
CREATE INDEX idx_ip_filters_type_active ON ip_filters(filter_type, is_active);
CREATE INDEX idx_ip_filters_expires ON ip_filters(expires_at) WHERE expires_at IS NOT NULL;

-- ============================================================
-- 4. Webhooks Management
-- ============================================================

-- Webhooks table: stores webhook configurations
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,       -- For HMAC signature verification
    events JSONB NOT NULL DEFAULT '[]',      -- Array of subscribed events: ["user.created", "user.blocked"]
    headers JSONB DEFAULT '{}',              -- Custom headers to send with webhook
    is_active BOOLEAN DEFAULT TRUE,
    retry_config JSONB DEFAULT '{"max_attempts": 3, "backoff_seconds": [60, 300, 900]}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_triggered_at TIMESTAMP
);

-- Webhook delivery logs: tracks all webhook deliveries
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,        -- e.g., "user.created", "user.login"
    payload JSONB NOT NULL,                  -- Event payload
    status VARCHAR(50) NOT NULL,             -- "pending", "success", "failed"
    http_status_code INTEGER,
    response_body TEXT,
    attempts INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create indexes for webhook queries
CREATE INDEX idx_webhooks_active ON webhooks(is_active);
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status, next_retry_at);
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC);

-- ============================================================
-- 5. Service Accounts (M2M Authentication)
-- ============================================================

-- Add account_type to users table
ALTER TABLE users ADD COLUMN account_type VARCHAR(20) DEFAULT 'human' CHECK (account_type IN ('human', 'service'));

-- Service accounts won't require email verification, phone, or 2FA
-- They will primarily use API keys for authentication

-- Create index for filtering by account type
CREATE INDEX idx_users_account_type ON users(account_type);

-- ============================================================
-- 6. Email Template Management
-- ============================================================

CREATE TABLE email_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL UNIQUE,        -- "verification", "password_reset", "welcome", "2fa"
    name VARCHAR(100) NOT NULL,              -- Display name
    subject VARCHAR(200) NOT NULL,           -- Email subject line
    html_body TEXT NOT NULL,                 -- HTML version of email
    text_body TEXT,                          -- Plain text fallback
    variables JSONB DEFAULT '[]',            -- Available variables: ["username", "code", "link"]
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Email template versions (for rollback capability)
CREATE TABLE email_template_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES email_templates(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    html_body TEXT NOT NULL,
    text_body TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_template_versions_template ON email_template_versions(template_id, created_at DESC);

-- ============================================================
-- 7. Branding & Customization
-- ============================================================

CREATE TABLE branding_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    primary_color VARCHAR(7) DEFAULT '#007bff',    -- Hex color
    secondary_color VARCHAR(7) DEFAULT '#6c757d',  -- Hex color
    background_color VARCHAR(7) DEFAULT '#ffffff', -- Hex color
    custom_css TEXT,                                -- Custom CSS for login page
    company_name VARCHAR(100),
    support_email VARCHAR(255),
    terms_url VARCHAR(500),
    privacy_url VARCHAR(500),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Insert default branding settings (single row table)
INSERT INTO branding_settings (id, company_name, support_email)
VALUES ('00000000-0000-0000-0000-000000000001', 'Auth Gateway', 'support@authgateway.com');

-- ============================================================
-- 8. Geo-Distribution Tracking
-- ============================================================

-- Add geo-location fields to audit_logs
ALTER TABLE audit_logs ADD COLUMN country_code VARCHAR(2);       -- ISO 3166-1 alpha-2 (e.g., "US", "GB")
ALTER TABLE audit_logs ADD COLUMN country_name VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN city VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN latitude DECIMAL(10, 7);
ALTER TABLE audit_logs ADD COLUMN longitude DECIMAL(10, 7);

-- Create index for geo queries
CREATE INDEX idx_audit_logs_country ON audit_logs(country_code);
CREATE INDEX idx_audit_logs_location ON audit_logs(latitude, longitude) WHERE latitude IS NOT NULL;

-- Login locations summary (materialized view or table for performance)
CREATE TABLE login_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    country_code VARCHAR(2) NOT NULL,
    country_name VARCHAR(100) NOT NULL,
    city VARCHAR(100),
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    login_count INTEGER DEFAULT 1,
    last_login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(country_code, city)
);

-- Create index for map queries
CREATE INDEX idx_login_locations_count ON login_locations(login_count DESC);

-- ============================================================
-- 9. System Settings & Maintenance Mode
-- ============================================================

CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    setting_type VARCHAR(50) DEFAULT 'string',  -- "string", "boolean", "integer", "json"
    is_public BOOLEAN DEFAULT FALSE,             -- Can be exposed to public API
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Insert default system settings
INSERT INTO system_settings (key, value, description, setting_type, is_public) VALUES
    ('maintenance_mode', 'false', 'Enable/disable maintenance mode', 'boolean', true),
    ('maintenance_message', 'System is under maintenance. Please try again later.', 'Message shown during maintenance', 'string', true),
    ('allow_new_registrations', 'true', 'Allow new user registrations', 'boolean', false),
    ('require_email_verification', 'true', 'Require email verification for new users', 'boolean', false),
    ('max_sessions_per_user', '10', 'Maximum concurrent sessions per user', 'integer', false),
    ('session_timeout_hours', '168', 'Session timeout in hours (default: 7 days)', 'integer', false);

-- ============================================================
-- 10. System Health Metrics
-- ============================================================

CREATE TABLE health_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20, 4) NOT NULL,
    metric_unit VARCHAR(50),                     -- "bytes", "percentage", "count", "milliseconds"
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for time-series queries
CREATE INDEX idx_health_metrics_name_time ON health_metrics(metric_name, recorded_at DESC);

-- Automatically partition by time for better performance (optional, for high-volume systems)
-- This is a simple example; production systems may use more sophisticated partitioning

-- ============================================================
-- 11. Seed Initial Permissions
-- ============================================================

-- Insert system permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    -- User management permissions
    ('users.create', 'users', 'create', 'Create new users'),
    ('users.read', 'users', 'read', 'View user information'),
    ('users.update', 'users', 'update', 'Update user information'),
    ('users.delete', 'users', 'delete', 'Delete users'),
    ('users.list', 'users', 'list', 'List all users'),

    -- Role management permissions
    ('roles.create', 'roles', 'create', 'Create new roles'),
    ('roles.read', 'roles', 'read', 'View role information'),
    ('roles.update', 'roles', 'update', 'Update roles'),
    ('roles.delete', 'roles', 'delete', 'Delete roles'),
    ('roles.list', 'roles', 'list', 'List all roles'),

    -- Permission management
    ('permissions.create', 'permissions', 'create', 'Create permissions'),
    ('permissions.read', 'permissions', 'read', 'View permissions'),
    ('permissions.update', 'permissions', 'update', 'Update permissions'),
    ('permissions.delete', 'permissions', 'delete', 'Delete permissions'),
    ('permissions.list', 'permissions', 'list', 'List all permissions'),

    -- API Key permissions
    ('api_keys.create', 'api_keys', 'create', 'Create API keys'),
    ('api_keys.read', 'api_keys', 'read', 'View API keys'),
    ('api_keys.update', 'api_keys', 'update', 'Update API keys'),
    ('api_keys.delete', 'api_keys', 'delete', 'Delete API keys'),
    ('api_keys.revoke', 'api_keys', 'revoke', 'Revoke API keys'),
    ('api_keys.list', 'api_keys', 'list', 'List all API keys'),

    -- Session management permissions
    ('sessions.read', 'sessions', 'read', 'View active sessions'),
    ('sessions.revoke', 'sessions', 'revoke', 'Revoke user sessions'),
    ('sessions.list', 'sessions', 'list', 'List all sessions'),

    -- Audit log permissions
    ('audit_logs.read', 'audit_logs', 'read', 'View audit logs'),
    ('audit_logs.list', 'audit_logs', 'list', 'List audit logs'),

    -- IP filter permissions
    ('ip_filters.create', 'ip_filters', 'create', 'Create IP filters'),
    ('ip_filters.read', 'ip_filters', 'read', 'View IP filters'),
    ('ip_filters.update', 'ip_filters', 'update', 'Update IP filters'),
    ('ip_filters.delete', 'ip_filters', 'delete', 'Delete IP filters'),
    ('ip_filters.list', 'ip_filters', 'list', 'List IP filters'),

    -- Webhook permissions
    ('webhooks.create', 'webhooks', 'create', 'Create webhooks'),
    ('webhooks.read', 'webhooks', 'read', 'View webhooks'),
    ('webhooks.update', 'webhooks', 'update', 'Update webhooks'),
    ('webhooks.delete', 'webhooks', 'delete', 'Delete webhooks'),
    ('webhooks.list', 'webhooks', 'list', 'List webhooks'),
    ('webhooks.test', 'webhooks', 'test', 'Test webhook delivery'),

    -- Email template permissions
    ('email_templates.create', 'email_templates', 'create', 'Create email templates'),
    ('email_templates.read', 'email_templates', 'read', 'View email templates'),
    ('email_templates.update', 'email_templates', 'update', 'Update email templates'),
    ('email_templates.delete', 'email_templates', 'delete', 'Delete email templates'),
    ('email_templates.list', 'email_templates', 'list', 'List email templates'),

    -- Branding permissions
    ('branding.read', 'branding', 'read', 'View branding settings'),
    ('branding.update', 'branding', 'update', 'Update branding settings'),

    -- System settings permissions
    ('system.read', 'system', 'read', 'View system settings'),
    ('system.update', 'system', 'update', 'Update system settings'),
    ('system.health', 'system', 'health', 'View system health metrics'),
    ('system.maintenance', 'system', 'maintenance', 'Control maintenance mode'),

    -- Statistics permissions
    ('stats.view', 'stats', 'view', 'View system statistics'),
    ('stats.export', 'stats', 'export', 'Export statistics data');

-- ============================================================
-- 12. Seed Initial Roles
-- ============================================================

-- Insert system roles
INSERT INTO roles (id, name, display_name, description, is_system_role) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Administrator', 'Full system access with all permissions', true),
    ('00000000-0000-0000-0000-000000000002', 'moderator', 'Moderator', 'User management and moderation capabilities', true),
    ('00000000-0000-0000-0000-000000000003', 'user', 'User', 'Standard user with basic permissions', true);

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000001', id FROM permissions;

-- Assign moderate permissions to moderator role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id FROM permissions
WHERE name IN (
    'users.read', 'users.list', 'users.update',
    'sessions.read', 'sessions.list', 'sessions.revoke',
    'audit_logs.read', 'audit_logs.list',
    'api_keys.read', 'api_keys.list',
    'stats.view'
);

-- Assign basic permissions to user role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000003', id FROM permissions
WHERE name IN (
    'api_keys.create', 'api_keys.read', 'api_keys.update', 'api_keys.delete',
    'sessions.read', 'sessions.revoke',
    'branding.read'
);

-- ============================================================
-- 13. Migrate Existing Users to New Role System
-- ============================================================

-- Update existing users with role_id based on their current role string
UPDATE users SET role_id = '00000000-0000-0000-0000-000000000001' WHERE role = 'admin';
UPDATE users SET role_id = '00000000-0000-0000-0000-000000000002' WHERE role = 'moderator';
UPDATE users SET role_id = '00000000-0000-0000-0000-000000000003' WHERE role = 'user';

-- ============================================================
-- 14. Update Triggers
-- ============================================================

-- Update trigger for roles
CREATE TRIGGER update_roles_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Update trigger for email_templates
CREATE TRIGGER update_email_templates_updated_at
BEFORE UPDATE ON email_templates
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Update trigger for webhooks
CREATE TRIGGER update_webhooks_updated_at
BEFORE UPDATE ON webhooks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Update trigger for ip_filters
CREATE TRIGGER update_ip_filters_updated_at
BEFORE UPDATE ON ip_filters
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 15. Create Views for Common Queries
-- ============================================================

-- View: User roles with permissions
CREATE VIEW user_role_permissions AS
SELECT
    u.id AS user_id,
    u.username,
    u.email,
    r.id AS role_id,
    r.name AS role_name,
    r.display_name AS role_display_name,
    COALESCE(json_agg(
        json_build_object(
            'permission_id', p.id,
            'permission_name', p.name,
            'resource', p.resource,
            'action', p.action
        ) ORDER BY p.name
    ) FILTER (WHERE p.id IS NOT NULL), '[]') AS permissions
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
GROUP BY u.id, u.username, u.email, r.id, r.name, r.display_name;

-- View: Active sessions with details
CREATE VIEW active_sessions AS
SELECT
    rt.id,
    rt.user_id,
    u.username,
    u.email,
    rt.device_type,
    rt.os,
    rt.browser,
    rt.ip_address,
    rt.session_name,
    rt.last_active_at,
    rt.created_at,
    rt.expires_at
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE rt.revoked_at IS NULL
ORDER BY rt.last_active_at DESC;

-- View: Webhook statistics
CREATE VIEW webhook_stats AS
SELECT
    w.id,
    w.name,
    w.url,
    w.is_active,
    COUNT(wd.id) AS total_deliveries,
    COUNT(wd.id) FILTER (WHERE wd.status = 'success') AS successful_deliveries,
    COUNT(wd.id) FILTER (WHERE wd.status = 'failed') AS failed_deliveries,
    COUNT(wd.id) FILTER (WHERE wd.status = 'pending') AS pending_deliveries,
    MAX(wd.created_at) AS last_delivery_at,
    w.last_triggered_at
FROM webhooks w
LEFT JOIN webhook_deliveries wd ON w.id = wd.webhook_id
GROUP BY w.id, w.name, w.url, w.is_active, w.last_triggered_at;

-- View: Login geo-distribution for dashboard map
CREATE VIEW login_geo_distribution AS
SELECT
    country_code,
    country_name,
    city,
    latitude,
    longitude,
    COUNT(*) AS login_count,
    MAX(created_at) AS last_login_at
FROM audit_logs
WHERE action = 'login'
  AND status = 'success'
  AND country_code IS NOT NULL
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY country_code, country_name, city, latitude, longitude
ORDER BY login_count DESC;

-- ============================================================
-- 16. Comments
-- ============================================================

COMMENT ON TABLE permissions IS 'Defines all available permissions in the system';
COMMENT ON TABLE roles IS 'Defines user roles with dynamic permission assignments';
COMMENT ON TABLE role_permissions IS 'Maps permissions to roles (many-to-many)';
COMMENT ON TABLE ip_filters IS 'IP whitelisting and blacklisting rules';
COMMENT ON TABLE webhooks IS 'Webhook configurations for event notifications';
COMMENT ON TABLE webhook_deliveries IS 'Tracks webhook delivery attempts and responses';
COMMENT ON TABLE email_templates IS 'Customizable email templates';
COMMENT ON TABLE branding_settings IS 'Branding and customization settings (single row table)';
COMMENT ON TABLE system_settings IS 'System-wide configuration settings';
COMMENT ON TABLE health_metrics IS 'System health and performance metrics';
COMMENT ON TABLE login_locations IS 'Aggregated login location data for performance';
