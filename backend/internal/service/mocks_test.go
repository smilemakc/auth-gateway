package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
)

// Manual mocks

type mockUserStore struct {
	GetByIDFunc             func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error)
	GetByEmailFunc          func(ctx context.Context, email string, isActive *bool) (*models.User, error)
	GetByUsernameFunc       func(ctx context.Context, username string, isActive *bool) (*models.User, error)
	CreateFunc              func(ctx context.Context, user *models.User) error
	UpdateFunc              func(ctx context.Context, user *models.User) error
	UpdatePasswordFunc      func(ctx context.Context, userID uuid.UUID, passwordHash string) error
	EmailExistsFunc         func(ctx context.Context, email string) (bool, error)
	UsernameExistsFunc      func(ctx context.Context, username string) (bool, error)
	PhoneExistsFunc         func(ctx context.Context, phone string) (bool, error)
	GetByIDWithRolesFunc    func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error)
	GetByEmailWithRolesFunc func(ctx context.Context, email string, isActive *bool) (*models.User, error)
	GetByPhoneFunc          func(ctx context.Context, phone string, isActive *bool) (*models.User, error)
	MarkEmailVerifiedFunc   func(ctx context.Context, userID uuid.UUID) error
	MarkPhoneVerifiedFunc   func(ctx context.Context, userID uuid.UUID) error
	ListFunc                func(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error)
	ListWithRolesFunc       func(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error)
	CountFunc               func(ctx context.Context, isActive *bool) (int, error)

	// 2FA methods
	UpdateTOTPSecretFunc func(ctx context.Context, userID uuid.UUID, secret string) error
	EnableTOTPFunc       func(ctx context.Context, userID uuid.UUID) error
	DisableTOTPFunc      func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockUserStore) GetByID(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) GetByEmail(ctx context.Context, email string, isActive *bool) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) GetByUsername(ctx context.Context, username string, isActive *bool) (*models.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}
func (m *mockUserStore) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}
func (m *mockUserStore) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, passwordHash)
	}
	return nil
}
func (m *mockUserStore) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(ctx, email)
	}
	return false, nil
}
func (m *mockUserStore) UsernameExists(ctx context.Context, username string) (bool, error) {
	if m.UsernameExistsFunc != nil {
		return m.UsernameExistsFunc(ctx, username)
	}
	return false, nil
}
func (m *mockUserStore) PhoneExists(ctx context.Context, phone string) (bool, error) {
	if m.PhoneExistsFunc != nil {
		return m.PhoneExistsFunc(ctx, phone)
	}
	return false, nil
}
func (m *mockUserStore) GetByIDWithRoles(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
	if m.GetByIDWithRolesFunc != nil {
		return m.GetByIDWithRolesFunc(ctx, id, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) GetByEmailWithRoles(ctx context.Context, email string, isActive *bool) (*models.User, error) {
	if m.GetByEmailWithRolesFunc != nil {
		return m.GetByEmailWithRolesFunc(ctx, email, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) GetByPhone(ctx context.Context, phone string, isActive *bool) (*models.User, error) {
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(ctx, phone, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	if m.MarkEmailVerifiedFunc != nil {
		return m.MarkEmailVerifiedFunc(ctx, userID)
	}
	return nil
}
func (m *mockUserStore) MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error {
	if m.MarkPhoneVerifiedFunc != nil {
		return m.MarkPhoneVerifiedFunc(ctx, userID)
	}
	return nil
}
func (m *mockUserStore) List(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) ListWithRoles(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, error) {
	if m.ListWithRolesFunc != nil {
		return m.ListWithRolesFunc(ctx, limit, offset, isActive)
	}
	return nil, nil
}
func (m *mockUserStore) Count(ctx context.Context, isActive *bool) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, isActive)
	}
	return 0, nil
}
func (m *mockUserStore) UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	if m.UpdateTOTPSecretFunc != nil {
		return m.UpdateTOTPSecretFunc(ctx, userID, secret)
	}
	return nil
}
func (m *mockUserStore) EnableTOTP(ctx context.Context, userID uuid.UUID) error {
	if m.EnableTOTPFunc != nil {
		return m.EnableTOTPFunc(ctx, userID)
	}
	return nil
}
func (m *mockUserStore) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	if m.DisableTOTPFunc != nil {
		return m.DisableTOTPFunc(ctx, userID)
	}
	return nil
}

