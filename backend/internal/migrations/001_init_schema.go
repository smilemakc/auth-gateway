package migrations

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating initial schema...")

		// ============================================================
		// 1. Enable UUID extension
		// ============================================================
		_, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
		if err != nil {
			return fmt.Errorf("failed to create uuid extension: %w", err)
		}

		// ============================================================
		// 2. Create all tables using model structs
		// ============================================================
		tablesInOrder := []interface{}{
			// Core auth tables (users first as it's referenced by many tables)
			(*models.User)(nil),

			// RBAC tables (create join table first, then tables with m2m relationships)
			(*models.Permission)(nil),
			(*models.RolePermission)(nil),
			(*models.Role)(nil),
			(*models.UserRole)(nil),

			// Auth-related tables
			(*models.RefreshToken)(nil),
			(*models.TokenBlacklist)(nil),
			(*models.OAuthAccount)(nil),
			(*models.OTP)(nil),
			(*models.BackupCode)(nil),
			(*models.APIKey)(nil),
			(*models.AuditLog)(nil),

			// Advanced feature tables
			(*models.IPFilter)(nil),
			(*models.Webhook)(nil),
			(*models.WebhookDelivery)(nil),
			(*models.EmailTemplate)(nil),
			(*models.EmailTemplateVersion)(nil),
			(*models.BrandingSettings)(nil),
			(*models.SystemSetting)(nil),
			(*models.HealthMetric)(nil),
			(*models.SMSSettings)(nil),
			(*models.SMSLog)(nil),
			(*models.LoginLocation)(nil),
		}

		for _, model := range tablesInOrder {
			_, err := db.NewCreateTable().
				Model(model).
				IfNotExists().
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to create table for %T: %w", model, err)
			}
		}

		// ============================================================
		// 3. Create indexes
		// ============================================================

		// Users indexes
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
			"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
			"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_users_account_type ON users(account_type)",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone) WHERE phone IS NOT NULL",
			"CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)",

			// Refresh tokens indexes
			"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)",
			"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_active ON refresh_tokens(user_id, revoked_at) WHERE revoked_at IS NULL",
			"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_last_active ON refresh_tokens(last_active_at DESC)",

			// OAuth accounts indexes
			"CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_accounts_provider ON oauth_accounts(provider, provider_user_id)",

			// Token blacklist indexes
			"CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_token_blacklist_token_hash ON token_blacklist(token_hash)",

			// OTP indexes
			"CREATE INDEX IF NOT EXISTS idx_otps_expires_at ON otps(expires_at) WHERE used = FALSE",
			"CREATE INDEX IF NOT EXISTS idx_otps_email_type ON otps(email, type) WHERE used = FALSE",
			"CREATE INDEX IF NOT EXISTS idx_otps_phone ON otps(phone)",
			"CREATE INDEX IF NOT EXISTS idx_otps_phone_type ON otps(phone, type) WHERE used = FALSE",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_otps_email_type_unique ON otps(email, type) WHERE email IS NOT NULL AND used = FALSE",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_otps_phone_type_unique ON otps(phone, type) WHERE phone IS NOT NULL AND used = FALSE",

			// Backup codes indexes
			"CREATE INDEX IF NOT EXISTS idx_backup_codes_user_id ON backup_codes(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_backup_codes_used ON backup_codes(used)",

			// API keys indexes
			"CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash)",
			"CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_api_keys_created_at ON api_keys(created_at)",

			// Audit logs indexes
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_country ON audit_logs(country_code)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_location ON audit_logs(latitude, longitude) WHERE latitude IS NOT NULL",

			// RBAC indexes
			"CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action)",
			"CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id)",
			"CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id)",

			// IP filters indexes
			"CREATE INDEX IF NOT EXISTS idx_ip_filters_type_active ON ip_filters(filter_type, is_active)",
			"CREATE INDEX IF NOT EXISTS idx_ip_filters_expires ON ip_filters(expires_at) WHERE expires_at IS NOT NULL",

			// Webhooks indexes
			"CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id)",
			"CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status, next_retry_at)",
			"CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC)",

			// Email templates indexes
			"CREATE INDEX IF NOT EXISTS idx_email_template_versions_template ON email_template_versions(template_id, created_at DESC)",

			// Login locations indexes
			"CREATE INDEX IF NOT EXISTS idx_login_locations_count ON login_locations(login_count DESC)",

			// Health metrics indexes
			"CREATE INDEX IF NOT EXISTS idx_health_metrics_name_time ON health_metrics(metric_name, recorded_at DESC)",

			// SMS settings indexes
			"CREATE INDEX IF NOT EXISTS idx_sms_settings_enabled ON sms_settings(enabled)",

			// SMS logs indexes
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_phone ON sms_logs(phone)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_type ON sms_logs(type)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_status ON sms_logs(status)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_created_at ON sms_logs(created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_user_id ON sms_logs(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_phone_created ON sms_logs(phone, created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_sms_logs_phone_type ON sms_logs(phone, type, created_at DESC)",
		}

		for _, indexSQL := range indexes {
			if _, err := db.ExecContext(ctx, indexSQL); err != nil {
				return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
			}
		}

		// ============================================================
		// 4. Create functions and triggers
		// ============================================================

		// Function to update updated_at timestamp
		_, err = db.ExecContext(ctx, `
			CREATE OR REPLACE FUNCTION update_updated_at_column()
			RETURNS TRIGGER AS $$
			BEGIN
				NEW.updated_at = CURRENT_TIMESTAMP;
				RETURN NEW;
			END;
			$$ language 'plpgsql'
		`)
		if err != nil {
			return fmt.Errorf("failed to create update_updated_at_column function: %w", err)
		}

		// Function to cleanup expired OTPs
		_, err = db.ExecContext(ctx, `
			CREATE OR REPLACE FUNCTION cleanup_expired_otps() RETURNS TRIGGER AS $$
			BEGIN
				DELETE FROM otps WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql
		`)
		if err != nil {
			return fmt.Errorf("failed to create cleanup_expired_otps function: %w", err)
		}

		// Create triggers
		triggers := []string{
			"CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_oauth_accounts_updated_at BEFORE UPDATE ON oauth_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_email_templates_updated_at BEFORE UPDATE ON email_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER update_ip_filters_updated_at BEFORE UPDATE ON ip_filters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
			"CREATE TRIGGER trigger_cleanup_expired_otps AFTER INSERT ON otps EXECUTE FUNCTION cleanup_expired_otps()",
		}

		for _, triggerSQL := range triggers {
			if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
				return fmt.Errorf("failed to create trigger: %w (SQL: %s)", err, triggerSQL)
			}
		}

		// ============================================================
		// 5. Create views
		// ============================================================

		// Note: user_role_permissions view is created in migration 002
		// which handles the transition from single role to multiple roles

		// Active sessions view
		_, err = db.ExecContext(ctx, `
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
			ORDER BY rt.last_active_at DESC
		`)
		if err != nil {
			return fmt.Errorf("failed to create active_sessions view: %w", err)
		}

		// Webhook statistics view
		_, err = db.ExecContext(ctx, `
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
			GROUP BY w.id, w.name, w.url, w.is_active, w.last_triggered_at
		`)
		if err != nil {
			return fmt.Errorf("failed to create webhook_stats view: %w", err)
		}

		// Login geo-distribution view
		_, err = db.ExecContext(ctx, `
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
			ORDER BY login_count DESC
		`)
		if err != nil {
			return fmt.Errorf("failed to create login_geo_distribution view: %w", err)
		}

		// ============================================================
		// 6. Add table and column comments
		// ============================================================

		comments := []string{
			// OTPs
			"COMMENT ON TABLE otps IS 'One-time passwords for email verification, password reset, and 2FA'",
			"COMMENT ON COLUMN otps.code IS 'Bcrypt hashed OTP code'",
			"COMMENT ON COLUMN otps.type IS 'Type of OTP: verification, password_reset, 2fa, login'",

			// API Keys
			"COMMENT ON TABLE api_keys IS 'Permanent API keys for external service integration'",
			"COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the API key'",
			"COMMENT ON COLUMN api_keys.key_prefix IS 'First 8 characters of the key for identification'",
			"COMMENT ON COLUMN api_keys.scopes IS 'JSON array of permission scopes'",
			"COMMENT ON COLUMN api_keys.expires_at IS 'NULL means the key never expires'",

			// RBAC
			"COMMENT ON TABLE permissions IS 'Defines all available permissions in the system'",
			"COMMENT ON TABLE roles IS 'Defines user roles with dynamic permission assignments'",
			"COMMENT ON TABLE role_permissions IS 'Maps permissions to roles (many-to-many)'",

			// Advanced features
			"COMMENT ON TABLE ip_filters IS 'IP whitelisting and blacklisting rules'",
			"COMMENT ON TABLE webhooks IS 'Webhook configurations for event notifications'",
			"COMMENT ON TABLE webhook_deliveries IS 'Tracks webhook delivery attempts and responses'",
			"COMMENT ON TABLE email_templates IS 'Customizable email templates'",
			"COMMENT ON TABLE branding_settings IS 'Branding and customization settings (single row table)'",
			"COMMENT ON TABLE system_settings IS 'System-wide configuration settings'",
			"COMMENT ON TABLE health_metrics IS 'System health and performance metrics'",
			"COMMENT ON TABLE login_locations IS 'Aggregated login location data for performance'",
		}

		for _, commentSQL := range comments {
			if _, err := db.ExecContext(ctx, commentSQL); err != nil {
				return fmt.Errorf("failed to add comment: %w (SQL: %s)", err, commentSQL)
			}
		}

		// ============================================================
		// 7. Insert seed data
		// ============================================================

		// Insert default branding settings
		_, err = db.ExecContext(ctx, `
			INSERT INTO branding_settings (company_name, support_email)
			VALUES ('Auth Gateway', 'support@authgateway.com')
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to insert default branding settings: %w", err)
		}

		// Insert default system settings
		_, err = db.ExecContext(ctx, `
			INSERT INTO system_settings (key, value, description, setting_type, is_public) VALUES
				('maintenance_mode', 'false', 'Enable/disable maintenance mode', 'boolean', true),
				('maintenance_message', 'System is under maintenance. Please try again later.', 'Message shown during maintenance', 'string', true),
				('allow_new_registrations', 'true', 'Allow new user registrations', 'boolean', false),
				('require_email_verification', 'true', 'Require email verification for new users', 'boolean', false),
				('max_sessions_per_user', '10', 'Maximum concurrent sessions per user', 'integer', false),
				('session_timeout_hours', '168', 'Session timeout in hours (default: 7 days)', 'integer', false)
			ON CONFLICT (key) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to insert default system settings: %w", err)
		}

		// Insert system permissions
		_, err = db.ExecContext(ctx, `
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
				('stats.export', 'stats', 'export', 'Export statistics data')
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to insert system permissions: %w", err)
		}

		// Insert system roles and get their IDs
		_, err = db.ExecContext(ctx, `
			INSERT INTO roles (name, display_name, description, is_system_role) VALUES
				('admin', 'Administrator', 'Full system access with all permissions', true),
				('moderator', 'Moderator', 'User management and moderation capabilities', true),
				('user', 'User', 'Standard user with basic permissions', true)
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to insert system roles: %w", err)
		}

		// Assign all permissions to admin role
		_, err = db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id
			FROM roles r
			CROSS JOIN permissions p
			WHERE r.name = 'admin'
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to assign permissions to admin role: %w", err)
		}

		// Assign moderate permissions to moderator role
		_, err = db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id
			FROM roles r
			CROSS JOIN permissions p
			WHERE r.name = 'moderator'
			  AND p.name IN (
				'users.read', 'users.list', 'users.update',
				'sessions.read', 'sessions.list', 'sessions.revoke',
				'audit_logs.read', 'audit_logs.list',
				'api_keys.read', 'api_keys.list',
				'stats.view'
			)
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to assign permissions to moderator role: %w", err)
		}

		// Assign basic permissions to user role
		_, err = db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id
			FROM roles r
			CROSS JOIN permissions p
			WHERE r.name = 'user'
			  AND p.name IN (
				'api_keys.create', 'api_keys.read', 'api_keys.update', 'api_keys.delete',
				'sessions.read', 'sessions.revoke',
				'branding.read'
			)
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to assign permissions to user role: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping schema...")

		// Drop views first
		views := []string{
			"DROP VIEW IF EXISTS login_geo_distribution CASCADE",
			"DROP VIEW IF EXISTS webhook_stats CASCADE",
			"DROP VIEW IF EXISTS active_sessions CASCADE",
		}

		for _, viewSQL := range views {
			if _, err := db.ExecContext(ctx, viewSQL); err != nil {
				return fmt.Errorf("failed to drop view: %w (SQL: %s)", err, viewSQL)
			}
		}

		// Drop tables in reverse order (respecting foreign keys)
		tablesInReverseOrder := []interface{}{
			(*models.LoginLocation)(nil),
			(*models.SMSLog)(nil),
			(*models.SMSSettings)(nil),
			(*models.HealthMetric)(nil),
			(*models.SystemSetting)(nil),
			(*models.BrandingSettings)(nil),
			(*models.EmailTemplateVersion)(nil),
			(*models.EmailTemplate)(nil),
			(*models.WebhookDelivery)(nil),
			(*models.Webhook)(nil),
			(*models.IPFilter)(nil),
			(*models.RolePermission)(nil),
			(*models.Role)(nil),
			(*models.Permission)(nil),
			(*models.AuditLog)(nil),
			(*models.APIKey)(nil),
			(*models.BackupCode)(nil),
			(*models.OTP)(nil),
			(*models.OAuthAccount)(nil),
			(*models.TokenBlacklist)(nil),
			(*models.RefreshToken)(nil),
			(*models.User)(nil),
		}

		for _, model := range tablesInReverseOrder {
			_, err := db.NewDropTable().
				Model(model).
				IfExists().
				Cascade().
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to drop table for %T: %w", model, err)
			}
		}

		// Drop functions
		functions := []string{
			"DROP FUNCTION IF EXISTS cleanup_expired_otps() CASCADE",
			"DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE",
		}

		for _, funcSQL := range functions {
			if _, err := db.ExecContext(ctx, funcSQL); err != nil {
				return fmt.Errorf("failed to drop function: %w (SQL: %s)", err, funcSQL)
			}
		}

		fmt.Println(" OK")
		return nil
	})
}
