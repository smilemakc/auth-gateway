package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/uptrace/bun"
)

// ===========================================================================
// UserStore mock
// ===========================================================================

type mockUserStoreHandler struct {
	GetByIDFunc              func(id uuid.UUID) (*models.User, error)
	GetByEmailFunc           func(email string) (*models.User, error)
	GetByUsernameFunc        func(username string) (*models.User, error)
	GetByPhoneFunc           func(phone string) (*models.User, error)
	CreateFunc               func(user *models.User) error
	UpdateFunc               func(user *models.User) error
	UpdatePasswordFunc       func(userID uuid.UUID, hash string) error
	EmailExistsFunc          func(email string) (bool, error)
	UsernameExistsFunc       func(username string) (bool, error)
	PhoneExistsFunc          func(phone string) (bool, error)
	MarkEmailVerifiedFunc    func(userID uuid.UUID) error
	MarkPhoneVerifiedFunc    func(userID uuid.UUID) error
	ListFunc                 func(opts ...service.UserListOption) ([]*models.User, error)
	CountFunc                func(isActive *bool) (int, error)
	GetUsersUpdatedAfterFunc func(after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error)
	UpdateTOTPSecretFunc     func(userID uuid.UUID, secret string) error
	EnableTOTPFunc           func(userID uuid.UUID) error
	DisableTOTPFunc          func(userID uuid.UUID) error
}

