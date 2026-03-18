package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oauthMockAppOAuthProviderStore is a configurable mock for AppOAuthProviderStore,
// used exclusively in OAuth service tests. The simpler mockAppOAuthProviderStore
// in application_service_test.go always returns nil and is not configurable.
type oauthMockAppOAuthProviderStore struct {
	GetByAppAndProviderFunc func(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error)
}

func (m *oauthMockAppOAuthProviderStore) Create(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}
func (m *oauthMockAppOAuthProviderStore) GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *oauthMockAppOAuthProviderStore) GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error) {
	if m.GetByAppAndProviderFunc != nil {
		return m.GetByAppAndProviderFunc(ctx, appID, provider)
	}
	return nil, nil
}
func (m *oauthMockAppOAuthProviderStore) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *oauthMockAppOAuthProviderStore) ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *oauthMockAppOAuthProviderStore) Update(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}
func (m *oauthMockAppOAuthProviderStore) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// oauthMockApplicationStore is a configurable mock for ApplicationStore in OAuth tests.
type oauthMockApplicationStore struct {
	GetApplicationByIDFunc func(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetUserProfileFunc     func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
}

func (m *oauthMockApplicationStore) CreateApplication(ctx context.Context, app *models.Application) error {
	return nil
}
func (m *oauthMockApplicationStore) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	if m.GetApplicationByIDFunc != nil {
		return m.GetApplicationByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *oauthMockApplicationStore) GetApplicationByName(ctx context.Context, name string) (*models.Application, error) {
	return nil, nil
}
func (m *oauthMockApplicationStore) UpdateApplication(ctx context.Context, app *models.Application) error {
	return nil
}
func (m *oauthMockApplicationStore) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *oauthMockApplicationStore) ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
	return nil, 0, nil
}
func (m *oauthMockApplicationStore) GetBySecretHash(ctx context.Context, hash string) (*models.Application, error) {
	return nil, nil
}
func (m *oauthMockApplicationStore) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	return nil, nil
}
func (m *oauthMockApplicationStore) CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error {
	return nil
}
func (m *oauthMockApplicationStore) CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}
func (m *oauthMockApplicationStore) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID, applicationID)
	}
	return nil, errors.New("not found")
}
func (m *oauthMockApplicationStore) UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}
func (m *oauthMockApplicationStore) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}
func (m *oauthMockApplicationStore) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *oauthMockApplicationStore) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
	return nil, 0, nil
}
func (m *oauthMockApplicationStore) UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}
func (m *oauthMockApplicationStore) BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	return nil
}
func (m *oauthMockApplicationStore) UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}

// setupOAuthService creates an OAuthService with configurable mocks for testing.
// It manually configures providers (bypassing env vars) to enable deterministic testing.
func setupOAuthService() (
	*OAuthService,
	*mockUserStore,
	*mockOAuthStore,
	*mockTokenStore,
	*mockAuditStore,
	*mockRBACStore,
	*mockTokenService,
	*mockHTTPClient,
) {
	mUser := &mockUserStore{}
	mOAuth := &mockOAuthStore{}
	mToken := &mockTokenStore{}
	mAudit := &mockAuditStore{}
	mRBAC := &mockRBACStore{}
	mJWT := &mockTokenService{}
	mHTTP := &mockHTTPClient{}
	mAppOAuth := &oauthMockAppOAuthProviderStore{}
	mApp := &oauthMockApplicationStore{}

	svc := NewOAuthService(
		mUser, mOAuth, mToken, mAudit, mRBAC,
		mJWT, nil, mHTTP, mAppOAuth, mApp,
		true, // jitProvisioning enabled
		nil,  // loginAlertService
	)

	// Manually configure a test provider to bypass env vars
	svc.providers[models.ProviderGoogle] = &OAuthProviderConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CallbackURL:  "http://localhost/callback",
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "profile", "email"},
	}

	return svc, mUser, mOAuth, mToken, mAudit, mRBAC, mJWT, mHTTP
}

// newJSONResponse creates an http.Response with JSON body for mock HTTP client.
func newJSONResponse(statusCode int, body interface{}) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// --- GetAuthURL Tests ---

