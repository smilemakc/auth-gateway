package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			ADD COLUMN IF NOT EXISTS secret_hash TEXT DEFAULT '',
			ADD COLUMN IF NOT EXISTS secret_prefix VARCHAR(12) DEFAULT '',
			ADD COLUMN IF NOT EXISTS secret_last_rotated_at TIMESTAMPTZ;
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			DROP COLUMN IF EXISTS secret_hash,
			DROP COLUMN IF EXISTS secret_prefix,
			DROP COLUMN IF EXISTS secret_last_rotated_at;
		`)
		return err
	})
}
