package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOAuthService_GetAuthURL(t *testing.T) {
	// Set valid config for Google
	t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("GOOGLE_CALLBACK_URL", "http://localhost/callback")

	svc := NewOAuthService(nil, nil, nil, nil, nil, nil, nil, nil, false)

	t.Run("Success_Google", func(t *testing.T) {
		url, err := svc.GetAuthURL(models.ProviderGoogle, "state-123")
		assert.NoError(t, err)
		assert.Contains(t, url, "accounts.google.com")
		assert.Contains(t, url, "client_id=")
		assert.Contains(t, url, "state=state-123")
	})

	t.Run("InvalidProvider", func(t *testing.T) {
		url, err := svc.GetAuthURL("invalid", "state")
		assert.ErrorIs(t, err, models.ErrInvalidProvider)
		assert.Empty(t, url)
	})
}

func TestOAuthService_ExchangeCode_And_GetUserInfo(t *testing.T) {
	mockHTTP := &mockHTTPClient{}
	svc := NewOAuthService(nil, nil, nil, nil, nil, nil, nil, mockHTTP, false)

	ctx := context.Background()

	t.Run("ExchangeCode_Success", func(t *testing.T) {
		mockHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.String(), "googleapis.com/token")

			respBody := `{"access_token": "access-token", "token_type": "Bearer", "expires_in": 3600}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
			}, nil
		}

		token, err := svc.ExchangeCode(ctx, models.ProviderGoogle, "auth-code")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "access-token", token.AccessToken)
	})

	t.Run("GetUserInfo_Success", func(t *testing.T) {
		mockHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.String(), "googleapis.com/oauth2/v2/userinfo")
			assert.Equal(t, "Bearer access-token", req.Header.Get("Authorization"))

			respBody := `{"id": "123", "email": "test@example.com", "name": "Test User"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
			}, nil
		}

		info, err := svc.GetUserInfo(ctx, models.ProviderGoogle, "access-token")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "123", info.ProviderUserID)
		assert.Equal(t, "test@example.com", info.Email)
	})
}

func TestOAuthService_HandleCallback(t *testing.T) {
	// Initialize mocks
	mockUserRepo := &mockUserStore{}
	mockOAuthRepo := &mockOAuthStore{}
	mockTokenRepo := &mockTokenStore{}
	mockRBACRepo := &mockRBACStore{}
	mockJWT := &mockJWTService{}
	mockHTTP := &mockHTTPClient{}
	// mockAudit := &mockAuditStore{} // Not used in current logic apparently

	svc := NewOAuthService(mockUserRepo, mockOAuthRepo, mockTokenRepo, nil, mockRBACRepo, mockJWT, nil, mockHTTP, true)
	ctx := context.Background()

	// Mock HTTP responses for ExchangeCode and GetUserInfo
	setupHTTPMocks := func() {
		mockHTTP.DoFunc = func(req *http.Request) (*http.Response, error) {
			if req.Method == "POST" {
				// Exchange Code
				respBody := `{"access_token": "access-token", "token_type": "Bearer", "expires_in": 3600}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(respBody)),
				}, nil
			} else {
				// User Info
				respBody := `{"id": "provider-123", "email": "new@example.com", "name": "New User"}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(respBody)),
				}, nil
			}
		}
	}

	t.Run("NewUser_Success", func(t *testing.T) {
		setupHTTPMocks()

		// 1. GetOAuthAccount -> Not found
		mockOAuthRepo.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
			return nil, nil // Not found
		}

		// 2. CreateUserFromOAuth checks
		mockUserRepo.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) {
			return false, nil
		}

		mockRBACRepo.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}

		mockUserRepo.CreateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, "new@example.com", user.Email)
			return nil
		}

		mockRBACRepo.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
			return nil
		}

		// 3. CreateOAuthAccount
		mockOAuthRepo.CreateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
			assert.Equal(t, "provider-123", account.ProviderUserID)
			return nil
		}

		// 4. GetUser (called after creation/update to get fresh user for token gen)
		mockUserRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID, includeRoles *bool) (*models.User, error) {
			return &models.User{ID: id, Email: "new@example.com", Roles: []models.Role{{Name: "user"}}}, nil
		}

		// 5. Generate Tokens
		mockJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) {
			return "new-jwt-access", nil
		}
		mockJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) {
			return "new-jwt-refresh", nil
		}
		mockJWT.GetRefreshTokenExpirationFunc = func() time.Duration {
			return time.Hour * 24
		}

		// 6. Save Refresh Token
		mockTokenRepo.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error {
			return nil
		}

		resp, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "127.0.0.1", "Mozilla/5.0")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.IsNewUser)
		assert.Equal(t, "new-jwt-access", resp.AccessToken)
	})

	t.Run("ExistingUser_Success", func(t *testing.T) {
		setupHTTPMocks()

		existingUserID := uuid.New()

		// 1. GetOAuthAccount -> Found
		mockOAuthRepo.GetOAuthAccountFunc = func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
			return &models.OAuthAccount{
				UserID:         existingUserID,
				Provider:       string(provider),
				ProviderUserID: providerUserID,
			}, nil
		}

		// 2. UpdateOAuthAccount
		mockOAuthRepo.UpdateOAuthAccountFunc = func(ctx context.Context, account *models.OAuthAccount) error {
			return nil
		}

		// 3. GetUser
		mockUserRepo.GetByIDFunc = func(ctx context.Context, id uuid.UUID, includeRoles *bool) (*models.User, error) {
			return &models.User{ID: id, Email: "new@example.com"}, nil
		}

		// 4. Generate Tokens
		mockJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) {
			return "jwt-access", nil
		}
		mockJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) {
			return "jwt-refresh", nil
		}
		mockJWT.GetRefreshTokenExpirationFunc = func() time.Duration {
			return time.Hour
		}

		// 5. Save Refresh Token
		mockTokenRepo.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error {
			return nil
		}

		resp, err := svc.HandleCallback(ctx, models.ProviderGoogle, "code", "127.0.0.1", "Mozilla/5.0")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsNewUser)
	})
}
