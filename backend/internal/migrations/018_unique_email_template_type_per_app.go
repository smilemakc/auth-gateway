package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Remove duplicate templates per (type, application_id), keeping the latest one
		_, err := db.ExecContext(ctx, `
			DELETE FROM email_templates
			WHERE id NOT IN (
				SELECT DISTINCT ON (type, COALESCE(application_id, '00000000-0000-0000-0000-000000000000'))
					id
				FROM email_templates
				ORDER BY type, COALESCE(application_id, '00000000-0000-0000-0000-000000000000'), updated_at DESC
			);
		`)
		if err != nil {
			return err
		}

		// Add unique constraint: one template per type per application (NULL app = global)
		_, err = db.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_email_templates_type_app_unique
			ON email_templates (type, COALESCE(application_id, '00000000-0000-0000-0000-000000000000'));
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS idx_email_templates_type_app_unique;
		`)
		return err
	})
}
