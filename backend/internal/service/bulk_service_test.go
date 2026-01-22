package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func setupBulkService() (*BulkService, *mockUserStore, *mockRBACStore) {
	mUser := &mockUserStore{}
	mRBAC := &mockRBACStore{}
	log := logger.New("test", logger.InfoLevel, false)
	svc := NewBulkService(mUser, mRBAC, log, 10)
	return svc, mUser, mRBAC
}

func TestBulkService_BulkCreateUsers(t *testing.T) {
	svc, mUser, _ := setupBulkService()
	ctx := context.Background()

	t.Run("create users successfully", func(t *testing.T) {
		req := &models.BulkCreateUsersRequest{
			Users: []models.BulkUserCreate{
				{
					Email:    "user1@example.com",
					Username: "user1",
					Password: "password123",
					FullName: "User One",
				},
				{
					Email:    "user2@example.com",
					Username: "user2",
					Password: "password123",
					FullName: "User Two",
				},
			},
		}

		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) {
			return false, nil
		}
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
			return false, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			user.ID = uuid.New()
			return nil
		}

		result, err := svc.BulkCreateUsers(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 2, result.Success)
		assert.Equal(t, 0, result.Failed)
	})

	t.Run("handle validation errors", func(t *testing.T) {
		req := &models.BulkCreateUsersRequest{
			Users: []models.BulkUserCreate{
				{
					Email:    "", // Invalid
					Username: "user1",
					Password: "password123",
				},
				{
					Email:    "user2@example.com",
					Username: "user2",
					Password: "password123",
				},
			},
		}

		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) {
			return false, nil
		}
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
			return false, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			user.ID = uuid.New()
			return nil
		}

		result, err := svc.BulkCreateUsers(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 1, result.Success)
		assert.Equal(t, 1, result.Failed)
		assert.Len(t, result.Errors, 1)
	})
}

func TestBulkService_BulkUpdateUsers(t *testing.T) {
	svc, mUser, _ := setupBulkService()
	ctx := context.Background()

	t.Run("update users successfully", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		req := &models.BulkUpdateUsersRequest{
			Users: []models.BulkUserUpdate{
				{
					ID:       userID1,
					FullName: stringPtr("Updated One"),
				},
				{
					ID:       userID2,
					FullName: stringPtr("Updated Two"),
				},
			},
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, Email: "test@example.com"}, nil
		}
		mUser.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}

		result, err := svc.BulkUpdateUsers(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 2, result.Success)
	})
}

func TestBulkService_BulkDeleteUsers(t *testing.T) {
	svc, mUser, _ := setupBulkService()
	ctx := context.Background()

	t.Run("delete users successfully", func(t *testing.T) {
		req := &models.BulkDeleteUsersRequest{
			UserIDs: []uuid.UUID{uuid.New(), uuid.New()},
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, IsActive: true}, nil
		}
		mUser.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}

		result, err := svc.BulkDeleteUsers(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 2, result.Success)
	})
}

func TestBulkService_BulkAssignRoles(t *testing.T) {
	svc, mUser, mRBAC := setupBulkService()
	ctx := context.Background()

	t.Run("assign roles successfully", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()
		roleID := uuid.New()
		assignedBy := uuid.New()

		req := &models.BulkAssignRolesRequest{
			UserIDs: []uuid.UUID{userID1, userID2},
			RoleIDs: []uuid.UUID{roleID},
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id}, nil
		}
		mRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id}, nil
		}
		mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
			return nil
		}

		result, err := svc.BulkAssignRoles(ctx, req, assignedBy)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 2, result.Success)
	})
}
