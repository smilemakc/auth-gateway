package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSCIMService() (*SCIMService, *mockUserStore, *mockGroupStore) {
	mUser := &mockUserStore{}
	mGroup := newMockGroupStore()
	log := logger.New("test", logger.InfoLevel, false)
	svc := NewSCIMService(mUser, mGroup, log, "https://api.example.com")
	return svc, mUser, mGroup
}

func TestSCIMService_GetUsers(t *testing.T) {
	svc, mUser, _ := setupSCIMService()
	ctx := context.Background()

	t.Run("get users with pagination", func(t *testing.T) {
		users := []*models.User{
			{ID: uuid.New(), Email: "user1@example.com", Username: "user1"},
			{ID: uuid.New(), Email: "user2@example.com", Username: "user2"},
		}

		mUser.ListFunc = func(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error) {
			return users, nil
		}
		mUser.CountFunc = func(ctx context.Context, isActive *bool) (int, error) {
			return 2, nil
		}

		result, err := svc.GetUsers(ctx, "", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalResults)
		assert.Len(t, result.Resources, 2)
	})
}

func TestSCIMService_GetUser(t *testing.T) {
	svc, mUser, _ := setupSCIMService()
	ctx := context.Background()

	t.Run("get existing user", func(t *testing.T) {
		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "user",
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, models.ErrNotFound
		}

		scimUser, err := svc.GetUser(ctx, userID.String())
		assert.NoError(t, err)
		assert.NotNil(t, scimUser)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return nil, models.ErrNotFound
		}

		_, err := svc.GetUser(ctx, uuid.New().String())
		assert.Error(t, err)
	})
}

func TestSCIMService_CreateUser(t *testing.T) {
	svc, mUser, _ := setupSCIMService()
	ctx := context.Background()

	t.Run("create user successfully", func(t *testing.T) {
		scimUser := &models.SCIMUser{
			Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
			UserName: "newuser@example.com",
			Emails: []models.SCIMEmail{
				{Value: "newuser@example.com", Primary: true},
			},
			Name: models.SCIMName{
				GivenName:  "New",
				FamilyName: "User",
			},
			Active: true,
		}

		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) {
			return false, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			user.ID = uuid.New()
			return nil
		}

		result, err := svc.CreateUser(ctx, scimUser)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestSCIMService_UpdateUser(t *testing.T) {
	svc, mUser, _ := setupSCIMService()
	ctx := context.Background()

	t.Run("update user successfully", func(t *testing.T) {
		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "user",
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, models.ErrNotFound
		}
		mUser.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}

		scimUser := &models.SCIMUser{
			Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
			UserName: "updated@example.com",
			Emails: []models.SCIMEmail{
				{Value: "updated@example.com", Primary: true},
			},
		}

		result, err := svc.UpdateUser(ctx, userID.String(), scimUser)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestSCIMService_DeleteUser(t *testing.T) {
	svc, mUser, _ := setupSCIMService()
	ctx := context.Background()

	t.Run("delete user successfully", func(t *testing.T) {
		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			IsActive: true,
		}

		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, models.ErrNotFound
		}
		mUser.UpdateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}

		err := svc.DeleteUser(ctx, userID.String())
		assert.NoError(t, err)
	})
}

func TestSCIMService_GetGroups(t *testing.T) {
	svc, _, mGroup := setupSCIMService()
	ctx := context.Background()

	t.Run("get groups with pagination", func(t *testing.T) {
		groups := []*models.Group{
			{ID: uuid.New(), Name: "group1", DisplayName: "Group 1"},
			{ID: uuid.New(), Name: "group2", DisplayName: "Group 2"},
		}

		// Use the List method from mockGroupStore
		for _, g := range groups {
			mGroup.Create(context.Background(), g)
		}

		result, err := svc.GetGroups(ctx, "", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.TotalResults, 2)
	})
}

func TestSCIMService_GetGroup(t *testing.T) {
	svc, _, mGroup := setupSCIMService()
	ctx := context.Background()

	t.Run("get existing group", func(t *testing.T) {
		group := &models.Group{
			Name:        "test-group",
			DisplayName: "Test Group",
		}
		err := mGroup.Create(ctx, group)
		require.NoError(t, err)

		scimGroup, err := svc.GetGroup(ctx, group.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, scimGroup)
	})
}
