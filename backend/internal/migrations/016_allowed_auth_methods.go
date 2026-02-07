package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			ADD COLUMN IF NOT EXISTS allowed_auth_methods JSONB DEFAULT '["password"]'::jsonb;
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			DROP COLUMN IF EXISTS allowed_auth_methods;
		`)
		return err
	})
}
