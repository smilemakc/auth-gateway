package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

// TestDatabase creates a test database connection
func TestDatabase(t *testing.T) (*repository.Database, func()) {
	t.Helper()

	// Use test database configuration
	cfg := &config.DatabaseConfig{
		Host:         getEnv("TEST_DB_HOST", "localhost"),
		Port:         getEnv("TEST_DB_PORT", "5432"),
		User:         getEnv("TEST_DB_USER", "postgres"),
		Password:     getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:       getEnv("TEST_DB_NAME", "auth_gateway_test"),
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	db, err := repository.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Logf("Error closing test database: %v", err)
		}
	}

	return db, cleanup
}

// SetupTestDB sets up a test database with migrations
func SetupTestDB(t *testing.T) (*repository.Database, func()) {
	t.Helper()

	db, cleanup := TestDatabase(t)

	// Run migrations
	_ = context.Background() // TODO: Use context for migrations

	// TODO: Run migrations here
	// For now, just verify connection
	if err := db.Health(); err != nil {
		t.Fatalf("Test database health check failed: %v", err)
	}

	return db, cleanup
}

// CleanupTestData cleans up test data from database
func CleanupTestData(t *testing.T, db *repository.Database, tables ...string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, table := range tables {
		_, err := db.NewDelete().Table(table).Exec(ctx)
		if err != nil {
			t.Logf("Error cleaning up table %s: %v", table, err)
		}
	}
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestRedis creates a test Redis connection (placeholder)
func TestRedis(t *testing.T) (interface{}, func()) {
	t.Helper()
	// TODO: Implement Redis test connection
	return nil, func() {}
}

// WaitForDB waits for database to be ready
func WaitForDB(db *repository.Database, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("database not ready within timeout")
		case <-ticker.C:
			if err := db.Health(); err == nil {
				return nil
			}
		}
	}
}
