package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// AuthServicer abstracts authentication operations
type AuthServicer interface {
	SignUp(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	SignIn(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	Verify2FALogin(ctx context.Context, twoFactorToken, code, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	Logout(ctx context.Context, accessToken, ip, userAgent string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error
	ResetPassword(ctx context.Context, userID uuid.UUID, newPassword, ip, userAgent string) error
	InitPasswordlessRegistration(ctx context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error
	CompletePasswordlessRegistration(ctx context.Context, req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	GenerateTokensForUser(ctx context.Context, user *models.User, ip, userAgent string) (*models.AuthResponse, error)
}

// UserServicer abstracts user profile and lookup operations
type UserServicer interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest, ip, userAgent string) (*models.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int, error)
}

// OTPServicer abstracts one-time password operations
type OTPServicer interface {
	GenerateOTPCode() (string, error)
	SendOTP(ctx context.Context, req *models.SendOTPRequest) error
	VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error)
	CleanupExpiredOTPs() error
}

// AdminUserServicer abstracts admin user management operations
type AdminUserServicer interface {
	ListUsers(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminUserListResponse, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error)
	CreateUser(ctx context.Context, req *models.AdminCreateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *models.AdminUpdateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	AdminReset2FA(ctx context.Context, userID, adminID uuid.UUID) error
	GetUserOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error)
	AssignRole(ctx context.Context, userID, roleID, adminID uuid.UUID) (*models.AdminUserResponse, error)
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*models.AdminUserResponse, error)
}

// AdminAPIKeyServicer abstracts admin API key management operations
type AdminAPIKeyServicer interface {
	ListAPIKeys(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminAPIKeyListResponse, error)
	RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error
}

// AdminAuditServicer abstracts admin audit log operations
type AdminAuditServicer interface {
	ListAuditLogs(ctx context.Context, page, pageSize int, userID *uuid.UUID) (*models.AuditLogListResponse, error)
}

// AdminStatsServicer abstracts admin statistics operations
type AdminStatsServicer interface {
	GetStats(ctx context.Context) (*models.AdminStatsResponse, error)
}

// AdminBulkServicer abstracts admin bulk import and sync operations
type AdminBulkServicer interface {
	SyncUsers(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error)
	ImportUsers(ctx context.Context, req *models.BulkImportUsersRequest, appID *uuid.UUID) (*models.ImportUsersResponse, error)
}

// AdminServicer composes all admin sub-interfaces for backward compatibility
type AdminServicer interface {
	AdminUserServicer
	AdminAPIKeyServicer
	AdminAuditServicer
	AdminStatsServicer
	AdminBulkServicer
}

// ApplicationServicer abstracts application lifecycle and profile operations
type ApplicationServicer interface {
	CreateApplication(ctx context.Context, req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	UpdateApplication(ctx context.Context, id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error)
	DeleteApplication(ctx context.Context, id uuid.UUID) error
	ListApplications(ctx context.Context, page, perPage int, isActive *bool) (*models.ApplicationListResponse, error)
	GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error)
	UpdateBranding(ctx context.Context, applicationID uuid.UUID, req *models.UpdateApplicationBrandingRequest) (*models.ApplicationBranding, error)
	GetOrCreateUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	UpdateUserProfile(ctx context.Context, userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error)
	ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error)
	ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error)
	BanUser(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error
	UnbanUser(ctx context.Context, userID, applicationID uuid.UUID) error
	DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error
	CheckUserAccess(ctx context.Context, userID, applicationID uuid.UUID) error
	IsAuthMethodAllowed(ctx context.Context, appID uuid.UUID, method string) error
	GenerateSecret(ctx context.Context, appID uuid.UUID) (string, error)
	RotateSecret(ctx context.Context, appID uuid.UUID) (string, error)
	ValidateSecret(ctx context.Context, secret string) (*models.Application, error)
	GetAuthConfig(ctx context.Context, app *models.Application) (*models.AuthConfigResponse, error)
}

// APIKeyServicer abstracts API key CRUD and validation operations
type APIKeyServicer interface {
	GenerateAPIKey() (string, error)
	Create(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error)
	ValidateAPIKey(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error)
	GetByID(ctx context.Context, userID, apiKeyID uuid.UUID) (*models.APIKey, error)
	List(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
	Update(ctx context.Context, userID, apiKeyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error)
	Revoke(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error
	Delete(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error
	HasScope(apiKey *models.APIKey, scope models.APIKeyScope) bool
}

// RBACServicer abstracts role-based access control operations
type RBACServicer interface {
	CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*models.Permission, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, req *models.UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	CheckUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	CheckUserAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	CheckUserAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error)
	GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	CreateRoleInApp(ctx context.Context, name, displayName, description string, appID *uuid.UUID) (*models.Role, error)
	ListRolesByApp(ctx context.Context, appID *uuid.UUID) ([]models.Role, error)
	ListPermissionsByApp(ctx context.Context, appID *uuid.UUID) ([]models.Permission, error)
	HasPermissionInApp(ctx context.Context, userID uuid.UUID, permissionName string, appID *uuid.UUID) (bool, error)
	GetUserRolesInApp(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) ([]models.Role, error)
	AssignRoleToUserInApp(ctx context.Context, userID, roleID, assignedBy uuid.UUID, appID *uuid.UUID) error
}

// SessionServicer abstracts session lifecycle operations
type SessionServicer interface {
	CreateSessionWithParams(ctx context.Context, params SessionCreationParams) (*models.Session, error)
	CreateSessionNonFatal(ctx context.Context, params SessionCreationParams) *models.Session
	CreateSessionFromRequest(ctx context.Context, userID uuid.UUID, accessToken string, refreshToken string, ipAddress string, userAgent string, tokenExpiration time.Duration) (*models.Session, error)
	CreateSessionFromRequestNonFatal(ctx context.Context, userID uuid.UUID, accessToken string, refreshToken string, ipAddress string, userAgent string, tokenExpiration time.Duration) *models.Session
	RefreshSession(ctx context.Context, params SessionRefreshParams) error
	RefreshSessionNonFatal(ctx context.Context, params SessionRefreshParams) bool
	RefreshSessionFromTokens(ctx context.Context, oldRefreshToken string, newRefreshToken string, newAccessToken string, newExpiresAt time.Time) error
	RefreshSessionFromTokensNonFatal(ctx context.Context, oldRefreshToken string, newRefreshToken string, newAccessToken string, newExpiresAt time.Time) bool
	GetUserSessions(ctx context.Context, userID uuid.UUID, page, perPage int) (*models.SessionListResponse, error)
	GetAllSessions(ctx context.Context, page, perPage int) (*models.SessionListResponse, error)
	GetSessionStats(ctx context.Context) (*models.SessionStats, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	AdminRevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error
	RevokeSessionByToken(ctx context.Context, token string) error
	UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error
	CleanupExpiredSessions(ctx context.Context) error
	GetUserSessionsByApp(ctx context.Context, userID, appID uuid.UUID) ([]models.Session, error)
	GetAppSessionsPaginated(ctx context.Context, appID uuid.UUID, page, perPage int) ([]models.Session, int, error)
}

// BlacklistServicer abstracts token blacklist operations
type BlacklistServicer interface {
	SyncFromDatabase(ctx context.Context) error
	GetSyncStats() SyncStats
	IsBlacklisted(ctx context.Context, tokenHash string) bool
	AddToBlacklist(ctx context.Context, tokenHash string, userID *uuid.UUID, ttl time.Duration) error
	AddAccessToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error
	AddRefreshToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error
	BlacklistSessionTokens(ctx context.Context, session *models.Session) error
	BlacklistAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

// RedisServicer abstracts Redis cache operations
type RedisServicer interface {
	Close() error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Health(ctx context.Context) error
	AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
	StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error
	GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error)
	DeletePendingRegistration(ctx context.Context, identifier string) error
	SAdd(ctx context.Context, key string, members ...string) error
	SIsMember(ctx context.Context, key string, member string) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)
}

// OAuthServicer abstracts OAuth consumer operations (Google, GitHub, etc.)
type OAuthServicer interface {
	GenerateState() (string, error)
	GetAuthURL(ctx context.Context, provider models.OAuthProvider, state string, appID *uuid.UUID) (string, error)
	ExchangeCode(ctx context.Context, provider models.OAuthProvider, code string, appID *uuid.UUID) (*OAuthTokenResponse, error)
	GetUserInfo(ctx context.Context, provider models.OAuthProvider, accessToken string, appID *uuid.UUID) (*models.OAuthUserInfo, error)
	HandleCallback(ctx context.Context, provider models.OAuthProvider, code, ipAddress, userAgent string, appID *uuid.UUID) (*models.OAuthLoginResponse, error)
}

// OAuthProviderServicer abstracts OAuth/OIDC provider operations
type OAuthProviderServicer interface {
	CreateClient(ctx context.Context, req *models.CreateOAuthClientRequest, ownerID *uuid.UUID) (*models.CreateOAuthClientResponse, error)
	GetClient(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error)
	GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error)
	UpdateClient(ctx context.Context, id uuid.UUID, req *models.UpdateOAuthClientRequest) (*models.OAuthClient, error)
	DeleteClient(ctx context.Context, id uuid.UUID) error
	ListClients(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error)
	RotateClientSecret(ctx context.Context, id uuid.UUID) (string, error)
	ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error)
	Authorize(ctx context.Context, req *models.AuthorizeRequest, userID uuid.UUID) (*models.AuthorizeResponse, error)
	ExchangeCode(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error)
	ClientCredentialsGrant(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error)
	RefreshToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error)
	DeviceAuthorization(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	PollDeviceToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error)
	ApproveDeviceCode(ctx context.Context, userID uuid.UUID, userCode string, approve bool) error
	IntrospectToken(ctx context.Context, token, tokenTypeHint string, clientID *string) (*models.IntrospectionResponse, error)
	RevokeToken(ctx context.Context, token, tokenTypeHint string, clientID *string) error
	GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfoResponse, error)
	GetDiscoveryDocument() *models.OIDCDiscoveryDocument
	GetJWKS() *models.JWKSDocument
	GetConsentInfo(ctx context.Context, clientID string, scopes []string) (*ConsentInfo, error)
	GrantConsent(ctx context.Context, userID uuid.UUID, clientID string, scopes []string) error
	RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error
	ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error)
	ListScopes(ctx context.Context) ([]*models.OAuthScope, error)
	CreateScope(ctx context.Context, scope *models.OAuthScope) error
	DeleteScope(ctx context.Context, id uuid.UUID) error
	ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error)
}

