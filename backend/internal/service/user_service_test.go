package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// GetProfile Tests
// ============================================================

func TestUserService_GetProfile(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			assert.Equal(t, userID, id)
			assert.Nil(t, isActive)
			return &models.User{
				ID:       userID,
				Email:    "test@example.com",
				FullName: "Test User",
			}, nil
		}

		user, err := service.GetProfile(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.FullName)
		// Verify PublicUser strips password hash
		assert.Empty(t, user.PasswordHash)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.GetProfile(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// ============================================================
// UpdateProfile Tests
// ============================================================

func TestUserService_UpdateProfile(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			FullName: "Updated Name",
		}

		callCount := 0
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			callCount++
			if callCount == 1 {
				return &models.User{
					ID:       userID,
					FullName: "Old Name",
				}, nil
			}
			// Second call - reload with roles
			return &models.User{
				ID:       userID,
				FullName: "Updated Name",
			}, nil
		}

		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, "Updated Name", user.FullName)
			return nil
		}

		mockAuditLogger.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionUpdateProfile, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "Updated Name", user.FullName)
	})

	t.Run("Success_UpdateProfilePictureURL", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			ProfilePictureURL: "https://example.com/avatar.jpg",
		}

		callCount := 0
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			callCount++
			if callCount == 1 {
				return &models.User{ID: userID, FullName: "Original"}, nil
			}
			return &models.User{ID: userID, FullName: "Original", ProfilePictureURL: "https://example.com/avatar.jpg"}, nil
		}

		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, "https://example.com/avatar.jpg", user.ProfilePictureURL)
			return nil
		}

		mockAuditLogger.LogFunc = func(params AuditLogParams) {}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "https://example.com/avatar.jpg", user.ProfilePictureURL)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := &models.UpdateUserRequest{FullName: "Test"}
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("UpdateFailure", func(t *testing.T) {
		req := &models.UpdateUserRequest{FullName: "Updated Name"}
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}

		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return errors.New("db error")
		}

		mockAuditLogger.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionUpdateProfile, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("ReloadAfterUpdateFails", func(t *testing.T) {
		req := &models.UpdateUserRequest{FullName: "Updated"}

		callCount := 0
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			callCount++
			if callCount == 1 {
				return &models.User{ID: userID, FullName: "Old"}, nil
			}
			// Second call fails
			return nil, errors.New("db error on reload")
		}
		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}
		mockAuditLogger.LogFunc = func(params AuditLogParams) {}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "db error on reload", err.Error())
	})
}

// ============================================================
// GetByID Tests
// ============================================================

func TestUserService_GetByID(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			assert.Equal(t, userID, id)
			assert.Nil(t, isActive)
			return &models.User{
				ID:           userID,
				Email:        "john@example.com",
				Username:     "john",
				PasswordHash: "secret_hash",
			}, nil
		}

		user, err := service.GetByID(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Empty(t, user.PasswordHash, "PublicUser should strip PasswordHash")
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// ============================================================
// GetByEmail Tests
// ============================================================

func TestUserService_GetByEmail(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockUserStore.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			assert.Equal(t, "test@example.com", email)
			assert.Nil(t, isActive)
			return &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: "hashed",
			}, nil
		}

		user, err := service.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Empty(t, user.PasswordHash, "PublicUser should strip PasswordHash")
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserStore.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// ============================================================
// List Tests
// ============================================================

func TestUserService_List(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		users := []*models.User{
			{ID: uuid.New(), Email: "a@test.com", PasswordHash: "hash1"},
			{ID: uuid.New(), Email: "b@test.com", PasswordHash: "hash2"},
			{ID: uuid.New(), Email: "c@test.com", PasswordHash: "hash3"},
		}
		mockUserStore.ListFunc = func(ctx context.Context, opts ...UserListOption) ([]*models.User, error) {
			return users, nil
		}

		result, err := service.List(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "a@test.com", result[0].Email)
		assert.Equal(t, "b@test.com", result[1].Email)
		assert.Equal(t, "c@test.com", result[2].Email)
		// Verify all are public users (PasswordHash stripped)
		for _, u := range result {
			assert.Empty(t, u.PasswordHash, "PublicUser should strip PasswordHash")
		}
	})

	t.Run("Empty", func(t *testing.T) {
		mockUserStore.ListFunc = func(ctx context.Context, opts ...UserListOption) ([]*models.User, error) {
			return []*models.User{}, nil
		}

		result, err := service.List(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		mockUserStore.ListFunc = func(ctx context.Context, opts ...UserListOption) ([]*models.User, error) {
			return nil, errors.New("db error")
		}

		result, err := service.List(ctx, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ============================================================
// Count Tests
// ============================================================

func TestUserService_Count(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserStore.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			assert.Nil(t, isActive)
			return 42, nil
		}

		count, err := service.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	t.Run("Zero", func(t *testing.T) {
		mockUserStore.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			return 0, nil
		}

		count, err := service.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Error", func(t *testing.T) {
		mockUserStore.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			return 0, errors.New("db error")
		}

		count, err := service.Count(ctx)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
	})
}

// ============================================================
// HTML Sanitization in UpdateProfile Tests
// ============================================================

func TestUserService_UpdateProfile_SanitizesHTML(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}
	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("SanitizesFullNameHTML", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			FullName: "<script>alert('xss')</script>John Doe",
		}

		callCount := 0
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			callCount++
			if callCount == 1 {
				return &models.User{ID: userID, FullName: "Old"}, nil
			}
			return &models.User{ID: userID, FullName: "John Doe"}, nil
		}

		var savedFullName string
		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			savedFullName = user.FullName
			return nil
		}
		mockAuditLogger.LogFunc = func(params AuditLogParams) {}

		_, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
		assert.NotContains(t, savedFullName, "<script>", "HTML script tags should be sanitized")
	})
}
