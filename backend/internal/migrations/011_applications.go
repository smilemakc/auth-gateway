package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating multi-application system tables...")
		return createApplicationsTables(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping multi-application system tables...")
		return dropApplicationsTables(ctx, db)
	})
}

func createApplicationsTables(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// 1. Create applications table
	// ============================================================
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS applications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			display_name VARCHAR(255) NOT NULL,
			description TEXT,
			homepage_url VARCHAR(512),
			callback_urls JSONB DEFAULT '[]'::jsonb,
			is_active BOOLEAN DEFAULT true,
			is_system BOOLEAN DEFAULT false,
			owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create applications table: %w", err)
	}

	// ============================================================
	// 2. Create application_branding table
	// ============================================================
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS application_branding (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id UUID NOT NULL UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
			logo_url VARCHAR(512),
			favicon_url VARCHAR(512),
			primary_color VARCHAR(7) DEFAULT '#3B82F6',
			secondary_color VARCHAR(7) DEFAULT '#8B5CF6',
			background_color VARCHAR(7) DEFAULT '#FFFFFF',
			custom_css TEXT,
			company_name VARCHAR(255),
			support_email VARCHAR(255),
			terms_url VARCHAR(512),
			privacy_url VARCHAR(512),
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create application_branding table: %w", err)
	}

	// ============================================================
	// 3. Create user_application_profiles table
	// ============================================================
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_application_profiles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			display_name VARCHAR(255),
			avatar_url VARCHAR(512),
			nickname VARCHAR(100),
			metadata JSONB DEFAULT '{}'::jsonb,
			app_roles JSONB DEFAULT '[]'::jsonb,
			is_active BOOLEAN DEFAULT true,
			is_banned BOOLEAN DEFAULT false,
			ban_reason TEXT,
			banned_at TIMESTAMP,
			banned_by UUID REFERENCES users(id) ON DELETE SET NULL,
			last_access_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, application_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_application_profiles table: %w", err)
	}

	// ============================================================
	// 4. Add application_id to oauth_clients
	// ============================================================
	_, err = db.ExecContext(ctx, `
		ALTER TABLE oauth_clients
		ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id) ON DELETE SET NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to add application_id to oauth_clients: %w", err)
	}

	// ============================================================
	// 5. Create indexes
	// ============================================================
	indexes := []string{
		// Applications indexes
		"CREATE INDEX IF NOT EXISTS idx_applications_name ON applications(name)",
		"CREATE INDEX IF NOT EXISTS idx_applications_owner_id ON applications(owner_id)",
		"CREATE INDEX IF NOT EXISTS idx_applications_is_active ON applications(is_active) WHERE is_active = true",

		// Application branding indexes
		"CREATE INDEX IF NOT EXISTS idx_app_branding_application_id ON application_branding(application_id)",

		// User application profiles indexes
		"CREATE INDEX IF NOT EXISTS idx_user_app_profiles_user_id ON user_application_profiles(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_app_profiles_app_id ON user_application_profiles(application_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_app_profiles_active ON user_application_profiles(is_active, is_banned)",
		"CREATE INDEX IF NOT EXISTS idx_user_app_profiles_last_access ON user_application_profiles(last_access_at DESC) WHERE last_access_at IS NOT NULL",

		// OAuth clients application_id index
		"CREATE INDEX IF NOT EXISTS idx_oauth_clients_application_id ON oauth_clients(application_id) WHERE application_id IS NOT NULL",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// 6. Create triggers for updated_at
	// ============================================================
	triggers := []string{
		"CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
		"CREATE TRIGGER update_application_branding_updated_at BEFORE UPDATE ON application_branding FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
		"CREATE TRIGGER update_user_application_profiles_updated_at BEFORE UPDATE ON user_application_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
	}

	for _, triggerSQL := range triggers {
		if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
			return fmt.Errorf("failed to create trigger: %w (SQL: %s)", err, triggerSQL)
		}
	}

	// ============================================================
	// 7. Add table and column comments
	// ============================================================
	comments := []string{
		"COMMENT ON TABLE applications IS 'Multi-application support - each app can have custom branding and user profiles'",
		"COMMENT ON COLUMN applications.name IS 'Unique application identifier (slug)'",
		"COMMENT ON COLUMN applications.callback_urls IS 'Array of allowed callback URLs for this application'",
		"COMMENT ON COLUMN applications.is_system IS 'System applications cannot be deleted'",

		"COMMENT ON TABLE application_branding IS 'Custom branding settings per application'",
		"COMMENT ON COLUMN application_branding.primary_color IS 'Primary brand color in hex format'",
		"COMMENT ON COLUMN application_branding.secondary_color IS 'Secondary brand color in hex format'",
		"COMMENT ON COLUMN application_branding.custom_css IS 'Custom CSS for application-specific styling'",

		"COMMENT ON TABLE user_application_profiles IS 'User profiles per application with app-specific roles and metadata'",
		"COMMENT ON COLUMN user_application_profiles.metadata IS 'Application-specific user metadata (JSON)'",
		"COMMENT ON COLUMN user_application_profiles.app_roles IS 'Application-specific roles (JSON array)'",
		"COMMENT ON COLUMN user_application_profiles.is_banned IS 'User banned from this specific application'",
		"COMMENT ON COLUMN user_application_profiles.last_access_at IS 'Last time user accessed this application'",

		"COMMENT ON COLUMN oauth_clients.application_id IS 'Links OAuth client to an application for multi-app support'",
	}

	for _, commentSQL := range comments {
		if _, err := db.ExecContext(ctx, commentSQL); err != nil {
			return fmt.Errorf("failed to add comment: %w (SQL: %s)", err, commentSQL)
		}
	}

	// ============================================================
	// 8. Seed system application
	// ============================================================
	_, err = db.ExecContext(ctx, `
		INSERT INTO applications (name, display_name, description, is_system, is_active)
		VALUES (
			'auth-gateway-admin',
			'Auth Gateway Admin',
			'System administration application for Auth Gateway',
			true,
			true
		)
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed system application: %w", err)
	}

	// ============================================================
	// 9. Create application permissions
	// ============================================================
	_, err = db.ExecContext(ctx, `
		INSERT INTO permissions (name, resource, action, description) VALUES
			('applications.create', 'applications', 'create', 'Create new applications'),
			('applications.read', 'applications', 'read', 'View application information'),
			('applications.update', 'applications', 'update', 'Update applications'),
			('applications.delete', 'applications', 'delete', 'Delete applications'),
			('applications.list', 'applications', 'list', 'List all applications'),

			('application_branding.read', 'application_branding', 'read', 'View application branding'),
			('application_branding.update', 'application_branding', 'update', 'Update application branding'),

			('user_app_profiles.create', 'user_app_profiles', 'create', 'Create user application profiles'),
			('user_app_profiles.read', 'user_app_profiles', 'read', 'View user application profiles'),
			('user_app_profiles.update', 'user_app_profiles', 'update', 'Update user application profiles'),
			('user_app_profiles.delete', 'user_app_profiles', 'delete', 'Delete user application profiles'),
			('user_app_profiles.list', 'user_app_profiles', 'list', 'List user application profiles'),
			('user_app_profiles.ban', 'user_app_profiles', 'ban', 'Ban user from application'),
			('user_app_profiles.unban', 'user_app_profiles', 'unban', 'Unban user from application')
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to create application permissions: %w", err)
	}

	// ============================================================
	// 10. Assign application permissions to admin role
	// ============================================================
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'admin'
		  AND p.resource IN ('applications', 'application_branding', 'user_app_profiles')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to assign application permissions to admin role: %w", err)
	}

	fmt.Println(" OK")
	return nil
}

func dropApplicationsTables(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// 1. Drop triggers first
	// ============================================================
	triggers := []string{
		"DROP TRIGGER IF EXISTS update_user_application_profiles_updated_at ON user_application_profiles",
		"DROP TRIGGER IF EXISTS update_application_branding_updated_at ON application_branding",
		"DROP TRIGGER IF EXISTS update_applications_updated_at ON applications",
	}

	for _, triggerSQL := range triggers {
		if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
			return fmt.Errorf("failed to drop trigger: %w (SQL: %s)", err, triggerSQL)
		}
	}

	// ============================================================
	// 2. Remove application_id column from oauth_clients
	// ============================================================
	_, err := db.ExecContext(ctx, `
		ALTER TABLE oauth_clients
		DROP COLUMN IF EXISTS application_id
	`)
	if err != nil {
		return fmt.Errorf("failed to drop application_id from oauth_clients: %w", err)
	}

	// ============================================================
	// 3. Drop tables in reverse order (respecting foreign keys)
	// ============================================================
	tables := []string{
		"DROP TABLE IF EXISTS user_application_profiles CASCADE",
		"DROP TABLE IF EXISTS application_branding CASCADE",
		"DROP TABLE IF EXISTS applications CASCADE",
	}

	for _, tableSQL := range tables {
		if _, err := db.ExecContext(ctx, tableSQL); err != nil {
			return fmt.Errorf("failed to drop table: %w (SQL: %s)", err, tableSQL)
		}
	}

	// ============================================================
	// 4. Delete application-related permissions
	// ============================================================
	_, err = db.ExecContext(ctx, `
		DELETE FROM role_permissions
		WHERE permission_id IN (
			SELECT id FROM permissions
			WHERE resource IN ('applications', 'application_branding', 'user_app_profiles')
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to delete application role permissions: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		DELETE FROM permissions
		WHERE resource IN ('applications', 'application_branding', 'user_app_profiles')
	`)
	if err != nil {
		return fmt.Errorf("failed to delete application permissions: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