type mockTokenStore struct {
	CreateRefreshTokenFunc  func(ctx context.Context, token *models.RefreshToken) error
	GetRefreshTokenFunc     func(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshTokenFunc  func(ctx context.Context, tokenHash string) error
	RevokeAllUserTokensFunc func(ctx context.Context, userID uuid.UUID) error
	AddToBlacklistFunc      func(ctx context.Context, token *models.TokenBlacklist) error
	IsBlacklistedFunc       func(ctx context.Context, tokenHash string) (bool, error)
}

func (m *mockTokenStore) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	if m.CreateRefreshTokenFunc != nil {
		return m.CreateRefreshTokenFunc(ctx, token)
	}
	return nil
}
func (m *mockTokenStore) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	if m.GetRefreshTokenFunc != nil {
		return m.GetRefreshTokenFunc(ctx, tokenHash)
	}
	return nil, nil
}
func (m *mockTokenStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.RevokeRefreshTokenFunc != nil {
		return m.RevokeRefreshTokenFunc(ctx, tokenHash)
	}
	return nil
}
func (m *mockTokenStore) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeAllUserTokensFunc != nil {
		return m.RevokeAllUserTokensFunc(ctx, userID)
	}
	return nil
}
func (m *mockTokenStore) AddToBlacklist(ctx context.Context, token *models.TokenBlacklist) error {
	if m.AddToBlacklistFunc != nil {
		return m.AddToBlacklistFunc(ctx, token)
	}
	return nil
}
func (m *mockTokenStore) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(ctx, tokenHash)
	}
	return false, nil
}

type mockRBACStore struct {
	// Permission Methods
	CreatePermissionFunc    func(ctx context.Context, permission *models.Permission) error
	GetPermissionByIDFunc   func(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetPermissionByNameFunc func(ctx context.Context, name string) (*models.Permission, error)
	ListPermissionsFunc     func(ctx context.Context) ([]models.Permission, error)
	UpdatePermissionFunc    func(ctx context.Context, id uuid.UUID, description string) error
	DeletePermissionFunc    func(ctx context.Context, id uuid.UUID) error

	// Role Methods
	CreateRoleFunc         func(ctx context.Context, role *models.Role) error
	GetRoleByIDFunc        func(ctx context.Context, id uuid.UUID) (*models.Role, error)
	GetRoleByNameFunc      func(ctx context.Context, name string) (*models.Role, error)
	ListRolesFunc          func(ctx context.Context) ([]models.Role, error)
	UpdateRoleFunc         func(ctx context.Context, id uuid.UUID, displayName, description string) error
	DeleteRoleFunc         func(ctx context.Context, id uuid.UUID) error
	SetRolePermissionsFunc func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// User-Role Methods
	AssignRoleToUserFunc   func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUserFunc func(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRolesFunc       func(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	SetUserRolesFunc       func(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error
	GetUsersWithRoleFunc   func(ctx context.Context, roleID uuid.UUID) ([]models.User, error)

	// Permission Checking Methods
	HasPermissionFunc       func(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)
	HasAnyPermissionFunc    func(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
	HasAllPermissionsFunc   func(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
	GetUserPermissionsFunc  func(ctx context.Context, userID uuid.UUID) ([]models.Permission, error)
	GetPermissionMatrixFunc func(ctx context.Context) (*models.PermissionMatrix, error)
}

// Permission Method Implementations
func (m *mockRBACStore) CreatePermission(ctx context.Context, permission *models.Permission) error {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(ctx, permission)
	}
	return nil
}
func (m *mockRBACStore) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	if m.GetPermissionByIDFunc != nil {
		return m.GetPermissionByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockRBACStore) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	if m.GetPermissionByNameFunc != nil {
		return m.GetPermissionByNameFunc(ctx, name)
	}
	return nil, nil
}
func (m *mockRBACStore) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	if m.ListPermissionsFunc != nil {
		return m.ListPermissionsFunc(ctx)
	}
	return nil, nil
}
func (m *mockRBACStore) UpdatePermission(ctx context.Context, id uuid.UUID, description string) error {
	if m.UpdatePermissionFunc != nil {
		return m.UpdatePermissionFunc(ctx, id, description)
	}
	return nil
}
func (m *mockRBACStore) DeletePermission(ctx context.Context, id uuid.UUID) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(ctx, id)
	}
	return nil
}

// Role Method Implementations
func (m *mockRBACStore) CreateRole(ctx context.Context, role *models.Role) error {
	if m.CreateRoleFunc != nil {
		return m.CreateRoleFunc(ctx, role)
	}
	return nil
}
func (m *mockRBACStore) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	if m.GetRoleByIDFunc != nil {
		return m.GetRoleByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockRBACStore) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	if m.GetRoleByNameFunc != nil {
		return m.GetRoleByNameFunc(ctx, name)
	}
	return nil, nil
}
func (m *mockRBACStore) ListRoles(ctx context.Context) ([]models.Role, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc(ctx)
	}
	return nil, nil
}
func (m *mockRBACStore) UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error {
	if m.UpdateRoleFunc != nil {
		return m.UpdateRoleFunc(ctx, id, displayName, description)
	}
	return nil
}
func (m *mockRBACStore) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if m.DeleteRoleFunc != nil {
		return m.DeleteRoleFunc(ctx, id)
	}
	return nil
}
func (m *mockRBACStore) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if m.SetRolePermissionsFunc != nil {
		return m.SetRolePermissionsFunc(ctx, roleID, permissionIDs)
	}
	return nil
}

