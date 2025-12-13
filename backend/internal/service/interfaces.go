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
	GetByID(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error)
	GetByEmail(ctx context.Context, email string, isActive *bool) (*models.User, error)
	GetByUsername(ctx context.Context, username string, isActive *bool) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	GetByIDWithRoles(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error)
	GetByEmailWithRoles(ctx context.Context, email string, isActive *bool) (*models.User, error)
	GetByPhone(ctx context.Context, phone string, isActive *bool) (*models.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error)
	ListWithRoles(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error)
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
}

// CacheService defines the interface for caching (Redis)
type CacheService interface {
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
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
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, error)
	GenerateTwoFactorToken(user *models.User) (string, error)
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
	GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error)
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
}

// JWTService defines the interface for JWT token operations
type JWTService interface {
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, error)
	GetAccessTokenExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
	ValidateRefreshToken(tokenString string) (*jwt.Claims, error)
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
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
	Update(ctx context.Context, apiKey *models.APIKey) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	CountActive(ctx context.Context, userID uuid.UUID) (int, error)
	ListAll(ctx context.Context) ([]*models.APIKey, error)
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetUserSessionsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error)
	GetAllSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error)
	RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error
	UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error
	GetSessionStats(ctx context.Context) (*models.SessionStats, error)
	DeleteExpiredSessions(ctx context.Context, olderThan time.Duration) error
}
