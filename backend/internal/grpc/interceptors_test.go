package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// ===================== Mock Implementations =====================

// mockAPIKeyStoreForGRPC implements service.APIKeyStore
type mockAPIKeyStoreForGRPC struct {
	GetByKeyHashFunc   func(ctx context.Context, keyHash string) (*models.APIKey, error)
	UpdateLastUsedFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAPIKeyStoreForGRPC) Create(ctx context.Context, apiKey *models.APIKey) error {
	return nil
}
func (m *mockAPIKeyStoreForGRPC) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreForGRPC) GetByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	if m.GetByKeyHashFunc != nil {
		return m.GetByKeyHashFunc(ctx, keyHash)
	}
	return nil, errors.New("not found")
}
func (m *mockAPIKeyStoreForGRPC) GetByUserID(ctx context.Context, userID uuid.UUID, opts ...service.APIKeyGetOption) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreForGRPC) Update(ctx context.Context, apiKey *models.APIKey) error {
	return nil
}
func (m *mockAPIKeyStoreForGRPC) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(ctx, id)
	}
	return nil
}
func (m *mockAPIKeyStoreForGRPC) Revoke(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockAPIKeyStoreForGRPC) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockAPIKeyStoreForGRPC) DeleteExpired(ctx context.Context) error        { return nil }
func (m *mockAPIKeyStoreForGRPC) Count(ctx context.Context, userID uuid.UUID, opts ...service.APIKeyGetOption) (int, error) {
	return 0, nil
}
func (m *mockAPIKeyStoreForGRPC) ListAll(ctx context.Context) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreForGRPC) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreForGRPC) GetByUserIDAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.APIKey, error) {
	return nil, nil
}

// mockUserStoreForGRPC implements service.UserStore
type mockUserStoreForGRPC struct {
	GetByIDFunc func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error)
}

func (m *mockUserStoreForGRPC) GetByID(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id, isActive, opts...)
	}
	return nil, nil
}
func (m *mockUserStoreForGRPC) GetByEmail(ctx context.Context, email string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	return nil, nil
}
func (m *mockUserStoreForGRPC) GetByUsername(ctx context.Context, username string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	return nil, nil
}
func (m *mockUserStoreForGRPC) GetByPhone(ctx context.Context, phone string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	return nil, nil
}
func (m *mockUserStoreForGRPC) Create(ctx context.Context, user *models.User) error      { return nil }
func (m *mockUserStoreForGRPC) Update(ctx context.Context, user *models.User) error      { return nil }
func (m *mockUserStoreForGRPC) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}
func (m *mockUserStoreForGRPC) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreForGRPC) UsernameExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreForGRPC) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreForGRPC) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockUserStoreForGRPC) MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockUserStoreForGRPC) List(ctx context.Context, opts ...service.UserListOption) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserStoreForGRPC) Count(ctx context.Context, isActive *bool) (int, error) {
	return 0, nil
}
func (m *mockUserStoreForGRPC) GetUsersUpdatedAfter(ctx context.Context, after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (m *mockUserStoreForGRPC) UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return nil
}
func (m *mockUserStoreForGRPC) EnableTOTP(ctx context.Context, userID uuid.UUID) error  { return nil }
func (m *mockUserStoreForGRPC) DisableTOTP(ctx context.Context, userID uuid.UUID) error { return nil }

// mockAuditLoggerForGRPC implements service.AuditLogger
type mockAuditLoggerForGRPC struct{}

func (m *mockAuditLoggerForGRPC) LogWithAction(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
}
func (m *mockAuditLoggerForGRPC) Log(params service.AuditLogParams) {}