// User-Role Method Implementations
func (m *mockRBACStore) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	if m.AssignRoleToUserFunc != nil {
		return m.AssignRoleToUserFunc(ctx, userID, roleID, assignedBy)
	}
	return nil
}
func (m *mockRBACStore) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.RemoveRoleFromUserFunc != nil {
		return m.RemoveRoleFromUserFunc(ctx, userID, roleID)
	}
	return nil
}
func (m *mockRBACStore) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRBACStore) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	if m.SetUserRolesFunc != nil {
		return m.SetUserRolesFunc(ctx, userID, roleIDs, assignedBy)
	}
	return nil
}
func (m *mockRBACStore) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error) {
	if m.GetUsersWithRoleFunc != nil {
		return m.GetUsersWithRoleFunc(ctx, roleID)
	}
	return nil, nil
}

// Permission Checking Implementations
func (m *mockRBACStore) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if m.HasPermissionFunc != nil {
		return m.HasPermissionFunc(ctx, userID, permissionName)
	}
	return false, nil
}
func (m *mockRBACStore) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	if m.HasAnyPermissionFunc != nil {
		return m.HasAnyPermissionFunc(ctx, userID, permissionNames)
	}
	return false, nil
}
func (m *mockRBACStore) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	if m.HasAllPermissionsFunc != nil {
		return m.HasAllPermissionsFunc(ctx, userID, permissionNames)
	}
	return false, nil
}
func (m *mockRBACStore) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	if m.GetUserPermissionsFunc != nil {
		return m.GetUserPermissionsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRBACStore) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	if m.GetPermissionMatrixFunc != nil {
		return m.GetPermissionMatrixFunc(ctx)
	}
	return nil, nil
}

type mockCacheService struct {
	IsBlacklistedFunc             func(ctx context.Context, tokenHash string) (bool, error)
	AddToBlacklistFunc            func(ctx context.Context, tokenHash string, expiration time.Duration) error
	IncrementRateLimitFunc        func(ctx context.Context, key string, window time.Duration) (int64, error)
	StorePendingRegistrationFunc  func(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error
	GetPendingRegistrationFunc    func(ctx context.Context, identifier string) (*models.PendingRegistration, error)
	DeletePendingRegistrationFunc func(ctx context.Context, identifier string) error
}

func (m *mockCacheService) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(ctx, tokenHash)
	}
	return false, nil
}
func (m *mockCacheService) AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error {
	if m.AddToBlacklistFunc != nil {
		return m.AddToBlacklistFunc(ctx, tokenHash, expiration)
	}
	return nil
}
func (m *mockCacheService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	if m.IncrementRateLimitFunc != nil {
		return m.IncrementRateLimitFunc(ctx, key, window)
	}
	return 1, nil
}
func (m *mockCacheService) StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
	if m.StorePendingRegistrationFunc != nil {
		return m.StorePendingRegistrationFunc(ctx, identifier, data, expiration)
	}
	return nil
}
func (m *mockCacheService) GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
	if m.GetPendingRegistrationFunc != nil {
		return m.GetPendingRegistrationFunc(ctx, identifier)
	}
	return nil, nil
}
func (m *mockCacheService) DeletePendingRegistration(ctx context.Context, identifier string) error {
	if m.DeletePendingRegistrationFunc != nil {
		return m.DeletePendingRegistrationFunc(ctx, identifier)
	}
	return nil
}

