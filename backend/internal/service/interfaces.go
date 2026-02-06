package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
)

// UserStore defines the interface for user storage
type UserStore interface {
	GetByID(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error)
	GetByEmail(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error)
	GetByUsername(ctx context.Context, username string, isActive *bool, opts ...UserGetOption) (*models.User, error)
	GetByPhone(ctx context.Context, phone string, isActive *bool, opts ...UserGetOption) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context, opts ...UserListOption) ([]*models.User, error)
	Count(ctx context.Context, isActive *bool) (int, error)
	// 2FA methods
	UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error
	EnableTOTP(ctx context.Context, userID uuid.UUID) error
	DisableTOTP(ctx context.Context, userID uuid.UUID) error
}

// TokenStore defines the interface for token storage
type TokenStore interface {
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	AddToBlacklist(ctx context.Context, token *models.TokenBlacklist) error
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	// GetAllActiveBlacklistEntries is optional - only implemented by concrete repository
	// Used for synchronization, so we use type assertion instead of interface
}

// BackupCodeStore defines the interface for backup code storage
type BackupCodeStore interface {
	CreateBatch(ctx context.Context, codes []*models.BackupCode) error
	GetUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*models.BackupCode, error)
	CountUnusedByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// RBACStore defines the interface for RBAC storage
type RBACStore interface {
	// Permission Methods
	CreatePermission(ctx context.Context, permission *models.Permission) error
	GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, description string) error
	DeletePermission(ctx context.Context, id uuid.UUID) error

	// Role Methods
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// User-Role Methods
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error
	GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error)

	// Permission Checking Methods
	HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)
	HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
	HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error)
	GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error)

	// Application-scoped Methods
	GetRoleByNameAndApp(ctx context.Context, name string, appID *uuid.UUID) (*models.Role, error)
	ListRolesByApp(ctx context.Context, appID *uuid.UUID) ([]models.Role, error)
	ListPermissionsByApp(ctx context.Context, appID *uuid.UUID) ([]models.Permission, error)
	HasPermissionInApp(ctx context.Context, userID uuid.UUID, permissionName string, appID *uuid.UUID) (bool, error)
	GetUserRolesInApp(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) ([]models.Role, error)
	AssignRoleToUserInApp(ctx context.Context, userID, roleID, assignedBy uuid.UUID, appID *uuid.UUID) error
}

// CacheService defines the interface for caching (Redis)
type CacheService interface {
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
	// Pending registration methods
	StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error
	GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error)
	DeletePendingRegistration(ctx context.Context, identifier string) error
}

// SMSLogStore defines the interface for SMS log storage
type SMSLogStore interface {
	Create(ctx context.Context, log *models.SMSLog) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error
	GetStats(ctx context.Context) (*models.SMSStatsResponse, error)
	DeleteOlderThan(ctx context.Context, duration time.Duration) (int64, error)
}

// TokenService defines the interface for JWT operations
type TokenService interface {
	GenerateAccessToken(user *models.User, applicationID ...*uuid.UUID) (string, error)
	GenerateRefreshToken(user *models.User, applicationID ...*uuid.UUID) (string, error)
	GenerateTwoFactorToken(user *models.User, applicationID ...*uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
	ValidateRefreshToken(tokenString string) (*jwt.Claims, error)
	ExtractClaims(tokenString string) (*jwt.Claims, error)
	GetAccessTokenExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
}

// AuditLogger defines the interface for auditing
type AuditLogger interface {
	LogWithAction(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{})
	Log(params AuditLogParams)
}

// OTPStore defines the interface for OTP storage
type OTPStore interface {
	Create(ctx context.Context, otp *models.OTP) error
	GetByEmailAndType(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	InvalidateAllForEmail(ctx context.Context, email string, otpType models.OTPType) error
	CountRecentByEmail(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error)
	DeleteExpired(ctx context.Context, olderThan time.Duration) error

	// Phone methods
	GetByPhoneAndType(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error)
	InvalidateAllForPhone(ctx context.Context, phone string, otpType models.OTPType) error
	CountRecentByPhone(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error)
}

// EmailSender defines the interface for email sending
type EmailSender interface {
	SendOTP(to, code, otpType string) error
	SendWelcome(to, username string) error
	Send(to, subject, htmlBody string) error
}

// IPFilterStore defines the interface for IP filter storage
type IPFilterStore interface {
	CreateIPFilter(ctx context.Context, filter *models.IPFilter) error
	GetIPFilterByID(ctx context.Context, id uuid.UUID) (*models.IPFilter, error)
	ListIPFilters(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error)
	UpdateIPFilter(ctx context.Context, id uuid.UUID, reason string, isActive bool) error
	DeleteIPFilter(ctx context.Context, id uuid.UUID) error
	GetActiveIPFilters(ctx context.Context) ([]models.IPFilter, error)
}

// OAuthStore defines the interface for OAuth account storage
type OAuthStore interface {
	CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error
	GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error)
	UpdateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error
	DeleteOAuthAccount(ctx context.Context, id uuid.UUID) error
	DeleteOAuthAccountsByProvider(ctx context.Context, userID uuid.UUID, provider string) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error)
	ListAll(ctx context.Context) ([]*models.OAuthAccount, error)
}

