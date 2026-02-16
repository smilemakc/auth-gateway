package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Enhanced mockApplicationStore with function fields for all methods ---

type mockApplicationStore struct {
	CreateApplicationFunc        func(ctx context.Context, app *models.Application) error
	GetApplicationByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetApplicationByNameFunc     func(ctx context.Context, name string) (*models.Application, error)
	UpdateApplicationFunc        func(ctx context.Context, app *models.Application) error
	DeleteApplicationFunc        func(ctx context.Context, id uuid.UUID) error
	ListApplicationsFunc         func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error)
	GetBySecretHashFunc          func(ctx context.Context, hash string) (*models.Application, error)
	GetBrandingFunc              func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error)
	CreateOrUpdateBrandingFunc   func(ctx context.Context, branding *models.ApplicationBranding) error
	CreateUserProfileFunc        func(ctx context.Context, profile *models.UserApplicationProfile) error
	GetUserProfileFunc           func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	UpdateUserProfileFunc        func(ctx context.Context, profile *models.UserApplicationProfile) error
	DeleteUserProfileFunc        func(ctx context.Context, userID, applicationID uuid.UUID) error
	ListUserProfilesFunc         func(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error)
	ListApplicationUsersFunc     func(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error)
	UpdateLastAccessFunc         func(ctx context.Context, userID, applicationID uuid.UUID) error
	BanUserFromApplicationFunc   func(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error
	UnbanUserFromApplicationFunc func(ctx context.Context, userID, applicationID uuid.UUID) error
}

func (m *mockApplicationStore) CreateApplication(ctx context.Context, app *models.Application) error {
	if m.CreateApplicationFunc != nil {
		return m.CreateApplicationFunc(ctx, app)
	}
	return nil
}

func (m *mockApplicationStore) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	if m.GetApplicationByIDFunc != nil {
		return m.GetApplicationByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) GetApplicationByName(ctx context.Context, name string) (*models.Application, error) {
	if m.GetApplicationByNameFunc != nil {
		return m.GetApplicationByNameFunc(ctx, name)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) UpdateApplication(ctx context.Context, app *models.Application) error {
	if m.UpdateApplicationFunc != nil {
		return m.UpdateApplicationFunc(ctx, app)
	}
	return nil
}

func (m *mockApplicationStore) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	if m.DeleteApplicationFunc != nil {
		return m.DeleteApplicationFunc(ctx, id)
	}
	return nil
}

func (m *mockApplicationStore) ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
	if m.ListApplicationsFunc != nil {
		return m.ListApplicationsFunc(ctx, page, perPage, isActive)
	}
	return nil, 0, nil
}

func (m *mockApplicationStore) GetBySecretHash(ctx context.Context, hash string) (*models.Application, error) {
	if m.GetBySecretHashFunc != nil {
		return m.GetBySecretHashFunc(ctx, hash)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	if m.GetBrandingFunc != nil {
		return m.GetBrandingFunc(ctx, applicationID)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error {
	if m.CreateOrUpdateBrandingFunc != nil {
		return m.CreateOrUpdateBrandingFunc(ctx, branding)
	}
	return nil
}

func (m *mockApplicationStore) CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	if m.CreateUserProfileFunc != nil {
		return m.CreateUserProfileFunc(ctx, profile)
	}
	return nil
}

func (m *mockApplicationStore) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID, applicationID)
	}
	return nil, errors.New("not found")
}

func (m *mockApplicationStore) UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(ctx, profile)
	}
	return nil
}

func (m *mockApplicationStore) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.DeleteUserProfileFunc != nil {
		return m.DeleteUserProfileFunc(ctx, userID, applicationID)
	}
	return nil
}

func (m *mockApplicationStore) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	if m.ListUserProfilesFunc != nil {
		return m.ListUserProfilesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockApplicationStore) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
	if m.ListApplicationUsersFunc != nil {
		return m.ListApplicationUsersFunc(ctx, applicationID, page, perPage)
	}
	return nil, 0, nil
}

func (m *mockApplicationStore) UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.UpdateLastAccessFunc != nil {
		return m.UpdateLastAccessFunc(ctx, userID, applicationID)
	}
	return nil
}

func (m *mockApplicationStore) BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	if m.BanUserFromApplicationFunc != nil {
		return m.BanUserFromApplicationFunc(ctx, userID, applicationID, bannedBy, reason)
	}
	return nil
}

