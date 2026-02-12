package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/keys"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// mockOAuthProviderStore implements OAuthProviderStore interface for testing
type mockOAuthProviderStore struct {
	// Client operations
	CreateClientFunc        func(ctx context.Context, client *models.OAuthClient) error
	GetClientByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error)
	GetClientByClientIDFunc func(ctx context.Context, clientID string) (*models.OAuthClient, error)
	UpdateClientFunc        func(ctx context.Context, client *models.OAuthClient) error
	DeleteClientFunc        func(ctx context.Context, id uuid.UUID) error
	HardDeleteClientFunc    func(ctx context.Context, id uuid.UUID) error
	ListClientsFunc         func(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error)
	ListActiveClientsFunc   func(ctx context.Context) ([]*models.OAuthClient, error)

	// Authorization code operations
	CreateAuthorizationCodeFunc         func(ctx context.Context, code *models.AuthorizationCode) error
	GetAuthorizationCodeFunc            func(ctx context.Context, codeHash string) (*models.AuthorizationCode, error)
	MarkAuthorizationCodeUsedFunc       func(ctx context.Context, id uuid.UUID) error
	DeleteExpiredAuthorizationCodesFunc func(ctx context.Context) (int64, error)

	// Access token operations
	CreateAccessTokenFunc           func(ctx context.Context, token *models.OAuthAccessToken) error
	GetAccessTokenFunc              func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error)
	GetAccessTokenByIDFunc          func(ctx context.Context, id uuid.UUID) (*models.OAuthAccessToken, error)
	RevokeAccessTokenFunc           func(ctx context.Context, tokenHash string) error
	RevokeAllUserAccessTokensFunc   func(ctx context.Context, userID, clientID uuid.UUID) error
	RevokeAllClientAccessTokensFunc func(ctx context.Context, clientID uuid.UUID) error
	DeleteExpiredAccessTokensFunc   func(ctx context.Context) (int64, error)

	// Refresh token operations
	CreateRefreshTokenFunc           func(ctx context.Context, token *models.OAuthRefreshToken) error
	GetRefreshTokenFunc              func(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error)
	RevokeRefreshTokenFunc           func(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokensFunc   func(ctx context.Context, userID, clientID uuid.UUID) error
	RevokeAllClientRefreshTokensFunc func(ctx context.Context, clientID uuid.UUID) error
	DeleteExpiredRefreshTokensFunc   func(ctx context.Context) (int64, error)

	// User consent operations
	CreateOrUpdateConsentFunc func(ctx context.Context, consent *models.UserConsent) error
	GetUserConsentFunc        func(ctx context.Context, userID, clientID uuid.UUID) (*models.UserConsent, error)
	RevokeConsentFunc         func(ctx context.Context, userID, clientID uuid.UUID) error
	ListUserConsentsFunc      func(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error)
	ListClientConsentsFunc    func(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error)

	// Device code operations
	CreateDeviceCodeFunc         func(ctx context.Context, code *models.DeviceCode) error
	GetDeviceCodeFunc            func(ctx context.Context, deviceCodeHash string) (*models.DeviceCode, error)
	GetDeviceCodeByUserCodeFunc  func(ctx context.Context, userCode string) (*models.DeviceCode, error)
	UpdateDeviceCodeStatusFunc   func(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, userID *uuid.UUID) error
	DeleteExpiredDeviceCodesFunc func(ctx context.Context) (int64, error)

	// Scope operations
	CreateScopeFunc      func(ctx context.Context, scope *models.OAuthScope) error
	GetScopeByNameFunc   func(ctx context.Context, name string) (*models.OAuthScope, error)
	ListScopesFunc       func(ctx context.Context) ([]*models.OAuthScope, error)
	ListSystemScopesFunc func(ctx context.Context) ([]*models.OAuthScope, error)
	DeleteScopeFunc      func(ctx context.Context, id uuid.UUID) error
}

// Implement all interface methods

