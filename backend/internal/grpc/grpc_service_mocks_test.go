package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// ===================== mockUserStoreGRPC =====================

type mockUserStoreGRPC struct {
	GetByIDFunc              func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error)
	GetByEmailFunc           func(ctx context.Context, email string, isActive *bool, opts ...service.UserGetOption) (*models.User, error)
	ListFunc                 func(ctx context.Context, opts ...service.UserListOption) ([]*models.User, error)
	CountFunc                func(ctx context.Context, isActive *bool) (int, error)
	GetUsersUpdatedAfterFunc func(ctx context.Context, after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error)
}

func (m *mockUserStoreGRPC) GetByID(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id, isActive, opts...)
	}
	return nil, models.ErrUserNotFound
}
func (m *mockUserStoreGRPC) GetByEmail(ctx context.Context, email string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email, isActive, opts...)
	}
	return nil, models.ErrUserNotFound
}
func (m *mockUserStoreGRPC) GetByUsername(ctx context.Context, username string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	return nil, models.ErrUserNotFound
}
func (m *mockUserStoreGRPC) GetByPhone(ctx context.Context, phone string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
	return nil, models.ErrUserNotFound
}
func (m *mockUserStoreGRPC) Create(ctx context.Context, user *models.User) error      { return nil }
func (m *mockUserStoreGRPC) Update(ctx context.Context, user *models.User) error      { return nil }
func (m *mockUserStoreGRPC) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	return nil
}
func (m *mockUserStoreGRPC) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreGRPC) UsernameExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreGRPC) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return false, nil
}
func (m *mockUserStoreGRPC) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockUserStoreGRPC) MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockUserStoreGRPC) List(ctx context.Context, opts ...service.UserListOption) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, opts...)
	}
	return nil, nil
}
func (m *mockUserStoreGRPC) Count(ctx context.Context, isActive *bool) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, isActive)
	}
	return 0, nil
}
func (m *mockUserStoreGRPC) GetUsersUpdatedAfter(ctx context.Context, after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	if m.GetUsersUpdatedAfterFunc != nil {
		return m.GetUsersUpdatedAfterFunc(ctx, after, appID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockUserStoreGRPC) UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return nil
}
func (m *mockUserStoreGRPC) EnableTOTP(ctx context.Context, userID uuid.UUID) error  { return nil }
func (m *mockUserStoreGRPC) DisableTOTP(ctx context.Context, userID uuid.UUID) error { return nil }

// ===================== mockTokenStoreGRPC =====================

type mockTokenStoreGRPC struct {
	IsBlacklistedFunc func(ctx context.Context, tokenHash string) (bool, error)
}

func (m *mockTokenStoreGRPC) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return nil
}
func (m *mockTokenStoreGRPC) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	return nil, nil
}
func (m *mockTokenStoreGRPC) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return nil
}
func (m *mockTokenStoreGRPC) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockTokenStoreGRPC) AddToBlacklist(ctx context.Context, token *models.TokenBlacklist) error {
	return nil
}
func (m *mockTokenStoreGRPC) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(ctx, tokenHash)
	}
	return false, nil
}

// ===================== mockRBACStoreGRPC =====================

type mockRBACStoreGRPC struct {
	HasPermissionFunc    func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	GetUserRolesFunc     func(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	GetUserRolesInAppFunc func(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) ([]models.Role, error)
}

func (m *mockRBACStoreGRPC) CreatePermission(ctx context.Context, permission *models.Permission) error {
	return nil
}
func (m *mockRBACStoreGRPC) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) UpdatePermission(ctx context.Context, id uuid.UUID, description string) error {
	return nil
}
func (m *mockRBACStoreGRPC) DeletePermission(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRBACStoreGRPC) ListPermissionsByApp(ctx context.Context, appID *uuid.UUID) ([]models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) CreateRole(ctx context.Context, role *models.Role) error { return nil }
func (m *mockRBACStoreGRPC) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) ListRoles(ctx context.Context) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error {
	return nil
}
func (m *mockRBACStoreGRPC) DeleteRole(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRBACStoreGRPC) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreGRPC) GetRoleByNameAndApp(ctx context.Context, name string, appID *uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) ListRolesByApp(ctx context.Context, appID *uuid.UUID) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreGRPC) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreGRPC) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRBACStoreGRPC) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreGRPC) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) GetUserRolesInApp(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) ([]models.Role, error) {
	if m.GetUserRolesInAppFunc != nil {
		return m.GetUserRolesInAppFunc(ctx, userID, appID)
	}
	return nil, nil
}
func (m *mockRBACStoreGRPC) AssignRoleToUserInApp(ctx context.Context, userID, roleID, assignedBy uuid.UUID, appID *uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreGRPC) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if m.HasPermissionFunc != nil {
		return m.HasPermissionFunc(ctx, userID, permissionName)
	}
	return false, nil
}
func (m *mockRBACStoreGRPC) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreGRPC) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreGRPC) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	return nil, nil
}
func (m *mockRBACStoreGRPC) HasPermissionInApp(ctx context.Context, userID uuid.UUID, permissionName string, appID *uuid.UUID) (bool, error) {
	return false, nil
}