// mockApplicationStoreForGRPC implements service.ApplicationStore
type mockApplicationStoreForGRPC struct {
	GetApplicationByIDFunc   func(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetBySecretHashFunc      func(ctx context.Context, hash string) (*models.Application, error)
	GetApplicationByNameFunc func(ctx context.Context, name string) (*models.Application, error)
}

func (m *mockApplicationStoreForGRPC) CreateApplication(ctx context.Context, app *models.Application) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	if m.GetApplicationByIDFunc != nil {
		return m.GetApplicationByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}
func (m *mockApplicationStoreForGRPC) GetApplicationByName(ctx context.Context, name string) (*models.Application, error) {
	if m.GetApplicationByNameFunc != nil {
		return m.GetApplicationByNameFunc(ctx, name)
	}
	return nil, errors.New("not found")
}
func (m *mockApplicationStoreForGRPC) UpdateApplication(ctx context.Context, app *models.Application) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
	return nil, 0, nil
}
func (m *mockApplicationStoreForGRPC) GetBySecretHash(ctx context.Context, hash string) (*models.Application, error) {
	if m.GetBySecretHashFunc != nil {
		return m.GetBySecretHashFunc(ctx, hash)
	}
	return nil, errors.New("not found")
}
func (m *mockApplicationStoreForGRPC) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	return nil, nil
}
func (m *mockApplicationStoreForGRPC) CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockApplicationStoreForGRPC) UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockApplicationStoreForGRPC) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
	return nil, 0, nil
}
func (m *mockApplicationStoreForGRPC) UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	return nil
}
func (m *mockApplicationStoreForGRPC) UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}

// mockAppOAuthProviderStoreForGRPC implements service.AppOAuthProviderStore
type mockAppOAuthProviderStoreForGRPC struct{}

func (m *mockAppOAuthProviderStoreForGRPC) Create(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}
func (m *mockAppOAuthProviderStoreForGRPC) GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *mockAppOAuthProviderStoreForGRPC) GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *mockAppOAuthProviderStoreForGRPC) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *mockAppOAuthProviderStoreForGRPC) ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error) {
	return nil, nil
}
func (m *mockAppOAuthProviderStoreForGRPC) Update(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	return nil
}
func (m *mockAppOAuthProviderStoreForGRPC) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// ===================== Helpers =====================

func testLogger() *logger.Logger {
	return logger.New("test", logger.DebugLevel, false)
}

func noopHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func errorHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, status.Error(codes.Internal, "handler error")
}

func panicHandler(ctx context.Context, req interface{}) (interface{}, error) {
	panic("test panic")
}

func contextCapturingHandler(captured *context.Context) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		*captured = ctx
		return "ok", nil
	}
}

func buildAPIKeyService(apiKeyStore *mockAPIKeyStoreForGRPC, userStore *mockUserStoreForGRPC) *service.APIKeyService {
	return service.NewAPIKeyService(apiKeyStore, userStore, &mockAuditLoggerForGRPC{})
}

func buildApplicationService(appStore *mockApplicationStoreForGRPC) *service.ApplicationService {
	return service.NewApplicationService(appStore, &mockAppOAuthProviderStoreForGRPC{}, testLogger())
}