func (m *mockOAuthProviderStore) CreateClient(ctx context.Context, client *models.OAuthClient) error {
	if m.CreateClientFunc != nil {
		return m.CreateClientFunc(ctx, client)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetClientByID(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
	if m.GetClientByIDFunc != nil {
		return m.GetClientByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	if m.GetClientByClientIDFunc != nil {
		return m.GetClientByClientIDFunc(ctx, clientID)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) UpdateClient(ctx context.Context, client *models.OAuthClient) error {
	if m.UpdateClientFunc != nil {
		return m.UpdateClientFunc(ctx, client)
	}
	return nil
}

func (m *mockOAuthProviderStore) DeleteClient(ctx context.Context, id uuid.UUID) error {
	if m.DeleteClientFunc != nil {
		return m.DeleteClientFunc(ctx, id)
	}
	return nil
}

func (m *mockOAuthProviderStore) HardDeleteClient(ctx context.Context, id uuid.UUID) error {
	if m.HardDeleteClientFunc != nil {
		return m.HardDeleteClientFunc(ctx, id)
	}
	return nil
}

func (m *mockOAuthProviderStore) ListClients(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error) {
	if m.ListClientsFunc != nil {
		return m.ListClientsFunc(ctx, page, perPage, opts...)
	}
	return nil, 0, nil
}

func (m *mockOAuthProviderStore) ListActiveClients(ctx context.Context) ([]*models.OAuthClient, error) {
	if m.ListActiveClientsFunc != nil {
		return m.ListActiveClientsFunc(ctx)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) CreateAuthorizationCode(ctx context.Context, code *models.AuthorizationCode) error {
	if m.CreateAuthorizationCodeFunc != nil {
		return m.CreateAuthorizationCodeFunc(ctx, code)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetAuthorizationCode(ctx context.Context, codeHash string) (*models.AuthorizationCode, error) {
	if m.GetAuthorizationCodeFunc != nil {
		return m.GetAuthorizationCodeFunc(ctx, codeHash)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) MarkAuthorizationCodeUsed(ctx context.Context, id uuid.UUID) error {
	if m.MarkAuthorizationCodeUsedFunc != nil {
		return m.MarkAuthorizationCodeUsedFunc(ctx, id)
	}
	return nil
}

func (m *mockOAuthProviderStore) DeleteExpiredAuthorizationCodes(ctx context.Context) (int64, error) {
	if m.DeleteExpiredAuthorizationCodesFunc != nil {
		return m.DeleteExpiredAuthorizationCodesFunc(ctx)
	}
	return 0, nil
}

func (m *mockOAuthProviderStore) CreateAccessToken(ctx context.Context, token *models.OAuthAccessToken) error {
	if m.CreateAccessTokenFunc != nil {
		return m.CreateAccessTokenFunc(ctx, token)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetAccessToken(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) GetAccessTokenByID(ctx context.Context, id uuid.UUID) (*models.OAuthAccessToken, error) {
	if m.GetAccessTokenByIDFunc != nil {
		return m.GetAccessTokenByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) RevokeAccessToken(ctx context.Context, tokenHash string) error {
	if m.RevokeAccessTokenFunc != nil {
		return m.RevokeAccessTokenFunc(ctx, tokenHash)
	}
	return nil
}

func (m *mockOAuthProviderStore) RevokeAllUserAccessTokens(ctx context.Context, userID, clientID uuid.UUID) error {
	if m.RevokeAllUserAccessTokensFunc != nil {
		return m.RevokeAllUserAccessTokensFunc(ctx, userID, clientID)
	}
	return nil
}

func (m *mockOAuthProviderStore) RevokeAllClientAccessTokens(ctx context.Context, clientID uuid.UUID) error {
	if m.RevokeAllClientAccessTokensFunc != nil {
		return m.RevokeAllClientAccessTokensFunc(ctx, clientID)
	}
	return nil
}

func (m *mockOAuthProviderStore) DeleteExpiredAccessTokens(ctx context.Context) (int64, error) {
	if m.DeleteExpiredAccessTokensFunc != nil {
		return m.DeleteExpiredAccessTokensFunc(ctx)
	}
	return 0, nil
}

func (m *mockOAuthProviderStore) CreateRefreshToken(ctx context.Context, token *models.OAuthRefreshToken) error {
	if m.CreateRefreshTokenFunc != nil {
		return m.CreateRefreshTokenFunc(ctx, token)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetRefreshToken(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error) {
	if m.GetRefreshTokenFunc != nil {
		return m.GetRefreshTokenFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.RevokeRefreshTokenFunc != nil {
		return m.RevokeRefreshTokenFunc(ctx, tokenHash)
	}
	return nil
}

func (m *mockOAuthProviderStore) RevokeAllUserRefreshTokens(ctx context.Context, userID, clientID uuid.UUID) error {
	if m.RevokeAllUserRefreshTokensFunc != nil {
		return m.RevokeAllUserRefreshTokensFunc(ctx, userID, clientID)
	}
	return nil
}

func (m *mockOAuthProviderStore) RevokeAllClientRefreshTokens(ctx context.Context, clientID uuid.UUID) error {
	if m.RevokeAllClientRefreshTokensFunc != nil {
		return m.RevokeAllClientRefreshTokensFunc(ctx, clientID)
	}
	return nil
}

func (m *mockOAuthProviderStore) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	if m.DeleteExpiredRefreshTokensFunc != nil {
		return m.DeleteExpiredRefreshTokensFunc(ctx)
	}
	return 0, nil
}

func (m *mockOAuthProviderStore) CreateOrUpdateConsent(ctx context.Context, consent *models.UserConsent) error {
	if m.CreateOrUpdateConsentFunc != nil {
		return m.CreateOrUpdateConsentFunc(ctx, consent)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetUserConsent(ctx context.Context, userID, clientID uuid.UUID) (*models.UserConsent, error) {
	if m.GetUserConsentFunc != nil {
		return m.GetUserConsentFunc(ctx, userID, clientID)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	if m.RevokeConsentFunc != nil {
		return m.RevokeConsentFunc(ctx, userID, clientID)
	}
	return nil
}

func (m *mockOAuthProviderStore) ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error) {
	if m.ListUserConsentsFunc != nil {
		return m.ListUserConsentsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error) {
	if m.ListClientConsentsFunc != nil {
		return m.ListClientConsentsFunc(ctx, clientID)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) CreateDeviceCode(ctx context.Context, code *models.DeviceCode) error {
	if m.CreateDeviceCodeFunc != nil {
		return m.CreateDeviceCodeFunc(ctx, code)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetDeviceCode(ctx context.Context, deviceCodeHash string) (*models.DeviceCode, error) {
	if m.GetDeviceCodeFunc != nil {
		return m.GetDeviceCodeFunc(ctx, deviceCodeHash)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) GetDeviceCodeByUserCode(ctx context.Context, userCode string) (*models.DeviceCode, error) {
	if m.GetDeviceCodeByUserCodeFunc != nil {
		return m.GetDeviceCodeByUserCodeFunc(ctx, userCode)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) UpdateDeviceCodeStatus(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, userID *uuid.UUID) error {
	if m.UpdateDeviceCodeStatusFunc != nil {
		return m.UpdateDeviceCodeStatusFunc(ctx, id, status, userID)
	}
	return nil
}

func (m *mockOAuthProviderStore) DeleteExpiredDeviceCodes(ctx context.Context) (int64, error) {
	if m.DeleteExpiredDeviceCodesFunc != nil {
		return m.DeleteExpiredDeviceCodesFunc(ctx)
	}
	return 0, nil
}

func (m *mockOAuthProviderStore) CreateScope(ctx context.Context, scope *models.OAuthScope) error {
	if m.CreateScopeFunc != nil {
		return m.CreateScopeFunc(ctx, scope)
	}
	return nil
}

func (m *mockOAuthProviderStore) GetScopeByName(ctx context.Context, name string) (*models.OAuthScope, error) {
	if m.GetScopeByNameFunc != nil {
		return m.GetScopeByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) ListScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	if m.ListScopesFunc != nil {
		return m.ListScopesFunc(ctx)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) ListSystemScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	if m.ListSystemScopesFunc != nil {
		return m.ListSystemScopesFunc(ctx)
	}
	return nil, nil
}

func (m *mockOAuthProviderStore) DeleteScope(ctx context.Context, id uuid.UUID) error {
	if m.DeleteScopeFunc != nil {
		return m.DeleteScopeFunc(ctx, id)
	}
	return nil
}

// mockKeyManager implements a minimal key manager for testing JWKS functionality
type mockKeyManager struct {
	jwks *keys.JWKS
}

func (m *mockKeyManager) GetJWKS() *keys.JWKS {
	if m.jwks != nil {
		return m.jwks
	}
	return &keys.JWKS{
		Keys: []keys.JWK{
			{
				KTY: "RSA",
				Use: "sig",
				Alg: "RS256",
				KID: "test-key-1",
				N:   "test-n-value",
				E:   "AQAB",
			},
		},
	}
}

// mockOIDCService implements a mock for OIDCService
type mockOIDCService struct {
	GenerateOAuthAccessTokenFunc func(userID *uuid.UUID, clientID string, scope string, roles []string, ttl time.Duration) (string, error)
	GenerateIDTokenFunc          func(userID uuid.UUID, clientID, nonce string, scopes []string, user *models.User, ttl time.Duration) (string, error)
	ValidateOAuthAccessTokenFunc func(tokenString string) (*jwt.OAuthAccessTokenClaims, error)
}

func (m *mockOIDCService) GenerateOAuthAccessToken(userID *uuid.UUID, clientID string, scope string, roles []string, ttl time.Duration) (string, error) {
	if m.GenerateOAuthAccessTokenFunc != nil {
		return m.GenerateOAuthAccessTokenFunc(userID, clientID, scope, roles, ttl)
	}
	return "mock_access_token", nil
}

func (m *mockOIDCService) GenerateIDToken(userID uuid.UUID, clientID, nonce string, scopes []string, user *models.User, ttl time.Duration) (string, error) {
	if m.GenerateIDTokenFunc != nil {
		return m.GenerateIDTokenFunc(userID, clientID, nonce, scopes, user, ttl)
	}
	return "mock_id_token", nil
}

func (m *mockOIDCService) ValidateOAuthAccessToken(tokenString string) (*jwt.OAuthAccessTokenClaims, error) {
	if m.ValidateOAuthAccessTokenFunc != nil {
		return m.ValidateOAuthAccessTokenFunc(tokenString)
	}
	return nil, errors.New("not implemented")
}

// Helper function to create a test service
func setupOAuthProviderService() (*OAuthProviderService, *mockOAuthProviderStore, *mockUserStore, *mockAuditStore) {
	mRepo := &mockOAuthProviderStore{}
	mUserRepo := &mockUserStore{}
	mAuditRepo := &mockAuditStore{}
	log := logger.New("test", logger.DebugLevel, false)

	// Create service without real OIDC/KeyManager (we'll test specific methods that don't need them)
	svc := &OAuthProviderService{
		repo:      mRepo,
		userRepo:  mUserRepo,
		auditRepo: mAuditRepo,
		logger:    log,
		issuer:    "https://auth.example.com",
		baseURL:   "https://auth.example.com",
	}

	return svc, mRepo, mUserRepo, mAuditRepo
}

// Helper to create a valid test client
func createTestClient(clientType string) *models.OAuthClient {
	clientID := uuid.New()
	secretHash := func() *string {
		if clientType == string(models.ClientTypeConfidential) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("agws_test_secret"), 10)
			h := string(hash)
			return &h
		}
		return nil
	}()

	return &models.OAuthClient{
		ID:               clientID,
		ClientID:         "agw_test_client_123",
		ClientSecretHash: secretHash,
		Name:             "Test Client",
		Description:      "A test OAuth client",
		ClientType:       clientType,
		RedirectURIs:     []string{"https://example.com/callback"},
		AllowedGrantTypes: []string{
			string(models.GrantTypeAuthorizationCode),
			string(models.GrantTypeRefreshToken),
			string(models.GrantTypeClientCredentials),
		},
		AllowedScopes:   []string{"openid", "profile", "email"},
		DefaultScopes:   []string{"openid"},
		AccessTokenTTL:  900,
		RefreshTokenTTL: 604800,
		IDTokenTTL:      3600,
		RequirePKCE:     false,
		RequireConsent:  true,
		FirstParty:      false,
		IsActive:        true,
	}
}

// ============================================================================
// ValidateClientCredentials Tests
// ============================================================================

func TestValidateClientCredentials_ShouldReturnClient_WhenCredentialsValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	secret := "agws_test_secret"
	secretHash, _ := bcrypt.GenerateFromPassword([]byte(secret), 10)
	hashStr := string(secretHash)
	client.ClientSecretHash = &hashStr

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, client.ClientID, secret)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, client.ClientID, result.ClientID)
	assert.Equal(t, client.Name, result.Name)
}

func TestValidateClientCredentials_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, "nonexistent", "secret")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
	assert.Nil(t, result)
}

func TestValidateClientCredentials_ShouldReturnError_WhenClientInactive(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	client.IsActive = false

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, client.ClientID, "secret")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
	assert.Nil(t, result)
}

func TestValidateClientCredentials_ShouldReturnError_WhenSecretIsWrong(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	secretHash, _ := bcrypt.GenerateFromPassword([]byte("correct_secret"), 10)
	hashStr := string(secretHash)
	client.ClientSecretHash = &hashStr

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, client.ClientID, "wrong_secret")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
	assert.Nil(t, result)
}

func TestValidateClientCredentials_ShouldReturnError_WhenSecretHashIsNil(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	client.ClientSecretHash = nil // Corrupt state

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, client.ClientID, "any_secret")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
	assert.Nil(t, result)
}

func TestValidateClientCredentials_ShouldSucceed_ForPublicClient_WithoutSecret(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypePublic))
	client.ClientSecretHash = nil

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	result, err := svc.ValidateClientCredentials(ctx, client.ClientID, "")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, client.ClientID, result.ClientID)
}

// ============================================================================
// GetDiscoveryDocument Tests
// ============================================================================

func TestGetDiscoveryDocument_ShouldReturnValidDocument(t *testing.T) {
	// Arrange
	svc, _, _, _ := setupOAuthProviderService()

	// Act
	doc := svc.GetDiscoveryDocument()

	// Assert
	require.NotNil(t, doc)
	assert.Equal(t, "https://auth.example.com", doc.Issuer)
	assert.Equal(t, "https://auth.example.com/oauth2/authorize", doc.AuthorizationEndpoint)
	assert.Equal(t, "https://auth.example.com/oauth2/token", doc.TokenEndpoint)
	assert.Equal(t, "https://auth.example.com/oauth2/userinfo", doc.UserInfoEndpoint)
	assert.Equal(t, "https://auth.example.com/.well-known/jwks.json", doc.JwksURI)
	assert.Equal(t, "https://auth.example.com/oauth2/revoke", doc.RevocationEndpoint)
	assert.Equal(t, "https://auth.example.com/oauth2/introspect", doc.IntrospectionEndpoint)
	assert.Equal(t, "https://auth.example.com/oauth2/device/code", doc.DeviceAuthorizationEndpoint)
}

func TestGetDiscoveryDocument_ShouldContainSupportedScopes(t *testing.T) {
	// Arrange
	svc, _, _, _ := setupOAuthProviderService()

	// Act
	doc := svc.GetDiscoveryDocument()

	// Assert
	expectedScopes := []string{"openid", "profile", "email", "phone", "address", "offline_access"}
	assert.ElementsMatch(t, expectedScopes, doc.ScopesSupported)
}

func TestGetDiscoveryDocument_ShouldContainSupportedGrantTypes(t *testing.T) {
	// Arrange
	svc, _, _, _ := setupOAuthProviderService()

	// Act
	doc := svc.GetDiscoveryDocument()

	// Assert
	assert.Contains(t, doc.GrantTypesSupported, "authorization_code")
	assert.Contains(t, doc.GrantTypesSupported, "refresh_token")
	assert.Contains(t, doc.GrantTypesSupported, "client_credentials")
	assert.Contains(t, doc.GrantTypesSupported, "urn:ietf:params:oauth:grant-type:device_code")
}

func TestGetDiscoveryDocument_ShouldContainSupportedCodeChallengeMethods(t *testing.T) {
	// Arrange
	svc, _, _, _ := setupOAuthProviderService()

	// Act
	doc := svc.GetDiscoveryDocument()

	// Assert
	assert.Contains(t, doc.CodeChallengeMethodsSupported, "S256")
	assert.Contains(t, doc.CodeChallengeMethodsSupported, "plain")
}

func TestGetDiscoveryDocument_ShouldContainTokenEndpointAuthMethods(t *testing.T) {
	// Arrange
	svc, _, _, _ := setupOAuthProviderService()

	// Act
	doc := svc.GetDiscoveryDocument()

	// Assert
	assert.Contains(t, doc.TokenEndpointAuthMethodsSupported, "client_secret_basic")
	assert.Contains(t, doc.TokenEndpointAuthMethodsSupported, "client_secret_post")
	assert.Contains(t, doc.TokenEndpointAuthMethodsSupported, "none")
}

// ============================================================================
// IntrospectToken Tests
// ============================================================================

func TestIntrospectToken_ShouldReturnActive_WhenAccessTokenValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	token := "valid_access_token"
	userID := uuid.New()
	clientID := uuid.New()

	accessToken := &models.OAuthAccessToken{
		ID:        uuid.New(),
		ClientID:  clientID,
		UserID:    &userID,
		Scope:     "openid profile",
		IsActive:  true,
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		Client: &models.OAuthClient{
			ClientID: "test_client",
		},
		User: &models.User{
			ID:       userID,
			Username: "testuser",
		},
	}

	mRepo.GetAccessTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
		return accessToken, nil
	}

	// Act
	result, err := svc.IntrospectToken(ctx, token, "access_token", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Active)
	assert.Equal(t, "openid profile", result.Scope)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, "test_client", result.ClientID)
	assert.Equal(t, userID.String(), result.Subject)
	assert.Equal(t, "testuser", result.Username)
}

func TestIntrospectToken_ShouldReturnActive_WhenRefreshTokenValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	token := "valid_refresh_token"
	userID := uuid.New()
	clientID := uuid.New()

	refreshToken := &models.OAuthRefreshToken{
		ID:        uuid.New(),
		ClientID:  clientID,
		UserID:    userID,
		Scope:     "openid profile",
		IsActive:  true,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		Client: &models.OAuthClient{
			ClientID: "test_client",
		},
		User: &models.User{
			ID:       userID,
			Username: "testuser",
		},
	}

	mRepo.GetAccessTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
		return nil, errors.New("not found")
	}
	mRepo.GetRefreshTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error) {
		return refreshToken, nil
	}

	// Act
	result, err := svc.IntrospectToken(ctx, token, "refresh_token", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Active)
	assert.Equal(t, "refresh_token", result.TokenType)
	assert.Equal(t, userID.String(), result.Subject)
}

func TestIntrospectToken_ShouldReturnInactive_WhenTokenNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetAccessTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
		return nil, errors.New("not found")
	}
	mRepo.GetRefreshTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error) {
		return nil, errors.New("not found")
	}

	// Act
	result, err := svc.IntrospectToken(ctx, "unknown_token", "", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Active)
}