func (m *mockUserStoreHandler) GetByID(_ context.Context, id uuid.UUID, _ *bool, _ ...service.UserGetOption) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}
func (m *mockUserStoreHandler) GetByEmail(_ context.Context, email string, _ *bool, _ ...service.UserGetOption) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}
func (m *mockUserStoreHandler) GetByUsername(_ context.Context, username string, _ *bool, _ ...service.UserGetOption) (*models.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(username)
	}
	return nil, nil
}
func (m *mockUserStoreHandler) GetByPhone(_ context.Context, phone string, _ *bool, _ ...service.UserGetOption) (*models.User, error) {
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(phone)
	}
	return nil, nil
}
func (m *mockUserStoreHandler) Create(_ context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}
func (m *mockUserStoreHandler) Update(_ context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}
func (m *mockUserStoreHandler) UpdatePassword(_ context.Context, userID uuid.UUID, hash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(userID, hash)
	}
	return nil
}
func (m *mockUserStoreHandler) EmailExists(_ context.Context, email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(email)
	}
	return false, nil
}
func (m *mockUserStoreHandler) UsernameExists(_ context.Context, username string) (bool, error) {
	if m.UsernameExistsFunc != nil {
		return m.UsernameExistsFunc(username)
	}
	return false, nil
}
func (m *mockUserStoreHandler) PhoneExists(_ context.Context, phone string) (bool, error) {
	if m.PhoneExistsFunc != nil {
		return m.PhoneExistsFunc(phone)
	}
	return false, nil
}
func (m *mockUserStoreHandler) MarkEmailVerified(_ context.Context, userID uuid.UUID) error {
	if m.MarkEmailVerifiedFunc != nil {
		return m.MarkEmailVerifiedFunc(userID)
	}
	return nil
}
func (m *mockUserStoreHandler) MarkPhoneVerified(_ context.Context, userID uuid.UUID) error {
	if m.MarkPhoneVerifiedFunc != nil {
		return m.MarkPhoneVerifiedFunc(userID)
	}
	return nil
}
func (m *mockUserStoreHandler) List(_ context.Context, opts ...service.UserListOption) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(opts...)
	}
	return nil, nil
}
func (m *mockUserStoreHandler) Count(_ context.Context, isActive *bool) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(isActive)
	}
	return 0, nil
}
func (m *mockUserStoreHandler) GetUsersUpdatedAfter(_ context.Context, after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	if m.GetUsersUpdatedAfterFunc != nil {
		return m.GetUsersUpdatedAfterFunc(after, appID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockUserStoreHandler) UpdateTOTPSecret(_ context.Context, userID uuid.UUID, secret string) error {
	if m.UpdateTOTPSecretFunc != nil {
		return m.UpdateTOTPSecretFunc(userID, secret)
	}
	return nil
}
func (m *mockUserStoreHandler) EnableTOTP(_ context.Context, userID uuid.UUID) error {
	if m.EnableTOTPFunc != nil {
		return m.EnableTOTPFunc(userID)
	}
	return nil
}
func (m *mockUserStoreHandler) DisableTOTP(_ context.Context, userID uuid.UUID) error {
	if m.DisableTOTPFunc != nil {
		return m.DisableTOTPFunc(userID)
	}
	return nil
}

// ===========================================================================
// TokenStore mock (TransactionalTokenStore)
// ===========================================================================

type mockTokenStoreHandler struct{}

func (m *mockTokenStoreHandler) CreateRefreshToken(_ context.Context, _ *models.RefreshToken) error {
	return nil
}
func (m *mockTokenStoreHandler) GetRefreshToken(_ context.Context, _ string) (*models.RefreshToken, error) {
	return nil, nil
}
func (m *mockTokenStoreHandler) RevokeRefreshToken(_ context.Context, _ string) error {
	return nil
}
func (m *mockTokenStoreHandler) RevokeAllUserTokens(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockTokenStoreHandler) AddToBlacklist(_ context.Context, _ *models.TokenBlacklist) error {
	return nil
}
func (m *mockTokenStoreHandler) IsBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockTokenStoreHandler) GetRefreshTokenForUpdate(_ context.Context, _ bun.Tx, _ string) (*models.RefreshToken, error) {
	return nil, nil
}
func (m *mockTokenStoreHandler) RevokeRefreshTokenWithTx(_ context.Context, _ bun.Tx, _ string) error {
	return nil
}
func (m *mockTokenStoreHandler) CreateRefreshTokenWithTx(_ context.Context, _ bun.Tx, _ *models.RefreshToken) error {
	return nil
}
func (m *mockTokenStoreHandler) GetAllActiveBlacklistEntries(_ context.Context) ([]*models.TokenBlacklist, error) {
	return nil, nil
}

// ===========================================================================
// TokenService (JWT) mock
// ===========================================================================

type mockTokenServiceHandler struct {
	GenerateAccessTokenFunc       func(user *models.User) (string, error)
	GenerateRefreshTokenFunc      func(user *models.User) (string, error)
	ValidateAccessTokenFunc       func(token string) (*jwt.Claims, error)
	ValidateRefreshTokenFunc      func(token string) error
	ExtractClaimsFunc             func(token string) (uuid.UUID, error)
	GetAccessTokenExpirationFunc  func() int64
	GetRefreshTokenExpirationFunc func() time.Duration
}

func (m *mockTokenServiceHandler) GenerateAccessToken(user *models.User, _ ...*uuid.UUID) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(user)
	}
	return "mock-access-token", nil
}
func (m *mockTokenServiceHandler) GenerateRefreshToken(user *models.User, _ ...*uuid.UUID) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(user)
	}
	return "mock-refresh-token", nil
}
func (m *mockTokenServiceHandler) GenerateTwoFactorToken(_ *models.User, _ ...*uuid.UUID) (string, error) {
	return "mock-2fa-token", nil
}
func (m *mockTokenServiceHandler) ValidateAccessToken(token string) (*jwt.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(token)
	}
	return nil, nil
}
func (m *mockTokenServiceHandler) ValidateRefreshToken(token string) (*jwt.Claims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return nil, m.ValidateRefreshTokenFunc(token)
	}
	return nil, nil
}
func (m *mockTokenServiceHandler) ExtractClaims(token string) (*jwt.Claims, error) {
	if m.ExtractClaimsFunc != nil {
		uid, err := m.ExtractClaimsFunc(token)
		if err != nil {
			return nil, err
		}
		return &jwt.Claims{UserID: uid}, nil
	}
	return &jwt.Claims{UserID: uuid.New()}, nil
}
func (m *mockTokenServiceHandler) GetAccessTokenExpiration() time.Duration {
	if m.GetAccessTokenExpirationFunc != nil {
		secs := m.GetAccessTokenExpirationFunc()
		return time.Duration(secs) * time.Second
	}
	return 15 * time.Minute
}
func (m *mockTokenServiceHandler) GetRefreshTokenExpiration() time.Duration {
	if m.GetRefreshTokenExpirationFunc != nil {
		return m.GetRefreshTokenExpirationFunc()
	}
	return 7 * 24 * time.Hour
}