type mockSMSLogStore struct {
	CreateFunc          func(ctx context.Context, log *models.SMSLog) error
	UpdateStatusFunc    func(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error
	GetStatsFunc        func(ctx context.Context) (*models.SMSStatsResponse, error)
	DeleteOlderThanFunc func(ctx context.Context, duration time.Duration) (int64, error)
}

func (m *mockSMSLogStore) Create(ctx context.Context, log *models.SMSLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	return nil
}
func (m *mockSMSLogStore) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, errorMsg)
	}
	return nil
}
func (m *mockSMSLogStore) GetStats(ctx context.Context) (*models.SMSStatsResponse, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}
	return nil, nil
}
func (m *mockSMSLogStore) DeleteOlderThan(ctx context.Context, duration time.Duration) (int64, error) {
	if m.DeleteOlderThanFunc != nil {
		return m.DeleteOlderThanFunc(ctx, duration)
	}
	return 0, nil
}

type mockTokenService struct {
	GenerateAccessTokenFunc       func(user *models.User) (string, error)
	GenerateRefreshTokenFunc      func(user *models.User) (string, error)
	GenerateTwoFactorTokenFunc    func(user *models.User) (string, error)
	ValidateAccessTokenFunc       func(tokenString string) (*jwt.Claims, error)
	ValidateRefreshTokenFunc      func(tokenString string) (*jwt.Claims, error)
	ExtractClaimsFunc             func(tokenString string) (*jwt.Claims, error)
	GetAccessTokenExpirationFunc  func() time.Duration
	GetRefreshTokenExpirationFunc func() time.Duration
}

func (m *mockTokenService) GenerateAccessToken(user *models.User) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(user)
	}
	return "", nil
}
func (m *mockTokenService) GenerateRefreshToken(user *models.User) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(user)
	}
	return "", nil
}
func (m *mockTokenService) GenerateTwoFactorToken(user *models.User) (string, error) {
	if m.GenerateTwoFactorTokenFunc != nil {
		return m.GenerateTwoFactorTokenFunc(user)
	}
	return "", nil
}
func (m *mockTokenService) ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return nil, nil
}
func (m *mockTokenService) ValidateRefreshToken(tokenString string) (*jwt.Claims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return m.ValidateRefreshTokenFunc(tokenString)
	}
	return nil, nil
}
func (m *mockTokenService) ExtractClaims(tokenString string) (*jwt.Claims, error) {
	if m.ExtractClaimsFunc != nil {
		return m.ExtractClaimsFunc(tokenString)
	}
	return nil, nil
}
func (m *mockTokenService) GetAccessTokenExpiration() time.Duration {
	if m.GetAccessTokenExpirationFunc != nil {
		return m.GetAccessTokenExpirationFunc()
	}
	return time.Hour // Default
}
func (m *mockTokenService) GetRefreshTokenExpiration() time.Duration {
	if m.GetRefreshTokenExpirationFunc != nil {
		return m.GetRefreshTokenExpirationFunc()
	}
	return 24 * time.Hour // Default
}

type mockAuditLogger struct {
	LogWithActionFunc func(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{})
	LogFunc           func(params AuditLogParams)
}

func (m *mockAuditLogger) LogWithAction(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
	if m.LogWithActionFunc != nil {
		m.LogWithActionFunc(userID, action, status, ip, userAgent, details)
	}
}

func (m *mockAuditLogger) Log(params AuditLogParams) {
	if m.LogFunc != nil {
		m.LogFunc(params)
	}
}

type mockOTPStore struct {
	CreateFunc                func(ctx context.Context, otp *models.OTP) error
	GetByEmailAndTypeFunc     func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error)
	MarkAsUsedFunc            func(ctx context.Context, id uuid.UUID) error
	InvalidateAllForEmailFunc func(ctx context.Context, email string, otpType models.OTPType) error
	CountRecentByEmailFunc    func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error)
	DeleteExpiredFunc         func(ctx context.Context, olderThan time.Duration) error
	GetByPhoneAndTypeFunc     func(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error)
	InvalidateAllForPhoneFunc func(ctx context.Context, phone string, otpType models.OTPType) error
	CountRecentByPhoneFunc    func(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error)
}