func TestIntrospectToken_ShouldReturnInactive_WhenAccessTokenExpired(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	accessToken := &models.OAuthAccessToken{
		ID:        uuid.New(),
		IsActive:  true,
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	mRepo.GetAccessTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
		return accessToken, nil
	}

	// Act
	result, err := svc.IntrospectToken(ctx, "expired_token", "access_token", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Active)
}

func TestIntrospectToken_ShouldReturnInactive_WhenAccessTokenRevoked(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	revokedAt := time.Now().Add(-time.Minute)
	accessToken := &models.OAuthAccessToken{
		ID:        uuid.New(),
		IsActive:  false,
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		RevokedAt: &revokedAt,
	}

	mRepo.GetAccessTokenFunc = func(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
		return accessToken, nil
	}

	// Act
	result, err := svc.IntrospectToken(ctx, "revoked_token", "access_token", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Active)
}

// ============================================================================
// RevokeToken Tests
// ============================================================================

func TestRevokeToken_ShouldRevokeAccessToken_WhenTokenTypeHintIsAccessToken(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	revokedCalled := false
	mRepo.RevokeAccessTokenFunc = func(ctx context.Context, tokenHash string) error {
		revokedCalled = true
		return nil
	}

	// Act
	err := svc.RevokeToken(ctx, "some_token", "access_token", nil)

	// Assert
	assert.NoError(t, err)
	assert.True(t, revokedCalled)
}