// TwoFactorServicer abstracts TOTP two-factor authentication operations
type TwoFactorServicer interface {
	SetupTOTP(ctx context.Context, userID uuid.UUID, password string) (*models.TwoFactorSetupResponse, error)
	VerifyTOTPSetup(ctx context.Context, userID uuid.UUID, code string) error
	VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	DisableTOTP(ctx context.Context, userID uuid.UUID, password, code string) error
	GetStatus(ctx context.Context, userID uuid.UUID) (*models.TwoFactorStatusResponse, error)
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, password string) ([]string, error)
}

// IPFilterServicer abstracts IP filtering operations
type IPFilterServicer interface {
	CreateIPFilter(ctx context.Context, req *models.CreateIPFilterRequest, createdBy uuid.UUID) (*models.IPFilter, error)
	GetIPFilter(ctx context.Context, id uuid.UUID) (*models.IPFilter, error)
	ListIPFilters(ctx context.Context, page, perPage int, filterType string) (*models.IPFilterListResponse, error)
	UpdateIPFilter(ctx context.Context, id uuid.UUID, req *models.UpdateIPFilterRequest) error
	DeleteIPFilter(ctx context.Context, id uuid.UUID) error
	CheckIPAllowed(ctx context.Context, ipAddress string) (*models.CheckIPResponse, error)
}

