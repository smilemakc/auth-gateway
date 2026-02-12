package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating compliance features tables...")

		// Create password_history table
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS password_history (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				password_hash TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create password_history table: %w", err)
		}

		// Create account_lockouts table
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS account_lockouts (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
				failed_attempts INTEGER NOT NULL DEFAULT 0,
				locked_until TIMESTAMP,
				last_failed_attempt TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create account_lockouts table: %w", err)
		}

		// Create indexes
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_password_history_user_id ON password_history(user_id);
			CREATE INDEX IF NOT EXISTS idx_password_history_created_at ON password_history(created_at);
			CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_id ON account_lockouts(user_id);
			CREATE INDEX IF NOT EXISTS idx_account_lockouts_locked_until ON account_lockouts(locked_until) WHERE locked_until IS NOT NULL;
		`)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Add password_expires_at to users table if not exists
		_, err = db.ExecContext(ctx, `
			ALTER TABLE users 
			ADD COLUMN IF NOT EXISTS password_expires_at TIMESTAMP;
		`)
		if err != nil {
			return fmt.Errorf("failed to add password_expires_at column: %w", err)
		}

		// Add password_changed_at to users table if not exists
		_, err = db.ExecContext(ctx, `
			ALTER TABLE users 
			ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMP;
		`)
		if err != nil {
			return fmt.Errorf("failed to add password_changed_at column: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping compliance features tables...")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS account_lockouts CASCADE;
			DROP TABLE IF EXISTS password_history CASCADE;
			ALTER TABLE users DROP COLUMN IF EXISTS password_changed_at;
			ALTER TABLE users DROP COLUMN IF EXISTS password_expires_at;
		`)
		if err != nil {
			return fmt.Errorf("failed to drop compliance features: %w", err)
		}

		fmt.Println(" OK")
		return nil
	})
}