func TestRevokeToken_ShouldRevokeRefreshToken_WhenTokenTypeHintIsRefreshToken(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	revokedCalled := false
	mRepo.RevokeRefreshTokenFunc = func(ctx context.Context, tokenHash string) error {
		revokedCalled = true
		return nil
	}

	// Act
	err := svc.RevokeToken(ctx, "some_token", "refresh_token", nil)

	// Assert
	assert.NoError(t, err)
	assert.True(t, revokedCalled)
}

func TestRevokeToken_ShouldTryBothTypes_WhenNoHintProvided(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	accessRevokeCalled := false
	refreshRevokeCalled := false

	mRepo.RevokeAccessTokenFunc = func(ctx context.Context, tokenHash string) error {
		accessRevokeCalled = true
		return errors.New("not found") // Simulate not found
	}
	mRepo.RevokeRefreshTokenFunc = func(ctx context.Context, tokenHash string) error {
		refreshRevokeCalled = true
		return nil
	}

	// Act
	err := svc.RevokeToken(ctx, "some_token", "", nil)

	// Assert
	assert.NoError(t, err)
	assert.True(t, accessRevokeCalled)
	assert.True(t, refreshRevokeCalled)
}

func TestRevokeToken_ShouldNotFail_WhenTokenNotFound(t *testing.T) {
	// Arrange - RFC 7009 specifies revocation should always succeed
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.RevokeAccessTokenFunc = func(ctx context.Context, tokenHash string) error {
		return errors.New("not found")
	}
	mRepo.RevokeRefreshTokenFunc = func(ctx context.Context, tokenHash string) error {
		return errors.New("not found")
	}

	// Act
	err := svc.RevokeToken(ctx, "nonexistent_token", "", nil)

	// Assert
	assert.NoError(t, err) // Should not return error per RFC 7009
}