// AuditServicer abstracts audit logging operations
type AuditServicer interface {
	Log(params AuditLogParams)
	LogSync(ctx context.Context, params AuditLogParams) error
	LogWithAction(userID *uuid.UUID, action string, status string, ip, userAgent string, details map[string]interface{})
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)
	Count(ctx context.Context) (int, error)
	CountByActionSince(ctx context.Context, action models.AuditAction, since time.Time) (int, error)
	DeleteOlderThan(ctx context.Context, days int) error
	ListByApp(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error)
}

// EmailProfileServicer abstracts email provider and profile operations
type EmailProfileServicer interface {
	CreateProvider(ctx context.Context, req *models.CreateEmailProviderRequest) (*models.EmailProvider, error)
	GetProvider(ctx context.Context, id uuid.UUID) (*models.EmailProviderResponse, error)
	ListProviders(ctx context.Context, appID *uuid.UUID) ([]*models.EmailProviderResponse, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProviderRequest) error
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	TestProvider(ctx context.Context, id uuid.UUID, testEmail string) error
	CreateProfile(ctx context.Context, req *models.CreateEmailProfileRequest) (*models.EmailProfile, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*models.EmailProfile, error)
	ListProfiles(ctx context.Context, appID *uuid.UUID) ([]*models.EmailProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req *models.UpdateEmailProfileRequest) error
	DeleteProfile(ctx context.Context, id uuid.UUID) error
	SetDefaultProfile(ctx context.Context, id uuid.UUID) error
	GetProfileTemplates(ctx context.Context, profileID uuid.UUID) ([]*models.EmailProfileTemplate, error)
	SetProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string, templateID uuid.UUID) error
	RemoveProfileTemplate(ctx context.Context, profileID uuid.UUID, otpType string) error
	SendOTPEmail(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error
	SendEmail(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error
	GetProfileStats(ctx context.Context, profileID uuid.UUID) (*models.EmailStatsResponse, error)
	TestProfile(ctx context.Context, profileID uuid.UUID, testEmail string) error
}

// WebhookServicer abstracts webhook management operations
type WebhookServicer interface {
	CreateWebhook(ctx context.Context, req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error)
	GetWebhook(ctx context.Context, id uuid.UUID) (*models.Webhook, error)
	ListWebhooks(ctx context.Context, page, perPage int) (*models.WebhookListResponse, error)
	UpdateWebhook(ctx context.Context, id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error
	DeleteWebhook(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	TriggerWebhook(ctx context.Context, eventType string, data map[string]interface{}) error
	ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error)
	TestWebhook(ctx context.Context, id uuid.UUID, req *models.TestWebhookRequest) error
	GetAvailableEvents() []string
	ListWebhooksByApp(ctx context.Context, appID uuid.UUID) ([]*models.Webhook, error)
}

// TokenExchangeServicer abstracts cross-application token exchange operations
type TokenExchangeServicer interface {
	CreateExchange(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error)
	RedeemExchange(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error)
}

// TelegramServicer abstracts Telegram bot and account operations
type TelegramServicer interface {
	CreateBot(ctx context.Context, appID uuid.UUID, req *models.CreateTelegramBotRequest) (*models.TelegramBot, error)
	GetBot(ctx context.Context, id uuid.UUID) (*models.TelegramBot, error)
	ListBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error)
	ListAuthBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error)
	UpdateBot(ctx context.Context, id uuid.UUID, req *models.UpdateTelegramBotRequest) (*models.TelegramBot, error)
	DeleteBot(ctx context.Context, id uuid.UUID) error
	GetOrCreateAccount(ctx context.Context, userID uuid.UUID, telegramUserID int64, username, firstName, lastName, photoURL string, authDate time.Time) (*models.UserTelegramAccount, error)
	ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramAccount, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	GrantBotAccess(ctx context.Context, userID, botID, accountID uuid.UUID) (*models.UserTelegramBotAccess, error)
	ListBotAccessByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramBotAccess, error)
	ListBotAccessByUserAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.UserTelegramBotAccess, error)
	UpdateBotAccess(ctx context.Context, id uuid.UUID, canSendMessages bool) error
	RevokeBotAccess(ctx context.Context, id uuid.UUID) error
	VerifyTelegramAuth(botToken string, data map[string]string) bool
}

