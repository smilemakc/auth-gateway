package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Adding unique constraint on token_blacklist.token_hash...")

		// Drop existing non-unique index if exists
		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS idx_token_blacklist_token_hash
		`)
		if err != nil {
			return fmt.Errorf("failed to drop old index: %w", err)
		}

		// Create unique index for ON CONFLICT to work
		_, err = db.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_token_blacklist_token_hash_unique
			ON token_blacklist(token_hash)
		`)
		if err != nil {
			return fmt.Errorf("failed to create unique index on token_hash: %w", err)
		}

		fmt.Println(" done.")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Reverting unique constraint on token_blacklist.token_hash...")

		// Drop unique index
		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS idx_token_blacklist_token_hash_unique
		`)
		if err != nil {
			return fmt.Errorf("failed to drop unique index: %w", err)
		}

		// Recreate non-unique index
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_token_blacklist_token_hash
			ON token_blacklist(token_hash)
		`)
		if err != nil {
			return fmt.Errorf("failed to recreate index: %w", err)
		}

		fmt.Println(" done.")
		return nil
	})
}