func makeAPIKeyForPlainKey(userID uuid.UUID, plainKey string, scopes []string) *models.APIKey {
	scopesJSON, _ := json.Marshal(scopes)
	return &models.APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		KeyHash:   utils.HashToken(plainKey),
		KeyPrefix: "agw_testpref",
		Scopes:    scopesJSON,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func makeUser(id uuid.UUID) *models.User {
	return &models.User{
		ID:       id,
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
	}
}

func makeApplication(id uuid.UUID) *models.Application {
	return &models.Application{
		ID:       id,
		Name:     "test-app",
		IsActive: true,
	}
}

// ===================== extractCredential Tests =====================

func TestExtractCredential_ShouldReturnAPIKey_WhenXApiKeyPresent(t *testing.T) {
	md := metadata.Pairs("x-api-key", "agw_testkey123")
	result := extractCredential(md)
	assert.Equal(t, "agw_testkey123", result)
}

func TestExtractCredential_ShouldReturnBearerToken_WhenAuthorizationPresent(t *testing.T) {
	md := metadata.Pairs("authorization", "Bearer some-token")
	result := extractCredential(md)
	assert.Equal(t, "some-token", result)
}

func TestExtractCredential_ShouldPreferAPIKey_WhenBothPresent(t *testing.T) {
	md := metadata.Pairs("x-api-key", "agw_testkey123", "authorization", "Bearer some-token")
	result := extractCredential(md)
	assert.Equal(t, "agw_testkey123", result)
}

func TestExtractCredential_ShouldReturnEmpty_WhenNoCredentials(t *testing.T) {
	md := metadata.Pairs("other-header", "value")
	result := extractCredential(md)
	assert.Empty(t, result)
}

func TestExtractCredential_ShouldReturnEmpty_WhenEmptyAPIKey(t *testing.T) {
	md := metadata.Pairs("x-api-key", "")
	result := extractCredential(md)
	assert.Empty(t, result)
}

func TestExtractCredential_ShouldReturnEmpty_WhenAuthorizationNotBearer(t *testing.T) {
	md := metadata.Pairs("authorization", "Basic dXNlcjpwYXNz")
	result := extractCredential(md)
	assert.Empty(t, result)
}

// ===================== containsScope Tests =====================

func TestContainsScope_ShouldReturnTrue_WhenScopeExists(t *testing.T) {
	scopes := []string{"users:read", "token:validate", "admin:all"}
	assert.True(t, containsScope(scopes, "token:validate"))
}

func TestContainsScope_ShouldReturnFalse_WhenScopeNotFound(t *testing.T) {
	scopes := []string{"users:read", "token:validate"}
	assert.False(t, containsScope(scopes, "admin:all"))
}

func TestContainsScope_ShouldReturnFalse_WhenScopesEmpty(t *testing.T) {
	assert.False(t, containsScope([]string{}, "users:read"))
}

func TestContainsScope_ShouldReturnFalse_WhenScopesNil(t *testing.T) {
	assert.False(t, containsScope(nil, "users:read"))
}

// ===================== GetApplicationIDFromGRPCContext Tests =====================

func TestGetApplicationIDFromGRPCContext_ShouldReturnID_WhenSet(t *testing.T) {
	appID := uuid.New().String()
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, appID)

	result := GetApplicationIDFromGRPCContext(ctx)
	require.NotNil(t, result)
	assert.Equal(t, appID, *result)
}

func TestGetApplicationIDFromGRPCContext_ShouldReturnNil_WhenNotSet(t *testing.T) {
	ctx := context.Background()
	result := GetApplicationIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

func TestGetApplicationIDFromGRPCContext_ShouldReturnNil_WhenEmptyString(t *testing.T) {
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, "")
	result := GetApplicationIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

func TestGetApplicationIDFromGRPCContext_ShouldReturnNil_WhenWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, 12345)
	result := GetApplicationIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

// ===================== GetApplicationUUIDFromGRPCContext Tests =====================

func TestGetApplicationUUIDFromGRPCContext_ShouldReturnUUID_WhenValidUUIDSet(t *testing.T) {
	appID := uuid.New()
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, appID.String())

	result := GetApplicationUUIDFromGRPCContext(ctx)
	require.NotNil(t, result)
	assert.Equal(t, appID, *result)
}

func TestGetApplicationUUIDFromGRPCContext_ShouldReturnNil_WhenInvalidUUID(t *testing.T) {
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, "not-a-uuid")
	result := GetApplicationUUIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

func TestGetApplicationUUIDFromGRPCContext_ShouldReturnNil_WhenNotSet(t *testing.T) {
	result := GetApplicationUUIDFromGRPCContext(context.Background())
	assert.Nil(t, result)
}

// ===================== ResolveApplicationID Tests =====================

func TestResolveApplicationID_ShouldReturnRequestID_WhenProvided(t *testing.T) {
	reqAppID := uuid.New().String()
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, uuid.New().String())

	result := ResolveApplicationID(ctx, reqAppID)
	assert.Equal(t, reqAppID, result)
}

func TestResolveApplicationID_ShouldReturnContextID_WhenRequestIDEmpty(t *testing.T) {
	ctxAppID := uuid.New().String()
	ctx := context.WithValue(context.Background(), GRPCApplicationIDKey, ctxAppID)

	result := ResolveApplicationID(ctx, "")
	assert.Equal(t, ctxAppID, result)
}

func TestResolveApplicationID_ShouldReturnEmpty_WhenNeitherAvailable(t *testing.T) {
	result := ResolveApplicationID(context.Background(), "")
	assert.Empty(t, result)
}