func (m *mockOTPStore) Create(ctx context.Context, otp *models.OTP) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, otp)
	}
	return nil
}
func (m *mockOTPStore) GetByEmailAndType(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
	if m.GetByEmailAndTypeFunc != nil {
		return m.GetByEmailAndTypeFunc(ctx, email, otpType)
	}
	return nil, nil
}
func (m *mockOTPStore) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	if m.MarkAsUsedFunc != nil {
		return m.MarkAsUsedFunc(ctx, id)
	}
	return nil
}
func (m *mockOTPStore) InvalidateAllForEmail(ctx context.Context, email string, otpType models.OTPType) error {
	if m.InvalidateAllForEmailFunc != nil {
		return m.InvalidateAllForEmailFunc(ctx, email, otpType)
	}
	return nil
}
func (m *mockOTPStore) CountRecentByEmail(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
	if m.CountRecentByEmailFunc != nil {
		return m.CountRecentByEmailFunc(ctx, email, otpType, duration)
	}
	return 0, nil
}
func (m *mockOTPStore) DeleteExpired(ctx context.Context, olderThan time.Duration) error {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx, olderThan)
	}
	return nil
}
func (m *mockOTPStore) GetByPhoneAndType(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error) {
	if m.GetByPhoneAndTypeFunc != nil {
		return m.GetByPhoneAndTypeFunc(ctx, phone, otpType)
	}
	return nil, nil
}
func (m *mockOTPStore) InvalidateAllForPhone(ctx context.Context, phone string, otpType models.OTPType) error {
	if m.InvalidateAllForPhoneFunc != nil {
		return m.InvalidateAllForPhoneFunc(ctx, phone, otpType)
	}
	return nil
}
func (m *mockOTPStore) CountRecentByPhone(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error) {
	if m.CountRecentByPhoneFunc != nil {
		return m.CountRecentByPhoneFunc(ctx, phone, otpType, duration)
	}
	return 0, nil
}

type mockEmailSender struct {
	SendOTPFunc     func(to, code, otpType string) error
	SendWelcomeFunc func(to, username string) error
	SendFunc        func(to, subject, htmlBody string) error
}

func (m *mockEmailSender) SendOTP(to, code, otpType string) error {
	if m.SendOTPFunc != nil {
		return m.SendOTPFunc(to, code, otpType)
	}
	return nil
}
func (m *mockEmailSender) SendWelcome(to, username string) error {
	if m.SendWelcomeFunc != nil {
		return m.SendWelcomeFunc(to, username)
	}
	return nil
}
func (m *mockEmailSender) Send(to, subject, htmlBody string) error {
	if m.SendFunc != nil {
		return m.SendFunc(to, subject, htmlBody)
	}
	return nil
}

type mockGeoLocationProvider struct {
	GetLocationFunc func(ip string) (*models.GeoLocation, error)
}

func (m *mockGeoLocationProvider) GetLocation(ip string) (*models.GeoLocation, error) {
	if m.GetLocationFunc != nil {
		return m.GetLocationFunc(ip)
	}
	return nil, nil
}

type mockIPFilterStore struct {
	CreateIPFilterFunc     func(ctx context.Context, filter *models.IPFilter) error
	GetIPFilterByIDFunc    func(ctx context.Context, id uuid.UUID) (*models.IPFilter, error)
	ListIPFiltersFunc      func(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error)
	UpdateIPFilterFunc     func(ctx context.Context, id uuid.UUID, reason string, isActive bool) error
	DeleteIPFilterFunc     func(ctx context.Context, id uuid.UUID) error
	GetActiveIPFiltersFunc func(ctx context.Context) ([]models.IPFilter, error)
}

