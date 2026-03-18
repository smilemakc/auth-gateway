package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Drop the old absolute unique constraint on email (created by Bun's "unique" tag).
		// This blocks users with empty email (e.g. phone-only or Telegram auth).
		_, err := db.ExecContext(ctx, `ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;`)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `DROP INDEX IF EXISTS users_email_key;`)
		if err != nil {
			return err
		}

		// Create a partial unique index: uniqueness is enforced only for non-empty emails.
		// Multiple rows with email='' are allowed.
		_, err = db.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
			ON users(email) WHERE email != '';
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_users_email_unique;`)
		if err != nil {
			return err
		}
		// Restore the original constraint (not just an index) for Bun ORM compatibility
		_, err = db.ExecContext(ctx, `
			ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
		`)
		return err
	})
}