// ===================== GetTenantIDFromGRPCContext Tests =====================

func TestGetTenantIDFromGRPCContext_ShouldReturnID_WhenSet(t *testing.T) {
	tenantID := "tenant-123"
	ctx := context.WithValue(context.Background(), GRPCTenantIDKey, tenantID)

	result := GetTenantIDFromGRPCContext(ctx)
	require.NotNil(t, result)
	assert.Equal(t, tenantID, *result)
}

func TestGetTenantIDFromGRPCContext_ShouldReturnNil_WhenNotSet(t *testing.T) {
	result := GetTenantIDFromGRPCContext(context.Background())
	assert.Nil(t, result)
}

func TestGetTenantIDFromGRPCContext_ShouldReturnNil_WhenEmptyString(t *testing.T) {
	ctx := context.WithValue(context.Background(), GRPCTenantIDKey, "")
	result := GetTenantIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

func TestGetTenantIDFromGRPCContext_ShouldReturnNil_WhenWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), GRPCTenantIDKey, 999)
	result := GetTenantIDFromGRPCContext(ctx)
	assert.Nil(t, result)
}

// ===================== contextExtractorInterceptor Tests =====================

func TestContextExtractorInterceptor_ShouldExtractTenantID_WhenPresent(t *testing.T) {
	interceptor := contextExtractorInterceptor(testLogger())
	tenantID := "tenant-abc"
	md := metadata.Pairs("x-tenant-id", tenantID)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var captured context.Context
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, contextCapturingHandler(&captured))

	require.NoError(t, err)
	result := GetTenantIDFromGRPCContext(captured)
	require.NotNil(t, result)
	assert.Equal(t, tenantID, *result)
}

func TestContextExtractorInterceptor_ShouldPassThrough_WhenNoMetadata(t *testing.T) {
	interceptor := contextExtractorInterceptor(testLogger())
	ctx := context.Background()

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, noopHandler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestContextExtractorInterceptor_ShouldIgnoreEmptyTenantID(t *testing.T) {
	interceptor := contextExtractorInterceptor(testLogger())
	md := metadata.Pairs("x-tenant-id", "")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var captured context.Context
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, contextCapturingHandler(&captured))

	require.NoError(t, err)
	result := GetTenantIDFromGRPCContext(captured)
	assert.Nil(t, result)
}

// ===================== loggingInterceptor Tests =====================

func TestLoggingInterceptor_ShouldPassThrough_WhenHandlerSucceeds(t *testing.T) {
	interceptor := loggingInterceptor(testLogger())

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}, noopHandler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestLoggingInterceptor_ShouldReturnError_WhenHandlerFails(t *testing.T) {
	interceptor := loggingInterceptor(testLogger())

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/GetUser"}, errorHandler)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

// ===================== recoveryInterceptor Tests =====================

func TestRecoveryInterceptor_ShouldPassThrough_WhenNoPanic(t *testing.T) {
	interceptor := recoveryInterceptor(testLogger())

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, noopHandler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestRecoveryInterceptor_ShouldCatchPanic_WhenHandlerPanics(t *testing.T) {
	interceptor := recoveryInterceptor(testLogger())

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Crash"}, panicHandler)

	require.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestRecoveryInterceptor_ShouldPreserveHandlerError_WhenNoPanic(t *testing.T) {
	interceptor := recoveryInterceptor(testLogger())

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, errorHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "handler error", st.Message())
}

// ===================== apiKeyAuthInterceptor Tests =====================

func TestAPIKeyAuthInterceptor_ShouldRejectRequest_WhenNoMetadata(t *testing.T) {
	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "missing API key")
}

func TestAPIKeyAuthInterceptor_ShouldRejectRequest_WhenNoCredential(t *testing.T) {
	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("other-header", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAPIKeyAuthInterceptor_ShouldRejectRequest_WhenMethodNotInScopeMap(t *testing.T) {
	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", "agw_testkey")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/UnknownMethod"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "not configured for access")
}

func TestAPIKeyAuthInterceptor_ShouldRejectRequest_WhenInvalidAPIKey(t *testing.T) {
	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, &mockUserStoreForGRPC{})
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", "agw_invalidkey")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid API key")
}