// ============================================================================
// ListClients Tests
// ============================================================================

func TestListClients_ShouldReturnClients_WhenClientsExist(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	expectedClients := []*models.OAuthClient{
		createTestClient(string(models.ClientTypeConfidential)),
		createTestClient(string(models.ClientTypePublic)),
	}

	mRepo.ListClientsFunc = func(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error) {
		return expectedClients, 2, nil
	}

	// Act
	clients, total, err := svc.ListClients(ctx, 1, 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, clients, 2)
	assert.Equal(t, 2, total)
}

func TestListClients_ShouldNormalizePagination_WhenInvalidValuesProvided(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.ListClientsFunc = func(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error) {
		assert.Equal(t, 1, page)     // Should be normalized to 1
		assert.Equal(t, 20, perPage) // Should be normalized to default 20
		return []*models.OAuthClient{}, 0, nil
	}

	// Act
	_, _, err := svc.ListClients(ctx, -5, 500) // Invalid values

	// Assert
	assert.NoError(t, err)
}

func TestListClients_ShouldFilterByOwner_WhenOwnerIDProvided(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	ownerID := uuid.New()
	mRepo.ListClientsFunc = func(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error) {
		o := BuildOAuthClientListOptions(opts)
		assert.NotNil(t, o.OwnerID)
		assert.Equal(t, ownerID, *o.OwnerID)
		return []*models.OAuthClient{}, 0, nil
	}

	// Act
	_, _, err := svc.ListClients(ctx, 1, 10, OAuthClientListOwner(ownerID))

	// Assert
	assert.NoError(t, err)
}