// ===================== mockAPIKeyServicerGRPC =====================

type mockAPIKeyServicerGRPC struct {
	ValidateAPIKeyFunc func(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error)
	HasScopeFunc       func(apiKey *models.APIKey, scope models.APIKeyScope) bool
}

func (m *mockAPIKeyServicerGRPC) GenerateAPIKey() (string, error) { return "", nil }
func (m *mockAPIKeyServicerGRPC) Create(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error) {
	return nil, nil
}
func (m *mockAPIKeyServicerGRPC) ValidateAPIKey(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error) {
	if m.ValidateAPIKeyFunc != nil {
		return m.ValidateAPIKeyFunc(ctx, plainKey)
	}
	return nil, nil, models.ErrInvalidCredentials
}
func (m *mockAPIKeyServicerGRPC) GetByID(ctx context.Context, userID, apiKeyID uuid.UUID) (*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyServicerGRPC) List(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyServicerGRPC) Update(ctx context.Context, userID, apiKeyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyServicerGRPC) Revoke(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *mockAPIKeyServicerGRPC) Delete(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *mockAPIKeyServicerGRPC) HasScope(apiKey *models.APIKey, scope models.APIKeyScope) bool {
	if m.HasScopeFunc != nil {
		return m.HasScopeFunc(apiKey, scope)
	}
	return true
}

// ===================== mockAuthServicerGRPC =====================