func TestAPIKeyAuthInterceptor_ShouldAuthenticate_WhenValidAPIKeyWithScope(t *testing.T) {
	userID := uuid.New()
	plainKey := "agw_testvalidkey"
	keyHash := utils.HashToken(plainKey)
	apiKey := makeAPIKeyForPlainKey(userID, plainKey, []string{"token:validate"})
	user := makeUser(userID)

	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			if hash == keyHash {
				return apiKey, nil
			}
			return nil, errors.New("not found")
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	userStore := &mockUserStoreForGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, userStore)
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", plainKey)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	var captured context.Context
	resp, err := interceptor(ctx, nil, info, contextCapturingHandler(&captured))

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)

	// Verify user context values
	assert.Equal(t, userID.String(), captured.Value(GRPCUserIDKey))
	assert.Equal(t, "test@example.com", captured.Value(GRPCUserEmailKey))
}

func TestAPIKeyAuthInterceptor_ShouldReject_WhenInsufficientScope(t *testing.T) {
	userID := uuid.New()
	plainKey := "agw_wrongscope"
	keyHash := utils.HashToken(plainKey)
	// Key has "users:read" scope but method requires "token:validate"
	apiKey := makeAPIKeyForPlainKey(userID, plainKey, []string{"users:read"})
	user := makeUser(userID)

	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			if hash == keyHash {
				return apiKey, nil
			}
			return nil, errors.New("not found")
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	userStore := &mockUserStoreForGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, userStore)
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", plainKey)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "insufficient scope")
}

func TestAPIKeyAuthInterceptor_ShouldSetAppID_WhenXApplicationIDPresent(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	plainKey := "agw_withappid"
	keyHash := utils.HashToken(plainKey)
	apiKey := makeAPIKeyForPlainKey(userID, plainKey, []string{"token:validate"})
	user := makeUser(userID)
	app := makeApplication(appID)

	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			if hash == keyHash {
				return apiKey, nil
			}
			return nil, errors.New("not found")
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	userStore := &mockUserStoreForGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}
	appStore := &mockApplicationStoreForGRPC{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			if id == appID {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, userStore)
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", plainKey, "x-application-id", appID.String())
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	var captured context.Context
	_, err := interceptor(ctx, nil, info, contextCapturingHandler(&captured))

	require.NoError(t, err)
	result := GetApplicationIDFromGRPCContext(captured)
	require.NotNil(t, result)
	assert.Equal(t, appID.String(), *result)
}

func TestAPIKeyAuthInterceptor_ShouldReject_WhenInvalidApplicationIDFormat(t *testing.T) {
	userID := uuid.New()
	plainKey := "agw_badappid"
	keyHash := utils.HashToken(plainKey)
	apiKey := makeAPIKeyForPlainKey(userID, plainKey, []string{"token:validate"})
	user := makeUser(userID)

	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			if hash == keyHash {
				return apiKey, nil
			}
			return nil, errors.New("not found")
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	userStore := &mockUserStoreForGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, userStore)
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", plainKey, "x-application-id", "not-a-valid-uuid")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "invalid application ID format")
}

func TestAPIKeyAuthInterceptor_ShouldReject_WhenApplicationNotFound(t *testing.T) {
	userID := uuid.New()
	unknownAppID := uuid.New()
	plainKey := "agw_appnotfound"
	keyHash := utils.HashToken(plainKey)
	apiKey := makeAPIKeyForPlainKey(userID, plainKey, []string{"token:validate"})
	user := makeUser(userID)

	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			if hash == keyHash {
				return apiKey, nil
			}
			return nil, errors.New("not found")
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	userStore := &mockUserStoreForGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}
	appStore := &mockApplicationStoreForGRPC{
		GetApplicationByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, userStore)
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", plainKey, "x-application-id", unknownAppID.String())
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "application not found")
}

