package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating email profiles system tables...")
		return createEmailProfilesTables(ctx, db)
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping email profiles system tables...")
		return dropEmailProfilesTables(ctx, db)
	})
}

func createEmailProfilesTables(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS email_providers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,

			smtp_host VARCHAR(255),
			smtp_port INTEGER,
			smtp_username VARCHAR(255),
			smtp_password VARCHAR(500),
			smtp_use_tls BOOLEAN DEFAULT true,

			sendgrid_api_key VARCHAR(500),

			aws_region VARCHAR(50),
			aws_access_key_id VARCHAR(255),
			aws_secret_access_key VARCHAR(500),

			mailgun_domain VARCHAR(255),
			mailgun_api_key VARCHAR(500),

			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_providers table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS email_profiles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			provider_id UUID NOT NULL REFERENCES email_providers(id) ON DELETE RESTRICT,
			from_email VARCHAR(255) NOT NULL,
			from_name VARCHAR(255) NOT NULL,
			reply_to VARCHAR(255),
			is_default BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_profiles table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS email_profile_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			profile_id UUID NOT NULL REFERENCES email_profiles(id) ON DELETE CASCADE,
			otp_type VARCHAR(50) NOT NULL,
			template_id UUID NOT NULL REFERENCES email_templates(id) ON DELETE RESTRICT,
			UNIQUE(profile_id, otp_type)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_profile_templates table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS email_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			profile_id UUID REFERENCES email_profiles(id) ON DELETE SET NULL,
			recipient_email VARCHAR(255) NOT NULL,
			subject VARCHAR(500) NOT NULL,
			template_type VARCHAR(50),
			provider_type VARCHAR(50) NOT NULL,
			message_id VARCHAR(255),
			status VARCHAR(20) NOT NULL,
			error_message TEXT,
			sent_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			ip_address VARCHAR(45)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_logs table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_email_providers_active ON email_providers(is_active);
		CREATE INDEX IF NOT EXISTS idx_email_profiles_default ON email_profiles(is_default) WHERE is_default = true;
		CREATE INDEX IF NOT EXISTS idx_email_profiles_provider ON email_profiles(provider_id);
		CREATE INDEX IF NOT EXISTS idx_email_profile_templates_profile ON email_profile_templates(profile_id);
		CREATE INDEX IF NOT EXISTS idx_email_logs_recipient ON email_logs(recipient_email);
		CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);
		CREATE INDEX IF NOT EXISTS idx_email_logs_created ON email_logs(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_email_logs_user ON email_logs(user_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION update_email_providers_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_email_providers_updated_at ON email_providers;
		CREATE TRIGGER trigger_update_email_providers_updated_at
			BEFORE UPDATE ON email_providers
			FOR EACH ROW
			EXECUTE FUNCTION update_email_providers_updated_at();

		CREATE OR REPLACE FUNCTION update_email_profiles_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_email_profiles_updated_at ON email_profiles;
		CREATE TRIGGER trigger_update_email_profiles_updated_at
			BEFORE UPDATE ON email_profiles
			FOR EACH ROW
			EXECUTE FUNCTION update_email_profiles_updated_at();
	`)
	if err != nil {
		return fmt.Errorf("failed to create update triggers: %w", err)
	}

	fmt.Println(" OK")
	return nil
}

func dropEmailProfilesTables(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS email_logs CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop email_logs table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS email_profile_templates CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop email_profile_templates table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS email_profiles CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop email_profiles table: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS email_providers CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to drop email_providers table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		DROP FUNCTION IF EXISTS update_email_providers_updated_at() CASCADE;
		DROP FUNCTION IF EXISTS update_email_profiles_updated_at() CASCADE;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop update functions: %w", err)
	}

	fmt.Println(" OK")
	return nil
}
