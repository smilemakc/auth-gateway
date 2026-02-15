package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			ADD COLUMN IF NOT EXISTS allowed_grpc_scopes JSONB DEFAULT '[]';
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			ALTER TABLE applications
			DROP COLUMN IF EXISTS allowed_grpc_scopes;
		`)
		return err
	})
}