func (m *mockApplicationStore) UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.UnbanUserFromApplicationFunc != nil {
		return m.UnbanUserFromApplicationFunc(ctx, userID, applicationID)
	}
	return nil
}

// --- Mock AppOAuthProviderStore ---

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

// --- Setup helper ---

func testLogger() *logger.Logger {
	return logger.New("test", logger.DebugLevel, false)
}

func setupApplicationService(mockAppRepo *mockApplicationStore) *ApplicationService {
	mockAppOAuthRepo := &mockAppOAuthProviderStore{}
	return NewApplicationService(mockAppRepo, mockAppOAuthRepo, testLogger())
}

// ============================================================
// CheckUserAccess tests (preserved from original)
// ============================================================

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

// ============================================================
// Commit 7: ApplicationService CRUD Tests
// ============================================================

func TestCreateApplication_ShouldReturnApp_WhenValidRequest(t *testing.T) {
	// Arrange
	appID := uuid.New()
	ownerID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
		CreateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			assert.Equal(t, "my-app", app.Name)
			assert.Equal(t, "My App", app.DisplayName)
			assert.True(t, app.IsActive)
			assert.False(t, app.IsSystem)
			assert.Equal(t, &ownerID, app.OwnerID)
			return nil
		},
		CreateOrUpdateBrandingFunc: func(ctx context.Context, branding *models.ApplicationBranding) error {
			assert.Equal(t, "#3B82F6", branding.PrimaryColor)
			return nil
		},
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "my-app",
				IsActive: true,
			}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        "my-app",
		DisplayName: "My App",
	}

	// Act
	app, secret, err := service.CreateApplication(ctx, req, &ownerID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "my-app", app.Name)
	assert.Equal(t, "My App", app.DisplayName)
	assert.True(t, app.IsActive)
	assert.NotNil(t, app.Branding)
	assert.NotEmpty(t, secret)
}

func TestCreateApplication_ShouldDefaultAuthMethods_WhenNoneProvided(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
		CreateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			assert.Equal(t, []string{"password"}, app.AllowedAuthMethods)
			return nil
		},
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{ID: id, Name: "test-app", IsActive: true}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        "test-app",
		DisplayName: "Test App",
	}

	// Act
	app, _, err := service.CreateApplication(ctx, req, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestCreateApplication_ShouldRespectIsActive_WhenExplicitlySetFalse(t *testing.T) {
	// Arrange
	isActive := false
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
		CreateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			assert.False(t, app.IsActive)
			return nil
		},
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{ID: id, Name: "inactive-app", IsActive: false}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        "inactive-app",
		DisplayName: "Inactive App",
		IsActive:    &isActive,
	}

	// Act
	_, _, err := service.CreateApplication(ctx, req, nil)

	// Assert
	require.NoError(t, err)
}

func TestCreateApplication_ShouldReturnError_WhenDuplicateName(t *testing.T) {
	// Arrange
	existingApp := &models.Application{
		ID:   uuid.New(),
		Name: "existing-app",
	}
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return existingApp, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        "existing-app",
		DisplayName: "Existing App",
	}

	// Act
	app, secret, err := service.CreateApplication(ctx, req, nil)

	// Assert
	assert.ErrorIs(t, err, ErrApplicationNameExists)
	assert.Nil(t, app)
	assert.Empty(t, secret)
}

func TestCreateApplication_ShouldReturnError_WhenInvalidName(t *testing.T) {
	service := setupApplicationService(&mockApplicationStore{})
	ctx := context.Background()

	testCases := []struct {
		name string
		slug string
	}{
		{"too short", "ab"},
		{"spaces", "my app"},
		{"special characters", "my_app!"},
		{"starts with hyphen", "-my-app"},
		{"ends with hyphen", "my-app-"},
		{"consecutive hyphens", "my--app"},
		{"underscores", "my_app"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &models.CreateApplicationRequest{
				Name:        tc.slug,
				DisplayName: "Test",
			}

			app, secret, err := service.CreateApplication(ctx, req, nil)

			assert.ErrorIs(t, err, ErrInvalidApplicationName)
			assert.Nil(t, app)
			assert.Empty(t, secret)
		})
	}
}

