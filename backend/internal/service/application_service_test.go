package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

type mockApplicationStore struct {
	GetUserProfileFunc func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
}

func (m *mockApplicationStore) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID, applicationID)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) CreateApplication(ctx context.Context, app *models.Application) error {
	return nil
}

func (m *mockApplicationStore) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	return nil, nil
}

func (m *mockApplicationStore) GetApplicationByName(ctx context.Context, name string) (*models.Application, error) {
	return nil, nil
}

func (m *mockApplicationStore) UpdateApplication(ctx context.Context, app *models.Application) error {
	return nil
}

func (m *mockApplicationStore) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockApplicationStore) ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
	return nil, 0, nil
}

func (m *mockApplicationStore) GetBySecretHash(ctx context.Context, hash string) (*models.Application, error) {
	return nil, nil
}

func (m *mockApplicationStore) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	return nil, nil
}

func (m *mockApplicationStore) CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error {
	return nil
}

func (m *mockApplicationStore) CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}

func (m *mockApplicationStore) UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}

func (m *mockApplicationStore) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}

func (m *mockApplicationStore) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	return nil, nil
}

func (m *mockApplicationStore) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
	return nil, 0, nil
}

func (m *mockApplicationStore) UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}

func (m *mockApplicationStore) BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	return nil
}

func (m *mockApplicationStore) UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}

type mockAppOAuthProviderStore struct{}

func (m *mockAppOAuthProviderStore) Create(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}

func (m *mockAppOAuthProviderStore) GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error) {
	return nil, nil
}

func (m *mockAppOAuthProviderStore) GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error) {
	return nil, nil
}

func (m *mockAppOAuthProviderStore) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}

func (m *mockAppOAuthProviderStore) ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}

func (m *mockAppOAuthProviderStore) Update(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}

func (m *mockAppOAuthProviderStore) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func setupApplicationService(mockAppRepo *mockApplicationStore) *ApplicationService {
	mockAppOAuthRepo := &mockAppOAuthProviderStore{}
	return NewApplicationService(mockAppRepo, mockAppOAuthRepo, nil)
}

func TestCheckUserAccess_MissingProfile_ReturnsError(t *testing.T) {
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("profile not found")
		},
	}
	service := setupApplicationService(mockAppRepo)

	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	err := service.CheckUserAccess(ctx, userID, appID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user has no access to application")
}

func TestCheckUserAccess_BannedUser_ReturnsError(t *testing.T) {
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: applicationID,
				IsBanned:      true,
				IsActive:      true,
			}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)

	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	err := service.CheckUserAccess(ctx, userID, appID)
	assert.Error(t, err)
	assert.Equal(t, ErrUserBannedFromApp, err)
}

func TestCheckUserAccess_ActiveUser_ReturnsNil(t *testing.T) {
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: applicationID,
				IsBanned:      false,
				IsActive:      true,
			}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)

	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	err := service.CheckUserAccess(ctx, userID, appID)
	assert.NoError(t, err)
}