// AuditStore defines the interface for audit log storage
type AuditStore interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error)
	GetFailedLoginAttempts(ctx context.Context, ipAddress string, limit int) ([]*models.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)
	Count(ctx context.Context) (int, error)
	DeleteOlderThan(ctx context.Context, days int) error
	CountByActionSince(ctx context.Context, action models.AuditAction, since time.Time) (int, error)
	ListByApp(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error)
}


// HTTPClient defines the interface for HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// APIKeyStore defines the interface for API key storage
type APIKeyStore interface {
	Create(ctx context.Context, apiKey *models.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, opts ...APIKeyGetOption) ([]*models.APIKey, error)
	Update(ctx context.Context, apiKey *models.APIKey) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	Count(ctx context.Context, userID uuid.UUID, opts ...APIKeyGetOption) (int, error)
	ListAll(ctx context.Context) ([]*models.APIKey, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.APIKey, error)
	GetByUserIDAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.APIKey, error)
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
	GetUserSessionsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error)
	GetAllSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error
	UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error
	UpdateSessionAccessTokenHash(ctx context.Context, sessionID uuid.UUID, accessTokenHash string) error
	RefreshSessionTokens(ctx context.Context, oldTokenHash, newTokenHash, newAccessTokenHash string, newExpiresAt time.Time) error
	GetSessionStats(ctx context.Context) (*models.SessionStats, error)
	DeleteExpiredSessions(ctx context.Context, olderThan time.Duration) error
	GetUserSessionsByApp(ctx context.Context, userID, appID uuid.UUID) ([]models.Session, error)
	GetAppSessionsPaginated(ctx context.Context, appID uuid.UUID, page, perPage int) ([]models.Session, int, error)
}