func (m *mockIPFilterStore) CreateIPFilter(ctx context.Context, filter *models.IPFilter) error {
	if m.CreateIPFilterFunc != nil {
		return m.CreateIPFilterFunc(ctx, filter)
	}
	return nil
}
func (m *mockIPFilterStore) GetIPFilterByID(ctx context.Context, id uuid.UUID) (*models.IPFilter, error) {
	if m.GetIPFilterByIDFunc != nil {
		return m.GetIPFilterByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockIPFilterStore) ListIPFilters(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
	if m.ListIPFiltersFunc != nil {
		return m.ListIPFiltersFunc(ctx, page, perPage, filterType)
	}
	return nil, 0, nil
}
func (m *mockIPFilterStore) UpdateIPFilter(ctx context.Context, id uuid.UUID, reason string, isActive bool) error {
	if m.UpdateIPFilterFunc != nil {
		return m.UpdateIPFilterFunc(ctx, id, reason, isActive)
	}
	return nil
}
func (m *mockIPFilterStore) DeleteIPFilter(ctx context.Context, id uuid.UUID) error {
	if m.DeleteIPFilterFunc != nil {
		return m.DeleteIPFilterFunc(ctx, id)
	}
	return nil
}
func (m *mockIPFilterStore) GetActiveIPFilters(ctx context.Context) ([]models.IPFilter, error) {
	if m.GetActiveIPFiltersFunc != nil {
		return m.GetActiveIPFiltersFunc(ctx)
	}
	return nil, nil
}

type mockOAuthStore struct {
	CreateOAuthAccountFunc            func(ctx context.Context, account *models.OAuthAccount) error
	GetOAuthAccountFunc               func(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error)
	GetOAuthAccountsByUserIDFunc      func(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error)
	UpdateOAuthAccountFunc            func(ctx context.Context, account *models.OAuthAccount) error
	DeleteOAuthAccountFunc            func(ctx context.Context, id uuid.UUID) error
	DeleteOAuthAccountsByProviderFunc func(ctx context.Context, userID uuid.UUID, provider string) error
	GetByUserIDFunc                   func(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error)
	ListAllFunc                       func(ctx context.Context) ([]*models.OAuthAccount, error)
}

func (m *mockOAuthStore) CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	if m.CreateOAuthAccountFunc != nil {
		return m.CreateOAuthAccountFunc(ctx, account)
	}
	return nil
}
func (m *mockOAuthStore) GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
	if m.GetOAuthAccountFunc != nil {
		return m.GetOAuthAccountFunc(ctx, provider, providerUserID)
	}
	return nil, nil
}
func (m *mockOAuthStore) GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	if m.GetOAuthAccountsByUserIDFunc != nil {
		return m.GetOAuthAccountsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockOAuthStore) UpdateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	if m.UpdateOAuthAccountFunc != nil {
		return m.UpdateOAuthAccountFunc(ctx, account)
	}
	return nil
}
func (m *mockOAuthStore) DeleteOAuthAccount(ctx context.Context, id uuid.UUID) error {
	if m.DeleteOAuthAccountFunc != nil {
		return m.DeleteOAuthAccountFunc(ctx, id)
	}
	return nil
}
func (m *mockOAuthStore) DeleteOAuthAccountsByProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	if m.DeleteOAuthAccountsByProviderFunc != nil {
		return m.DeleteOAuthAccountsByProviderFunc(ctx, userID, provider)
	}
	return nil
}
func (m *mockOAuthStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockOAuthStore) ListAll(ctx context.Context) ([]*models.OAuthAccount, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

type mockAuditStore struct {
	CreateFunc                 func(ctx context.Context, log *models.AuditLog) error
	GetByUserIDFunc            func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	GetByActionFunc            func(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error)
	GetFailedLoginAttemptsFunc func(ctx context.Context, ipAddress string, limit int) ([]*models.AuditLog, error)
	ListFunc                   func(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)
	CountFunc                  func(ctx context.Context) (int, error)
	DeleteOlderThanFunc        func(ctx context.Context, days int) error
	CountByActionSinceFunc     func(ctx context.Context, action models.AuditAction, since time.Time) (int, error)
}

func (m *mockAuditStore) Create(ctx context.Context, log *models.AuditLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	return nil
}
func (m *mockAuditStore) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, nil
}
func (m *mockAuditStore) GetByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error) {
	if m.GetByActionFunc != nil {
		return m.GetByActionFunc(ctx, action, limit, offset)
	}
	return nil, nil
}
func (m *mockAuditStore) GetFailedLoginAttempts(ctx context.Context, ipAddress string, limit int) ([]*models.AuditLog, error) {
	if m.GetFailedLoginAttemptsFunc != nil {
		return m.GetFailedLoginAttemptsFunc(ctx, ipAddress, limit)
	}
	return nil, nil
}
func (m *mockAuditStore) List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, nil
}
func (m *mockAuditStore) Count(ctx context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}
func (m *mockAuditStore) DeleteOlderThan(ctx context.Context, days int) error {
	if m.DeleteOlderThanFunc != nil {
		return m.DeleteOlderThanFunc(ctx, days)
	}
	return nil
}
func (m *mockAuditStore) CountByActionSince(ctx context.Context, action models.AuditAction, since time.Time) (int, error) {
	if m.CountByActionSinceFunc != nil {
		return m.CountByActionSinceFunc(ctx, action, since)
	}
	return 0, nil
}