// ===========================================================================
// RBACStore mock
// ===========================================================================

type mockRBACStoreHandler struct {
	ListPermissionsFunc      func() ([]models.Permission, error)
	ListPermissionsByAppFunc func(appID *uuid.UUID) ([]models.Permission, error)
	ListRolesFunc            func() ([]models.Role, error)
	ListRolesByAppFunc       func(appID *uuid.UUID) ([]models.Role, error)
	CreatePermissionFunc     func(p *models.Permission) error
	CreateRoleFunc           func(r *models.Role) error
	GetPermissionByIDFunc    func(id uuid.UUID) (*models.Permission, error)
	GetRoleByIDFunc          func(id uuid.UUID) (*models.Role, error)
	GetRoleByNameFunc        func(name string) (*models.Role, error)
	UpdatePermissionFunc     func(id uuid.UUID, desc string) error
	UpdateRoleFunc           func(id uuid.UUID, displayName, description string) error
	DeletePermissionFunc     func(id uuid.UUID) error
	DeleteRoleFunc           func(id uuid.UUID) error
	GetPermissionMatrixFunc  func() (*models.PermissionMatrix, error)
	GetUserRolesFunc         func() ([]models.Role, error)
	GetUsersWithRoleFunc     func(roleID uuid.UUID) ([]models.User, error)
}

