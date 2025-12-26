package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating groups tables...")
		return createGroupsTable(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping groups tables...")
		return dropGroupsTable(ctx, db)
	})
}

// createGroupsTable creates the groups and user_groups tables
func createGroupsTable(ctx context.Context, db *bun.DB) error {
	// Create groups table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			display_name VARCHAR(255) NOT NULL,
			description TEXT,
			parent_group_id UUID REFERENCES groups(id) ON DELETE SET NULL,
			is_system_group BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create groups table: %w", err)
	}

	// Create user_groups junction table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, group_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_groups table: %w", err)
	}

	// Create indexes for performance
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_groups_parent_group_id ON groups(parent_group_id);
		CREATE INDEX IF NOT EXISTS idx_groups_name ON groups(name);
		CREATE INDEX IF NOT EXISTS idx_user_groups_user_id ON user_groups(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_groups_group_id ON user_groups(group_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create trigger to update updated_at timestamp
	_, err = db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION update_groups_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_groups_updated_at ON groups;
		CREATE TRIGGER trigger_update_groups_updated_at
			BEFORE UPDATE ON groups
			FOR EACH ROW
			EXECUTE FUNCTION update_groups_updated_at();
	`)
	if err != nil {
		return fmt.Errorf("failed to create update trigger: %w", err)
	}

	fmt.Println(" OK")
	return nil
}

// dropGroupsTable drops the groups and user_groups tables
func dropGroupsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS user_groups CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop user_groups table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS groups CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop groups table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP FUNCTION IF EXISTS update_groups_updated_at() CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop update function: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