// ============================================================================
// GrantConsent Tests
// ============================================================================

func TestGrantConsent_ShouldCreateConsent_WhenClientValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, mAuditRepo := setupOAuthProviderService()
	ctx := context.Background()

	userID := uuid.New()
	clientID := "test_client"
	scopes := []string{"openid", "profile"}

	client := createTestClient(string(models.ClientTypeConfidential))
	client.ClientID = clientID

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, cID string) (*models.OAuthClient, error) {
		return client, nil
	}

	consentCreated := false
	mRepo.CreateOrUpdateConsentFunc = func(ctx context.Context, consent *models.UserConsent) error {
		consentCreated = true
		assert.Equal(t, userID, consent.UserID)
		assert.Equal(t, client.ID, consent.ClientID)
		assert.ElementsMatch(t, scopes, consent.Scopes)
		return nil
	}

	mAuditRepo.CreateFunc = func(ctx context.Context, log *models.AuditLog) error {
		return nil
	}

	// Act
	err := svc.GrantConsent(ctx, userID, clientID, scopes)

	// Assert
	assert.NoError(t, err)
	assert.True(t, consentCreated)
}

func TestGrantConsent_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, cID string) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	err := svc.GrantConsent(ctx, uuid.New(), "nonexistent", []string{"openid"})

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
}

// ============================================================================
// RevokeConsent Tests
// ============================================================================

func TestRevokeConsent_ShouldRevokeTokensAndConsent(t *testing.T) {
	// Arrange
	svc, mRepo, _, mAuditRepo := setupOAuthProviderService()
	ctx := context.Background()

	userID := uuid.New()
	clientID := uuid.New()

	accessTokensRevoked := false
	refreshTokensRevoked := false
	consentRevoked := false

	mRepo.RevokeAllUserAccessTokensFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		accessTokensRevoked = true
		return nil
	}
	mRepo.RevokeAllUserRefreshTokensFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		refreshTokensRevoked = true
		return nil
	}
	mRepo.RevokeConsentFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		consentRevoked = true
		return nil
	}
	mAuditRepo.CreateFunc = func(ctx context.Context, log *models.AuditLog) error {
		return nil
	}

	// Act
	err := svc.RevokeConsent(ctx, userID, clientID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, accessTokensRevoked)
	assert.True(t, refreshTokensRevoked)
	assert.True(t, consentRevoked)
}

func TestRevokeConsent_ShouldContinue_WhenTokenRevocationFails(t *testing.T) {
	// Arrange
	svc, mRepo, _, mAuditRepo := setupOAuthProviderService()
	ctx := context.Background()

	consentRevoked := false

	mRepo.RevokeAllUserAccessTokensFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		return errors.New("failed to revoke access tokens")
	}
	mRepo.RevokeAllUserRefreshTokensFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		return errors.New("failed to revoke refresh tokens")
	}
	mRepo.RevokeConsentFunc = func(ctx context.Context, uID, cID uuid.UUID) error {
		consentRevoked = true
		return nil
	}
	mAuditRepo.CreateFunc = func(ctx context.Context, log *models.AuditLog) error {
		return nil
	}

	// Act
	err := svc.RevokeConsent(ctx, uuid.New(), uuid.New())

	// Assert
	assert.NoError(t, err)
	assert.True(t, consentRevoked, "Consent should still be revoked even if token revocation fails")
}

// ============================================================================
// ListUserConsents Tests
// ============================================================================

func TestListUserConsents_ShouldReturnConsents_WhenConsentsExist(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	userID := uuid.New()
	expectedConsents := []*models.UserConsent{
		{
			ID:       uuid.New(),
			UserID:   userID,
			ClientID: uuid.New(),
			Scopes:   []string{"openid", "profile"},
		},
		{
			ID:       uuid.New(),
			UserID:   userID,
			ClientID: uuid.New(),
			Scopes:   []string{"openid", "email"},
		},
	}

	mRepo.ListUserConsentsFunc = func(ctx context.Context, uID uuid.UUID) ([]*models.UserConsent, error) {
		return expectedConsents, nil
	}

	// Act
	consents, err := svc.ListUserConsents(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, consents, 2)
}

// ============================================================================
// CreateScope Tests
// ============================================================================

func TestCreateScope_ShouldCreateScope_WhenValidInput(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	scope := &models.OAuthScope{
		ID:          uuid.New(),
		Name:        "custom:read",
		DisplayName: "Read Custom Data",
		Description: "Allows reading custom data",
	}

	scopeCreated := false
	mRepo.CreateScopeFunc = func(ctx context.Context, s *models.OAuthScope) error {
		scopeCreated = true
		return nil
	}

	// Act
	err := svc.CreateScope(ctx, scope)

	// Assert
	assert.NoError(t, err)
	assert.True(t, scopeCreated)
}

// ============================================================================
// DeleteScope Tests
// ============================================================================