func (m *mockRBACStoreHandler) CreatePermission(_ context.Context, p *models.Permission) error {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(p)
	}
	return nil
}
func (m *mockRBACStoreHandler) GetPermissionByID(_ context.Context, id uuid.UUID) (*models.Permission, error) {
	if m.GetPermissionByIDFunc != nil {
		return m.GetPermissionByIDFunc(id)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) GetPermissionByName(_ context.Context, _ string) (*models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreHandler) ListPermissions(_ context.Context) ([]models.Permission, error) {
	if m.ListPermissionsFunc != nil {
		return m.ListPermissionsFunc()
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) UpdatePermission(_ context.Context, id uuid.UUID, desc string) error {
	if m.UpdatePermissionFunc != nil {
		return m.UpdatePermissionFunc(id, desc)
	}
	return nil
}
func (m *mockRBACStoreHandler) DeletePermission(_ context.Context, id uuid.UUID) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(id)
	}
	return nil
}
func (m *mockRBACStoreHandler) ListPermissionsByApp(_ context.Context, appID *uuid.UUID) ([]models.Permission, error) {
	if m.ListPermissionsByAppFunc != nil {
		return m.ListPermissionsByAppFunc(appID)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) CreateRole(_ context.Context, r *models.Role) error {
	if m.CreateRoleFunc != nil {
		return m.CreateRoleFunc(r)
	}
	return nil
}
func (m *mockRBACStoreHandler) GetRoleByID(_ context.Context, id uuid.UUID) (*models.Role, error) {
	if m.GetRoleByIDFunc != nil {
		return m.GetRoleByIDFunc(id)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) GetRoleByName(_ context.Context, name string) (*models.Role, error) {
	if m.GetRoleByNameFunc != nil {
		return m.GetRoleByNameFunc(name)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) ListRoles(_ context.Context) ([]models.Role, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc()
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) UpdateRole(_ context.Context, id uuid.UUID, displayName, description string) error {
	if m.UpdateRoleFunc != nil {
		return m.UpdateRoleFunc(id, displayName, description)
	}
	return nil
}
func (m *mockRBACStoreHandler) DeleteRole(_ context.Context, id uuid.UUID) error {
	if m.DeleteRoleFunc != nil {
		return m.DeleteRoleFunc(id)
	}
	return nil
}
func (m *mockRBACStoreHandler) SetRolePermissions(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreHandler) AssignRoleToUser(_ context.Context, _, _, _ uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreHandler) RemoveRoleFromUser(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreHandler) GetUserRoles(_ context.Context, _ uuid.UUID) ([]models.Role, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc()
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) SetUserRoles(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockRBACStoreHandler) GetUsersWithRole(_ context.Context, roleID uuid.UUID) ([]models.User, error) {
	if m.GetUsersWithRoleFunc != nil {
		return m.GetUsersWithRoleFunc(roleID)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) HasPermission(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreHandler) HasAnyPermission(_ context.Context, _ uuid.UUID, _ []string) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreHandler) HasAllPermissions(_ context.Context, _ uuid.UUID, _ []string) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreHandler) GetUserPermissions(_ context.Context, _ uuid.UUID) ([]models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStoreHandler) GetPermissionMatrix(_ context.Context) (*models.PermissionMatrix, error) {
	if m.GetPermissionMatrixFunc != nil {
		return m.GetPermissionMatrixFunc()
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) GetRoleByNameAndApp(_ context.Context, _ string, _ *uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreHandler) ListRolesByApp(_ context.Context, appID *uuid.UUID) ([]models.Role, error) {
	if m.ListRolesByAppFunc != nil {
		return m.ListRolesByAppFunc(appID)
	}
	return nil, nil
}
func (m *mockRBACStoreHandler) HasPermissionInApp(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockRBACStoreHandler) GetUserRolesInApp(_ context.Context, _ uuid.UUID, _ *uuid.UUID) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRBACStoreHandler) AssignRoleToUserInApp(_ context.Context, _, _, _ uuid.UUID, _ *uuid.UUID) error {
	return nil
}

// ===========================================================================
// AuditLogger mock
// ===========================================================================

type mockAuditLoggerHandler struct {
	LogCalled bool
}

func (m *mockAuditLoggerHandler) LogWithAction(_ *uuid.UUID, _, _, _, _ string, _ map[string]interface{}) {
	m.LogCalled = true
}
func (m *mockAuditLoggerHandler) Log(_ service.AuditLogParams) {
	m.LogCalled = true
}

// ===========================================================================
// CacheService mock
// ===========================================================================

type mockCacheServiceHandler struct{}

func (m *mockCacheServiceHandler) IsBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockCacheServiceHandler) AddToBlacklist(_ context.Context, _ string, _ time.Duration) error {
	return nil
}
func (m *mockCacheServiceHandler) IncrementRateLimit(_ context.Context, _ string, _ time.Duration) (int64, error) {
	return 1, nil
}
func (m *mockCacheServiceHandler) StorePendingRegistration(_ context.Context, _ string, _ *models.PendingRegistration, _ time.Duration) error {
	return nil
}
func (m *mockCacheServiceHandler) GetPendingRegistration(_ context.Context, _ string) (*models.PendingRegistration, error) {
	return nil, nil
}
func (m *mockCacheServiceHandler) DeletePendingRegistration(_ context.Context, _ string) error {
	return nil
}

// ===========================================================================
// BlacklistChecker mock
// ===========================================================================

type mockBlacklistCheckerHandler struct{}

func (m *mockBlacklistCheckerHandler) IsBlacklisted(_ context.Context, _ string) bool {
	return false
}
func (m *mockBlacklistCheckerHandler) AddToBlacklist(_ context.Context, _ string, _ *uuid.UUID, _ time.Duration) error {
	return nil
}
func (m *mockBlacklistCheckerHandler) AddAccessToken(_ context.Context, _ string, _ *uuid.UUID) error {
	return nil
}
func (m *mockBlacklistCheckerHandler) AddRefreshToken(_ context.Context, _ string, _ *uuid.UUID) error {
	return nil
}
func (m *mockBlacklistCheckerHandler) BlacklistSessionTokens(_ context.Context, _ *models.Session) error {
	return nil
}
func (m *mockBlacklistCheckerHandler) BlacklistAllUserSessions(_ context.Context, _ uuid.UUID) error {
	return nil
}

// ===========================================================================
// SessionManager mock
// ===========================================================================

type mockSessionManagerHandler struct{}

func (m *mockSessionManagerHandler) CreateSessionNonFatal(_ context.Context, _ service.SessionCreationParams) *models.Session {
	return nil
}
func (m *mockSessionManagerHandler) RefreshSessionNonFatal(_ context.Context, _ service.SessionRefreshParams) bool {
	return false
}

// ===========================================================================
// TransactionDB mock
// ===========================================================================

type mockTransactionDBHandler struct{}

func (m *mockTransactionDBHandler) RunInTx(_ context.Context, fn func(context.Context, bun.Tx) error) error {
	return fn(context.Background(), bun.Tx{})
}

// ===========================================================================
// OTPStore mock
// ===========================================================================

type mockOTPStoreHandler struct {
	CreateFunc                func(otp *models.OTP) error
	GetByEmailAndTypeFunc     func(email string, otpType models.OTPType) (*models.OTP, error)
	MarkAsUsedFunc            func(id uuid.UUID) error
	InvalidateAllForEmailFunc func(email string, otpType models.OTPType) error
	CountRecentByEmailFunc    func(email string, otpType models.OTPType, duration time.Duration) (int, error)
}

func (m *mockOTPStoreHandler) Create(_ context.Context, otp *models.OTP) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(otp)
	}
	return nil
}
func (m *mockOTPStoreHandler) GetByEmailAndType(_ context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
	if m.GetByEmailAndTypeFunc != nil {
		return m.GetByEmailAndTypeFunc(email, otpType)
	}
	return nil, nil
}
func (m *mockOTPStoreHandler) MarkAsUsed(_ context.Context, id uuid.UUID) error {
	if m.MarkAsUsedFunc != nil {
		return m.MarkAsUsedFunc(id)
	}
	return nil
}
func (m *mockOTPStoreHandler) InvalidateAllForEmail(_ context.Context, email string, otpType models.OTPType) error {
	if m.InvalidateAllForEmailFunc != nil {
		return m.InvalidateAllForEmailFunc(email, otpType)
	}
	return nil
}
func (m *mockOTPStoreHandler) CountRecentByEmail(_ context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
	if m.CountRecentByEmailFunc != nil {
		return m.CountRecentByEmailFunc(email, otpType, duration)
	}
	return 0, nil
}
func (m *mockOTPStoreHandler) DeleteExpired(_ context.Context, _ time.Duration) error {
	return nil
}
func (m *mockOTPStoreHandler) GetByPhoneAndType(_ context.Context, _ string, _ models.OTPType) (*models.OTP, error) {
	return nil, nil
}
func (m *mockOTPStoreHandler) InvalidateAllForPhone(_ context.Context, _ string, _ models.OTPType) error {
	return nil
}
func (m *mockOTPStoreHandler) CountRecentByPhone(_ context.Context, _ string, _ models.OTPType, _ time.Duration) (int, error) {
	return 0, nil
}

