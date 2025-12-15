package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Adding access_token_hash column to sessions table...")

		// Add access_token_hash column for immediate session revocation
		_, err := db.ExecContext(ctx, `
			ALTER TABLE sessions
			ADD COLUMN IF NOT EXISTS access_token_hash VARCHAR(255)
		`)
		if err != nil {
			return fmt.Errorf("failed to add access_token_hash column: %w", err)
		}

		// Create index for efficient lookup by access token hash
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_sessions_access_token_hash
			ON sessions(access_token_hash)
			WHERE access_token_hash IS NOT NULL AND revoked_at IS NULL
		`)
		if err != nil {
			return fmt.Errorf("failed to create access_token_hash index: %w", err)
		}

		fmt.Println(" done.")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Removing access_token_hash column from sessions table...")

		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS idx_sessions_access_token_hash
		`)
		if err != nil {
			return fmt.Errorf("failed to drop index: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			ALTER TABLE sessions
			DROP COLUMN IF EXISTS access_token_hash
		`)
		if err != nil {
			return fmt.Errorf("failed to drop access_token_hash column: %w", err)
		}

		fmt.Println(" done.")
		return nil
	})
}