func TestDeleteScope_ShouldDeleteScope_WhenScopeExists(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	scopeID := uuid.New()
	scopeDeleted := false

	mRepo.DeleteScopeFunc = func(ctx context.Context, id uuid.UUID) error {
		scopeDeleted = true
		assert.Equal(t, scopeID, id)
		return nil
	}

	// Act
	err := svc.DeleteScope(ctx, scopeID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, scopeDeleted)
}

// ============================================================================
// ListScopes Tests
// ============================================================================

func TestListScopes_ShouldReturnAllScopes(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	expectedScopes := []*models.OAuthScope{
		{ID: uuid.New(), Name: "openid", DisplayName: "OpenID"},
		{ID: uuid.New(), Name: "profile", DisplayName: "Profile"},
		{ID: uuid.New(), Name: "email", DisplayName: "Email"},
	}

	mRepo.ListScopesFunc = func(ctx context.Context) ([]*models.OAuthScope, error) {
		return expectedScopes, nil
	}

	// Act
	scopes, err := svc.ListScopes(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, scopes, 3)
}

// ============================================================================
// GetConsentInfo Tests
// ============================================================================

func TestGetConsentInfo_ShouldReturnConsentInfo_WhenClientValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	scopes := []string{"openid", "profile"}

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, cID string) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.GetScopeByNameFunc = func(ctx context.Context, name string) (*models.OAuthScope, error) {
		return &models.OAuthScope{
			Name:        name,
			DisplayName: name + " display",
			Description: name + " description",
		}, nil
	}

	// Act
	info, err := svc.GetConsentInfo(ctx, client.ClientID, scopes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, client, info.Client)
	assert.Len(t, info.RequestedScopes, 2)
}

func TestGetConsentInfo_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, cID string) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	info, err := svc.GetConsentInfo(ctx, "nonexistent", []string{"openid"})

	// Assert
	assert.ErrorIs(t, err, ErrInvalidClient)
	assert.Nil(t, info)
}

func TestGetConsentInfo_ShouldHandleUnknownScope_Gracefully(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, cID string) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.GetScopeByNameFunc = func(ctx context.Context, name string) (*models.OAuthScope, error) {
		return nil, errors.New("not found") // Unknown scope
	}

	// Act
	info, err := svc.GetConsentInfo(ctx, client.ClientID, []string{"unknown_scope"})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Len(t, info.RequestedScopes, 1)
	assert.Equal(t, "unknown_scope", info.RequestedScopes[0].Name)
	assert.Equal(t, "unknown_scope", info.RequestedScopes[0].DisplayName)
}

// ============================================================================
// ListClientConsents Tests
// ============================================================================

func TestListClientConsents_ShouldReturnConsents_WhenConsentsExist(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	clientID := uuid.New()
	expectedConsents := []*models.UserConsent{
		{ID: uuid.New(), UserID: uuid.New(), ClientID: clientID},
		{ID: uuid.New(), UserID: uuid.New(), ClientID: clientID},
	}

	mRepo.ListClientConsentsFunc = func(ctx context.Context, cID uuid.UUID) ([]*models.UserConsent, error) {
		return expectedConsents, nil
	}

	// Act
	consents, err := svc.ListClientConsents(ctx, clientID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, consents, 2)
}

// ============================================================================
// GetClient Tests
// ============================================================================

func TestGetClient_ShouldReturnClient_WhenClientExists(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	expectedClient := createTestClient(string(models.ClientTypeConfidential))

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return expectedClient, nil
	}

	// Act
	client, err := svc.GetClient(ctx, expectedClient.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedClient.ID, client.ID)
	assert.Equal(t, expectedClient.Name, client.Name)
}

func TestGetClient_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	client, err := svc.GetClient(ctx, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, client)
}

// ============================================================================
// DeleteClient Tests
// ============================================================================

func TestDeleteClient_ShouldDeleteClient_WhenClientExists(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	clientDeleted := false

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.DeleteClientFunc = func(ctx context.Context, id uuid.UUID) error {
		clientDeleted = true
		return nil
	}

	// Act
	err := svc.DeleteClient(ctx, client.ID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, clientDeleted)
}

func TestDeleteClient_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	err := svc.DeleteClient(ctx, uuid.New())

	// Assert
	assert.Error(t, err)
}

// ============================================================================
// UpdateClient Tests
// ============================================================================

func TestUpdateClient_ShouldUpdateClient_WhenValidRequest(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	updateReq := &models.UpdateOAuthClientRequest{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.UpdateClientFunc = func(ctx context.Context, c *models.OAuthClient) error {
		assert.Equal(t, "Updated Name", c.Name)
		assert.Equal(t, "Updated Description", c.Description)
		return nil
	}

	// Act
	updatedClient, err := svc.UpdateClient(ctx, client.ID, updateReq)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedClient.Name)
	assert.Equal(t, "Updated Description", updatedClient.Description)
}

func TestUpdateClient_ShouldPartiallyUpdate_WhenSomeFieldsProvided(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))
	originalDescription := client.Description

	updateReq := &models.UpdateOAuthClientRequest{
		Name: "Only Name Updated",
	}

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.UpdateClientFunc = func(ctx context.Context, c *models.OAuthClient) error {
		return nil
	}

	// Act
	updatedClient, err := svc.UpdateClient(ctx, client.ID, updateReq)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Only Name Updated", updatedClient.Name)
	assert.Equal(t, originalDescription, updatedClient.Description)
}

// ============================================================================
// RotateClientSecret Tests
// ============================================================================

func TestRotateClientSecret_ShouldGenerateNewSecret_WhenClientIsConfidential(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypeConfidential))

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return client, nil
	}
	mRepo.UpdateClientFunc = func(ctx context.Context, c *models.OAuthClient) error {
		assert.NotNil(t, c.ClientSecretHash)
		return nil
	}

	// Act
	newSecret, err := svc.RotateClientSecret(ctx, client.ID)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, newSecret)
	assert.Contains(t, newSecret, "agws_") // Check prefix
}

