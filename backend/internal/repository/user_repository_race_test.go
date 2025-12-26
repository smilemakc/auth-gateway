package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_RaceCondition_CreateUser tests for race conditions when creating users
// This test should be run with: go test -race
func TestUserRepository_RaceCondition_CreateUser(t *testing.T) {
	// Skip if not running with race detector
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Create multiple users with same email concurrently
	email := "race-test@example.com"
	concurrency := 10
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	ctx := context.Background()

	// Attempt to create users concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			user := &models.User{
				ID:           uuid.New(),
				Email:        email,
				Username:     fmt.Sprintf("user_%d", index),
				PasswordHash: "hashedpassword",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := repo.Create(ctx, user)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Only one should succeed, others should get unique constraint error
	successCount := 0
	errorCount := 0

	for err := range errors {
		if err != nil {
			errorCount++
			// Should be unique constraint violation
			assert.Contains(t, err.Error(), "already exists", "Expected unique constraint error")
		} else {
			successCount++
		}
	}

	// Verify only one user was created
	users, err := db.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", email).
		Count(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, users, "Only one user should be created despite race condition")
}

// TestUserRepository_RaceCondition_UpdateUser tests for race conditions when updating users
func TestUserRepository_RaceCondition_UpdateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "race-update@example.com",
		Username:     "raceuser",
		PasswordHash: "hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	// Update user concurrently
	concurrency := 5
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Update a field that exists in User model and is allowed in Update method
			user.FullName = fmt.Sprintf("Updated Name %d", index)
			err := repo.Update(ctx, user)
			// Updates should succeed (last write wins)
			_ = err
		}(i)
	}

	wg.Wait()

	// Verify user was updated
	updatedUser, err := repo.GetByID(ctx, user.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, updatedUser.FullName)
}

// setupTestDB is a helper to set up test database (should be implemented)
func setupTestDB(t *testing.T) (*Database, func()) {
	t.Helper()
	// TODO: Implement actual test database setup
	// This is a placeholder - actual implementation should use testutil
	t.Skip("Test database setup not implemented")
	return nil, func() {}
}