func TestCreateApplication_ShouldReturnError_WhenRepoCreateFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
		CreateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return errors.New("database connection lost")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.CreateApplicationRequest{
		Name:        "fail-app",
		DisplayName: "Fail App",
	}

	// Act
	app, secret, err := service.CreateApplication(ctx, req, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create application")
	assert.Nil(t, app)
	assert.Empty(t, secret)
}

func TestCreateApplication_ShouldNormalizeName_WhenUppercaseProvided(t *testing.T) {
	// Arrange
	var capturedName string
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
		CreateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			capturedName = app.Name
			return nil
		},
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{ID: id, Name: capturedName, IsActive: true}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// The slug "MY-APP" gets lowercased to "my-app" which is valid
	req := &models.CreateApplicationRequest{
		Name:        "MY-APP",
		DisplayName: "My App",
	}

	// Act
	app, _, err := service.CreateApplication(ctx, req, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "my-app", capturedName)
}

func TestGetByID_ShouldReturnApp_WhenExists(t *testing.T) {
	// Arrange
	appID := uuid.New()
	expectedApp := &models.Application{
		ID:          appID,
		Name:        "test-app",
		DisplayName: "Test App",
		IsActive:    true,
	}
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			assert.Equal(t, appID, id)
			return expectedApp, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	app, err := service.GetByID(ctx, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, appID, app.ID)
	assert.Equal(t, "test-app", app.Name)
}

func TestGetByID_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	app, err := service.GetByID(ctx, uuid.New())

	// Assert
	assert.ErrorIs(t, err, ErrApplicationNotFound)
	assert.Nil(t, app)
}

func TestGetByName_ShouldReturnApp_WhenExists(t *testing.T) {
	// Arrange
	expectedApp := &models.Application{
		ID:          uuid.New(),
		Name:        "test-app",
		DisplayName: "Test App",
	}
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			assert.Equal(t, "test-app", name)
			return expectedApp, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	app, err := service.GetByName(ctx, "test-app")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "test-app", app.Name)
}

func TestGetByName_ShouldNormalizeInput_WhenUppercaseProvided(t *testing.T) {
	// Arrange
	var queriedName string
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			queriedName = name
			return &models.Application{Name: name}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	_, err := service.GetByName(ctx, "  TEST-APP  ")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-app", queriedName)
}

func TestGetByName_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByNameFunc: func(ctx context.Context, name string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	app, err := service.GetByName(ctx, "nonexistent")

	// Assert
	assert.ErrorIs(t, err, ErrApplicationNotFound)
	assert.Nil(t, app)
}

func TestUpdateApplication_ShouldReturnUpdatedApp_WhenValid(t *testing.T) {
	// Arrange
	appID := uuid.New()
	existingApp := &models.Application{
		ID:          appID,
		Name:        "test-app",
		DisplayName: "Old Name",
		Description: "Old Desc",
		IsActive:    true,
	}
	updatedApp := &models.Application{
		ID:          appID,
		Name:        "test-app",
		DisplayName: "New Name",
		Description: "New Desc",
		IsActive:    true,
	}

	getCallCount := 0
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			getCallCount++
			if getCallCount == 1 {
				return existingApp, nil
			}
			return updatedApp, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			assert.Equal(t, "New Name", app.DisplayName)
			assert.Equal(t, "New Desc", app.Description)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationRequest{
		DisplayName: "New Name",
		Description: "New Desc",
	}

	// Act
	result, err := service.UpdateApplication(ctx, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "New Name", result.DisplayName)
	assert.Equal(t, "New Desc", result.Description)
}

func TestUpdateApplication_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationRequest{
		DisplayName: "New Name",
	}

	// Act
	result, err := service.UpdateApplication(ctx, uuid.New(), req)

	// Assert
	assert.ErrorIs(t, err, ErrApplicationNotFound)
	assert.Nil(t, result)
}