func TestOAuthService_GetAuthURL(t *testing.T) {
	t.Run("ShouldReturnURL_WhenValidGoogleProvider", func(t *testing.T) {
		// Arrange
		svc, _, _, _, _, _, _, _ := setupOAuthService()
		ctx := context.Background()

		// Act
		authURL, err := svc.GetAuthURL(ctx, models.ProviderGoogle, "test-state", nil)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, authURL, "https://accounts.google.com/o/oauth2/v2/auth")
		assert.Contains(t, authURL, "client_id=test-client-id")
		assert.Contains(t, authURL, "state=test-state")
		assert.Contains(t, authURL, "redirect_uri=")
		assert.Contains(t, authURL, "response_type=code")
		// Google-specific params
		assert.Contains(t, authURL, "access_type=offline")
		assert.Contains(t, authURL, "prompt=consent")
	})

	t.Run("ShouldReturnError_WhenInvalidProvider", func(t *testing.T) {
		// Arrange
		svc, _, _, _, _, _, _, _ := setupOAuthService()
		ctx := context.Background()

		// Act
		authURL, err := svc.GetAuthURL(ctx, models.OAuthProvider("invalid"), "test-state", nil)

		// Assert
		assert.ErrorIs(t, err, models.ErrInvalidProvider)
		assert.Empty(t, authURL)
	})

	t.Run("ShouldReturnTelegramURL_WhenTelegramProvider", func(t *testing.T) {
		// Arrange
		svc, _, _, _, _, _, _, _ := setupOAuthService()
		ctx := context.Background()

		// Configure Telegram provider
		svc.providers[models.ProviderTelegram] = &OAuthProviderConfig{
			ClientID:    "bot-token-123",
			CallbackURL: "https://example.com/telegram/callback",
		}

		// Act
		authURL, err := svc.GetAuthURL(ctx, models.ProviderTelegram, "test-state", nil)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, authURL, "bot_id=bot-token-123")
		// Telegram does not use traditional OAuth params
		assert.NotContains(t, authURL, "response_type")
	})
}

// --- ExchangeCode Tests ---

func TestOAuthService_ExchangeCode_ShouldReturnTokens_WhenSuccess(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.String(), "oauth2.googleapis.com/token")
		return newJSONResponse(http.StatusOK, OAuthTokenResponse{
			AccessToken:  "oauth-access-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "oauth-refresh-token",
		}), nil
	}

	// Act
	result, err := svc.ExchangeCode(ctx, models.ProviderGoogle, "auth-code-123", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "oauth-access-token", result.AccessToken)
	assert.Equal(t, "oauth-refresh-token", result.RefreshToken)
	assert.Equal(t, 3600, result.ExpiresIn)
}

func TestOAuthService_ExchangeCode_ShouldReturnError_WhenHTTPFails(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}

	// Act
	result, err := svc.ExchangeCode(ctx, models.ProviderGoogle, "auth-code-123", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to exchange code")
}

func TestOAuthService_ExchangeCode_ShouldReturnError_WhenProviderReturnsNon200(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		return newJSONResponse(http.StatusBadRequest, map[string]string{"error": "invalid_grant"}), nil
	}

	// Act
	result, err := svc.ExchangeCode(ctx, models.ProviderGoogle, "expired-code", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "token exchange failed with status: 400")
}

func TestOAuthService_ExchangeCode_ShouldReturnError_WhenInvalidProvider(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	ctx := context.Background()

	// Act
	result, err := svc.ExchangeCode(ctx, models.OAuthProvider("nonexistent"), "code", nil)

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidProvider)
	assert.Nil(t, result)
}

// --- HandleCallback Tests ---

