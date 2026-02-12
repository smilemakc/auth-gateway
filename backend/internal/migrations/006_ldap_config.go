package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating LDAP configuration tables...")
		return createLDAPConfigTables(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping LDAP configuration tables...")
		return dropLDAPConfigTables(ctx, db)
	})
}

// createLDAPConfigTables creates the ldap_configs and ldap_sync_logs tables
func createLDAPConfigTables(ctx context.Context, db *bun.DB) error {
	// Create ldap_configs table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS ldap_configs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			server VARCHAR(255) NOT NULL,
			port INTEGER NOT NULL DEFAULT 389,
			use_tls BOOLEAN NOT NULL DEFAULT FALSE,
			use_ssl BOOLEAN NOT NULL DEFAULT FALSE,
			insecure BOOLEAN NOT NULL DEFAULT FALSE,
			bind_dn VARCHAR(500) NOT NULL,
			bind_password VARCHAR(500) NOT NULL,
			base_dn VARCHAR(500) NOT NULL,
			user_search_base VARCHAR(500),
			group_search_base VARCHAR(500),
			user_search_filter VARCHAR(500) NOT NULL DEFAULT '(objectClass=person)',
			group_search_filter VARCHAR(500) NOT NULL DEFAULT '(objectClass=group)',
			user_id_attribute VARCHAR(100) NOT NULL DEFAULT 'uid',
			user_email_attribute VARCHAR(100) NOT NULL DEFAULT 'mail',
			user_name_attribute VARCHAR(100) NOT NULL DEFAULT 'cn',
			user_dn_attribute VARCHAR(100) NOT NULL DEFAULT 'dn',
			group_id_attribute VARCHAR(100) NOT NULL DEFAULT 'cn',
			group_name_attribute VARCHAR(100) NOT NULL DEFAULT 'cn',
			group_member_attribute VARCHAR(100) NOT NULL DEFAULT 'member',
			sync_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			sync_interval BIGINT NOT NULL DEFAULT 3600000000000,
			last_sync_at TIMESTAMP,
			next_sync_at TIMESTAMP,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			last_test_at TIMESTAMP,
			last_test_result VARCHAR(50),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ldap_configs table: %w", err)
	}

	// Create ldap_sync_logs table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS ldap_sync_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ldap_config_id UUID NOT NULL REFERENCES ldap_configs(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL,
			users_synced INTEGER NOT NULL DEFAULT 0,
			users_created INTEGER NOT NULL DEFAULT 0,
			users_updated INTEGER NOT NULL DEFAULT 0,
			users_deleted INTEGER NOT NULL DEFAULT 0,
			groups_synced INTEGER NOT NULL DEFAULT 0,
			groups_created INTEGER NOT NULL DEFAULT 0,
			groups_updated INTEGER NOT NULL DEFAULT 0,
			error_message TEXT,
			started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			duration_ms BIGINT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ldap_sync_logs table: %w", err)
	}

	// Create indexes
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_ldap_configs_is_active ON ldap_configs(is_active);
		CREATE INDEX IF NOT EXISTS idx_ldap_sync_logs_config_id ON ldap_sync_logs(ldap_config_id);
		CREATE INDEX IF NOT EXISTS idx_ldap_sync_logs_started_at ON ldap_sync_logs(started_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create trigger to update updated_at timestamp
	_, err = db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION update_ldap_configs_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_ldap_configs_updated_at ON ldap_configs;
		CREATE TRIGGER trigger_update_ldap_configs_updated_at
			BEFORE UPDATE ON ldap_configs
			FOR EACH ROW
			EXECUTE FUNCTION update_ldap_configs_updated_at();
	`)
	if err != nil {
		return fmt.Errorf("failed to create update trigger: %w", err)
	}

	fmt.Println(" OK")
	return nil
}

// dropLDAPConfigTables drops the LDAP configuration tables
func dropLDAPConfigTables(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS ldap_sync_logs CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop ldap_sync_logs table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS ldap_configs CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop ldap_configs table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP FUNCTION IF EXISTS update_ldap_configs_updated_at() CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop update function: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