func TestUpdateApplication_ShouldReturnError_WhenRepoUpdateFails(t *testing.T) {
	// Arrange
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{ID: appID, Name: "test-app"}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationRequest{
		DisplayName: "New Name",
	}

	// Act
	result, err := service.UpdateApplication(ctx, appID, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update application")
	assert.Nil(t, result)
}

func TestUpdateApplication_ShouldOnlyUpdateProvidedFields(t *testing.T) {
	// Arrange
	appID := uuid.New()
	isActive := false
	existingApp := &models.Application{
		ID:                 appID,
		Name:               "test-app",
		DisplayName:        "Original",
		Description:        "Original Desc",
		HomepageURL:        "https://original.com",
		CallbackURLs:       []string{"https://original.com/callback"},
		IsActive:           true,
		AllowedAuthMethods: []string{"password"},
	}

	var capturedApp *models.Application
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			// Return a copy to avoid mutations affecting assertions
			appCopy := *existingApp
			return &appCopy, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			capturedApp = app
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Only update IsActive; everything else should remain the same
	req := &models.UpdateApplicationRequest{
		IsActive: &isActive,
	}

	// Act
	_, err := service.UpdateApplication(ctx, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, capturedApp)
	assert.Equal(t, "Original", capturedApp.DisplayName)
	assert.Equal(t, "Original Desc", capturedApp.Description)
	assert.Equal(t, "https://original.com", capturedApp.HomepageURL)
	assert.False(t, capturedApp.IsActive)
}

func TestDeleteApplication_ShouldSucceed_WhenAppExists(t *testing.T) {
	// Arrange
	appID := uuid.New()
	var capturedApp *models.Application
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "test-app",
				IsActive: true,
				IsSystem: false,
			}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			capturedApp = app
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteApplication(ctx, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, capturedApp)
	assert.False(t, capturedApp.IsActive, "delete should set IsActive to false")
}

func TestDeleteApplication_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteApplication(ctx, uuid.New())

	// Assert
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestDeleteApplication_ShouldReturnError_WhenSystemApp(t *testing.T) {
	// Arrange
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "system-app",
				IsSystem: true,
			}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteApplication(ctx, appID)

	// Assert
	assert.ErrorIs(t, err, ErrCannotDeleteSystemApp)
}

func TestDeleteApplication_ShouldReturnError_WhenRepoUpdateFails(t *testing.T) {
	// Arrange
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "test-app",
				IsSystem: false,
			}, nil
		},
		UpdateApplicationFunc: func(ctx context.Context, app *models.Application) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteApplication(ctx, appID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete application")
}

func TestListApplications_ShouldReturnPaginatedResults(t *testing.T) {
	// Arrange
	app1 := &models.Application{ID: uuid.New(), Name: "app-1"}
	app2 := &models.Application{ID: uuid.New(), Name: "app-2"}
	app3 := &models.Application{ID: uuid.New(), Name: "app-3"}

	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, perPage)
			return []*models.Application{app1, app2, app3}, 3, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 1, 20, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Applications, 3)
	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, 1, result.TotalPages)
}

func TestListApplications_ShouldNormalizePagination_WhenInvalidValues(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
			assert.Equal(t, 1, page, "page should default to 1 when < 1")
			assert.Equal(t, 20, perPage, "perPage should default to 20 when < 1")
			return []*models.Application{}, 0, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 0, 0, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
}

func TestListApplications_ShouldCapPerPage_WhenExceeds100(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
			assert.Equal(t, 20, perPage, "perPage should default to 20 when > 100")
			return []*models.Application{}, 0, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 1, 200, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestListApplications_ShouldCalculateTotalPages_Correctly(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
			return []*models.Application{{ID: uuid.New()}}, 45, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 1, 20, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalPages, "45 items / 20 per page = 3 pages (ceil)")
}

func TestListApplications_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
			return nil, 0, errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 1, 20, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list applications")
	assert.Nil(t, result)
}