// SMSServicer abstracts SMS OTP operations
type SMSServicer interface {
	GenerateOTPCode() (string, error)
	SendOTP(ctx context.Context, req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error)
	VerifyOTP(ctx context.Context, req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error)
	GetStats(ctx context.Context) (*models.SMSStatsResponse, error)
	CleanupOldLogs(ctx context.Context, duration time.Duration) (int64, error)
}

// SCIMServicer abstracts SCIM 2.0 provisioning operations
type SCIMServicer interface {
	GetUsers(ctx context.Context, filter string, startIndex, count int) (*models.SCIMListResponse, error)
	GetUser(ctx context.Context, id string) (*models.SCIMUser, error)
	CreateUser(ctx context.Context, scimUser *models.SCIMUser) (*models.SCIMUser, error)
	UpdateUser(ctx context.Context, id string, scimUser *models.SCIMUser) (*models.SCIMUser, error)
	PatchUser(ctx context.Context, id string, patchReq *models.SCIMPatchRequest) (*models.SCIMUser, error)
	DeleteUser(ctx context.Context, id string) error
	GetGroups(ctx context.Context, filter string, startIndex, count int) (*models.SCIMListResponse, error)
	GetGroup(ctx context.Context, id string) (*models.SCIMGroup, error)
	GetServiceProviderConfig(ctx context.Context) *models.SCIMServiceProviderConfig
	GetSchemas(ctx context.Context) []*models.SCIMSchema
}

