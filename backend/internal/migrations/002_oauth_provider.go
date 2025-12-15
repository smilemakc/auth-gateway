package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating OAuth 2.0 / OIDC provider schema...")

		// ============================================================
		// 1. Create oauth_clients table
		// ============================================================
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS oauth_clients (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				client_id VARCHAR(64) UNIQUE NOT NULL,
				client_secret_hash VARCHAR(255),
				name VARCHAR(255) NOT NULL,
				description TEXT,
				logo_url VARCHAR(512),
				client_type VARCHAR(20) DEFAULT 'confidential' CHECK (client_type IN ('confidential', 'public')),
				redirect_uris JSONB DEFAULT '[]'::jsonb,
				allowed_grant_types JSONB DEFAULT '["authorization_code"]'::jsonb,
				allowed_scopes JSONB DEFAULT '["openid"]'::jsonb,
				default_scopes JSONB DEFAULT '["openid"]'::jsonb,
				access_token_ttl INTEGER DEFAULT 900,
				refresh_token_ttl INTEGER DEFAULT 604800,
				id_token_ttl INTEGER DEFAULT 3600,
				require_pkce BOOLEAN DEFAULT false,
				require_consent BOOLEAN DEFAULT true,
				first_party BOOLEAN DEFAULT false,
				owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
				is_active BOOLEAN DEFAULT true,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create oauth_clients table: %w", err)
		}

		// ============================================================
		// 2. Create authorization_codes table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS authorization_codes (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				code_hash VARCHAR(64) UNIQUE NOT NULL,
				client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				redirect_uri VARCHAR(2048) NOT NULL,
				scope VARCHAR(1024) NOT NULL,
				code_challenge VARCHAR(128),
				code_challenge_method VARCHAR(10) CHECK (code_challenge_method IN ('S256', 'plain')),
				nonce VARCHAR(255),
				used BOOLEAN DEFAULT false,
				expires_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create authorization_codes table: %w", err)
		}

		// ============================================================
		// 3. Create oauth_access_tokens table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS oauth_access_tokens (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				token_hash VARCHAR(64) UNIQUE NOT NULL,
				client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
				user_id UUID REFERENCES users(id) ON DELETE CASCADE,
				scope VARCHAR(1024) NOT NULL,
				is_active BOOLEAN DEFAULT true,
				expires_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				revoked_at TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create oauth_access_tokens table: %w", err)
		}

		// ============================================================
		// 4. Create oauth_refresh_tokens table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				token_hash VARCHAR(64) UNIQUE NOT NULL,
				access_token_id UUID NOT NULL REFERENCES oauth_access_tokens(id) ON DELETE CASCADE,
				client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				scope VARCHAR(1024) NOT NULL,
				is_active BOOLEAN DEFAULT true,
				expires_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				revoked_at TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create oauth_refresh_tokens table: %w", err)
		}

		// ============================================================
		// 5. Create user_consents table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS user_consents (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
				scopes JSONB NOT NULL,
				granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				revoked_at TIMESTAMP,
				UNIQUE(user_id, client_id)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create user_consents table: %w", err)
		}

		// ============================================================
		// 6. Create device_codes table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS device_codes (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				device_code_hash VARCHAR(64) UNIQUE NOT NULL,
				user_code VARCHAR(10) UNIQUE NOT NULL,
				client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
				user_id UUID REFERENCES users(id) ON DELETE SET NULL,
				scope VARCHAR(1024) NOT NULL,
				status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'authorized', 'denied', 'expired')),
				verification_uri VARCHAR(512) NOT NULL,
				verification_uri_complete VARCHAR(512),
				expires_at TIMESTAMP NOT NULL,
				interval INTEGER DEFAULT 5,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create device_codes table: %w", err)
		}

		// ============================================================
		// 7. Create oauth_scopes table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS oauth_scopes (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				name VARCHAR(100) UNIQUE NOT NULL,
				display_name VARCHAR(255) NOT NULL,
				description TEXT,
				is_default BOOLEAN DEFAULT false,
				is_system BOOLEAN DEFAULT true,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create oauth_scopes table: %w", err)
		}

		// ============================================================
		// 8. Create indexes
		// ============================================================
		indexes := []string{
			// oauth_clients indexes
			"CREATE INDEX IF NOT EXISTS idx_oauth_clients_client_id ON oauth_clients(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_clients_owner_id ON oauth_clients(owner_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_clients_is_active ON oauth_clients(is_active)",

			// authorization_codes indexes
			"CREATE INDEX IF NOT EXISTS idx_authorization_codes_code_hash ON authorization_codes(code_hash)",
			"CREATE INDEX IF NOT EXISTS idx_authorization_codes_expires_at ON authorization_codes(expires_at) WHERE used = false",
			"CREATE INDEX IF NOT EXISTS idx_authorization_codes_client_id ON authorization_codes(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_authorization_codes_user_id ON authorization_codes(user_id)",

			// oauth_access_tokens indexes
			"CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_token_hash ON oauth_access_tokens(token_hash)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_user_id ON oauth_access_tokens(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_client_id ON oauth_access_tokens(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_expires_at ON oauth_access_tokens(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_active ON oauth_access_tokens(is_active, expires_at)",

			// oauth_refresh_tokens indexes
			"CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_token_hash ON oauth_refresh_tokens(token_hash)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_access_token_id ON oauth_refresh_tokens(access_token_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_client_id ON oauth_refresh_tokens(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_user_id ON oauth_refresh_tokens(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_active ON oauth_refresh_tokens(is_active, expires_at)",

			// user_consents indexes
			"CREATE INDEX IF NOT EXISTS idx_user_consents_user_id ON user_consents(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_consents_client_id ON user_consents(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_consents_revoked ON user_consents(revoked_at) WHERE revoked_at IS NULL",

			// device_codes indexes
			"CREATE INDEX IF NOT EXISTS idx_device_codes_device_code_hash ON device_codes(device_code_hash)",
			"CREATE INDEX IF NOT EXISTS idx_device_codes_user_code ON device_codes(user_code)",
			"CREATE INDEX IF NOT EXISTS idx_device_codes_client_id ON device_codes(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_device_codes_user_id ON device_codes(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_device_codes_status ON device_codes(status, expires_at)",

			// oauth_scopes indexes
			"CREATE INDEX IF NOT EXISTS idx_oauth_scopes_name ON oauth_scopes(name)",
			"CREATE INDEX IF NOT EXISTS idx_oauth_scopes_is_default ON oauth_scopes(is_default) WHERE is_default = true",
		}

		for _, indexSQL := range indexes {
			if _, err := db.ExecContext(ctx, indexSQL); err != nil {
				return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
			}
		}

		// ============================================================
		// 9. Create triggers for updated_at
		// ============================================================
		triggers := []string{
			"CREATE TRIGGER update_oauth_clients_updated_at BEFORE UPDATE ON oauth_clients FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
		}

		for _, triggerSQL := range triggers {
			if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
				return fmt.Errorf("failed to create trigger: %w (SQL: %s)", err, triggerSQL)
			}
		}

		// ============================================================
		// 10. Create cleanup function for expired authorization codes
		// ============================================================
		_, err = db.ExecContext(ctx, `
			CREATE OR REPLACE FUNCTION cleanup_expired_auth_codes() RETURNS TRIGGER AS $$
			BEGIN
				DELETE FROM authorization_codes WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '1 hour';
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql
		`)
		if err != nil {
			return fmt.Errorf("failed to create cleanup_expired_auth_codes function: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			CREATE TRIGGER trigger_cleanup_expired_auth_codes
			AFTER INSERT ON authorization_codes
			EXECUTE FUNCTION cleanup_expired_auth_codes()
		`)
		if err != nil {
			return fmt.Errorf("failed to create trigger for auth code cleanup: %w", err)
		}

		// ============================================================
		// 11. Add table and column comments
		// ============================================================
		comments := []string{
			"COMMENT ON TABLE oauth_clients IS 'OAuth 2.0 / OIDC client applications'",
			"COMMENT ON COLUMN oauth_clients.client_id IS 'Public client identifier (e.g., my-app)'",
			"COMMENT ON COLUMN oauth_clients.client_secret_hash IS 'BCrypt hash of client secret, NULL for public clients'",
			"COMMENT ON COLUMN oauth_clients.client_type IS 'Client type: confidential (has secret) or public (no secret)'",
			"COMMENT ON COLUMN oauth_clients.redirect_uris IS 'Array of allowed redirect URIs'",
			"COMMENT ON COLUMN oauth_clients.require_pkce IS 'Require PKCE for authorization code flow'",
			"COMMENT ON COLUMN oauth_clients.require_consent IS 'Require user consent for this client'",
			"COMMENT ON COLUMN oauth_clients.first_party IS 'First-party apps skip consent screen'",

			"COMMENT ON TABLE authorization_codes IS 'Temporary authorization codes for OAuth 2.0 flow (10 min TTL, single use)'",
			"COMMENT ON COLUMN authorization_codes.code_hash IS 'SHA-256 hash of the authorization code'",
			"COMMENT ON COLUMN authorization_codes.code_challenge IS 'PKCE code challenge'",
			"COMMENT ON COLUMN authorization_codes.code_challenge_method IS 'PKCE challenge method: S256 or plain'",
			"COMMENT ON COLUMN authorization_codes.nonce IS 'OIDC nonce for ID token validation'",

			"COMMENT ON TABLE oauth_access_tokens IS 'OAuth 2.0 access tokens for introspection and revocation'",
			"COMMENT ON COLUMN oauth_access_tokens.token_hash IS 'SHA-256 hash of the access token'",
			"COMMENT ON COLUMN oauth_access_tokens.user_id IS 'NULL for client_credentials grant'",

			"COMMENT ON TABLE oauth_refresh_tokens IS 'OAuth 2.0 refresh tokens'",
			"COMMENT ON COLUMN oauth_refresh_tokens.token_hash IS 'SHA-256 hash of the refresh token'",

			"COMMENT ON TABLE user_consents IS 'User consent records for OAuth clients'",
			"COMMENT ON COLUMN user_consents.scopes IS 'Array of granted scopes'",

			"COMMENT ON TABLE device_codes IS 'Device Authorization Grant codes (RFC 8628)'",
			"COMMENT ON COLUMN device_codes.user_code IS 'User-friendly code like ABCD-1234'",
			"COMMENT ON COLUMN device_codes.status IS 'pending, authorized, denied, or expired'",
			"COMMENT ON COLUMN device_codes.interval IS 'Polling interval in seconds'",

			"COMMENT ON TABLE oauth_scopes IS 'Defined OAuth 2.0 / OIDC scopes'",
			"COMMENT ON COLUMN oauth_scopes.is_system IS 'System-defined scope (cannot be deleted)'",
		}

		for _, commentSQL := range comments {
			if _, err := db.ExecContext(ctx, commentSQL); err != nil {
				return fmt.Errorf("failed to add comment: %w (SQL: %s)", err, commentSQL)
			}
		}

		// ============================================================
		// 12. Seed standard OIDC scopes
		// ============================================================
		_, err = db.ExecContext(ctx, `
			INSERT INTO oauth_scopes (name, display_name, description, is_default, is_system) VALUES
				('openid', 'OpenID', 'Required for OpenID Connect authentication', true, true),
				('profile', 'Profile', 'Access to basic profile information (name, username)', false, true),
				('email', 'Email Address', 'Access to email address', false, true),
				('offline_access', 'Offline Access', 'Request refresh token for offline access', false, true),
				('users:read', 'Read Users', 'Read user information', false, true),
				('users:write', 'Write Users', 'Create and update users', false, true),
				('api_keys:read', 'Read API Keys', 'View API keys', false, true),
				('api_keys:write', 'Write API Keys', 'Create and manage API keys', false, true),
				('sessions:read', 'Read Sessions', 'View active sessions', false, true),
				('sessions:write', 'Write Sessions', 'Manage user sessions', false, true)
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to seed OAuth scopes: %w", err)
		}

		// ============================================================
		// 13. Create OAuth permissions
		// ============================================================
		_, err = db.ExecContext(ctx, `
			INSERT INTO permissions (name, resource, action, description) VALUES
				('oauth_clients.create', 'oauth_clients', 'create', 'Create OAuth clients'),
				('oauth_clients.read', 'oauth_clients', 'read', 'View OAuth client information'),
				('oauth_clients.update', 'oauth_clients', 'update', 'Update OAuth clients'),
				('oauth_clients.delete', 'oauth_clients', 'delete', 'Delete OAuth clients'),
				('oauth_clients.list', 'oauth_clients', 'list', 'List all OAuth clients'),
				('oauth_clients.rotate_secret', 'oauth_clients', 'rotate_secret', 'Rotate client secret'),

				('oauth_tokens.revoke', 'oauth_tokens', 'revoke', 'Revoke OAuth tokens'),
				('oauth_tokens.introspect', 'oauth_tokens', 'introspect', 'Introspect OAuth tokens'),

				('oauth_scopes.create', 'oauth_scopes', 'create', 'Create OAuth scopes'),
				('oauth_scopes.read', 'oauth_scopes', 'read', 'View OAuth scopes'),
				('oauth_scopes.update', 'oauth_scopes', 'update', 'Update OAuth scopes'),
				('oauth_scopes.delete', 'oauth_scopes', 'delete', 'Delete OAuth scopes'),
				('oauth_scopes.list', 'oauth_scopes', 'list', 'List OAuth scopes')
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to create OAuth permissions: %w", err)
		}

		// ============================================================
		// 14. Assign OAuth permissions to admin role
		// ============================================================
		_, err = db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id
			FROM roles r
			CROSS JOIN permissions p
			WHERE r.name = 'admin'
			  AND p.resource IN ('oauth_clients', 'oauth_tokens', 'oauth_scopes')
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to assign OAuth permissions to admin role: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping OAuth 2.0 / OIDC provider schema...")

		// Drop triggers first
		triggers := []string{
			"DROP TRIGGER IF EXISTS trigger_cleanup_expired_auth_codes ON authorization_codes",
			"DROP TRIGGER IF EXISTS update_oauth_clients_updated_at ON oauth_clients",
		}

		for _, triggerSQL := range triggers {
			if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
				return fmt.Errorf("failed to drop trigger: %w (SQL: %s)", err, triggerSQL)
			}
		}

		// Drop function
		_, err := db.ExecContext(ctx, "DROP FUNCTION IF EXISTS cleanup_expired_auth_codes() CASCADE")
		if err != nil {
			return fmt.Errorf("failed to drop cleanup_expired_auth_codes function: %w", err)
		}

		// Drop tables in reverse order (respecting foreign keys)
		tables := []string{
			"DROP TABLE IF EXISTS oauth_refresh_tokens CASCADE",
			"DROP TABLE IF EXISTS oauth_access_tokens CASCADE",
			"DROP TABLE IF EXISTS device_codes CASCADE",
			"DROP TABLE IF EXISTS user_consents CASCADE",
			"DROP TABLE IF EXISTS authorization_codes CASCADE",
			"DROP TABLE IF EXISTS oauth_scopes CASCADE",
			"DROP TABLE IF EXISTS oauth_clients CASCADE",
		}

		for _, tableSQL := range tables {
			if _, err := db.ExecContext(ctx, tableSQL); err != nil {
				return fmt.Errorf("failed to drop table: %w (SQL: %s)", err, tableSQL)
			}
		}

		// Delete OAuth-related permissions
		_, err = db.ExecContext(ctx, `
			DELETE FROM role_permissions
			WHERE permission_id IN (
				SELECT id FROM permissions
				WHERE resource IN ('oauth_clients', 'oauth_tokens', 'oauth_scopes')
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to delete OAuth role permissions: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			DELETE FROM permissions
			WHERE resource IN ('oauth_clients', 'oauth_tokens', 'oauth_scopes')
		`)
		if err != nil {
			return fmt.Errorf("failed to delete OAuth permissions: %w", err)
		}

		fmt.Println(" OK")
		return nil
	})
}