func TestListApplications_ShouldFilterByActiveStatus(t *testing.T) {
	// Arrange
	isActive := true
	mockAppRepo := &mockApplicationStore{
		ListApplicationsFunc: func(ctx context.Context, page, perPage int, active *bool) ([]*models.Application, int, error) {
			require.NotNil(t, active)
			assert.True(t, *active)
			return []*models.Application{{ID: uuid.New(), IsActive: true}}, 1, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplications(ctx, 1, 20, &isActive)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Applications, 1)
}

// ============================================================
// Commit 8: Branding & User Profile Tests
// ============================================================

func TestGetBranding_ShouldReturnBranding_WhenExists(t *testing.T) {
	// Arrange
	appID := uuid.New()
	expectedBranding := &models.ApplicationBranding{
		ID:            uuid.New(),
		ApplicationID: appID,
		PrimaryColor:  "#FF0000",
		CompanyName:   "Test Corp",
	}
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			assert.Equal(t, appID, applicationID)
			return expectedBranding, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	branding, err := service.GetBranding(ctx, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, branding)
	assert.Equal(t, "#FF0000", branding.PrimaryColor)
	assert.Equal(t, "Test Corp", branding.CompanyName)
}

func TestGetBranding_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	branding, err := service.GetBranding(ctx, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get branding")
	assert.Nil(t, branding)
}

func TestUpdateBranding_ShouldUpdateExistingBranding_WhenBrandingExists(t *testing.T) {
	// Arrange
	appID := uuid.New()
	existingBranding := &models.ApplicationBranding{
		ID:              uuid.New(),
		ApplicationID:   appID,
		PrimaryColor:    "#3B82F6",
		SecondaryColor:  "#8B5CF6",
		BackgroundColor: "#FFFFFF",
	}
	updatedBranding := &models.ApplicationBranding{
		ID:              existingBranding.ID,
		ApplicationID:   appID,
		PrimaryColor:    "#FF0000",
		SecondaryColor:  "#8B5CF6",
		BackgroundColor: "#FFFFFF",
		CompanyName:     "Acme",
	}

	getCallCount := 0
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			getCallCount++
			if getCallCount == 1 {
				return existingBranding, nil
			}
			return updatedBranding, nil
		},
		CreateOrUpdateBrandingFunc: func(ctx context.Context, branding *models.ApplicationBranding) error {
			assert.Equal(t, "#FF0000", branding.PrimaryColor)
			assert.Equal(t, "Acme", branding.CompanyName)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationBrandingRequest{
		PrimaryColor: "#FF0000",
		CompanyName:  "Acme",
	}

	// Act
	result, err := service.UpdateBranding(ctx, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "#FF0000", result.PrimaryColor)
	assert.Equal(t, "Acme", result.CompanyName)
}

func TestUpdateBranding_ShouldCreateDefaultBranding_WhenNotExists(t *testing.T) {
	// Arrange
	appID := uuid.New()
	getCallCount := 0
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			getCallCount++
			if getCallCount == 1 {
				return nil, errors.New("not found")
			}
			return &models.ApplicationBranding{
				ApplicationID:   appID,
				PrimaryColor:    "#FF0000",
				SecondaryColor:  "#8B5CF6",
				BackgroundColor: "#FFFFFF",
				CompanyName:     "New Corp",
			}, nil
		},
		CreateOrUpdateBrandingFunc: func(ctx context.Context, branding *models.ApplicationBranding) error {
			// Should have defaults for non-provided fields
			assert.Equal(t, "#FF0000", branding.PrimaryColor)
			assert.Equal(t, "#8B5CF6", branding.SecondaryColor)
			assert.Equal(t, "#FFFFFF", branding.BackgroundColor)
			assert.Equal(t, "New Corp", branding.CompanyName)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationBrandingRequest{
		PrimaryColor: "#FF0000",
		CompanyName:  "New Corp",
	}

	// Act
	result, err := service.UpdateBranding(ctx, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestUpdateBranding_ShouldReturnError_WhenSaveFails(t *testing.T) {
	// Arrange
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			return &models.ApplicationBranding{
				ID:            uuid.New(),
				ApplicationID: appID,
				PrimaryColor:  "#3B82F6",
			}, nil
		},
		CreateOrUpdateBrandingFunc: func(ctx context.Context, branding *models.ApplicationBranding) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationBrandingRequest{
		PrimaryColor: "#FF0000",
	}

	// Act
	result, err := service.UpdateBranding(ctx, appID, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update branding")
	assert.Nil(t, result)
}

func TestUpdateBranding_ShouldUpdateAllFields_WhenAllProvided(t *testing.T) {
	// Arrange
	appID := uuid.New()
	var capturedBranding *models.ApplicationBranding
	mockAppRepo := &mockApplicationStore{
		GetBrandingFunc: func(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
			return &models.ApplicationBranding{
				ID:              uuid.New(),
				ApplicationID:   appID,
				PrimaryColor:    "#000000",
				SecondaryColor:  "#000000",
				BackgroundColor: "#000000",
			}, nil
		},
		CreateOrUpdateBrandingFunc: func(ctx context.Context, branding *models.ApplicationBranding) error {
			capturedBranding = branding
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateApplicationBrandingRequest{
		LogoURL:         "https://example.com/logo.png",
		FaviconURL:      "https://example.com/favicon.ico",
		PrimaryColor:    "#FF0000",
		SecondaryColor:  "#00FF00",
		BackgroundColor: "#0000FF",
		CustomCSS:       ".test { color: red; }",
		CompanyName:     "Acme Corp",
		SupportEmail:    "support@acme.com",
		TermsURL:        "https://acme.com/terms",
		PrivacyURL:      "https://acme.com/privacy",
	}

	// Act
	_, err := service.UpdateBranding(ctx, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, capturedBranding)
	assert.Equal(t, "https://example.com/logo.png", capturedBranding.LogoURL)
	assert.Equal(t, "https://example.com/favicon.ico", capturedBranding.FaviconURL)
	assert.Equal(t, "#FF0000", capturedBranding.PrimaryColor)
	assert.Equal(t, "#00FF00", capturedBranding.SecondaryColor)
	assert.Equal(t, "#0000FF", capturedBranding.BackgroundColor)
	assert.Equal(t, ".test { color: red; }", capturedBranding.CustomCSS)
	assert.Equal(t, "Acme Corp", capturedBranding.CompanyName)
	assert.Equal(t, "support@acme.com", capturedBranding.SupportEmail)
	assert.Equal(t, "https://acme.com/terms", capturedBranding.TermsURL)
	assert.Equal(t, "https://acme.com/privacy", capturedBranding.PrivacyURL)
}

func TestGetOrCreateUserProfile_ShouldReturnExisting_WhenProfileExists(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	now := time.Now().Add(-24 * time.Hour)
	existingProfile := &models.UserApplicationProfile{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: appID,
		IsActive:      true,
		IsBanned:      false,
		LastAccessAt:  &now,
	}
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return existingProfile, nil
		},
		UpdateLastAccessFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, appID, aid)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	profile, err := service.GetOrCreateUserProfile(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, appID, profile.ApplicationID)
	// LastAccessAt should be updated to approximately now
	require.NotNil(t, profile.LastAccessAt)
	assert.WithinDuration(t, time.Now(), *profile.LastAccessAt, 2*time.Second)
}

func TestGetOrCreateUserProfile_ShouldCreateNew_WhenNotExists(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	var capturedProfile *models.UserApplicationProfile
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
		CreateUserProfileFunc: func(ctx context.Context, profile *models.UserApplicationProfile) error {
			capturedProfile = profile
			assert.Equal(t, userID, profile.UserID)
			assert.Equal(t, appID, profile.ApplicationID)
			assert.True(t, profile.IsActive)
			assert.False(t, profile.IsBanned)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	profile, err := service.GetOrCreateUserProfile(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.NotNil(t, capturedProfile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, appID, profile.ApplicationID)
	assert.True(t, profile.IsActive)
	assert.False(t, profile.IsBanned)
	require.NotNil(t, profile.LastAccessAt)
}

func TestGetOrCreateUserProfile_ShouldReturnError_WhenCreateFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
		CreateUserProfileFunc: func(ctx context.Context, profile *models.UserApplicationProfile) error {
			return errors.New("unique constraint violation")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	profile, err := service.GetOrCreateUserProfile(ctx, uuid.New(), uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user profile")
	assert.Nil(t, profile)
}

func TestGetOrCreateUserProfile_ShouldNotFail_WhenUpdateLastAccessFails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	existingProfile := &models.UserApplicationProfile{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: appID,
		IsActive:      true,
	}
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return existingProfile, nil
		},
		UpdateLastAccessFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			return errors.New("redis timeout")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act - should succeed despite UpdateLastAccess failure
	profile, err := service.GetOrCreateUserProfile(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, profile)
}

func TestGetUserProfile_ShouldReturnProfile_WhenExists(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	expectedProfile := &models.UserApplicationProfile{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: appID,
		IsActive:      true,
	}
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return expectedProfile, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	profile, err := service.GetUserProfile(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, appID, profile.ApplicationID)
}

func TestGetUserProfile_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	profile, err := service.GetUserProfile(ctx, uuid.New(), uuid.New())

	// Assert
	assert.ErrorIs(t, err, ErrUserProfileNotFound)
	assert.Nil(t, profile)
}

func TestUpdateUserProfile_ShouldUpdateProfile_WhenValid(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	displayName := "New Display Name"
	nickname := "newbie"
	existingProfile := &models.UserApplicationProfile{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: appID,
		IsActive:      true,
	}
	updatedProfile := &models.UserApplicationProfile{
		ID:            existingProfile.ID,
		UserID:        userID,
		ApplicationID: appID,
		DisplayName:   &displayName,
		Nickname:      &nickname,
		IsActive:      true,
	}

	getCallCount := 0
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			getCallCount++
			if getCallCount == 1 {
				return existingProfile, nil
			}
			return updatedProfile, nil
		},
		UpdateUserProfileFunc: func(ctx context.Context, profile *models.UserApplicationProfile) error {
			assert.Equal(t, &displayName, profile.DisplayName)
			assert.Equal(t, &nickname, profile.Nickname)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateUserAppProfileRequest{
		DisplayName: &displayName,
		Nickname:    &nickname,
	}

	// Act
	result, err := service.UpdateUserProfile(ctx, userID, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, &displayName, result.DisplayName)
}

func TestUpdateUserProfile_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	displayName := "Test"
	req := &models.UpdateUserAppProfileRequest{
		DisplayName: &displayName,
	}

	// Act
	result, err := service.UpdateUserProfile(ctx, uuid.New(), uuid.New(), req)

	// Assert
	assert.ErrorIs(t, err, ErrUserProfileNotFound)
	assert.Nil(t, result)
}

func TestUpdateUserProfile_ShouldReturnError_WhenRepoUpdateFails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: appID,
			}, nil
		},
		UpdateUserProfileFunc: func(ctx context.Context, profile *models.UserApplicationProfile) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	displayName := "Test"
	req := &models.UpdateUserAppProfileRequest{
		DisplayName: &displayName,
	}

	// Act
	result, err := service.UpdateUserProfile(ctx, userID, appID, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user profile")
	assert.Nil(t, result)
}