// OAuthProviderStore defines the interface for OAuth provider operations
type OAuthProviderStore interface {
	// Client operations
	CreateClient(ctx context.Context, client *models.OAuthClient) error
	GetClientByID(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error)
	GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error)
	UpdateClient(ctx context.Context, client *models.OAuthClient) error
	DeleteClient(ctx context.Context, id uuid.UUID) error
	HardDeleteClient(ctx context.Context, id uuid.UUID) error
	ListClients(ctx context.Context, page, perPage int, opts ...OAuthClientListOption) ([]*models.OAuthClient, int, error)
	ListActiveClients(ctx context.Context) ([]*models.OAuthClient, error)

	// Authorization code operations
	CreateAuthorizationCode(ctx context.Context, code *models.AuthorizationCode) error
	GetAuthorizationCode(ctx context.Context, codeHash string) (*models.AuthorizationCode, error)
	MarkAuthorizationCodeUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredAuthorizationCodes(ctx context.Context) (int64, error)

	// Access token operations
	CreateAccessToken(ctx context.Context, token *models.OAuthAccessToken) error
	GetAccessToken(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error)
	GetAccessTokenByID(ctx context.Context, id uuid.UUID) (*models.OAuthAccessToken, error)
	RevokeAccessToken(ctx context.Context, tokenHash string) error
	RevokeAllUserAccessTokens(ctx context.Context, userID, clientID uuid.UUID) error
	RevokeAllClientAccessTokens(ctx context.Context, clientID uuid.UUID) error
	DeleteExpiredAccessTokens(ctx context.Context) (int64, error)

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, token *models.OAuthRefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID, clientID uuid.UUID) error
	RevokeAllClientRefreshTokens(ctx context.Context, clientID uuid.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)

	// User consent operations
	CreateOrUpdateConsent(ctx context.Context, consent *models.UserConsent) error
	GetUserConsent(ctx context.Context, userID, clientID uuid.UUID) (*models.UserConsent, error)
	RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error
	ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error)
	ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error)

	// Device code operations (RFC 8628)
	CreateDeviceCode(ctx context.Context, code *models.DeviceCode) error
	GetDeviceCode(ctx context.Context, deviceCodeHash string) (*models.DeviceCode, error)
	GetDeviceCodeByUserCode(ctx context.Context, userCode string) (*models.DeviceCode, error)
	UpdateDeviceCodeStatus(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, userID *uuid.UUID) error
	DeleteExpiredDeviceCodes(ctx context.Context) (int64, error)

	// Scope operations
	CreateScope(ctx context.Context, scope *models.OAuthScope) error
	GetScopeByName(ctx context.Context, name string) (*models.OAuthScope, error)
	ListScopes(ctx context.Context) ([]*models.OAuthScope, error)
	ListSystemScopes(ctx context.Context) ([]*models.OAuthScope, error)
	DeleteScope(ctx context.Context, id uuid.UUID) error
}

type BlackListStore interface {
	IsBlacklisted(ctx context.Context, tokenHash string) bool
	AddToBlacklist(ctx context.Context, tokenHash string, userID *uuid.UUID, ttl time.Duration) error
	AddAccessToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error
	AddRefreshToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error
	BlacklistSessionTokens(ctx context.Context, session *models.Session) error
}

// ApplicationStore defines the interface for application storage operations
type ApplicationStore interface {
	// Application CRUD
	CreateApplication(ctx context.Context, app *models.Application) error
	GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error)
	GetApplicationByName(ctx context.Context, name string) (*models.Application, error)
	UpdateApplication(ctx context.Context, app *models.Application) error
	DeleteApplication(ctx context.Context, id uuid.UUID) error
	ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error)

	// Application Branding
	GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error)
	CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error

	// User Application Profile
	CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error
	GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error
	DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error
	ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error)
	ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error)
	UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error

	// User banning
	BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error
	UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error
}

// AppOAuthProviderStore - per-app OAuth provider configuration
type AppOAuthProviderStore interface {
	Create(ctx context.Context, provider *models.ApplicationOAuthProvider) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error)
	GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error)
	ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error)
	Update(ctx context.Context, provider *models.ApplicationOAuthProvider) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TelegramBotStore - Telegram bots per application
type TelegramBotStore interface {
	Create(ctx context.Context, bot *models.TelegramBot) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.TelegramBot, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error)
	ListAuthBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error)
	Update(ctx context.Context, bot *models.TelegramBot) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserTelegramStore - user Telegram accounts and bot access
type UserTelegramStore interface {
	// Accounts
	CreateAccount(ctx context.Context, account *models.UserTelegramAccount) error
	GetAccountByUserAndTgID(ctx context.Context, userID uuid.UUID, telegramUserID int64) (*models.UserTelegramAccount, error)
	GetAccountByTgID(ctx context.Context, telegramUserID int64) (*models.UserTelegramAccount, error)
	ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramAccount, error)
	UpdateAccount(ctx context.Context, account *models.UserTelegramAccount) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	// Bot Access
	CreateBotAccess(ctx context.Context, access *models.UserTelegramBotAccess) error
	GetBotAccess(ctx context.Context, userID, botID uuid.UUID) (*models.UserTelegramBotAccess, error)
	ListBotAccessByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramBotAccess, error)
	ListBotAccessByUserAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.UserTelegramBotAccess, error)
	UpdateBotAccess(ctx context.Context, access *models.UserTelegramBotAccess) error
	DeleteBotAccess(ctx context.Context, id uuid.UUID) error
}