func TestOAuthService_HandleCallback_ShouldCreateNewUser_WhenNoExistingOAuthAccount(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, mToken, _, mRBAC, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	// Step 1: ExchangeCode - mock HTTP for token exchange
	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken:  "oauth-access",
				RefreshToken: "oauth-refresh",
				ExpiresIn:    3600,
			}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":      "google-uid-123",
			"email":   "newuser@gmail.com",
			"name":    "New Google User",
			"picture": "https://example.com/photo.jpg",
		}), nil
	}

	// Step 2: GetOAuthAccount - return nil (no existing account)
	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		assert.Equal(t, "google", provider)
		assert.Equal(t, "google-uid-123", providerUserID)
		return nil, nil
	}

	// Step 3: createUserFromOAuth - mock user creation flow
	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
		return false, nil
	}
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		assert.Equal(t, "user", name)
		return &models.Role{ID: roleID, Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		user.ID = userID
		assert.Equal(t, "newuser@gmail.com", user.Email)
		assert.True(t, user.IsActive)
		assert.True(t, user.EmailVerified)
		assert.Empty(t, user.PasswordHash) // OAuth users have no password
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, uID, rID, assignedBy uuid.UUID) error {
		return nil
	}

	// Step 4: CreateOAuthAccount
	mOAuth.CreateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		assert.Equal(t, "google", account.Provider)
		assert.Equal(t, "google-uid-123", account.ProviderUserID)
		assert.Equal(t, "oauth-access", account.AccessToken)
		assert.Equal(t, "oauth-refresh", account.RefreshToken)
		assert.NotNil(t, account.TokenExpiresAt)
		return nil
	}

	// Step 5: GetByID (after callback to get user for JWT)
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{
			ID:       userID,
			Email:    "newuser@gmail.com",
			Username: "new_google_user",
			IsActive: true,
		}, nil
	}

	// Step 6: JWT token generation
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "jwt-access-token", nil
	}
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "jwt-refresh-token", nil
	}
	mJWT.GetRefreshTokenExpirationFunc = func() time.Duration {
		return 7 * 24 * time.Hour
	}
	mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error {
		assert.Equal(t, userID, token.UserID)
		return nil
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "auth-code", "1.2.3.4", "Mozilla/5.0", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "jwt-access-token", result.AccessToken)
	assert.Equal(t, "jwt-refresh-token", result.RefreshToken)
	assert.True(t, result.IsNewUser)
	assert.NotNil(t, result.User)
}

func TestOAuthService_HandleCallback_ShouldLinkExistingUser_WhenOAuthAccountExists(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, mToken, _, _, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()
	oauthAccountID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken:  "new-oauth-access",
				RefreshToken: "new-oauth-refresh",
				ExpiresIn:    3600,
			}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-456",
			"email": "existing@gmail.com",
			"name":  "Existing User",
		}), nil
	}

	// Existing OAuth account found
	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID:             oauthAccountID,
			UserID:         userID,
			Provider:       "google",
			ProviderUserID: "google-uid-456",
			AccessToken:    "old-access",
			RefreshToken:   "old-refresh",
		}, nil
	}

	// Update OAuth account with new tokens
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		assert.Equal(t, oauthAccountID, account.ID)
		assert.Equal(t, "new-oauth-access", account.AccessToken)
		assert.Equal(t, "new-oauth-refresh", account.RefreshToken)
		assert.NotNil(t, account.TokenExpiresAt)
		return nil
	}

	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		assert.Equal(t, userID, id)
		return &models.User{
			ID:       userID,
			Email:    "existing@gmail.com",
			Username: "existing_user",
			IsActive: true,
		}, nil
	}

	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "jwt-access", nil
	}
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "jwt-refresh", nil
	}
	mJWT.GetRefreshTokenExpirationFunc = func() time.Duration {
		return 7 * 24 * time.Hour
	}
	mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error {
		return nil
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "auth-code", "1.2.3.4", "Mozilla/5.0", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "jwt-access", result.AccessToken)
	assert.Equal(t, "jwt-refresh", result.RefreshToken)
	assert.False(t, result.IsNewUser)
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenCodeExchangeFails(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "bad-code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to exchange code")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenUserInfoFails(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken: "token",
			}), nil
		}
		return nil, errors.New("user info endpoint down")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user info")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenUserInfoReturnsNon200(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken: "token",
			}), nil
		}
		return newJSONResponse(http.StatusUnauthorized, map[string]string{"error": "invalid_token"}), nil
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user info request failed with status: 401")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenJITProvisioningDisabled(t *testing.T) {
	// Arrange
	svc, _, mOAuth, _, _, _, _, mHTTP := setupOAuthService()
	svc.jitProvisioning = false // Disable JIT provisioning
	ctx := context.Background()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-789",
			"email": "unknown@gmail.com",
			"name":  "Unknown User",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return nil, nil
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*models.AppError)
	require.True(t, ok)
	assert.Equal(t, 403, appErr.Code)
	assert.Contains(t, appErr.Message, "Automatic user creation is disabled")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenCreateOAuthAccountFails(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, _, _, mRBAC, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token", ExpiresIn: 3600}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-create-err",
			"email": "createerr@gmail.com",
			"name":  "Create Error",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return nil, nil
	}
	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error { return nil }
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }

	mOAuth.CreateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		return errors.New("database constraint violation")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database constraint violation")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenUpdateOAuthAccountFails(t *testing.T) {
	// Arrange
	svc, _, mOAuth, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-update-err",
			"email": "updateerr@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			Provider:       "google",
			ProviderUserID: "google-uid-update-err",
		}, nil
	}

	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		return errors.New("db write error")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db write error")
}