// ===========================================================================
// AuditStore mock (for AdminHandler)
// ===========================================================================

type mockAuditStoreHandler struct {
	CreateFunc             func(log *models.AuditLog) error
	ListFunc               func(limit, offset int) ([]*models.AuditLog, error)
	CountFunc              func() (int, error)
	GetByUserIDFunc        func(userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	CountByActionSinceFunc func(action models.AuditAction, since time.Time) (int, error)
	ListByAppFunc          func(appID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error)
}

func (m *mockAuditStoreHandler) Create(_ context.Context, log *models.AuditLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(log)
	}
	return nil
}
func (m *mockAuditStoreHandler) GetByUserID(_ context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(userID, limit, offset)
	}
	return nil, nil
}
func (m *mockAuditStoreHandler) GetByAction(_ context.Context, _ string, _, _ int) ([]*models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditStoreHandler) GetFailedLoginAttempts(_ context.Context, _ string, _ int) ([]*models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditStoreHandler) List(_ context.Context, limit, offset int) ([]*models.AuditLog, error) {
	if m.ListFunc != nil {
		return m.ListFunc(limit, offset)
	}
	return nil, nil
}
func (m *mockAuditStoreHandler) Count(_ context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc()
	}
	return 0, nil
}
func (m *mockAuditStoreHandler) DeleteOlderThan(_ context.Context, _ int) error {
	return nil
}
func (m *mockAuditStoreHandler) CountByActionSince(_ context.Context, action models.AuditAction, since time.Time) (int, error) {
	if m.CountByActionSinceFunc != nil {
		return m.CountByActionSinceFunc(action, since)
	}
	return 0, nil
}
func (m *mockAuditStoreHandler) ListByApp(_ context.Context, appID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error) {
	if m.ListByAppFunc != nil {
		return m.ListByAppFunc(appID, limit, offset)
	}
	return nil, 0, nil
}