type mockJWTService struct {
	GenerateAccessTokenFunc       func(user *models.User) (string, error)
	GenerateRefreshTokenFunc      func(user *models.User) (string, error)
	GetAccessTokenExpirationFunc  func() time.Duration
	GetRefreshTokenExpirationFunc func() time.Duration
	ValidateAccessTokenFunc       func(tokenString string) (*jwt.Claims, error)
	ValidateRefreshTokenFunc      func(tokenString string) (*jwt.Claims, error)
}

func (m *mockJWTService) GenerateAccessToken(user *models.User) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(user)
	}
	return "", nil
}
func (m *mockJWTService) GenerateRefreshToken(user *models.User) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(user)
	}
	return "", nil
}
func (m *mockJWTService) GetAccessTokenExpiration() time.Duration {
	if m.GetAccessTokenExpirationFunc != nil {
		return m.GetAccessTokenExpirationFunc()
	}
	return time.Hour
}
func (m *mockJWTService) GetRefreshTokenExpiration() time.Duration {
	if m.GetRefreshTokenExpirationFunc != nil {
		return m.GetRefreshTokenExpirationFunc()
	}
	return time.Hour * 24
}
func (m *mockJWTService) ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return nil, nil
}
func (m *mockJWTService) ValidateRefreshToken(tokenString string) (*jwt.Claims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return m.ValidateRefreshTokenFunc(tokenString)
	}
	return nil, nil
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

type mockAPIKeyStore struct {
	CreateFunc            func(ctx context.Context, apiKey *models.APIKey) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*models.APIKey, error)
	GetByKeyHashFunc      func(ctx context.Context, keyHash string) (*models.APIKey, error)
	GetByUserIDFunc       func(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
	GetActiveByUserIDFunc func(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
	UpdateFunc            func(ctx context.Context, apiKey *models.APIKey) error
	UpdateLastUsedFunc    func(ctx context.Context, id uuid.UUID) error
	RevokeFunc            func(ctx context.Context, id uuid.UUID) error
	DeleteFunc            func(ctx context.Context, id uuid.UUID) error
	DeleteExpiredFunc     func(ctx context.Context) error
	CountFunc             func(ctx context.Context, userID uuid.UUID) (int, error)
	CountActiveFunc       func(ctx context.Context, userID uuid.UUID) (int, error)
	ListAllFunc           func(ctx context.Context) ([]*models.APIKey, error)
}

func (m *mockAPIKeyStore) Create(ctx context.Context, apiKey *models.APIKey) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, apiKey)
	}
	return nil
}
func (m *mockAPIKeyStore) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockAPIKeyStore) GetByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	if m.GetByKeyHashFunc != nil {
		return m.GetByKeyHashFunc(ctx, keyHash)
	}
	return nil, nil
}
func (m *mockAPIKeyStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockAPIKeyStore) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	if m.GetActiveByUserIDFunc != nil {
		return m.GetActiveByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockAPIKeyStore) Update(ctx context.Context, apiKey *models.APIKey) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, apiKey)
	}
	return nil
}