func TestOAuthService_HandleCallback_ShouldHandleExpiresInZero_WhenNoExpiration(t *testing.T) {
	// Arrange: OAuth response with ExpiresIn = 0 should not set TokenExpiresAt
	svc, mUser, mOAuth, mToken, _, mRBAC, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken: "token",
				ExpiresIn:   0, // No expiration
			}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-noexp",
			"email": "noexp@gmail.com",
			"name":  "No Expiry",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return nil, nil
	}
	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		user.ID = userID
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, uID, rID, assignedBy uuid.UUID) error { return nil }

	mOAuth.CreateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		assert.Nil(t, account.TokenExpiresAt, "TokenExpiresAt should be nil when ExpiresIn is 0")
		return nil
	}
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{ID: userID, Email: "noexp@gmail.com"}, nil
	}
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "at", nil }
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "rt", nil }
	mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
	mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsNewUser)
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenGetUserByIDFails(t *testing.T) {
	// Arrange: Existing account found and updated but GetByID fails after
	svc, mUser, mOAuth, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-getfail",
			"email": "getfail@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID:             uuid.New(),
			UserID:         userID,
			Provider:       "google",
			ProviderUserID: "google-uid-getfail",
		}, nil
	}
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error { return nil }

	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return nil, errors.New("user not found in database")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found in database")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenAccessTokenGenerationFails(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, _, _, _, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-jwtfail",
			"email": "jwtfail@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-uid-jwtfail",
		}, nil
	}
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error { return nil }
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{ID: userID, Email: "jwtfail@gmail.com"}, nil
	}
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "", errors.New("signing key error")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "signing key error")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenRefreshTokenGenerationFails(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, _, _, _, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-rtfail",
			"email": "rtfail@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-uid-rtfail",
		}, nil
	}
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error { return nil }
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{ID: userID, Email: "rtfail@gmail.com"}, nil
	}
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "at", nil
	}
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) {
		return "", errors.New("refresh token generation error")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "refresh token generation error")
}

func TestOAuthService_HandleCallback_ShouldReturnError_WhenSaveRefreshTokenFails(t *testing.T) {
	// Arrange
	svc, mUser, mOAuth, mToken, _, _, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{AccessToken: "token"}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-savefail",
			"email": "savefail@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-uid-savefail",
		}, nil
	}
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error { return nil }
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{ID: userID, Email: "savefail@gmail.com"}, nil
	}
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "at", nil }
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "rt", nil }
	mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
	mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error {
		return errors.New("failed to save refresh token")
	}

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save refresh token")
}

func TestOAuthService_HandleCallback_ShouldKeepExistingRefreshToken_WhenNewTokenEmpty(t *testing.T) {
	// Arrange: When provider returns empty refresh token, keep old one
	svc, mUser, mOAuth, mToken, _, _, mJWT, mHTTP := setupOAuthService()
	ctx := context.Background()
	userID := uuid.New()

	callCount := 0
	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return newJSONResponse(http.StatusOK, OAuthTokenResponse{
				AccessToken:  "new-access",
				RefreshToken: "", // Empty - should keep old
				ExpiresIn:    3600,
			}), nil
		}
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "google-uid-keeprt",
			"email": "keeprt@gmail.com",
		}), nil
	}

	mOAuth.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
		return &models.OAuthAccount{
			ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-uid-keeprt",
			RefreshToken: "old-refresh-to-keep",
		}, nil
	}
	mOAuth.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
		assert.Equal(t, "new-access", account.AccessToken)
		assert.Equal(t, "old-refresh-to-keep", account.RefreshToken, "should keep old refresh token when new is empty")
		return nil
	}
	mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	mJWT.GenerateAccessTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "at", nil }
	mJWT.GenerateRefreshTokenFunc = func(user *models.User, appID ...*uuid.UUID) (string, error) { return "rt", nil }
	mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
	mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }

	// Act
	result, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "1.2.3.4", "ua", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsNewUser)
}

