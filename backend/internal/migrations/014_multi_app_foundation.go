package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating multi-application foundation tables...")
		return createMultiAppFoundationTables(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping multi-application foundation tables...")
		return dropMultiAppFoundationTables(ctx, db)
	})
}

func createMultiAppFoundationTables(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// PART 1: Create 4 new tables
	// ============================================================

	// 1. Create application_oauth_providers table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS application_oauth_providers (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			provider       VARCHAR(50)  NOT NULL,
			client_id      VARCHAR(500) NOT NULL,
			client_secret  TEXT         NOT NULL,
			callback_url   TEXT         NOT NULL,
			scopes         JSONB            DEFAULT '[]',
			auth_url       TEXT             DEFAULT '',
			token_url      TEXT             DEFAULT '',
			user_info_url  TEXT             DEFAULT '',
			is_active      BOOLEAN          DEFAULT true,
			created_at     TIMESTAMPTZ      DEFAULT NOW(),
			updated_at     TIMESTAMPTZ      DEFAULT NOW(),
			UNIQUE(application_id, provider)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create application_oauth_providers table: %w", err)
	}

	// 2. Create telegram_bots table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS telegram_bots (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			bot_token      TEXT         NOT NULL,
			bot_username   VARCHAR(255) NOT NULL,
			display_name   VARCHAR(255) NOT NULL,
			is_auth_bot    BOOLEAN          DEFAULT false,
			is_active      BOOLEAN          DEFAULT true,
			created_at     TIMESTAMPTZ      DEFAULT NOW(),
			updated_at     TIMESTAMPTZ      DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create telegram_bots table: %w", err)
	}

	// 3. Create user_telegram_accounts table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_telegram_accounts (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			telegram_user_id BIGINT       NOT NULL,
			username         VARCHAR(255),
			first_name       VARCHAR(255) NOT NULL,
			last_name        VARCHAR(255),
			photo_url        TEXT,
			auth_date        TIMESTAMPTZ  NOT NULL,
			created_at       TIMESTAMPTZ      DEFAULT NOW(),
			updated_at       TIMESTAMPTZ      DEFAULT NOW(),
			UNIQUE(user_id, telegram_user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_telegram_accounts table: %w", err)
	}

	// 4. Create user_telegram_bot_access table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_telegram_bot_access (
			id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			telegram_bot_id     UUID NOT NULL REFERENCES telegram_bots(id) ON DELETE CASCADE,
			telegram_account_id UUID NOT NULL REFERENCES user_telegram_accounts(id) ON DELETE CASCADE,
			can_send_messages   BOOLEAN          DEFAULT true,
			authorized_via      BOOLEAN          DEFAULT false,
			created_at          TIMESTAMPTZ      DEFAULT NOW(),
			UNIQUE(user_id, telegram_bot_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_telegram_bot_access table: %w", err)
	}

	// ============================================================
	// Create indexes for new tables
	// ============================================================
	newTableIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_app_oauth_providers_app ON application_oauth_providers(application_id)",
		"CREATE INDEX IF NOT EXISTS idx_telegram_bots_app ON telegram_bots(application_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_tg_accounts_user ON user_telegram_accounts(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_tg_accounts_tg_id ON user_telegram_accounts(telegram_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_tg_bot_access_user ON user_telegram_bot_access(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_tg_bot_access_bot ON user_telegram_bot_access(telegram_bot_id)",
	}

	for _, indexSQL := range newTableIndexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// PART 2: Add nullable application_id to existing tables
	// ============================================================
	alterTableStatements := []string{
		"ALTER TABLE sessions ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE roles ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE permissions ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE email_providers ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE email_profiles ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
		"ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id)",
	}

	for _, alterSQL := range alterTableStatements {
		if _, err := db.ExecContext(ctx, alterSQL); err != nil {
			return fmt.Errorf("failed to alter table: %w (SQL: %s)", err, alterSQL)
		}
	}

	// ============================================================
	// Update unique constraints for roles and permissions
	// ============================================================
	_, err = db.ExecContext(ctx, `ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_name_key`)
	if err != nil {
		return fmt.Errorf("failed to drop roles_name_key constraint: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_roles_name`)
	if err != nil {
		return fmt.Errorf("failed to drop idx_roles_name: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE UNIQUE INDEX idx_roles_name_app
		ON roles(name, COALESCE(application_id, '00000000-0000-0000-0000-000000000000'::uuid))
	`)
	if err != nil {
		return fmt.Errorf("failed to create idx_roles_name_app: %w", err)
	}

	_, err = db.ExecContext(ctx, `ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_name_key`)
	if err != nil {
		return fmt.Errorf("failed to drop permissions_name_key constraint: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_permissions_name`)
	if err != nil {
		return fmt.Errorf("failed to drop idx_permissions_name: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE UNIQUE INDEX idx_permissions_name_app
		ON permissions(name, COALESCE(application_id, '00000000-0000-0000-0000-000000000000'::uuid))
	`)
	if err != nil {
		return fmt.Errorf("failed to create idx_permissions_name_app: %w", err)
	}

	// ============================================================
	// Create partial indexes for filtering
	// ============================================================
	partialIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_sessions_app ON sessions(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_keys_app ON api_keys(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_roles_app ON roles(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_permissions_app ON permissions(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_webhooks_app ON webhooks(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_app ON audit_logs(application_id) WHERE application_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_app ON user_roles(application_id) WHERE application_id IS NOT NULL",
	}

	for _, indexSQL := range partialIndexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create partial index: %w (SQL: %s)", err, indexSQL)
		}
	}

	fmt.Println(" OK")
	return nil
}

func dropMultiAppFoundationTables(ctx context.Context, db *bun.DB) error {
	// ============================================================
	// 1. Drop partial indexes
	// ============================================================
	partialIndexes := []string{
		"DROP INDEX IF EXISTS idx_user_roles_app",
		"DROP INDEX IF EXISTS idx_audit_logs_app",
		"DROP INDEX IF EXISTS idx_webhooks_app",
		"DROP INDEX IF EXISTS idx_permissions_app",
		"DROP INDEX IF EXISTS idx_roles_app",
		"DROP INDEX IF EXISTS idx_api_keys_app",
		"DROP INDEX IF EXISTS idx_sessions_app",
	}

	for _, indexSQL := range partialIndexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to drop partial index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// 2. Restore original unique indexes for roles and permissions
	// ============================================================
	_, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_permissions_name_app`)
	if err != nil {
		return fmt.Errorf("failed to drop idx_permissions_name_app: %w", err)
	}

	_, err = db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name)`)
	if err != nil {
		return fmt.Errorf("failed to create idx_permissions_name: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_roles_name_app`)
	if err != nil {
		return fmt.Errorf("failed to drop idx_roles_name_app: %w", err)
	}

	_, err = db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles(name)`)
	if err != nil {
		return fmt.Errorf("failed to create idx_roles_name: %w", err)
	}

	// ============================================================
	// 3. Remove application_id columns from existing tables
	// ============================================================
	alterTableStatements := []string{
		"ALTER TABLE user_roles DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE audit_logs DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE email_profiles DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE email_providers DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE webhooks DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE permissions DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE roles DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE api_keys DROP COLUMN IF EXISTS application_id",
		"ALTER TABLE sessions DROP COLUMN IF EXISTS application_id",
	}

	for _, alterSQL := range alterTableStatements {
		if _, err := db.ExecContext(ctx, alterSQL); err != nil {
			return fmt.Errorf("failed to alter table: %w (SQL: %s)", err, alterSQL)
		}
	}

	// ============================================================
	// 4. Drop indexes for new tables
	// ============================================================
	newTableIndexes := []string{
		"DROP INDEX IF EXISTS idx_user_tg_bot_access_bot",
		"DROP INDEX IF EXISTS idx_user_tg_bot_access_user",
		"DROP INDEX IF EXISTS idx_user_tg_accounts_tg_id",
		"DROP INDEX IF EXISTS idx_user_tg_accounts_user",
		"DROP INDEX IF EXISTS idx_telegram_bots_app",
		"DROP INDEX IF EXISTS idx_app_oauth_providers_app",
	}

	for _, indexSQL := range newTableIndexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to drop index: %w (SQL: %s)", err, indexSQL)
		}
	}

	// ============================================================
	// 5. Drop new tables in reverse order (respecting foreign keys)
	// ============================================================
	tables := []string{
		"DROP TABLE IF EXISTS user_telegram_bot_access CASCADE",
		"DROP TABLE IF EXISTS user_telegram_accounts CASCADE",
		"DROP TABLE IF EXISTS telegram_bots CASCADE",
		"DROP TABLE IF EXISTS application_oauth_providers CASCADE",
	}

	for _, tableSQL := range tables {
		if _, err := db.ExecContext(ctx, tableSQL); err != nil {
			return fmt.Errorf("failed to drop table: %w (SQL: %s)", err, tableSQL)
		}
	}

	fmt.Println(" OK")
	return nil
}