type mockAuthServicerGRPC struct {
	SignUpFunc                          func(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	SignInFunc                          func(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	InitPasswordlessRegistrationFunc    func(ctx context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error
	CompletePasswordlessRegistrationFunc func(ctx context.Context, req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
}

func (m *mockAuthServicerGRPC) SignUp(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	if m.SignUpFunc != nil {
		return m.SignUpFunc(ctx, req, ip, userAgent, deviceInfo, appID)
	}
	return nil, models.ErrInvalidCredentials
}
func (m *mockAuthServicerGRPC) SignIn(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	if m.SignInFunc != nil {
		return m.SignInFunc(ctx, req, ip, userAgent, deviceInfo, appID)
	}
	return nil, models.ErrInvalidCredentials
}
func (m *mockAuthServicerGRPC) Verify2FALogin(ctx context.Context, twoFactorToken, code, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthServicerGRPC) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthServicerGRPC) Logout(ctx context.Context, accessToken, ip, userAgent string) error {
	return nil
}
func (m *mockAuthServicerGRPC) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error {
	return nil
}
func (m *mockAuthServicerGRPC) ResetPassword(ctx context.Context, userID uuid.UUID, newPassword, ip, userAgent string) error {
	return nil
}
func (m *mockAuthServicerGRPC) InitPasswordlessRegistration(ctx context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error {
	if m.InitPasswordlessRegistrationFunc != nil {
		return m.InitPasswordlessRegistrationFunc(ctx, req, ip, userAgent)
	}
	return nil
}
func (m *mockAuthServicerGRPC) CompletePasswordlessRegistration(ctx context.Context, req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	if m.CompletePasswordlessRegistrationFunc != nil {
		return m.CompletePasswordlessRegistrationFunc(ctx, req, ip, userAgent, deviceInfo)
	}
	return nil, nil
}
func (m *mockAuthServicerGRPC) GenerateTokensForUser(ctx context.Context, user *models.User, ip, userAgent string) (*models.AuthResponse, error) {
	return nil, nil
}

// ===================== mockOTPServicerGRPC =====================

type mockOTPServicerGRPC struct {
	SendOTPFunc   func(ctx context.Context, req *models.SendOTPRequest) error
	VerifyOTPFunc func(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error)
}

func (m *mockOTPServicerGRPC) GenerateOTPCode() (string, error) { return "123456", nil }
func (m *mockOTPServicerGRPC) SendOTP(ctx context.Context, req *models.SendOTPRequest) error {
	if m.SendOTPFunc != nil {
		return m.SendOTPFunc(ctx, req)
	}
	return nil
}
func (m *mockOTPServicerGRPC) VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	if m.VerifyOTPFunc != nil {
		return m.VerifyOTPFunc(ctx, req)
	}
	return &models.VerifyOTPResponse{Valid: true}, nil
}
func (m *mockOTPServicerGRPC) CleanupExpiredOTPs() error { return nil }

// ===================== mockOAuthProviderServicerGRPC =====================

type mockOAuthProviderServicerGRPC struct {
	IntrospectTokenFunc         func(ctx context.Context, token, tokenTypeHint string, clientID *string) (*models.IntrospectionResponse, error)
	ValidateClientCredentialsFunc func(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error)
	GetClientByClientIDFunc     func(ctx context.Context, clientID string) (*models.OAuthClient, error)
}

func (m *mockOAuthProviderServicerGRPC) CreateClient(ctx context.Context, req *models.CreateOAuthClientRequest, ownerID *uuid.UUID) (*models.CreateOAuthClientResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) GetClient(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	if m.GetClientByClientIDFunc != nil {
		return m.GetClientByClientIDFunc(ctx, clientID)
	}
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) UpdateClient(ctx context.Context, id uuid.UUID, req *models.UpdateOAuthClientRequest) (*models.OAuthClient, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) DeleteClient(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) ListClients(ctx context.Context, page, perPage int, opts ...service.OAuthClientListOption) ([]*models.OAuthClient, int, error) {
	return nil, 0, nil
}
func (m *mockOAuthProviderServicerGRPC) RotateClientSecret(ctx context.Context, id uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockOAuthProviderServicerGRPC) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error) {
	if m.ValidateClientCredentialsFunc != nil {
		return m.ValidateClientCredentialsFunc(ctx, clientID, clientSecret)
	}
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) Authorize(ctx context.Context, req *models.AuthorizeRequest, userID uuid.UUID) (*models.AuthorizeResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) ExchangeCode(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) ClientCredentialsGrant(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) RefreshToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) DeviceAuthorization(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) PollDeviceToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) ApproveDeviceCode(ctx context.Context, userID uuid.UUID, userCode string, approve bool) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) IntrospectToken(ctx context.Context, token, tokenTypeHint string, clientID *string) (*models.IntrospectionResponse, error) {
	if m.IntrospectTokenFunc != nil {
		return m.IntrospectTokenFunc(ctx, token, tokenTypeHint, clientID)
	}
	return &models.IntrospectionResponse{Active: false}, nil
}
func (m *mockOAuthProviderServicerGRPC) RevokeToken(ctx context.Context, token, tokenTypeHint string, clientID *string) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfoResponse, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) GetDiscoveryDocument() *models.OIDCDiscoveryDocument {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) GetJWKS() *models.JWKSDocument { return nil }
func (m *mockOAuthProviderServicerGRPC) GetConsentInfo(ctx context.Context, clientID string, scopes []string) (*service.ConsentInfo, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) GrantConsent(ctx context.Context, userID uuid.UUID, clientID string, scopes []string) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) ListScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	return nil, nil
}
func (m *mockOAuthProviderServicerGRPC) CreateScope(ctx context.Context, scope *models.OAuthScope) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) DeleteScope(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockOAuthProviderServicerGRPC) ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error) {
	return nil, nil
}

// ===================== mockEmailProfileServicerGRPC =====================

type mockEmailProfileServicerGRPC struct {
	SendEmailFunc func(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error
}

func (m *mockEmailProfileServicerGRPC) CreateProvider(ctx context.Context, req *models.CreateEmailProviderRequest) (*models.EmailProvider, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) GetProvider(ctx context.Context, id uuid.UUID) (*models.EmailProviderResponse, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) ListProviders(ctx context.Context, appID *uuid.UUID) ([]*models.EmailProviderResponse, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProviderRequest) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) TestProvider(ctx context.Context, id uuid.UUID, testEmail string) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) CreateProfile(ctx context.Context, req *models.CreateEmailProfileRequest) (*models.EmailProfile, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) GetProfile(ctx context.Context, id uuid.UUID) (*models.EmailProfile, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) ListProfiles(ctx context.Context, appID *uuid.UUID) ([]*models.EmailProfile, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) UpdateProfile(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProfileRequest) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) SetDefaultProfile(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) GetProfileTemplates(ctx context.Context, profileID uuid.UUID) ([]*models.EmailProfileTemplate, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) SetProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string, templateID uuid.UUID) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) RemoveProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) SendOTPEmail(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error {
	return nil
}
func (m *mockEmailProfileServicerGRPC) SendEmail(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(ctx, profileID, applicationID, toEmail, templateType, variables)
	}
	return nil
}
func (m *mockEmailProfileServicerGRPC) GetProfileStats(ctx context.Context, profileID uuid.UUID) (*models.EmailStatsResponse, error) {
	return nil, nil
}
func (m *mockEmailProfileServicerGRPC) TestProfile(ctx context.Context, profileID uuid.UUID, testEmail string) error {
	return nil
}