// --- parseUserInfo Tests ---

func TestOAuthService_ParseUserInfo_ShouldParseGoogleResponse(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":      "g-123",
		"email":   "user@gmail.com",
		"name":    "Google User",
		"picture": "https://photo.url",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderGoogle, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "g-123", info.ProviderUserID)
	assert.Equal(t, "user@gmail.com", info.Email)
	assert.Equal(t, "Google User", info.Name)
	assert.Equal(t, "https://photo.url", info.ProfilePicture)
	assert.Equal(t, "google", info.Provider)
}

func TestOAuthService_ParseUserInfo_ShouldParseYandexResponse(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":            "y-456",
		"default_email": "user@yandex.ru",
		"real_name":     "Yandex User",
		"login":         "yandexlogin",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderYandex, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "y-456", info.ProviderUserID)
	assert.Equal(t, "user@yandex.ru", info.Email)
	assert.Equal(t, "Yandex User", info.Name)
	assert.Equal(t, "yandexlogin", info.Username)
}

func TestOAuthService_ParseUserInfo_ShouldParseGitHubResponse(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":         float64(12345), // GitHub returns numeric id
		"email":      "dev@github.com",
		"name":       "GitHub Dev",
		"login":      "ghdev",
		"avatar_url": "https://github.com/avatar.jpg",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderGitHub, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "12345", info.ProviderUserID) // Converted via fmt.Sprintf
	assert.Equal(t, "dev@github.com", info.Email)
	assert.Equal(t, "GitHub Dev", info.Name)
	assert.Equal(t, "ghdev", info.Username)
	assert.Equal(t, "https://github.com/avatar.jpg", info.ProfilePicture)
}

func TestOAuthService_ParseUserInfo_ShouldParseInstagramResponse(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":       "ig-789",
		"username": "insta_user",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderInstagram, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "ig-789", info.ProviderUserID)
	assert.Equal(t, "insta_user", info.Username)
	assert.Equal(t, "insta_user", info.Name) // Instagram uses username as name
	assert.Empty(t, info.Email)              // Instagram doesn't provide email
}

func TestOAuthService_ParseUserInfo_ShouldParseTelegramResponse(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":         float64(999),
		"first_name": "Tele",
		"last_name":  "Gram",
		"username":   "telegram_user",
		"photo_url":  "https://t.me/photo.jpg",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderTelegram, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "999", info.ProviderUserID) // Converted via fmt.Sprintf
	assert.Equal(t, "Tele Gram", info.Name)     // first_name + last_name
	assert.Equal(t, "telegram_user", info.Username)
	assert.Equal(t, "https://t.me/photo.jpg", info.ProfilePicture)
}

func TestOAuthService_ParseUserInfo_ShouldParseTelegramResponse_WhenNoLastName(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":         float64(111),
		"first_name": "OnlyFirst",
		"username":   "firstonly",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderTelegram, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "OnlyFirst", info.Name) // No last name appended
}

func TestOAuthService_ParseUserInfo_ShouldParseOneCResponse_WithSubField(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"sub":                "1c-sub-001",
		"email":              "user@1c.local",
		"name":               "OneC User",
		"preferred_username": "onec_user",
		"picture":            "https://1c.local/photo.png",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderOneC, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "1c-sub-001", info.ProviderUserID)
	assert.Equal(t, "user@1c.local", info.Email)
	assert.Equal(t, "OneC User", info.Name)
	assert.Equal(t, "onec_user", info.Username)
	assert.Equal(t, "https://1c.local/photo.png", info.ProfilePicture)
}

func TestOAuthService_ParseUserInfo_ShouldParseOneCResponse_WithFallbackIDFields(t *testing.T) {
	// Arrange: When "sub" is missing, fall back to "id", then "user_id"
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"user_id": "1c-user-id-999",
		"email":   "fallback@1c.local",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderOneC, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "1c-user-id-999", info.ProviderUserID)
}

