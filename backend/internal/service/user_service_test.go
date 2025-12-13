package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUserService_GetProfile(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockUserStore.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
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
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserStore.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.GetProfile(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewUserService(mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()
	req := &models.UpdateUserRequest{
		FullName: "Updated Name",
	}

	t.Run("Success", func(t *testing.T) {
		// Mock GetByID
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{
				ID:       userID,
				FullName: "Old Name",
			}, nil
		}

		// Mock Update
		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, "Updated Name", user.FullName)
			return nil
		}

		// Mock GetByIDWithRoles (reload)
		mockUserStore.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{
				ID:       userID,
				FullName: "Updated Name",
			}, nil
		}

		// Mock Audit
		mockAuditLogger.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionUpdateProfile, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Updated Name", user.FullName)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("UpdateFailure", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}

		mockUserStore.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return errors.New("db error")
		}

		mockAuditLogger.LogWithActionFunc = func(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
			assert.Equal(t, models.ActionUpdateProfile, models.AuditAction(action))
			assert.Equal(t, models.StatusFailed, models.AuditStatus(status))
		}

		// Note: The service uses s.logAudit helper, which calls s.auditService.Log(AuditLogParams).
		// BUT wait, logAudit helper in UserService calls s.auditService.Log(params).
		// In my test above `logAudit` calls `s.auditService.Log`, NOT `LogWithAction`.
		// But I mocked `LogFunc` in the Success case.
		// In UpdateFailure, `s.logAudit` is called.
		// `UserService.logAudit` -> `s.auditService.Log(params)`.
		// So checking `LogWithActionFunc` is WRONG because UserService calls `.Log()`.
		// `AuditService` (the struct) has `Log` and `LogWithAction`.
		// `AuditLogger` (the interface) has `Log(params)`.
		// `UserService` calls `s.auditService.Log`.

		// Fix mock for failure case to check LogFunc
		mockAuditLogger.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionUpdateProfile, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		user, err := service.UpdateProfile(ctx, userID, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