type mockBackupCodeStore struct {
	CreateBatchFunc         func(ctx context.Context, codes []*models.BackupCode) error
	GetUnusedByUserIDFunc   func(ctx context.Context, userID uuid.UUID) ([]*models.BackupCode, error)
	CountUnusedByUserIDFunc func(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsUsedFunc          func(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserIDFunc   func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockBackupCodeStore) CreateBatch(ctx context.Context, codes []*models.BackupCode) error {
	if m.CreateBatchFunc != nil {
		return m.CreateBatchFunc(ctx, codes)
	}
	return nil
}
func (m *mockBackupCodeStore) GetUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*models.BackupCode, error) {
	if m.GetUnusedByUserIDFunc != nil {
		return m.GetUnusedByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockBackupCodeStore) CountUnusedByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.CountUnusedByUserIDFunc != nil {
		return m.CountUnusedByUserIDFunc(ctx, userID)
	}
	return 0, nil
}
func (m *mockBackupCodeStore) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	if m.MarkAsUsedFunc != nil {
		return m.MarkAsUsedFunc(ctx, id)
	}
	return nil
}
func (m *mockBackupCodeStore) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteAllByUserIDFunc != nil {
		return m.DeleteAllByUserIDFunc(ctx, userID)
	}
	return nil
}
func (m *mockAPIKeyStore) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(ctx, id)
	}
	return nil
}
func (m *mockAPIKeyStore) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(ctx, id)
	}
	return nil
}
func (m *mockAPIKeyStore) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *mockAPIKeyStore) DeleteExpired(ctx context.Context) error {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx)
	}
	return nil
}
func (m *mockAPIKeyStore) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, userID)
	}
	return 0, nil
}
func (m *mockAPIKeyStore) CountActive(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.CountActiveFunc != nil {
		return m.CountActiveFunc(ctx, userID)
	}
	return 0, nil
}
func (m *mockAPIKeyStore) ListAll(ctx context.Context) ([]*models.APIKey, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

type mockSessionStore struct {
	CreateSessionFunc                func(ctx context.Context, session *models.Session) error
	GetSessionByIDFunc               func(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetSessionByTokenHashFunc        func(ctx context.Context, tokenHash string) (*models.Session, error)
	GetUserSessionsFunc              func(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
	GetUserSessionsPaginatedFunc     func(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error)
	GetAllSessionsPaginatedFunc      func(ctx context.Context, page, perPage int) ([]models.Session, int, error)
	RevokeSessionFunc                func(ctx context.Context, id uuid.UUID) error
	RevokeUserSessionFunc            func(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeAllUserSessionsFunc        func(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error
	UpdateSessionNameFunc            func(ctx context.Context, sessionID uuid.UUID, name string) error
	UpdateSessionAccessTokenHashFunc func(ctx context.Context, sessionID uuid.UUID, accessTokenHash string) error
	RefreshSessionTokensFunc         func(ctx context.Context, oldTokenHash, newTokenHash, newAccessTokenHash string, newExpiresAt time.Time) error
	GetSessionStatsFunc              func(ctx context.Context) (*models.SessionStats, error)
	DeleteExpiredSessionsFunc        func(ctx context.Context, olderThan time.Duration) error
}

func (m *mockSessionStore) CreateSession(ctx context.Context, session *models.Session) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, session)
	}
	return nil
}
func (m *mockSessionStore) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockSessionStore) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	if m.GetSessionByTokenHashFunc != nil {
		return m.GetSessionByTokenHashFunc(ctx, tokenHash)
	}
	return nil, nil
}
func (m *mockSessionStore) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	if m.GetUserSessionsFunc != nil {
		return m.GetUserSessionsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockSessionStore) GetUserSessionsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	if m.GetUserSessionsPaginatedFunc != nil {
		return m.GetUserSessionsPaginatedFunc(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockSessionStore) GetAllSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error) {
	if m.GetAllSessionsPaginatedFunc != nil {
		return m.GetAllSessionsPaginatedFunc(ctx, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockSessionStore) RevokeSession(ctx context.Context, id uuid.UUID) error {
	if m.RevokeSessionFunc != nil {
		return m.RevokeSessionFunc(ctx, id)
	}
	return nil
}
func (m *mockSessionStore) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if m.RevokeUserSessionFunc != nil {
		return m.RevokeUserSessionFunc(ctx, userID, sessionID)
	}
	return nil
}
func (m *mockSessionStore) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error {
	if m.RevokeAllUserSessionsFunc != nil {
		return m.RevokeAllUserSessionsFunc(ctx, userID, exceptSessionID)
	}
	return nil
}
func (m *mockSessionStore) UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error {
	if m.UpdateSessionNameFunc != nil {
		return m.UpdateSessionNameFunc(ctx, sessionID, name)
	}
	return nil
}
func (m *mockSessionStore) UpdateSessionAccessTokenHash(ctx context.Context, sessionID uuid.UUID, accessTokenHash string) error {
	if m.UpdateSessionAccessTokenHashFunc != nil {
		return m.UpdateSessionAccessTokenHashFunc(ctx, sessionID, accessTokenHash)
	}
	return nil
}
func (m *mockSessionStore) RefreshSessionTokens(ctx context.Context, oldTokenHash, newTokenHash, newAccessTokenHash string, newExpiresAt time.Time) error {
	if m.RefreshSessionTokensFunc != nil {
		return m.RefreshSessionTokensFunc(ctx, oldTokenHash, newTokenHash, newAccessTokenHash, newExpiresAt)
	}
	return nil
}
func (m *mockSessionStore) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	if m.GetSessionStatsFunc != nil {
		return m.GetSessionStatsFunc(ctx)
	}
	return nil, nil
}
func (m *mockSessionStore) DeleteExpiredSessions(ctx context.Context, olderThan time.Duration) error {
	if m.DeleteExpiredSessionsFunc != nil {
		return m.DeleteExpiredSessionsFunc(ctx, olderThan)
	}
	return nil
}
