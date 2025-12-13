package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestAdminService_GetStats(t *testing.T) {
	mockUser := &mockUserStore{}
	mockAPIKey := &mockAPIKeyStore{}
	mockAudit := &mockAuditStore{}
	mockOAuth := &mockOAuthStore{}
	mockRBAC := &mockRBACStore{}

	svc := NewAdminService(mockUser, mockAPIKey, mockAudit, mockOAuth, mockRBAC)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Mock User Stats
		mockUser.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			return 10, nil
		}
		mockUser.ListFunc = func(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error) {
			return []*models.User{
				{ID: uuid.New(), IsActive: true, EmailVerified: true},
				{ID: uuid.New(), IsActive: false},
			}, nil
		}

		// Mock RBAC Stats (User Roles)
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
			return []models.Role{{Name: "admin"}, {Name: "user"}}, nil
		}

		// Mock API Keys
		mockAPIKey.ListAllFunc = func(ctx context.Context) ([]*models.APIKey, error) {
			return []*models.APIKey{{IsActive: true}, {IsActive: false}}, nil
		}

		// Mock OAuth Accounts
		mockOAuth.ListAllFunc = func(ctx context.Context) ([]*models.OAuthAccount, error) {
			return []*models.OAuthAccount{{}}, nil
		}

		// Mock Audit Logs
		mockAudit.CountByActionSinceFunc = func(ctx context.Context, action models.AuditAction, since time.Time) (int, error) {
			return 5, nil
		}

		stats, err := svc.GetStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 10, stats.TotalUsers)
		assert.Equal(t, 1, stats.ActiveUsers) // From the list of 2
		assert.Equal(t, 1, stats.VerifiedEmailUsers)
		assert.Equal(t, 2, stats.TotalAPIKeys)
		assert.Equal(t, 1, stats.ActiveAPIKeys)
		assert.Equal(t, 1, stats.TotalOAuthAccounts)
		assert.Equal(t, 5, stats.RecentLogins)
		assert.Equal(t, 2, stats.UsersByRole["admin"])
		assert.Equal(t, 2, stats.UsersByRole["user"])
	})
}

func TestAdminService_ListUsers(t *testing.T) {
	mockUser := &mockUserStore{}
	mockAPIKey := &mockAPIKeyStore{}
	mockOAuth := &mockOAuthStore{}

	svc := NewAdminService(mockUser, mockAPIKey, nil, mockOAuth, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockUser.ListWithRolesFunc = func(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error) {
			return []*models.User{{ID: userID, Username: "test"}}, nil
		}
		mockUser.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			return 1, nil
		}
		mockAPIKey.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*models.APIKey, error) {
			return []*models.APIKey{{}}, nil
		}
		mockOAuth.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*models.OAuthAccount, error) {
			return []*models.OAuthAccount{{}}, nil
		}

		resp, err := svc.ListUsers(ctx, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Users))
		assert.Equal(t, "test", resp.Users[0].Username)
		assert.Equal(t, 1, resp.Users[0].APIKeysCount)
		assert.Equal(t, 1, resp.Users[0].OAuthAccountsCount)
	})
}

func TestAdminService_UpdateUser(t *testing.T) {
	mockUser := &mockUserStore{}
	mockRBAC := &mockRBACStore{}
	mockAPIKey := &mockAPIKeyStore{} // Added to satisfy constructor
	mockOAuth := &mockOAuthStore{}   // Added to satisfy constructor

	svc := NewAdminService(mockUser, mockAPIKey, nil, mockOAuth, mockRBAC)
	ctx := context.Background()

	t.Run("Success_UpdateRolesAndActive", func(t *testing.T) {
		userID := uuid.New()
		adminID := uuid.New()
		roleID := uuid.New()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, includeRoles *bool) (*models.User, error) {
			return &models.User{ID: id, IsActive: true}, nil
		}
		mockUser.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, IsActive: false}, nil // Return updated state simulation
		}

		mockRBAC.SetUserRolesFunc = func(ctx context.Context, uid uuid.UUID, rIDs []uuid.UUID, assignedBy uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, roleID, rIDs[0])
			assert.Equal(t, adminID, assignedBy)
			return nil
		}

		mockUser.UpdateFunc = func(ctx context.Context, user *models.User) error {
			assert.False(t, user.IsActive)
			return nil
		}

		// Mock APIKey/OAuth calls in GetUser
		mockAPIKey.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*models.APIKey, error) { return nil, nil }
		mockOAuth.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) ([]*models.OAuthAccount, error) { return nil, nil }

		req := &models.AdminUpdateUserRequest{
			RoleIDs:  &[]uuid.UUID{roleID},
			IsActive: utils.Ptr(false),
		}

		resp, err := svc.UpdateUser(ctx, userID, req, adminID)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsActive)
	})
}
