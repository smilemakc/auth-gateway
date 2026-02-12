package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/smilemakc/auth-gateway/internal/config"
)

// Database represents the database connection
type Database struct {
	*bun.DB
	sqlDB *sql.DB // Keep reference to sql.DB for stats
}

// NewDatabase creates a new database connection using bun ORM
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	// Create pgdriver connector
	pgconn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)),
		pgdriver.WithUser(cfg.User),
		pgdriver.WithPassword(cfg.Password),
		pgdriver.WithDatabase(cfg.DBName),
		pgdriver.WithInsecure(cfg.SSLMode == "disable"),
	)

	// Create sql.DB from connector
	sqldb := sql.OpenDB(pgconn)

	// Set connection pool settings (keep existing values)
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create bun.DB with PostgreSQL dialect
	bunDB := bun.NewDB(sqldb, pgdialect.New())

	// Add query logger only in development/staging
	// Check both explicit flag and environment
	enableQueryLog := cfg.EnableQueryLog
	if !enableQueryLog {
		// Fallback to environment check if flag not set
		env := os.Getenv("ENV")
		if env == "" {
			env = os.Getenv("SERVER_ENV")
		}
		enableQueryLog = env == "development" || env == "dev" || env == "staging"
	}

	if enableQueryLog {
		bunDB.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	// Register models in order: join tables FIRST, then base models
	// This is required because bun processes m2m relations when registering models
	// and needs the join table models to be already registered
	// 1. Register all join tables (m2m) FIRST
	bunDB.RegisterModel((*models.UserRole)(nil))
	bunDB.RegisterModel((*models.UserGroup)(nil))
	bunDB.RegisterModel((*models.RolePermission)(nil))
	// 2. Register base models (they can now safely use m2m relations)
	bunDB.RegisterModel((*models.Permission)(nil))
	bunDB.RegisterModel((*models.Role)(nil))
	bunDB.RegisterModel((*models.User)(nil))
	bunDB.RegisterModel((*models.Group)(nil))
	return &Database{DB: bunDB, sqlDB: sqldb}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Health checks the database health
func (d *Database) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.DB.PingContext(ctx)
}

// Stats returns database connection pool statistics
func (d *Database) Stats() sql.DBStats {
	if d.sqlDB != nil {
		return d.sqlDB.Stats()
	}
	return sql.DBStats{}
}

// RunInTx runs a function within a database transaction
// If the function returns an error, the transaction is rolled back
func (d *Database) RunInTx(ctx context.Context, fn func(context.Context, bun.Tx) error) error {
	return d.DB.RunInTx(ctx, nil, fn)
}

// WithTransaction is a helper method that runs a function within a transaction
// It provides a cleaner API and handles transaction lifecycle
func (d *Database) WithTransaction(ctx context.Context, fn func(context.Context, bun.Tx) error) error {
	return d.RunInTx(ctx, fn)
}

// RetryTransaction retries a transaction function on deadlock or serialization errors
// maxRetries: maximum number of retry attempts (default: 3)
// backoff: time to wait between retries (default: 100ms, exponential backoff)
func (d *Database) RetryTransaction(ctx context.Context, fn func(context.Context, bun.Tx) error, maxRetries int, initialBackoff time.Duration) error {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if initialBackoff <= 0 {
		initialBackoff = 100 * time.Millisecond
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := d.RunInTx(ctx, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable (deadlock, serialization failure, etc.)
		if !isRetryableError(err) {
			return err
		}

		// Don't sleep on last attempt
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// Exponential backoff
				backoff = backoff * 2
			}
		}
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError checks if an error is retryable (deadlock, serialization failure, etc.)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a RetryableError
	if IsRetryableError(err) {
		return true
	}

	errStr := err.Error()

	// PostgreSQL error codes for retryable errors:
	// 40001 - serialization_failure
	// 40P01 - deadlock_detected
	// 08006 - connection_failure
	// 08003 - connection_does_not_exist
	retryablePatterns := []string{
		"serialization_failure",
		"deadlock_detected",
		"deadlock",
		"connection_failure",
		"connection_does_not_exist",
		"could not serialize",
		"deadlock detected",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