func TestUpdateUserProfile_ShouldUpdateAllFields_WhenAllProvided(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	displayName := "John Doe"
	avatarURL := "https://example.com/avatar.jpg"
	nickname := "johnd"
	metadata := []byte(`{"level":10}`)
	appRoles := []string{"admin", "editor"}
	isActive := false
	isBanned := true
	banReason := "Violation of terms"

	var capturedProfile *models.UserApplicationProfile
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: appID,
				IsActive:      true,
			}, nil
		},
		UpdateUserProfileFunc: func(ctx context.Context, profile *models.UserApplicationProfile) error {
			capturedProfile = profile
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	req := &models.UpdateUserAppProfileRequest{
		DisplayName: &displayName,
		AvatarURL:   &avatarURL,
		Nickname:    &nickname,
		Metadata:    metadata,
		AppRoles:    appRoles,
		IsActive:    &isActive,
		IsBanned:    &isBanned,
		BanReason:   &banReason,
	}

	// Act
	_, err := service.UpdateUserProfile(ctx, userID, appID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, capturedProfile)
	assert.Equal(t, &displayName, capturedProfile.DisplayName)
	assert.Equal(t, &avatarURL, capturedProfile.AvatarURL)
	assert.Equal(t, &nickname, capturedProfile.Nickname)
	assert.Equal(t, metadata, capturedProfile.Metadata)
	assert.Equal(t, appRoles, capturedProfile.AppRoles)
	assert.False(t, capturedProfile.IsActive)
	assert.True(t, capturedProfile.IsBanned)
	assert.Equal(t, &banReason, capturedProfile.BanReason)
}