// ===================== mockAdminServicerGRPC =====================

type mockAdminServicerGRPC struct {
	CreateUserFunc func(ctx context.Context, req *models.AdminCreateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error)
	GetUserFunc    func(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error)
	ListUsersFunc  func(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminUserListResponse, error)
	SyncUsersFunc  func(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error)
	ImportUsersFunc func(ctx context.Context, req *models.BulkImportUsersRequest, appID *uuid.UUID) (*models.ImportUsersResponse, error)
}

func (m *mockAdminServicerGRPC) ListUsers(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminUserListResponse, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, appID, page, pageSize)
	}
	return nil, nil
}
func (m *mockAdminServicerGRPC) GetUser(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockAdminServicerGRPC) CreateUser(ctx context.Context, req *models.AdminCreateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, req, adminID)
	}
	return nil, nil
}
func (m *mockAdminServicerGRPC) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.AdminUpdateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) DeleteUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *mockAdminServicerGRPC) AdminReset2FA(ctx context.Context, userID, adminID uuid.UUID) error {
	return nil
}
func (m *mockAdminServicerGRPC) GetUserOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) AssignRole(ctx context.Context, userID, roleID, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*models.AdminUserResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) ListAPIKeys(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminAPIKeyListResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return nil
}
func (m *mockAdminServicerGRPC) ListAuditLogs(ctx context.Context, page, pageSize int, userID *uuid.UUID) (*models.AuditLogListResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) GetStats(ctx context.Context) (*models.AdminStatsResponse, error) {
	return nil, nil
}
func (m *mockAdminServicerGRPC) SyncUsers(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error) {
	if m.SyncUsersFunc != nil {
		return m.SyncUsersFunc(ctx, updatedAfter, appID, limit, offset)
	}
	return nil, nil
}
func (m *mockAdminServicerGRPC) ImportUsers(ctx context.Context, req *models.BulkImportUsersRequest, appID *uuid.UUID) (*models.ImportUsersResponse, error) {
	if m.ImportUsersFunc != nil {
		return m.ImportUsersFunc(ctx, req, appID)
	}
	return nil, nil
}

// ===================== mockApplicationServicerGRPC =====================

type mockApplicationServicerGRPC struct {
	GetByIDFunc              func(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetAuthConfigFunc        func(ctx context.Context, app *models.Application) (*models.AuthConfigResponse, error)
	GetOrCreateUserProfileFunc func(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	UpdateUserProfileFunc    func(ctx context.Context, userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error)
	ListApplicationUsersFunc func(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error)
	BanUserFunc              func(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error
	UnbanUserFunc            func(ctx context.Context, userID, applicationID uuid.UUID) error
	DeleteUserProfileFunc    func(ctx context.Context, userID, applicationID uuid.UUID) error
}

func (m *mockApplicationServicerGRPC) CreateApplication(ctx context.Context, req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
	return nil, "", nil
}
func (m *mockApplicationServicerGRPC) GetByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockApplicationServicerGRPC) GetByName(ctx context.Context, name string) (*models.Application, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) UpdateApplication(ctx context.Context, id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockApplicationServicerGRPC) ListApplications(ctx context.Context, page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) UpdateBranding(ctx context.Context, applicationID uuid.UUID, req *models.UpdateApplicationBrandingRequest) (*models.ApplicationBranding, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) GetOrCreateUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	if m.GetOrCreateUserProfileFunc != nil {
		return m.GetOrCreateUserProfileFunc(ctx, userID, applicationID)
	}
	return nil, nil
}
func (m *mockApplicationServicerGRPC) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) UpdateUserProfile(ctx context.Context, userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error) {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(ctx, userID, applicationID, req)
	}
	return nil, nil
}
func (m *mockApplicationServicerGRPC) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error) {
	if m.ListApplicationUsersFunc != nil {
		return m.ListApplicationUsersFunc(ctx, applicationID, page, perPage)
	}
	return nil, nil
}
func (m *mockApplicationServicerGRPC) BanUser(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	if m.BanUserFunc != nil {
		return m.BanUserFunc(ctx, userID, applicationID, bannedBy, reason)
	}
	return nil
}
func (m *mockApplicationServicerGRPC) UnbanUser(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.UnbanUserFunc != nil {
		return m.UnbanUserFunc(ctx, userID, applicationID)
	}
	return nil
}
func (m *mockApplicationServicerGRPC) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.DeleteUserProfileFunc != nil {
		return m.DeleteUserProfileFunc(ctx, userID, applicationID)
	}
	return nil
}
func (m *mockApplicationServicerGRPC) CheckUserAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	return nil
}
func (m *mockApplicationServicerGRPC) IsAuthMethodAllowed(ctx context.Context, appID uuid.UUID, method string) error {
	return nil
}
func (m *mockApplicationServicerGRPC) GenerateSecret(ctx context.Context, appID uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockApplicationServicerGRPC) RotateSecret(ctx context.Context, appID uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockApplicationServicerGRPC) ValidateSecret(ctx context.Context, secret string) (*models.Application, error) {
	return nil, nil
}
func (m *mockApplicationServicerGRPC) GetAuthConfig(ctx context.Context, app *models.Application) (*models.AuthConfigResponse, error) {
	if m.GetAuthConfigFunc != nil {
		return m.GetAuthConfigFunc(ctx, app)
	}
	return nil, nil
}