// BulkServicer abstracts bulk user operations
type BulkServicer interface {
	BulkCreateUsers(ctx context.Context, req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error)
	BulkUpdateUsers(ctx context.Context, req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error)
	BulkDeleteUsers(ctx context.Context, req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error)
	BulkAssignRoles(ctx context.Context, req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error)
}

// GroupServicer abstracts user group operations
type GroupServicer interface {
	CreateGroup(ctx context.Context, req *models.CreateGroupRequest) (*models.Group, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*models.Group, error)
	ListGroups(ctx context.Context, page, pageSize int) (*models.GroupListResponse, error)
	UpdateGroup(ctx context.Context, id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	AddGroupMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error
	RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error)
	GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*models.Group, error)
	GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int, error)
	EvaluateDynamicGroupMembers(ctx context.Context, group *models.Group, allUsers []*models.User) []uuid.UUID
	GetGroupPermissions(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
	SyncDynamicGroupMembers(ctx context.Context, groupID uuid.UUID) error
}

// LDAPServicer abstracts LDAP integration operations
type LDAPServicer interface {
	CreateConfig(ctx context.Context, req *models.CreateLDAPConfigRequest) (*models.LDAPConfig, error)
	GetConfig(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error)
	GetActiveConfig(ctx context.Context) (*models.LDAPConfig, error)
	ListConfigs(ctx context.Context) ([]*models.LDAPConfig, error)
	UpdateConfig(ctx context.Context, id uuid.UUID, req *models.UpdateLDAPConfigRequest) (*models.LDAPConfig, error)
	DeleteConfig(ctx context.Context, id uuid.UUID) error
	TestConnection(ctx context.Context, req *models.LDAPTestConnectionRequest) (*models.LDAPTestConnectionResponse, error)
	Sync(ctx context.Context, configID uuid.UUID, req *models.LDAPSyncRequest) (*models.LDAPSyncResponse, error)
	GetSyncLogs(ctx context.Context, configID uuid.UUID, page, pageSize int) ([]*models.LDAPSyncLog, int, error)
}

// SAMLServicer abstracts SAML identity provider operations
type SAMLServicer interface {
	CreateSP(ctx context.Context, req *models.CreateSAMLSPRequest) (*models.SAMLServiceProvider, error)
	GetSP(ctx context.Context, id uuid.UUID) (*models.SAMLServiceProvider, error)
	ListSPs(ctx context.Context, page, pageSize int) ([]*models.SAMLServiceProvider, int, error)
	UpdateSP(ctx context.Context, id uuid.UUID, req *models.UpdateSAMLSPRequest) (*models.SAMLServiceProvider, error)
	DeleteSP(ctx context.Context, id uuid.UUID) error
	GetSPByEntityID(ctx context.Context, entityID string) (*models.SAMLServiceProvider, error)
	GetMetadata() (*SAMLMetadata, error)
	CreateAssertion(ctx context.Context, userID uuid.UUID, sp *models.SAMLServiceProvider) (*SAMLResponse, error)
}

// MigrationServicer abstracts data migration operations
type MigrationServicer interface {
	ImportUsers(ctx context.Context, appID uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error)
	ImportOAuthAccounts(ctx context.Context, entries []models.ImportOAuthEntry) (*models.ImportResult, error)
	ImportRoles(ctx context.Context, appID uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error)
}