func TestAPIKeyAuthInterceptor_ShouldAuth_WhenValidAppSecret(t *testing.T) {
	appID := uuid.New()
	secret := "app_testsecret123"
	secretHash := utils.HashToken(secret)

	app := &models.Application{
		ID:         appID,
		Name:       "test-app",
		IsActive:   true,
		SecretHash: secretHash,
	}

	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			if hash == secretHash {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", secret)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	var captured context.Context
	resp, err := interceptor(ctx, nil, info, contextCapturingHandler(&captured))

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)

	result := GetApplicationIDFromGRPCContext(captured)
	require.NotNil(t, result)
	assert.Equal(t, appID.String(), *result)
}

func TestAPIKeyAuthInterceptor_ShouldReject_WhenInvalidAppSecret(t *testing.T) {
	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", "app_invalidsecret")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid application secret")
}

func TestAPIKeyAuthInterceptor_ShouldReject_WhenAppSecretScopeRestricted(t *testing.T) {
	appID := uuid.New()
	secret := "app_scopedapp"
	secretHash := utils.HashToken(secret)

	// App only allows "users:read" but method needs "token:validate"
	app := &models.Application{
		ID:                appID,
		Name:              "scoped-app",
		IsActive:          true,
		SecretHash:        secretHash,
		AllowedGRPCScopes: []string{"users:read"},
	}

	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			if hash == secretHash {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", secret)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "application not authorized for scope")
}

func TestAPIKeyAuthInterceptor_ShouldAllow_WhenAppSecretHasMatchingScope(t *testing.T) {
	appID := uuid.New()
	secret := "app_allowedscope"
	secretHash := utils.HashToken(secret)

	app := &models.Application{
		ID:                appID,
		Name:              "allowed-app",
		IsActive:          true,
		SecretHash:        secretHash,
		AllowedGRPCScopes: []string{"token:validate", "users:read"},
	}

	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			if hash == secretHash {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", secret)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	resp, err := interceptor(ctx, nil, info, noopHandler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestAPIKeyAuthInterceptor_ShouldAllow_WhenAppSecretHasNoScopeRestrictions(t *testing.T) {
	appID := uuid.New()
	secret := "app_fullaccess"
	secretHash := utils.HashToken(secret)

	app := &models.Application{
		ID:                appID,
		Name:              "full-access-app",
		IsActive:          true,
		SecretHash:        secretHash,
		AllowedGRPCScopes: []string{}, // empty = full access
	}

	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			if hash == secretHash {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", secret)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	resp, err := interceptor(ctx, nil, info, noopHandler)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestAPIKeyAuthInterceptor_ShouldUseBearer_WhenAuthorizationHeader(t *testing.T) {
	apiKeyStore := &mockAPIKeyStoreForGRPC{
		GetByKeyHashFunc: func(ctx context.Context, hash string) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	apiKeyService := buildAPIKeyService(apiKeyStore, &mockUserStoreForGRPC{})
	appService := buildApplicationService(&mockApplicationStoreForGRPC{})
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	// Use Authorization header with Bearer token
	md := metadata.Pairs("authorization", "Bearer agw_bearerkey")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/ValidateToken"}

	// Should try to validate "agw_bearerkey" as API key and fail
	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// ===================== methodScopes Tests =====================

func TestMethodScopes_ShouldContainAllExpectedMethods(t *testing.T) {
	expectedMethods := []string{
		"/auth.AuthService/ValidateToken",
		"/auth.AuthService/IntrospectToken",
		"/auth.AuthService/GetUser",
		"/auth.AuthService/CheckPermission",
		"/auth.AuthService/GetApplicationAuthConfig",
		"/auth.AuthService/GetUserApplicationProfile",
		"/auth.AuthService/UpdateUserProfile",
		"/auth.AuthService/CreateUserProfile",
		"/auth.AuthService/DeleteUserProfile",
		"/auth.AuthService/BanUser",
		"/auth.AuthService/UnbanUser",
		"/auth.AuthService/ListApplicationUsers",
		"/auth.AuthService/CreateUser",
		"/auth.AuthService/Login",
		"/auth.AuthService/SendOTP",
		"/auth.AuthService/VerifyOTP",
		"/auth.AuthService/LoginWithOTP",
		"/auth.AuthService/VerifyLoginOTP",
		"/auth.AuthService/RegisterWithOTP",
		"/auth.AuthService/VerifyRegistrationOTP",
		"/auth.AuthService/InitPasswordlessRegistration",
		"/auth.AuthService/CompletePasswordlessRegistration",
		"/auth.AuthService/SyncUsers",
		"/auth.AuthService/SendEmail",
		"/auth.AuthService/IntrospectOAuthToken",
		"/auth.AuthService/ValidateOAuthClient",
		"/auth.AuthService/GetOAuthClient",
		"/auth.AuthService/CreateTokenExchange",
		"/auth.AuthService/RedeemTokenExchange",
	}

	for _, method := range expectedMethods {
		_, ok := methodScopes[method]
		assert.True(t, ok, "method %s should be in methodScopes map", method)
	}
}

func TestMethodScopes_ShouldNotContainGetUserTelegramBots(t *testing.T) {
	_, ok := methodScopes["/auth.AuthService/GetUserTelegramBots"]
	assert.False(t, ok, "GetUserTelegramBots should not be in methodScopes (deny-by-default)")
}

func TestMethodScopes_ShouldHaveCorrectScopes(t *testing.T) {
	tests := []struct {
		method string
		scope  models.APIKeyScope
	}{
		{"/auth.AuthService/ValidateToken", models.ScopeValidateToken},
		{"/auth.AuthService/IntrospectToken", models.ScopeIntrospectToken},
		{"/auth.AuthService/GetUser", models.ScopeReadUsers},
		{"/auth.AuthService/CheckPermission", models.ScopeReadUsers},
		{"/auth.AuthService/Login", models.ScopeAuthLogin},
		{"/auth.AuthService/CreateUser", models.ScopeAuthRegister},
		{"/auth.AuthService/SendOTP", models.ScopeAuthOTP},
		{"/auth.AuthService/VerifyOTP", models.ScopeAuthOTP},
		{"/auth.AuthService/SyncUsers", models.ScopeSyncUsers},
		{"/auth.AuthService/SendEmail", models.ScopeEmailSend},
		{"/auth.AuthService/IntrospectOAuthToken", models.ScopeOAuthRead},
		{"/auth.AuthService/CreateTokenExchange", models.ScopeExchangeManage},
		{"/auth.AuthService/RedeemTokenExchange", models.ScopeExchangeManage},
		{"/auth.AuthService/GetUserApplicationProfile", models.ScopeReadProfile},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			scope, ok := methodScopes[tt.method]
			require.True(t, ok, "method %s should exist in methodScopes", tt.method)
			assert.Equal(t, tt.scope, scope, "method %s has wrong scope", tt.method)
		})
	}
}

// ===================== Deny-by-default behavior =====================

func TestAPIKeyAuthInterceptor_DenyByDefault_ShouldRejectUnknownMethod(t *testing.T) {
	appID := uuid.New()
	secret := "app_denytest"
	secretHash := utils.HashToken(secret)
	app := &models.Application{
		ID:         appID,
		Name:       "deny-test",
		IsActive:   true,
		SecretHash: secretHash,
	}

	appStore := &mockApplicationStoreForGRPC{
		GetBySecretHashFunc: func(ctx context.Context, hash string) (*models.Application, error) {
			if hash == secretHash {
				return app, nil
			}
			return nil, errors.New("not found")
		},
	}

	apiKeyService := buildAPIKeyService(&mockAPIKeyStoreForGRPC{}, &mockUserStoreForGRPC{})
	appService := buildApplicationService(appStore)
	interceptor := apiKeyAuthInterceptor(apiKeyService, appService, testLogger())

	md := metadata.Pairs("x-api-key", secret)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/SomeNewMethodNotInMap"}

	_, err := interceptor(ctx, nil, info, noopHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "not configured for access")
}

// ===================== Interceptor Chain Integration Tests =====================

func TestInterceptorChain_ShouldRecoverFromPanicInLoggedRequest(t *testing.T) {
	log := testLogger()
	recovery := recoveryInterceptor(log)
	logging := loggingInterceptor(log)

	// Simulate chain: logging -> recovery -> handler
	chainedHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return recovery(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/test"}, panicHandler)
	}

	resp, err := logging(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, chainedHandler)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
