package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"

	"github.com/smilemakc/auth-gateway/internal/migrations"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
	Long:  "Manage database schema migrations using Bun's migration system",
}

var migrateInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize migration system",
	Long:  "Create the migration tracking table in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		migrator := migrate.NewMigrator(db.DB, migrations.Migrations)

		if err := migrator.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize migrations: %w", err)
		}

		fmt.Println("✓ Migration system initialized")
		return nil
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply pending migrations",
	Long:  "Apply all pending migrations or a specific number of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		count, _ := cmd.Flags().GetInt("count")
		migrator := migrate.NewMigrator(db.DB, migrations.Migrations)

		if count > 0 {
			// Apply specific number of migrations
			for i := 0; i < count; i++ {
				group, err := migrator.Migrate(ctx)
				if err != nil {
					return fmt.Errorf("migration failed: %w", err)
				}
				if group.IsZero() {
					fmt.Println("No more pending migrations")
					break
				}
				fmt.Printf("✓ Applied migration: %s\n", group.Migrations[0].Name)
			}
		} else {
			// Apply all pending migrations
			group, err := migrator.Migrate(ctx)
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
			if group.IsZero() {
				fmt.Println("No pending migrations")
			} else {
				fmt.Printf("✓ Applied %d migration(s)\n", len(group.Migrations))
			}
		}

		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback migrations",
	Long:  "Rollback the last migration or a specific number of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		count, _ := cmd.Flags().GetInt("count")
		migrator := migrate.NewMigrator(db.DB, migrations.Migrations)

		for i := 0; i < count; i++ {
			group, err := migrator.Rollback(ctx)
			if err != nil {
				return fmt.Errorf("rollback failed: %w", err)
			}
			if group.IsZero() {
				fmt.Println("No migrations to rollback")
				break
			}
			fmt.Printf("✓ Rolled back: %s\n", group.Migrations[0].Name)
		}

		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Display the status of all migrations (applied or pending)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		migrator := migrate.NewMigrator(db.DB, migrations.Migrations)

		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get migration status: %w", err)
		}

		fmt.Println("Migration Status:")
		fmt.Println("================")
		for _, m := range ms {
			status := "✗ pending"
			if m.GroupID > 0 {
				status = fmt.Sprintf("✓ applied (%s)", m.MigratedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("  %s  %s\n", status, m.Name)
		}

		return nil
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create new migration file",
	Long:  "Create a new migration file (SQL or Go) with the specified name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		migrationType, _ := cmd.Flags().GetString("type")

		// Validate migration type
		if migrationType != "sql" && migrationType != "go" {
			return fmt.Errorf("invalid type: %s (must be 'sql' or 'go')", migrationType)
		}

		// Get next migration number
		migrationsDir := "internal/migrations"
		files, err := os.ReadDir(migrationsDir)
		if err != nil {
			return fmt.Errorf("failed to read migrations directory: %w", err)
		}

		nextNum := 1
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".go") && f.Name() != "migrations.go" {
				// Extract number from filename (e.g., "001_init_schema.go" -> 1)
				parts := strings.Split(f.Name(), "_")
				if len(parts) > 0 {
					num, _ := strconv.Atoi(parts[0])
					if num >= nextNum {
						nextNum = num + 1
					}
				}
			}
			if strings.HasSuffix(f.Name(), ".up.sql") {
				// Extract number from filename (e.g., "001_init.up.sql" -> 1)
				parts := strings.Split(f.Name(), "_")
				if len(parts) > 0 {
					num, _ := strconv.Atoi(parts[0])
					if num >= nextNum {
						nextNum = num + 1
					}
				}
			}
		}

		filename := fmt.Sprintf("%03d_%s", nextNum, name)

		if migrationType == "sql" {
			// Create SQL migration files
			upFile := filepath.Join(migrationsDir, filename+".up.sql")
			downFile := filepath.Join(migrationsDir, filename+".down.sql")

			upContent := []byte(`-- Migration: ` + name + `
-- Description: Add your description here

-- Add your SQL statements here

`)

			downContent := []byte(`-- Rollback migration: ` + name + `
-- Description: Rollback changes from the up migration

-- Add your rollback SQL statements here

`)

			if err := os.WriteFile(upFile, upContent, 0644); err != nil {
				return fmt.Errorf("failed to create up migration file: %w", err)
			}

			if err := os.WriteFile(downFile, downContent, 0644); err != nil {
				return fmt.Errorf("failed to create down migration file: %w", err)
			}

			fmt.Printf("✓ Created SQL migrations:\n")
			fmt.Printf("  %s\n", upFile)
			fmt.Printf("  %s\n", downFile)
		} else if migrationType == "go" {
			// Create Go migration file
			goFile := filepath.Join(migrationsDir, filename+".go")

			template := `package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ` + name + `...")

		// TODO: Implement up migration
		// Example: Create a new table
		// _, err := db.NewCreateTable().
		// 	Model((*models.YourModel)(nil)).
		// 	IfNotExists().
		// 	Exec(ctx)
		// if err != nil {
		// 	return fmt.Errorf("failed to create table: %w", err)
		// }

		// Example: Execute raw SQL
		// _, err := db.ExecContext(ctx, "CREATE INDEX idx_example ON table_name(column_name)")
		// if err != nil {
		// 	return fmt.Errorf("failed to create index: %w", err)
		// }

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ` + name + `...")

		// TODO: Implement down migration (rollback)
		// Example: Drop a table
		// _, err := db.NewDropTable().
		// 	Model((*models.YourModel)(nil)).
		// 	IfExists().
		// 	Cascade().
		// 	Exec(ctx)
		// if err != nil {
		// 	return fmt.Errorf("failed to drop table: %w", err)
		// }

		fmt.Println(" OK")
		return nil
	})
}
`

			if err := os.WriteFile(goFile, []byte(template), 0644); err != nil {
				return fmt.Errorf("failed to create Go migration file: %w", err)
			}

			fmt.Printf("✓ Created Go migration:\n")
			fmt.Printf("  %s\n", goFile)
		}

		return nil
	},
}

func init() {
	// Add subcommands to migrate command
	migrateCmd.AddCommand(migrateInitCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)

	// Add flags
	migrateUpCmd.Flags().IntP("count", "c", 0, "Number of migrations to apply (0 = all)")
	migrateDownCmd.Flags().IntP("count", "c", 1, "Number of migrations to rollback")
	migrateCreateCmd.Flags().StringP("type", "t", "go", "Migration type: sql or go")
}