func TestListUserProfiles_ShouldReturnProfiles_WhenExists(t *testing.T) {
	// Arrange
	userID := uuid.New()
	profiles := []*models.UserApplicationProfile{
		{ID: uuid.New(), UserID: userID, ApplicationID: uuid.New()},
		{ID: uuid.New(), UserID: userID, ApplicationID: uuid.New()},
	}
	mockAppRepo := &mockApplicationStore{
		ListUserProfilesFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.UserApplicationProfile, error) {
			assert.Equal(t, userID, uid)
			return profiles, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListUserProfiles(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListUserProfiles_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListUserProfilesFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.UserApplicationProfile, error) {
			return nil, errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListUserProfiles(ctx, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list user profiles")
	assert.Nil(t, result)
}

func TestListApplicationUsers_ShouldReturnPaginatedResults(t *testing.T) {
	// Arrange
	appID := uuid.New()
	profiles := []*models.UserApplicationProfile{
		{ID: uuid.New(), UserID: uuid.New(), ApplicationID: appID},
		{ID: uuid.New(), UserID: uuid.New(), ApplicationID: appID},
	}
	mockAppRepo := &mockApplicationStore{
		ListApplicationUsersFunc: func(ctx context.Context, aid uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
			assert.Equal(t, appID, aid)
			return profiles, 50, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplicationUsers(ctx, appID, 1, 20)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Profiles, 2)
	assert.Equal(t, 50, result.Total)
	assert.Equal(t, 3, result.TotalPages)
}

func TestListApplicationUsers_ShouldNormalizePagination(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		ListApplicationUsersFunc: func(ctx context.Context, aid uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, perPage)
			return []*models.UserApplicationProfile{}, 0, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	result, err := service.ListApplicationUsers(ctx, uuid.New(), -5, 500)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
}

func TestBanUser_ShouldSucceed_WhenUserNotBanned(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	bannedBy := uuid.New()
	reason := "spam"

	var capturedReason string
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: appID,
				IsBanned:      false,
			}, nil
		},
		BanUserFromApplicationFunc: func(ctx context.Context, uid, aid, by uuid.UUID, r string) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, appID, aid)
			assert.Equal(t, bannedBy, by)
			capturedReason = r
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.BanUser(ctx, userID, appID, bannedBy, reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "spam", capturedReason)
}

func TestBanUser_ShouldReturnNil_WhenAlreadyBanned(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:       uuid.New(),
				IsBanned: true,
			}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.BanUser(ctx, uuid.New(), uuid.New(), uuid.New(), "reason")

	// Assert
	assert.NoError(t, err)
}

