package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/smilemakc/auth-gateway/internal/config"
)

// Database represents the database connection
type Database struct {
	*bun.DB
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
	db := bun.NewDB(sqldb, pgdialect.New())

	// Add query logger for development (optional)
	// TODO: Add environment check when DatabaseConfig has Environment field
	// Uncomment below to enable query logging:
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	return &Database{db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Health checks the database health
func (d *Database) Health() error {
	ctx := context.Background()
	return d.DB.PingContext(ctx)
}