// ===================== mockRedisServicerGRPC =====================

type mockRedisServicerGRPC struct {
	IsBlacklistedFunc func(ctx context.Context, tokenHash string) (bool, error)
}

func (m *mockRedisServicerGRPC) Close() error { return nil }
func (m *mockRedisServicerGRPC) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (m *mockRedisServicerGRPC) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (m *mockRedisServicerGRPC) Delete(ctx context.Context, keys ...string) error { return nil }
func (m *mockRedisServicerGRPC) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (m *mockRedisServicerGRPC) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}
func (m *mockRedisServicerGRPC) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return false, nil
}
func (m *mockRedisServicerGRPC) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}
func (m *mockRedisServicerGRPC) Health(ctx context.Context) error { return nil }
func (m *mockRedisServicerGRPC) AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error {
	return nil
}
func (m *mockRedisServicerGRPC) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(ctx, tokenHash)
	}
	return false, nil
}
func (m *mockRedisServicerGRPC) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	return 0, nil
}
func (m *mockRedisServicerGRPC) StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
	return nil
}
func (m *mockRedisServicerGRPC) GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
	return nil, nil
}
func (m *mockRedisServicerGRPC) DeletePendingRegistration(ctx context.Context, identifier string) error {
	return nil
}
func (m *mockRedisServicerGRPC) SAdd(ctx context.Context, key string, members ...string) error {
	return nil
}
func (m *mockRedisServicerGRPC) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	return false, nil
}
func (m *mockRedisServicerGRPC) SMembers(ctx context.Context, key string) ([]string, error) {
	return nil, nil
}

// ===================== mockTokenExchangeServicerGRPC =====================

type mockTokenExchangeServicerGRPC struct {
	CreateExchangeFunc func(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error)
	RedeemExchangeFunc func(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error)
}

func (m *mockTokenExchangeServicerGRPC) CreateExchange(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
	if m.CreateExchangeFunc != nil {
		return m.CreateExchangeFunc(ctx, req, sourceAppID)
	}
	return nil, nil
}
func (m *mockTokenExchangeServicerGRPC) RedeemExchange(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
	if m.RedeemExchangeFunc != nil {
		return m.RedeemExchangeFunc(ctx, req, redeemingAppID)
	}
	return nil, nil
}

// ===================== Test Helper =====================

func newTestAuthHandlerV2(
	jwtSvc *jwt.Service,
	opts ...func(*AuthHandlerV2),
) *AuthHandlerV2 {
	h := NewAuthHandlerV2(
		jwtSvc,
		&mockUserStoreGRPC{},
		&mockTokenStoreGRPC{},
		&mockRBACStoreGRPC{},
		&mockAPIKeyServicerGRPC{},
		&mockAuthServicerGRPC{},
		&mockOAuthProviderServicerGRPC{},
		&mockOTPServicerGRPC{},
		&mockEmailProfileServicerGRPC{},
		&mockAdminServicerGRPC{},
		&mockApplicationServicerGRPC{},
		&mockRedisServicerGRPC{},
		&mockTokenExchangeServicerGRPC{},
		logger.New("test", logger.ErrorLevel, false),
	)
	for _, opt := range opts {
		opt(h)
	}
	return h
}
