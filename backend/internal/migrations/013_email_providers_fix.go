package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Fixing email providers schema...")
		return fixEmailProvidersSchema(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Reverting email providers schema fix...")
		return revertEmailProvidersSchemaFix(ctx, db)
	})
}

func fixEmailProvidersSchema(ctx context.Context, db *bun.DB) error {
	// Add created_by column to email_providers
	_, err := db.ExecContext(ctx, `
		ALTER TABLE email_providers
		ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to add created_by to email_providers: %w", err)
	}

	// Rename AWS columns to SES columns in email_providers
	_, err = db.ExecContext(ctx, `
		DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'aws_region') THEN
				ALTER TABLE email_providers RENAME COLUMN aws_region TO ses_region;
			END IF;
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'aws_access_key_id') THEN
				ALTER TABLE email_providers RENAME COLUMN aws_access_key_id TO ses_access_key_id;
			END IF;
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'aws_secret_access_key') THEN
				ALTER TABLE email_providers RENAME COLUMN aws_secret_access_key TO ses_secret_access_key;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to rename AWS columns to SES: %w", err)
	}

	// Add created_by column to email_profiles
	_, err = db.ExecContext(ctx, `
		ALTER TABLE email_profiles
		ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to add created_by to email_profiles: %w", err)
	}

	// Add created_by and timestamps to email_profile_templates
	_, err = db.ExecContext(ctx, `
		ALTER TABLE email_profile_templates
		ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL,
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to add columns to email_profile_templates: %w", err)
	}

	fmt.Println(" OK")
	return nil
}

func revertEmailProvidersSchemaFix(ctx context.Context, db *bun.DB) error {
	// Remove added columns
	_, err := db.ExecContext(ctx, `
		ALTER TABLE email_providers DROP COLUMN IF EXISTS created_by;
		ALTER TABLE email_profiles DROP COLUMN IF EXISTS created_by;
		ALTER TABLE email_profile_templates DROP COLUMN IF EXISTS created_by;
		ALTER TABLE email_profile_templates DROP COLUMN IF EXISTS created_at;
		ALTER TABLE email_profile_templates DROP COLUMN IF EXISTS updated_at;
	`)
	if err != nil {
		return fmt.Errorf("failed to remove columns: %w", err)
	}

	// Rename SES columns back to AWS columns
	_, err = db.ExecContext(ctx, `
		DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'ses_region') THEN
				ALTER TABLE email_providers RENAME COLUMN ses_region TO aws_region;
			END IF;
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'ses_access_key_id') THEN
				ALTER TABLE email_providers RENAME COLUMN ses_access_key_id TO aws_access_key_id;
			END IF;
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'email_providers' AND column_name = 'ses_secret_access_key') THEN
				ALTER TABLE email_providers RENAME COLUMN ses_secret_access_key TO aws_secret_access_key;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to rename SES columns back to AWS: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