func TestBanUser_ShouldReturnError_WhenProfileNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.BanUser(ctx, uuid.New(), uuid.New(), uuid.New(), "reason")

	// Assert
	assert.ErrorIs(t, err, ErrUserProfileNotFound)
}

func TestBanUser_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:       uuid.New(),
				IsBanned: false,
			}, nil
		},
		BanUserFromApplicationFunc: func(ctx context.Context, uid, aid, by uuid.UUID, r string) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.BanUser(ctx, uuid.New(), uuid.New(), uuid.New(), "reason")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ban user")
}

func TestUnbanUser_ShouldSucceed_WhenUserIsBanned(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:            uuid.New(),
				UserID:        userID,
				ApplicationID: appID,
				IsBanned:      true,
			}, nil
		},
		UnbanUserFromApplicationFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, appID, aid)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.UnbanUser(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
}

func TestUnbanUser_ShouldReturnNil_WhenNotBanned(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:       uuid.New(),
				IsBanned: false,
			}, nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.UnbanUser(ctx, uuid.New(), uuid.New())

	// Assert
	assert.NoError(t, err)
}

func TestUnbanUser_ShouldReturnError_WhenProfileNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("not found")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.UnbanUser(ctx, uuid.New(), uuid.New())

	// Assert
	assert.ErrorIs(t, err, ErrUserProfileNotFound)
}

func TestUnbanUser_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		GetUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				ID:       uuid.New(),
				IsBanned: true,
			}, nil
		},
		UnbanUserFromApplicationFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			return errors.New("database error")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.UnbanUser(ctx, uuid.New(), uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unban user")
}

func TestDeleteUserProfile_ShouldDelegateToRepo(t *testing.T) {
	// Arrange
	userID := uuid.New()
	appID := uuid.New()
	mockAppRepo := &mockApplicationStore{
		DeleteUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, appID, aid)
			return nil
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteUserProfile(ctx, userID, appID)

	// Assert
	require.NoError(t, err)
}

func TestDeleteUserProfile_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockAppRepo := &mockApplicationStore{
		DeleteUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) error {
			return errors.New("foreign key constraint")
		},
	}
	service := setupApplicationService(mockAppRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteUserProfile(ctx, uuid.New(), uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "foreign key constraint")
}