func TestOAuthService_ParseUserInfo_ShouldParseOneCResponse_WithIDFallback(t *testing.T) {
	// Arrange: "sub" is empty but "id" is present
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"id":    "1c-id-fallback",
		"email": "idfallback@1c.local",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderOneC, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "1c-id-fallback", info.ProviderUserID)
}

func TestOAuthService_ParseUserInfo_ShouldParseOneCResponse_WithComposedName(t *testing.T) {
	// Arrange: When "name" is missing, compose from given_name + family_name
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"sub":         "1c-composed",
		"given_name":  "Ivan",
		"family_name": "Petrov",
		"username":    "ipetrov",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderOneC, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Ivan Petrov", info.Name)
	assert.Equal(t, "ipetrov", info.Username)
}

func TestOAuthService_ParseUserInfo_ShouldParseOneCResponse_WithOnlyFamilyName(t *testing.T) {
	// Arrange: Only family_name present, no given_name
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{
		"sub":         "1c-family-only",
		"family_name": "Petrov",
	}

	// Act
	info, err := svc.parseUserInfo(models.ProviderOneC, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Petrov", info.Name)
}

func TestOAuthService_ParseUserInfo_ShouldHandleMissingFields(t *testing.T) {
	// Arrange: Empty data for Google should return empty strings (not panic)
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	data := map[string]interface{}{}

	// Act
	info, err := svc.parseUserInfo(models.ProviderGoogle, data)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, info.ProviderUserID)
	assert.Empty(t, info.Email)
	assert.Empty(t, info.Name)
	assert.Equal(t, "google", info.Provider)
}

// --- createUserFromOAuth Tests ---

func TestOAuthService_CreateUserFromOAuth_ShouldGeneratePlaceholderEmail_WhenNoEmail(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		assert.Equal(t, "instagram_ig-12345678@oauth.local", user.Email)
		assert.False(t, user.EmailVerified) // No real email provided
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "ig-12345678",
		Provider:       "instagram",
		Username:       "insta_user",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestOAuthService_CreateUserFromOAuth_ShouldGenerateUsername_WhenNoUsernameOrName(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		// Username should be generated from provider + first 8 chars of providerUserID
		assert.Contains(t, user.Username, "google")
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "google-long-uid-123456",
		Email:          "test@gmail.com",
		Provider:       "google",
		// No Username or Name
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestOAuthService_CreateUserFromOAuth_ShouldResolveUsernameConflicts(t *testing.T) {
	// Arrange: First username check returns true (exists), second returns false
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	checkCount := 0
	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
		checkCount++
		if checkCount == 1 {
			return true, nil // First check: username exists
		}
		return false, nil // Second check: with suffix, available
	}
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		// Username should have a numeric suffix appended
		assert.Contains(t, user.Username, "1")
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "github-uid",
		Email:          "dev@github.com",
		Name:           "Dev User",
		Username:       "devuser",
		Provider:       "github",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, 2, checkCount)
}

func TestOAuthService_CreateUserFromOAuth_ShouldReturnError_WhenRoleNotFound(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return nil, errors.New("role not found")
	}

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "uid-123",
		Email:          "test@test.com",
		Provider:       "google",
		Username:       "testuser",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get default role")
}

func TestOAuthService_CreateUserFromOAuth_ShouldReturnError_WhenUserCreateFails(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		return errors.New("unique constraint violation")
	}

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "uid-456",
		Email:          "dup@test.com",
		Provider:       "google",
		Username:       "dupuser",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestOAuthService_CreateUserFromOAuth_ShouldReturnError_WhenUsernameCheckFails(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, _, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
		return false, errors.New("database error")
	}

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "uid-789",
		Email:          "dbfail@test.com",
		Provider:       "google",
		Username:       "dbfailuser",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestOAuthService_CreateUserFromOAuth_ShouldReturnError_WhenRoleAssignmentFails(t *testing.T) {
	// Arrange
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error { return nil }
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
		return errors.New("role assignment error")
	}

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "uid-role-fail",
		Email:          "rolefail@test.com",
		Provider:       "google",
		Username:       "rolefailuser",
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to assign default role")
}

