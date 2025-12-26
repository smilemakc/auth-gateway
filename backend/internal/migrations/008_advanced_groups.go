package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Adding advanced group features...")

		// Add is_dynamic column
		_, err := db.ExecContext(ctx, `
			ALTER TABLE groups 
			ADD COLUMN IF NOT EXISTS is_dynamic BOOLEAN NOT NULL DEFAULT FALSE;
		`)
		if err != nil {
			return fmt.Errorf("failed to add is_dynamic column: %w", err)
		}

		// Add membership_rules column (JSONB)
		_, err = db.ExecContext(ctx, `
			ALTER TABLE groups 
			ADD COLUMN IF NOT EXISTS membership_rules JSONB;
		`)
		if err != nil {
			return fmt.Errorf("failed to add membership_rules column: %w", err)
		}

		// Add permission_ids column (UUID array)
		_, err = db.ExecContext(ctx, `
			ALTER TABLE groups 
			ADD COLUMN IF NOT EXISTS permission_ids UUID[];
		`)
		if err != nil {
			return fmt.Errorf("failed to add permission_ids column: %w", err)
		}

		// Create index for dynamic groups
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_groups_is_dynamic ON groups(is_dynamic) WHERE is_dynamic = TRUE;
		`)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Removing advanced group features...")

		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS idx_groups_is_dynamic;
			ALTER TABLE groups DROP COLUMN IF EXISTS permission_ids;
			ALTER TABLE groups DROP COLUMN IF EXISTS membership_rules;
			ALTER TABLE groups DROP COLUMN IF EXISTS is_dynamic;
		`)
		if err != nil {
			return fmt.Errorf("failed to remove advanced group features: %w", err)
		}

		fmt.Println(" OK")
		return nil
	})
}