// ===========================================================================
// APIKeyStore mock
// ===========================================================================

type mockAPIKeyStoreHandler struct {
	ListAllFunc   func() ([]*models.APIKey, error)
	RevokeFunc    func(id uuid.UUID) error
	ListByAppFunc func(appID uuid.UUID) ([]*models.APIKey, error)
}

func (m *mockAPIKeyStoreHandler) Create(_ context.Context, _ *models.APIKey) error {
	return nil
}
func (m *mockAPIKeyStoreHandler) GetByID(_ context.Context, _ uuid.UUID) (*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreHandler) GetByKeyHash(_ context.Context, _ string) (*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreHandler) GetByUserID(_ context.Context, _ uuid.UUID, _ ...service.APIKeyGetOption) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreHandler) Update(_ context.Context, _ *models.APIKey) error {
	return nil
}
func (m *mockAPIKeyStoreHandler) UpdateLastUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockAPIKeyStoreHandler) Revoke(_ context.Context, id uuid.UUID) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(id)
	}
	return nil
}
func (m *mockAPIKeyStoreHandler) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockAPIKeyStoreHandler) DeleteExpired(_ context.Context) error {
	return nil
}
func (m *mockAPIKeyStoreHandler) Count(_ context.Context, _ uuid.UUID, _ ...service.APIKeyGetOption) (int, error) {
	return 0, nil
}
func (m *mockAPIKeyStoreHandler) ListAll(_ context.Context) ([]*models.APIKey, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc()
	}
	return nil, nil
}
func (m *mockAPIKeyStoreHandler) GetByUserIDAndApp(_ context.Context, _, _ uuid.UUID) ([]*models.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyStoreHandler) ListByApp(_ context.Context, appID uuid.UUID) ([]*models.APIKey, error) {
	if m.ListByAppFunc != nil {
		return m.ListByAppFunc(appID)
	}
	return nil, nil
}

// ===========================================================================
// OAuthStore mock
// ===========================================================================

type mockOAuthStoreHandler struct {
	GetByUserIDFunc func(userID uuid.UUID) ([]*models.OAuthAccount, error)
}

func (m *mockOAuthStoreHandler) CreateOAuthAccount(_ context.Context, _ *models.OAuthAccount) error {
	return nil
}
func (m *mockOAuthStoreHandler) GetOAuthAccount(_ context.Context, _, _ string) (*models.OAuthAccount, error) {
	return nil, nil
}
func (m *mockOAuthStoreHandler) UpdateOAuthAccount(_ context.Context, _ *models.OAuthAccount) error {
	return nil
}
func (m *mockOAuthStoreHandler) DeleteOAuthAccount(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockOAuthStoreHandler) DeleteOAuthAccountsByProvider(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockOAuthStoreHandler) GetByUserID(_ context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(userID)
	}
	return nil, nil
}
func (m *mockOAuthStoreHandler) ListAll(_ context.Context) ([]*models.OAuthAccount, error) {
	return nil, nil
}

// ===========================================================================
// BackupCodeStore mock
// ===========================================================================

type mockBackupCodeStoreHandler struct{}

func (m *mockBackupCodeStoreHandler) CreateBatch(_ context.Context, _ []*models.BackupCode) error {
	return nil
}
func (m *mockBackupCodeStoreHandler) GetUnusedByUserID(_ context.Context, _ uuid.UUID) ([]*models.BackupCode, error) {
	return nil, nil
}
func (m *mockBackupCodeStoreHandler) CountUnusedByUserID(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockBackupCodeStoreHandler) MarkAsUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockBackupCodeStoreHandler) DeleteAllByUserID(_ context.Context, _ uuid.UUID) error {
	return nil
}

// ===========================================================================
// SessionStore mock (for SessionService used in AdvancedAdminHandler)
// ===========================================================================