// TemplateServicer abstracts email template operations
type TemplateServicer interface {
	CreateEmailTemplate(ctx context.Context, req *models.CreateEmailTemplateRequest, createdBy uuid.UUID) (*models.EmailTemplate, error)
	GetEmailTemplate(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error)
	GetEmailTemplateByType(ctx context.Context, templateType string) (*models.EmailTemplate, error)
	ListEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error)
	UpdateEmailTemplate(ctx context.Context, id uuid.UUID, req *models.UpdateEmailTemplateRequest, updatedBy uuid.UUID) error
	DeleteEmailTemplate(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	PreviewEmailTemplate(ctx context.Context, req *models.PreviewEmailTemplateRequest) (*models.PreviewEmailTemplateResponse, error)
	RenderTemplate(ctx context.Context, templateType string, variables map[string]interface{}) (subject, htmlBody, textBody string, err error)
	GetAvailableTemplateTypes() []string
	GetDefaultVariables(templateType string) []string
	CreateEmailTemplateForApp(ctx context.Context, applicationID uuid.UUID, req *models.CreateEmailTemplateRequest, createdBy uuid.UUID) (*models.EmailTemplate, error)
	ListEmailTemplatesForApp(ctx context.Context, applicationID uuid.UUID) ([]models.EmailTemplate, error)
	GetEmailTemplateByTypeAndApp(ctx context.Context, templateType string, applicationID uuid.UUID) (*models.EmailTemplate, error)
	UpdateEmailTemplateForApp(ctx context.Context, applicationID, templateID uuid.UUID, req *models.UpdateEmailTemplateRequest, updatedBy uuid.UUID) error
	RenderTemplateForApp(ctx context.Context, templateType string, applicationID uuid.UUID, variables map[string]interface{}) (subject, htmlBody, textBody string, err error)
	InitializeTemplatesForApp(ctx context.Context, applicationID uuid.UUID, createdBy uuid.UUID) error
}

// AppOAuthProviderServicer abstracts per-application OAuth provider configuration
type AppOAuthProviderServicer interface {
	Create(ctx context.Context, appID uuid.UUID, req *models.CreateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error)
	GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error)
	ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GeoServicer abstracts geo-location lookup operations
type GeoServicer interface {
	GetLocation(ctx context.Context, ip string) *models.GeoLocation
	GetLocationAsync(ip string, callback func(*models.GeoLocation))
}

// BrandingRepositoryInterface abstracts branding settings storage for handler/middleware layer
type BrandingRepositoryInterface interface {
	GetBrandingSettings(ctx context.Context) (*models.BrandingSettings, error)
	UpdateBrandingSettings(ctx context.Context, settings *models.BrandingSettings, updatedBy uuid.UUID) error
}

// SystemRepositoryInterface abstracts system settings and health metrics storage for handler/middleware layer
type SystemRepositoryInterface interface {
	GetSetting(ctx context.Context, key string) (*models.SystemSetting, error)
	GetAllSettings(ctx context.Context) ([]models.SystemSetting, error)
	GetPublicSettings(ctx context.Context) ([]models.SystemSetting, error)
	UpdateSetting(ctx context.Context, key, value string, updatedBy *uuid.UUID) error
	CreateSetting(ctx context.Context, setting *models.SystemSetting) error
	DeleteSetting(ctx context.Context, key string) error
	RecordHealthMetric(ctx context.Context, metric *models.HealthMetric) error
	GetRecentMetrics(ctx context.Context, metricName string, limit int) ([]models.HealthMetric, error)
	DeleteOldMetrics(ctx context.Context, olderThanDays int) error
}

// GeoRepositoryInterface abstracts geo-location analytics storage for handler/middleware layer
type GeoRepositoryInterface interface {
	GetLoginGeoDistribution(ctx context.Context, days int) ([]models.LoginLocation, error)
	GetTopCountries(ctx context.Context, limit, days int) ([]models.CountryStats, error)
	GetTopCities(ctx context.Context, limit, days int) ([]models.CityStats, error)
	UpdateOrCreateLoginLocation(ctx context.Context, location *models.LoginLocation) error
}

// SMSSettingsRepositoryInterface abstracts SMS settings storage for handler/middleware layer
type SMSSettingsRepositoryInterface interface {
	Create(ctx context.Context, settings *models.SMSSettings) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SMSSettings, error)
	GetActive(ctx context.Context) (*models.SMSSettings, error)
	GetAll(ctx context.Context) ([]*models.SMSSettings, error)
	Update(ctx context.Context, id uuid.UUID, settings *models.SMSSettings) error
	Delete(ctx context.Context, id uuid.UUID) error
	DisableAll(ctx context.Context) error
}
