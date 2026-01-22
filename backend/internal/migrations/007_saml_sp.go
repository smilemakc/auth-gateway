package migrations

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating saml_service_providers table...")

		_, err := db.NewCreateTable().
			Model((*models.SAMLServiceProvider)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create saml_service_providers table: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Dropping saml_service_providers table...")

		_, err := db.NewDropTable().
			Model((*models.SAMLServiceProvider)(nil)).
			IfExists().
			Cascade().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop saml_service_providers table: %w", err)
		}

		fmt.Println(" OK")
		return nil
	})
}
