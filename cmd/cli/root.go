package main

import (
	"fmt"
	"os"

	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
	db  *repository.Database
)

var rootCmd = &cobra.Command{
	Use:   "auth-gateway-cli",
	Short: "Auth Gateway CLI management tool",
	Long:  `A command-line tool for managing Auth Gateway users, roles, and other administrative tasks.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip initialization for help commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize database
		db, err = repository.NewDatabase(&cfg.Database)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Clean up database connection
		if db != nil {
			db.Close()
		}
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