type mockSessionStoreHandler struct {
	GetUserSessionsPaginatedFunc func(userID uuid.UUID, page, perPage int) ([]models.Session, int, error)
	GetAllSessionsPaginatedFunc  func(page, perPage int) ([]models.Session, int, error)
	GetSessionByIDFunc           func(id uuid.UUID) (*models.Session, error)
	RevokeSessionFunc            func(id uuid.UUID) error
	RevokeUserSessionFunc        func(userID, sessionID uuid.UUID) error
	RevokeAllUserSessionsFunc    func(userID uuid.UUID, except *uuid.UUID) error
	GetSessionStatsFunc          func() (*models.SessionStats, error)
	GetUserSessionsFunc          func(userID uuid.UUID) ([]models.Session, error)
	GetAppSessionsPaginatedFunc  func(appID uuid.UUID, page, perPage int) ([]models.Session, int, error)
}

func (m *mockSessionStoreHandler) CreateSession(_ context.Context, _ *models.Session) error {
	return nil
}
func (m *mockSessionStoreHandler) GetSessionByID(_ context.Context, id uuid.UUID) (*models.Session, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(id)
	}
	return nil, nil
}
func (m *mockSessionStoreHandler) GetSessionByTokenHash(_ context.Context, _ string) (*models.Session, error) {
	return nil, nil
}
func (m *mockSessionStoreHandler) GetUserSessions(_ context.Context, userID uuid.UUID) ([]models.Session, error) {
	if m.GetUserSessionsFunc != nil {
		return m.GetUserSessionsFunc(userID)
	}
	return nil, nil
}
func (m *mockSessionStoreHandler) GetUserSessionsPaginated(_ context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	if m.GetUserSessionsPaginatedFunc != nil {
		return m.GetUserSessionsPaginatedFunc(userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockSessionStoreHandler) GetAllSessionsPaginated(_ context.Context, page, perPage int) ([]models.Session, int, error) {
	if m.GetAllSessionsPaginatedFunc != nil {
		return m.GetAllSessionsPaginatedFunc(page, perPage)
	}
	return nil, 0, nil
}
func (m *mockSessionStoreHandler) RevokeSession(_ context.Context, id uuid.UUID) error {
	if m.RevokeSessionFunc != nil {
		return m.RevokeSessionFunc(id)
	}
	return nil
}
func (m *mockSessionStoreHandler) RevokeUserSession(_ context.Context, userID, sessionID uuid.UUID) error {
	if m.RevokeUserSessionFunc != nil {
		return m.RevokeUserSessionFunc(userID, sessionID)
	}
	return nil
}
func (m *mockSessionStoreHandler) RevokeAllUserSessions(_ context.Context, userID uuid.UUID, except *uuid.UUID) error {
	if m.RevokeAllUserSessionsFunc != nil {
		return m.RevokeAllUserSessionsFunc(userID, except)
	}
	return nil
}
func (m *mockSessionStoreHandler) UpdateSessionName(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockSessionStoreHandler) UpdateSessionAccessTokenHash(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockSessionStoreHandler) RefreshSessionTokens(_ context.Context, _, _, _ string, _ time.Time) error {
	return nil
}
func (m *mockSessionStoreHandler) GetSessionStats(_ context.Context) (*models.SessionStats, error) {
	if m.GetSessionStatsFunc != nil {
		return m.GetSessionStatsFunc()
	}
	return nil, nil
}
func (m *mockSessionStoreHandler) DeleteExpiredSessions(_ context.Context, _ time.Duration) error {
	return nil
}
func (m *mockSessionStoreHandler) GetUserSessionsByApp(_ context.Context, _, _ uuid.UUID) ([]models.Session, error) {
	return nil, nil
}
func (m *mockSessionStoreHandler) GetAppSessionsPaginated(_ context.Context, appID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	if m.GetAppSessionsPaginatedFunc != nil {
		return m.GetAppSessionsPaginatedFunc(appID, page, perPage)
	}
	return nil, 0, nil
}

// ===========================================================================
// IPFilterStore mock
// ===========================================================================

type mockIPFilterStoreHandler struct {
	ListIPFiltersFunc  func(page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error)
	CreateIPFilterFunc func(filter *models.IPFilter) error
	DeleteIPFilterFunc func(id uuid.UUID) error
}

func (m *mockIPFilterStoreHandler) CreateIPFilter(_ context.Context, filter *models.IPFilter) error {
	if m.CreateIPFilterFunc != nil {
		return m.CreateIPFilterFunc(filter)
	}
	return nil
}
func (m *mockIPFilterStoreHandler) GetIPFilterByID(_ context.Context, _ uuid.UUID) (*models.IPFilter, error) {
	return nil, nil
}
func (m *mockIPFilterStoreHandler) ListIPFilters(_ context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
	if m.ListIPFiltersFunc != nil {
		return m.ListIPFiltersFunc(page, perPage, filterType)
	}
	return nil, 0, nil
}
func (m *mockIPFilterStoreHandler) UpdateIPFilter(_ context.Context, _ uuid.UUID, _ string, _ bool) error {
	return nil
}
func (m *mockIPFilterStoreHandler) DeleteIPFilter(_ context.Context, id uuid.UUID) error {
	if m.DeleteIPFilterFunc != nil {
		return m.DeleteIPFilterFunc(id)
	}
	return nil
}
func (m *mockIPFilterStoreHandler) GetActiveIPFilters(_ context.Context) ([]models.IPFilter, error) {
	return nil, nil
}

// ===========================================================================
// BlackListStore mock (for SessionService)
// ===========================================================================

type mockBlackListStoreHandler struct{}

func (m *mockBlackListStoreHandler) IsBlacklisted(_ context.Context, _ string) bool {
	return false
}
func (m *mockBlackListStoreHandler) AddToBlacklist(_ context.Context, _ string, _ *uuid.UUID, _ time.Duration) error {
	return nil
}
func (m *mockBlackListStoreHandler) AddAccessToken(_ context.Context, _ string, _ *uuid.UUID) error {
	return nil
}
func (m *mockBlackListStoreHandler) AddRefreshToken(_ context.Context, _ string, _ *uuid.UUID) error {
	return nil
}
func (m *mockBlackListStoreHandler) BlacklistSessionTokens(_ context.Context, _ *models.Session) error {
	return nil
}
func (m *mockBlackListStoreHandler) BlacklistAllUserSessions(_ context.Context, _ uuid.UUID) error {
	return nil
}

// ===========================================================================
// ApplicationStore mock (for AdminService)
// ===========================================================================

type mockAppStoreHandler struct{}

func (m *mockAppStoreHandler) CreateApplication(_ context.Context, _ *models.Application) error {
	return nil
}
func (m *mockAppStoreHandler) GetApplicationByID(_ context.Context, _ uuid.UUID) (*models.Application, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) GetApplicationByName(_ context.Context, _ string) (*models.Application, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) UpdateApplication(_ context.Context, _ *models.Application) error {
	return nil
}
func (m *mockAppStoreHandler) DeleteApplication(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockAppStoreHandler) ListApplications(_ context.Context, _, _ int, _ *bool) ([]*models.Application, int, error) {
	return nil, 0, nil
}
func (m *mockAppStoreHandler) GetBySecretHash(_ context.Context, _ string) (*models.Application, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) GetBranding(_ context.Context, _ uuid.UUID) (*models.ApplicationBranding, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) CreateOrUpdateBranding(_ context.Context, _ *models.ApplicationBranding) error {
	return nil
}
func (m *mockAppStoreHandler) CreateUserProfile(_ context.Context, _ *models.UserApplicationProfile) error {
	return nil
}
func (m *mockAppStoreHandler) GetUserProfile(_ context.Context, _, _ uuid.UUID) (*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) UpdateUserProfile(_ context.Context, _ *models.UserApplicationProfile) error {
	return nil
}
func (m *mockAppStoreHandler) DeleteUserProfile(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockAppStoreHandler) ListUserProfiles(_ context.Context, _ uuid.UUID) ([]*models.UserApplicationProfile, error) {
	return nil, nil
}
func (m *mockAppStoreHandler) ListApplicationUsers(_ context.Context, _ uuid.UUID, _, _ int) ([]*models.UserApplicationProfile, int, error) {
	return nil, 0, nil
}
func (m *mockAppStoreHandler) UpdateLastAccess(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockAppStoreHandler) BanUserFromApplication(_ context.Context, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockAppStoreHandler) UnbanUserFromApplication(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