func TestOAuthService_CreateUserFromOAuth_ShouldUseName_WhenUsernameEmpty(t *testing.T) {
	// Arrange: When Username is empty, should use Name as username
	svc, mUser, _, _, _, mRBAC, _, _ := setupOAuthService()
	ctx := context.Background()

	mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
	mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	var createdUsername string
	mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
		createdUsername = user.Username
		return nil
	}
	mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }

	userInfo := &models.OAuthUserInfo{
		ProviderUserID: "uid-name-as-username",
		Email:          "name@test.com",
		Name:           "John Doe",
		Provider:       "google",
		// No Username
	}

	// Act
	user, err := svc.createUserFromOAuth(ctx, userInfo)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEmpty(t, createdUsername)
}

// --- GetUserInfo Tests ---

func TestOAuthService_GetUserInfo_ShouldReturnParsedInfo_WhenSuccess(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Bearer test-access-token", req.Header.Get("Authorization"))
		return newJSONResponse(http.StatusOK, map[string]interface{}{
			"id":    "g-001",
			"email": "info@gmail.com",
			"name":  "Info User",
		}), nil
	}

	// Act
	info, err := svc.GetUserInfo(ctx, models.ProviderGoogle, "test-access-token", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "g-001", info.ProviderUserID)
	assert.Equal(t, "info@gmail.com", info.Email)
	assert.Equal(t, "Info User", info.Name)
}

func TestOAuthService_GetUserInfo_ShouldReturnError_WhenHTTPFails(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, mHTTP := setupOAuthService()
	ctx := context.Background()

	mHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	}

	// Act
	info, err := svc.GetUserInfo(ctx, models.ProviderGoogle, "token", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "failed to get user info")
}

func TestOAuthService_GetUserInfo_ShouldReturnError_WhenInvalidProvider(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()
	ctx := context.Background()

	// Act
	info, err := svc.GetUserInfo(ctx, models.OAuthProvider("unknown"), "token", nil)

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidProvider)
	assert.Nil(t, info)
}

// --- Helper function tests ---

func TestOAuthService_GenerateState_ShouldReturnNonEmptyString(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()

	// Act
	state, err := svc.GenerateState()

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.True(t, len(state) > 20, "state should be long enough for security")
}

func TestOAuthService_GenerateState_ShouldReturnUniqueValues(t *testing.T) {
	// Arrange
	svc, _, _, _, _, _, _, _ := setupOAuthService()

	// Act
	state1, err1 := svc.GenerateState()
	state2, err2 := svc.GenerateState()

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, state1, state2, "consecutive states should be unique")
}

func TestGetString_ShouldReturnValue_WhenKeyExists(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	assert.Equal(t, "value", getString(data, "key"))
}

func TestGetString_ShouldReturnEmpty_WhenKeyMissing(t *testing.T) {
	data := map[string]interface{}{}
	assert.Equal(t, "", getString(data, "missing"))
}

func TestGetString_ShouldReturnEmpty_WhenValueNotString(t *testing.T) {
	data := map[string]interface{}{
		"key": 12345,
	}
	assert.Equal(t, "", getString(data, "key"))
}

func TestJoinScopes_ShouldJoinWithSpaces(t *testing.T) {
	assert.Equal(t, "openid profile email", joinScopes([]string{"openid", "profile", "email"}))
}

func TestJoinScopes_ShouldHandleEmpty(t *testing.T) {
	assert.Equal(t, "", joinScopes([]string{}))
}

func TestJoinScopes_ShouldHandleSingle(t *testing.T) {
	assert.Equal(t, "openid", joinScopes([]string{"openid"}))
}

func TestSplitScopes_ShouldSplitBySpace(t *testing.T) {
	assert.Equal(t, []string{"openid", "profile", "email"}, splitScopes("openid profile email"))
}

func TestSplitScopes_ShouldHandleMultipleSpaces(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, splitScopes("a  b"))
}

func TestSplitScopes_ShouldHandleEmpty(t *testing.T) {
	result := splitScopes("")
	assert.Nil(t, result)
}

func TestSplitScopes_ShouldHandleSingle(t *testing.T) {
	assert.Equal(t, []string{"openid"}, splitScopes("openid"))
}