func TestRotateClientSecret_ShouldReturnError_WhenClientIsPublic(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	client := createTestClient(string(models.ClientTypePublic))

	mRepo.GetClientByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
		return client, nil
	}

	// Act
	newSecret, err := svc.RotateClientSecret(ctx, client.ID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot rotate secret for public client")
	assert.Empty(t, newSecret)
}

// ============================================================================
// ApproveDeviceCode Tests
// ============================================================================

func TestApproveDeviceCode_ShouldApprove_WhenDeviceCodeValid(t *testing.T) {
	// Arrange
	svc, mRepo, _, mAuditRepo := setupOAuthProviderService()
	ctx := context.Background()

	userID := uuid.New()
	userCode := "ABCD-EFGH"

	deviceCode := &models.DeviceCode{
		ID:        uuid.New(),
		UserCode:  userCode,
		Status:    models.DeviceCodeStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		ClientID:  uuid.New(),
	}

	mRepo.GetDeviceCodeByUserCodeFunc = func(ctx context.Context, code string) (*models.DeviceCode, error) {
		return deviceCode, nil
	}

	statusUpdated := false
	mRepo.UpdateDeviceCodeStatusFunc = func(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, uID *uuid.UUID) error {
		statusUpdated = true
		assert.Equal(t, models.DeviceCodeStatusAuthorized, status)
		assert.NotNil(t, uID)
		assert.Equal(t, userID, *uID)
		return nil
	}
	mAuditRepo.CreateFunc = func(ctx context.Context, log *models.AuditLog) error {
		return nil
	}

	// Act
	err := svc.ApproveDeviceCode(ctx, userID, userCode, true)

	// Assert
	assert.NoError(t, err)
	assert.True(t, statusUpdated)
}

func TestApproveDeviceCode_ShouldDeny_WhenApproveFalse(t *testing.T) {
	// Arrange
	svc, mRepo, _, mAuditRepo := setupOAuthProviderService()
	ctx := context.Background()

	userID := uuid.New()
	userCode := "ABCD-EFGH"

	deviceCode := &models.DeviceCode{
		ID:        uuid.New(),
		UserCode:  userCode,
		Status:    models.DeviceCodeStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	mRepo.GetDeviceCodeByUserCodeFunc = func(ctx context.Context, code string) (*models.DeviceCode, error) {
		return deviceCode, nil
	}

	mRepo.UpdateDeviceCodeStatusFunc = func(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, uID *uuid.UUID) error {
		assert.Equal(t, models.DeviceCodeStatusDenied, status)
		assert.Nil(t, uID)
		return nil
	}
	mAuditRepo.CreateFunc = func(ctx context.Context, log *models.AuditLog) error {
		return nil
	}

	// Act
	err := svc.ApproveDeviceCode(ctx, userID, userCode, false)

	// Assert
	assert.NoError(t, err)
}

func TestApproveDeviceCode_ShouldReturnError_WhenDeviceCodeNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetDeviceCodeByUserCodeFunc = func(ctx context.Context, code string) (*models.DeviceCode, error) {
		return nil, errors.New("not found")
	}

	// Act
	err := svc.ApproveDeviceCode(ctx, uuid.New(), "INVALID", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device code not found")
}

func TestApproveDeviceCode_ShouldReturnError_WhenDeviceCodeExpired(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	deviceCode := &models.DeviceCode{
		ID:        uuid.New(),
		UserCode:  "ABCD-EFGH",
		Status:    models.DeviceCodeStatusPending,
		ExpiresAt: time.Now().Add(-time.Minute), // Expired
	}

	mRepo.GetDeviceCodeByUserCodeFunc = func(ctx context.Context, code string) (*models.DeviceCode, error) {
		return deviceCode, nil
	}

	// Act
	err := svc.ApproveDeviceCode(ctx, uuid.New(), "ABCD-EFGH", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device code expired")
}

func TestApproveDeviceCode_ShouldReturnError_WhenDeviceCodeAlreadyProcessed(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	deviceCode := &models.DeviceCode{
		ID:        uuid.New(),
		UserCode:  "ABCD-EFGH",
		Status:    models.DeviceCodeStatusAuthorized, // Already processed
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	mRepo.GetDeviceCodeByUserCodeFunc = func(ctx context.Context, code string) (*models.DeviceCode, error) {
		return deviceCode, nil
	}

	// Act
	err := svc.ApproveDeviceCode(ctx, uuid.New(), "ABCD-EFGH", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device code already processed")
}

// ============================================================================
// GetClientByClientID Tests
// ============================================================================

func TestGetClientByClientID_ShouldReturnClient_WhenClientExists(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	expectedClient := createTestClient(string(models.ClientTypeConfidential))

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return expectedClient, nil
	}

	// Act
	client, err := svc.GetClientByClientID(ctx, expectedClient.ClientID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedClient.ClientID, client.ClientID)
}

func TestGetClientByClientID_ShouldReturnError_WhenClientNotFound(t *testing.T) {
	// Arrange
	svc, mRepo, _, _ := setupOAuthProviderService()
	ctx := context.Background()

	mRepo.GetClientByClientIDFunc = func(ctx context.Context, clientID string) (*models.OAuthClient, error) {
		return nil, errors.New("not found")
	}

	// Act
	client, err := svc.GetClientByClientID(ctx, "nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, client)
}
