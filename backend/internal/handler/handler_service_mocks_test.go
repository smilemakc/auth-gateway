package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// ===========================================================================
// mockAPIKeyServicer
// ===========================================================================

type mockAPIKeyServicer struct {
	GenerateAPIKeyFunc func() (string, error)
	CreateFunc         func(userID uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error)
	ValidateAPIKeyFunc func(plainKey string) (*models.APIKey, *models.User, error)
	GetByIDFunc        func(userID, apiKeyID uuid.UUID) (*models.APIKey, error)
	ListFunc           func(userID uuid.UUID) ([]*models.APIKey, error)
	UpdateFunc         func(userID, apiKeyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error)
	RevokeFunc         func(userID, apiKeyID uuid.UUID, ip, userAgent string) error
	DeleteFunc         func(userID, apiKeyID uuid.UUID, ip, userAgent string) error
	HasScopeFunc       func(apiKey *models.APIKey, scope models.APIKeyScope) bool
}

func (m *mockAPIKeyServicer) GenerateAPIKey() (string, error) {
	if m.GenerateAPIKeyFunc != nil {
		return m.GenerateAPIKeyFunc()
	}
	return "agw_test_key", nil
}

func (m *mockAPIKeyServicer) Create(_ context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(userID, req, ip, userAgent)
	}
	return nil, nil
}

func (m *mockAPIKeyServicer) ValidateAPIKey(_ context.Context, plainKey string) (*models.APIKey, *models.User, error) {
	if m.ValidateAPIKeyFunc != nil {
		return m.ValidateAPIKeyFunc(plainKey)
	}
	return nil, nil, nil
}

func (m *mockAPIKeyServicer) GetByID(_ context.Context, userID, apiKeyID uuid.UUID) (*models.APIKey, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(userID, apiKeyID)
	}
	return nil, nil
}

func (m *mockAPIKeyServicer) List(_ context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	if m.ListFunc != nil {
		return m.ListFunc(userID)
	}
	return nil, nil
}

func (m *mockAPIKeyServicer) Update(_ context.Context, userID, apiKeyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(userID, apiKeyID, req, ip, userAgent)
	}
	return nil, nil
}

func (m *mockAPIKeyServicer) Revoke(_ context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(userID, apiKeyID, ip, userAgent)
	}
	return nil
}

func (m *mockAPIKeyServicer) Delete(_ context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(userID, apiKeyID, ip, userAgent)
	}
	return nil
}

func (m *mockAPIKeyServicer) HasScope(apiKey *models.APIKey, scope models.APIKeyScope) bool {
	if m.HasScopeFunc != nil {
		return m.HasScopeFunc(apiKey, scope)
	}
	return false
}

// ===========================================================================
// mockTwoFactorServicer
// ===========================================================================

type mockTwoFactorServicer struct {
	SetupTOTPFunc              func(userID uuid.UUID, password string) (*models.TwoFactorSetupResponse, error)
	VerifyTOTPSetupFunc        func(userID uuid.UUID, code string) error
	VerifyTOTPFunc             func(userID uuid.UUID, code string) (bool, error)
	DisableTOTPFunc            func(userID uuid.UUID, password, code string) error
	GetStatusFunc              func(userID uuid.UUID) (*models.TwoFactorStatusResponse, error)
	RegenerateBackupCodesFunc  func(userID uuid.UUID, password string) ([]string, error)
}

func (m *mockTwoFactorServicer) SetupTOTP(_ context.Context, userID uuid.UUID, password string) (*models.TwoFactorSetupResponse, error) {
	if m.SetupTOTPFunc != nil {
		return m.SetupTOTPFunc(userID, password)
	}
	return nil, nil
}

func (m *mockTwoFactorServicer) VerifyTOTPSetup(_ context.Context, userID uuid.UUID, code string) error {
	if m.VerifyTOTPSetupFunc != nil {
		return m.VerifyTOTPSetupFunc(userID, code)
	}
	return nil
}

func (m *mockTwoFactorServicer) VerifyTOTP(_ context.Context, userID uuid.UUID, code string) (bool, error) {
	if m.VerifyTOTPFunc != nil {
		return m.VerifyTOTPFunc(userID, code)
	}
	return false, nil
}

func (m *mockTwoFactorServicer) DisableTOTP(_ context.Context, userID uuid.UUID, password, code string) error {
	if m.DisableTOTPFunc != nil {
		return m.DisableTOTPFunc(userID, password, code)
	}
	return nil
}

func (m *mockTwoFactorServicer) GetStatus(_ context.Context, userID uuid.UUID) (*models.TwoFactorStatusResponse, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(userID)
	}
	return nil, nil
}

func (m *mockTwoFactorServicer) RegenerateBackupCodes(_ context.Context, userID uuid.UUID, password string) ([]string, error) {
	if m.RegenerateBackupCodesFunc != nil {
		return m.RegenerateBackupCodesFunc(userID, password)
	}
	return nil, nil
}

// ===========================================================================
// mockTokenExchangeServicer
// ===========================================================================

type mockTokenExchangeServicer struct {
	CreateExchangeFunc func(req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error)
	RedeemExchangeFunc func(req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error)
}

func (m *mockTokenExchangeServicer) CreateExchange(_ context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
	if m.CreateExchangeFunc != nil {
		return m.CreateExchangeFunc(req, sourceAppID)
	}
	return nil, nil
}

func (m *mockTokenExchangeServicer) RedeemExchange(_ context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
	if m.RedeemExchangeFunc != nil {
		return m.RedeemExchangeFunc(req, redeemingAppID)
	}
	return nil, nil
}

// ===========================================================================
// mockOTPServicer
// ===========================================================================

type mockOTPServicer struct {
	GenerateOTPCodeFunc func() (string, error)
	SendOTPFunc         func(req *models.SendOTPRequest) error
	VerifyOTPFunc       func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error)
}

func (m *mockOTPServicer) GenerateOTPCode() (string, error) {
	if m.GenerateOTPCodeFunc != nil {
		return m.GenerateOTPCodeFunc()
	}
	return "123456", nil
}

func (m *mockOTPServicer) SendOTP(_ context.Context, req *models.SendOTPRequest) error {
	if m.SendOTPFunc != nil {
		return m.SendOTPFunc(req)
	}
	return nil
}

func (m *mockOTPServicer) VerifyOTP(_ context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	if m.VerifyOTPFunc != nil {
		return m.VerifyOTPFunc(req)
	}
	return nil, nil
}

func (m *mockOTPServicer) CleanupExpiredOTPs() error {
	return nil
}

// ===========================================================================
// mockAuthServicer
// ===========================================================================

type mockAuthServicer struct {
	SignUpFunc                          func(req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	SignInFunc                          func(req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error)
	Verify2FALoginFunc                  func(twoFactorToken, code, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	RefreshTokenFunc                    func(refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	LogoutFunc                          func(accessToken, ip, userAgent string) error
	ChangePasswordFunc                  func(userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error
	ResetPasswordFunc                   func(userID uuid.UUID, newPassword, ip, userAgent string) error
	InitPasswordlessRegistrationFunc    func(req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error
	CompletePasswordlessRegistrationFunc func(req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error)
	GenerateTokensForUserFunc           func(user *models.User, ip, userAgent string) (*models.AuthResponse, error)
}

func (m *mockAuthServicer) SignUp(_ context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	if m.SignUpFunc != nil {
		return m.SignUpFunc(req, ip, userAgent, deviceInfo, appID)
	}
	return nil, nil
}

func (m *mockAuthServicer) SignIn(_ context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	if m.SignInFunc != nil {
		return m.SignInFunc(req, ip, userAgent, deviceInfo, appID)
	}
	return nil, nil
}

func (m *mockAuthServicer) Verify2FALogin(_ context.Context, twoFactorToken, code, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	if m.Verify2FALoginFunc != nil {
		return m.Verify2FALoginFunc(twoFactorToken, code, ip, userAgent, deviceInfo)
	}
	return nil, nil
}

func (m *mockAuthServicer) RefreshToken(_ context.Context, refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(refreshToken, ip, userAgent, deviceInfo)
	}
	return nil, nil
}

func (m *mockAuthServicer) Logout(_ context.Context, accessToken, ip, userAgent string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(accessToken, ip, userAgent)
	}
	return nil
}

func (m *mockAuthServicer) ChangePassword(_ context.Context, userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(userID, oldPassword, newPassword, ip, userAgent)
	}
	return nil
}

func (m *mockAuthServicer) ResetPassword(_ context.Context, userID uuid.UUID, newPassword, ip, userAgent string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(userID, newPassword, ip, userAgent)
	}
	return nil
}

func (m *mockAuthServicer) InitPasswordlessRegistration(_ context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error {
	if m.InitPasswordlessRegistrationFunc != nil {
		return m.InitPasswordlessRegistrationFunc(req, ip, userAgent)
	}
	return nil
}

func (m *mockAuthServicer) CompletePasswordlessRegistration(_ context.Context, req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	if m.CompletePasswordlessRegistrationFunc != nil {
		return m.CompletePasswordlessRegistrationFunc(req, ip, userAgent, deviceInfo)
	}
	return nil, nil
}

func (m *mockAuthServicer) GenerateTokensForUser(_ context.Context, user *models.User, ip, userAgent string) (*models.AuthResponse, error) {
	if m.GenerateTokensForUserFunc != nil {
		return m.GenerateTokensForUserFunc(user, ip, userAgent)
	}
	return nil, nil
}

// ===========================================================================
// mockUserServicer
// ===========================================================================

type mockUserServicer struct {
	GetProfileFunc    func(userID uuid.UUID) (*models.User, error)
	UpdateProfileFunc func(userID uuid.UUID, req *models.UpdateUserRequest, ip, userAgent string) (*models.User, error)
	GetByIDFunc       func(userID uuid.UUID) (*models.User, error)
	GetByEmailFunc    func(email string) (*models.User, error)
	ListFunc          func(limit, offset int) ([]*models.User, error)
	CountFunc         func() (int, error)
}

func (m *mockUserServicer) GetProfile(_ context.Context, userID uuid.UUID) (*models.User, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(userID)
	}
	return nil, nil
}

func (m *mockUserServicer) UpdateProfile(_ context.Context, userID uuid.UUID, req *models.UpdateUserRequest, ip, userAgent string) (*models.User, error) {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(userID, req, ip, userAgent)
	}
	return nil, nil
}

func (m *mockUserServicer) GetByID(_ context.Context, userID uuid.UUID) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(userID)
	}
	return nil, nil
}

func (m *mockUserServicer) GetByEmail(_ context.Context, email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

func (m *mockUserServicer) List(_ context.Context, limit, offset int) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(limit, offset)
	}
	return nil, nil
}

func (m *mockUserServicer) Count(_ context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc()
	}
	return 0, nil
}

// ===========================================================================
// mockWebhookServicer
// ===========================================================================

type mockWebhookServicer struct {
	CreateWebhookFunc         func(req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error)
	GetWebhookFunc            func(id uuid.UUID) (*models.Webhook, error)
	ListWebhooksFunc          func(page, perPage int) (*models.WebhookListResponse, error)
	UpdateWebhookFunc         func(id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error
	DeleteWebhookFunc         func(id uuid.UUID, deletedBy uuid.UUID) error
	TriggerWebhookFunc        func(eventType string, data map[string]interface{}) error
	ListWebhookDeliveriesFunc func(webhookID uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error)
	TestWebhookFunc           func(id uuid.UUID, req *models.TestWebhookRequest) error
	ListWebhooksByAppFunc     func(appID uuid.UUID) ([]*models.Webhook, error)
}

func (m *mockWebhookServicer) CreateWebhook(_ context.Context, req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error) {
	if m.CreateWebhookFunc != nil {
		return m.CreateWebhookFunc(req, createdBy)
	}
	return nil, "", nil
}

func (m *mockWebhookServicer) GetWebhook(_ context.Context, id uuid.UUID) (*models.Webhook, error) {
	if m.GetWebhookFunc != nil {
		return m.GetWebhookFunc(id)
	}
	return nil, nil
}

func (m *mockWebhookServicer) ListWebhooks(_ context.Context, page, perPage int) (*models.WebhookListResponse, error) {
	if m.ListWebhooksFunc != nil {
		return m.ListWebhooksFunc(page, perPage)
	}
	return nil, nil
}

func (m *mockWebhookServicer) UpdateWebhook(_ context.Context, id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error {
	if m.UpdateWebhookFunc != nil {
		return m.UpdateWebhookFunc(id, req, updatedBy)
	}
	return nil
}

func (m *mockWebhookServicer) DeleteWebhook(_ context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	if m.DeleteWebhookFunc != nil {
		return m.DeleteWebhookFunc(id, deletedBy)
	}
	return nil
}

func (m *mockWebhookServicer) TriggerWebhook(_ context.Context, eventType string, data map[string]interface{}) error {
	if m.TriggerWebhookFunc != nil {
		return m.TriggerWebhookFunc(eventType, data)
	}
	return nil
}

func (m *mockWebhookServicer) ListWebhookDeliveries(_ context.Context, webhookID uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error) {
	if m.ListWebhookDeliveriesFunc != nil {
		return m.ListWebhookDeliveriesFunc(webhookID, page, perPage)
	}
	return nil, nil
}

func (m *mockWebhookServicer) TestWebhook(_ context.Context, id uuid.UUID, req *models.TestWebhookRequest) error {
	if m.TestWebhookFunc != nil {
		return m.TestWebhookFunc(id, req)
	}
	return nil
}

func (m *mockWebhookServicer) GetAvailableEvents() []string {
	return []string{}
}

func (m *mockWebhookServicer) ListWebhooksByApp(_ context.Context, appID uuid.UUID) ([]*models.Webhook, error) {
	if m.ListWebhooksByAppFunc != nil {
		return m.ListWebhooksByAppFunc(appID)
	}
	return nil, nil
}

// ===========================================================================
// mockBulkServicer
// ===========================================================================

type mockBulkServicer struct {
	BulkCreateUsersFunc func(req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error)
	BulkUpdateUsersFunc func(req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error)
	BulkDeleteUsersFunc func(req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error)
	BulkAssignRolesFunc func(req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error)
}

func (m *mockBulkServicer) BulkCreateUsers(_ context.Context, req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error) {
	if m.BulkCreateUsersFunc != nil {
		return m.BulkCreateUsersFunc(req)
	}
	return nil, nil
}

func (m *mockBulkServicer) BulkUpdateUsers(_ context.Context, req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error) {
	if m.BulkUpdateUsersFunc != nil {
		return m.BulkUpdateUsersFunc(req)
	}
	return nil, nil
}

func (m *mockBulkServicer) BulkDeleteUsers(_ context.Context, req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error) {
	if m.BulkDeleteUsersFunc != nil {
		return m.BulkDeleteUsersFunc(req)
	}
	return nil, nil
}

func (m *mockBulkServicer) BulkAssignRoles(_ context.Context, req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error) {
	if m.BulkAssignRolesFunc != nil {
		return m.BulkAssignRolesFunc(req, assignedBy)
	}
	return nil, nil
}

// ===========================================================================
// mockSMSServicer
// ===========================================================================

type mockSMSServicer struct {
	GenerateOTPCodeFunc func() (string, error)
	SendOTPFunc         func(req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error)
	VerifyOTPFunc       func(req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error)
	GetStatsFunc        func() (*models.SMSStatsResponse, error)
}

func (m *mockSMSServicer) GenerateOTPCode() (string, error) {
	if m.GenerateOTPCodeFunc != nil {
		return m.GenerateOTPCodeFunc()
	}
	return "123456", nil
}

func (m *mockSMSServicer) SendOTP(_ context.Context, req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error) {
	if m.SendOTPFunc != nil {
		return m.SendOTPFunc(req, ipAddress)
	}
	return nil, nil
}

func (m *mockSMSServicer) VerifyOTP(_ context.Context, req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error) {
	if m.VerifyOTPFunc != nil {
		return m.VerifyOTPFunc(req)
	}
	return nil, nil
}

func (m *mockSMSServicer) GetStats(_ context.Context) (*models.SMSStatsResponse, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc()
	}
	return nil, nil
}

func (m *mockSMSServicer) CleanupOldLogs(_ context.Context, duration time.Duration) (int64, error) {
	return 0, nil
}

// ===========================================================================
// mockMigrationServicer
// ===========================================================================

type mockMigrationServicer struct {
	ImportUsersFunc         func(appID uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error)
	ImportOAuthAccountsFunc func(entries []models.ImportOAuthEntry) (*models.ImportResult, error)
	ImportRolesFunc         func(appID uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error)
}

func (m *mockMigrationServicer) ImportUsers(_ context.Context, appID uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error) {
	if m.ImportUsersFunc != nil {
		return m.ImportUsersFunc(appID, entries)
	}
	return nil, nil
}

func (m *mockMigrationServicer) ImportOAuthAccounts(_ context.Context, entries []models.ImportOAuthEntry) (*models.ImportResult, error) {
	if m.ImportOAuthAccountsFunc != nil {
		return m.ImportOAuthAccountsFunc(entries)
	}
	return nil, nil
}

func (m *mockMigrationServicer) ImportRoles(_ context.Context, appID uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error) {
	if m.ImportRolesFunc != nil {
		return m.ImportRolesFunc(appID, entries)
	}
	return nil, nil
}

// ===========================================================================
// mockGroupServicer
// ===========================================================================

type mockGroupServicer struct {
	CreateGroupFunc               func(req *models.CreateGroupRequest) (*models.Group, error)
	GetGroupFunc                  func(id uuid.UUID) (*models.Group, error)
	ListGroupsFunc                func(page, pageSize int) (*models.GroupListResponse, error)
	UpdateGroupFunc               func(id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error)
	DeleteGroupFunc               func(id uuid.UUID) error
	AddGroupMembersFunc           func(groupID uuid.UUID, userIDs []uuid.UUID) error
	RemoveGroupMemberFunc         func(groupID, userID uuid.UUID) error
	GetGroupMembersFunc           func(groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error)
	GetUserGroupsFunc             func(userID uuid.UUID) ([]*models.Group, error)
	GetGroupMemberCountFunc       func(groupID uuid.UUID) (int, error)
	GetGroupPermissionsFunc       func(groupID uuid.UUID) ([]uuid.UUID, error)
	SyncDynamicGroupMembersFunc   func(groupID uuid.UUID) error
}

func (m *mockGroupServicer) CreateGroup(_ context.Context, req *models.CreateGroupRequest) (*models.Group, error) {
	if m.CreateGroupFunc != nil {
		return m.CreateGroupFunc(req)
	}
	return nil, nil
}

func (m *mockGroupServicer) GetGroup(_ context.Context, id uuid.UUID) (*models.Group, error) {
	if m.GetGroupFunc != nil {
		return m.GetGroupFunc(id)
	}
	return nil, nil
}

func (m *mockGroupServicer) ListGroups(_ context.Context, page, pageSize int) (*models.GroupListResponse, error) {
	if m.ListGroupsFunc != nil {
		return m.ListGroupsFunc(page, pageSize)
	}
	return nil, nil
}

func (m *mockGroupServicer) UpdateGroup(_ context.Context, id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error) {
	if m.UpdateGroupFunc != nil {
		return m.UpdateGroupFunc(id, req)
	}
	return nil, nil
}

func (m *mockGroupServicer) DeleteGroup(_ context.Context, id uuid.UUID) error {
	if m.DeleteGroupFunc != nil {
		return m.DeleteGroupFunc(id)
	}
	return nil
}

func (m *mockGroupServicer) AddGroupMembers(_ context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	if m.AddGroupMembersFunc != nil {
		return m.AddGroupMembersFunc(groupID, userIDs)
	}
	return nil
}

func (m *mockGroupServicer) RemoveGroupMember(_ context.Context, groupID, userID uuid.UUID) error {
	if m.RemoveGroupMemberFunc != nil {
		return m.RemoveGroupMemberFunc(groupID, userID)
	}
	return nil
}

func (m *mockGroupServicer) GetGroupMembers(_ context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
	if m.GetGroupMembersFunc != nil {
		return m.GetGroupMembersFunc(groupID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockGroupServicer) GetUserGroups(_ context.Context, userID uuid.UUID) ([]*models.Group, error) {
	if m.GetUserGroupsFunc != nil {
		return m.GetUserGroupsFunc(userID)
	}
	return nil, nil
}

func (m *mockGroupServicer) GetGroupMemberCount(_ context.Context, groupID uuid.UUID) (int, error) {
	if m.GetGroupMemberCountFunc != nil {
		return m.GetGroupMemberCountFunc(groupID)
	}
	return 0, nil
}

func (m *mockGroupServicer) EvaluateDynamicGroupMembers(_ context.Context, group *models.Group, allUsers []*models.User) []uuid.UUID {
	return nil
}

func (m *mockGroupServicer) GetGroupPermissions(_ context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetGroupPermissionsFunc != nil {
		return m.GetGroupPermissionsFunc(groupID)
	}
	return nil, nil
}

func (m *mockGroupServicer) SyncDynamicGroupMembers(_ context.Context, groupID uuid.UUID) error {
	if m.SyncDynamicGroupMembersFunc != nil {
		return m.SyncDynamicGroupMembersFunc(groupID)
	}
	return nil
}

// ===========================================================================
// mockRedisServicer
// ===========================================================================

type mockRedisServicer struct {
	GetFunc            func(key string) (string, error)
	SetFunc            func(key string, value interface{}, expiration time.Duration) error
	DeleteFunc         func(keys ...string) error
	IsBlacklistedFunc  func(tokenHash string) (bool, error)
	AddToBlacklistFunc func(tokenHash string, expiration time.Duration) error
}

func (m *mockRedisServicer) Close() error {
	return nil
}

func (m *mockRedisServicer) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, expiration)
	}
	return nil
}

func (m *mockRedisServicer) Get(_ context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return "", nil
}

func (m *mockRedisServicer) Delete(_ context.Context, keys ...string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(keys...)
	}
	return nil
}

func (m *mockRedisServicer) Exists(_ context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockRedisServicer) Increment(_ context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *mockRedisServicer) SetNX(_ context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return false, nil
}

func (m *mockRedisServicer) Expire(_ context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *mockRedisServicer) Health(_ context.Context) error {
	return nil
}

func (m *mockRedisServicer) AddToBlacklist(_ context.Context, tokenHash string, expiration time.Duration) error {
	if m.AddToBlacklistFunc != nil {
		return m.AddToBlacklistFunc(tokenHash, expiration)
	}
	return nil
}

func (m *mockRedisServicer) IsBlacklisted(_ context.Context, tokenHash string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(tokenHash)
	}
	return false, nil
}

func (m *mockRedisServicer) IncrementRateLimit(_ context.Context, key string, window time.Duration) (int64, error) {
	return 0, nil
}

func (m *mockRedisServicer) StorePendingRegistration(_ context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
	return nil
}

func (m *mockRedisServicer) GetPendingRegistration(_ context.Context, identifier string) (*models.PendingRegistration, error) {
	return nil, nil
}

func (m *mockRedisServicer) DeletePendingRegistration(_ context.Context, identifier string) error {
	return nil
}

func (m *mockRedisServicer) SAdd(_ context.Context, key string, members ...string) error {
	return nil
}

func (m *mockRedisServicer) SIsMember(_ context.Context, key string, member string) (bool, error) {
	return false, nil
}

func (m *mockRedisServicer) SMembers(_ context.Context, key string) ([]string, error) {
	return nil, nil
}

// ===========================================================================
// mockEmailProfileServicer
// ===========================================================================

type mockEmailProfileServicer struct {
	SendEmailFunc    func(profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error
	SendOTPEmailFunc func(profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error
}

func (m *mockEmailProfileServicer) CreateProvider(_ context.Context, req *models.CreateEmailProviderRequest) (*models.EmailProvider, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) GetProvider(_ context.Context, id uuid.UUID) (*models.EmailProviderResponse, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) ListProviders(_ context.Context, appID *uuid.UUID) ([]*models.EmailProviderResponse, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) UpdateProvider(_ context.Context, id uuid.UUID, req *models.UpdateEmailProviderRequest) error {
	return nil
}

func (m *mockEmailProfileServicer) DeleteProvider(_ context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEmailProfileServicer) TestProvider(_ context.Context, id uuid.UUID, testEmail string) error {
	return nil
}

func (m *mockEmailProfileServicer) CreateProfile(_ context.Context, req *models.CreateEmailProfileRequest) (*models.EmailProfile, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) GetProfile(_ context.Context, id uuid.UUID) (*models.EmailProfile, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) ListProfiles(_ context.Context, appID *uuid.UUID) ([]*models.EmailProfile, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) UpdateProfile(_ context.Context, id uuid.UUID, req *models.UpdateEmailProfileRequest) error {
	return nil
}

func (m *mockEmailProfileServicer) DeleteProfile(_ context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEmailProfileServicer) SetDefaultProfile(_ context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEmailProfileServicer) GetProfileTemplates(_ context.Context, profileID uuid.UUID) ([]*models.EmailProfileTemplate, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) SetProfileTemplate(_ context.Context, profileID uuid.UUID, otpType string, templateID uuid.UUID) error {
	return nil
}

func (m *mockEmailProfileServicer) RemoveProfileTemplate(_ context.Context, profileID uuid.UUID, otpType string) error {
	return nil
}

func (m *mockEmailProfileServicer) SendOTPEmail(_ context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error {
	if m.SendOTPEmailFunc != nil {
		return m.SendOTPEmailFunc(profileID, applicationID, toEmail, otpType, code)
	}
	return nil
}

func (m *mockEmailProfileServicer) SendEmail(_ context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(profileID, applicationID, toEmail, templateType, variables)
	}
	return nil
}

func (m *mockEmailProfileServicer) GetProfileStats(_ context.Context, profileID uuid.UUID) (*models.EmailStatsResponse, error) {
	return nil, nil
}

func (m *mockEmailProfileServicer) TestProfile(_ context.Context, profileID uuid.UUID, testEmail string) error {
	return nil
}

// ===========================================================================
// mockApplicationServicer
// ===========================================================================

type mockApplicationServicer struct {
	CreateApplicationFunc   func(req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error)
	GetByIDFunc             func(id uuid.UUID) (*models.Application, error)
	GetByNameFunc           func(name string) (*models.Application, error)
	UpdateApplicationFunc   func(id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error)
	DeleteApplicationFunc   func(id uuid.UUID) error
	ListApplicationsFunc    func(page, perPage int, isActive *bool) (*models.ApplicationListResponse, error)
	GetBrandingFunc         func(applicationID uuid.UUID) (*models.ApplicationBranding, error)
	UpdateBrandingFunc      func(applicationID uuid.UUID, req *models.UpdateApplicationBrandingRequest) (*models.ApplicationBranding, error)
	GetUserProfileFunc      func(userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error)
	UpdateUserProfileFunc   func(userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error)
	ListUserProfilesFunc    func(userID uuid.UUID) ([]*models.UserApplicationProfile, error)
	ListApplicationUsersFunc func(applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error)
	BanUserFunc             func(userID, applicationID, bannedBy uuid.UUID, reason string) error
	UnbanUserFunc           func(userID, applicationID uuid.UUID) error
	DeleteUserProfileFunc   func(userID, applicationID uuid.UUID) error
	CheckUserAccessFunc     func(userID, applicationID uuid.UUID) error
	ValidateSecretFunc      func(secret string) (*models.Application, error)
	GetAuthConfigFunc       func(app *models.Application) (*models.AuthConfigResponse, error)
	GenerateSecretFunc      func(appID uuid.UUID) (string, error)
	RotateSecretFunc        func(appID uuid.UUID) (string, error)
}

func (m *mockApplicationServicer) CreateApplication(_ context.Context, req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
	if m.CreateApplicationFunc != nil {
		return m.CreateApplicationFunc(req, ownerID)
	}
	return nil, "", nil
}

func (m *mockApplicationServicer) GetByID(_ context.Context, id uuid.UUID) (*models.Application, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *mockApplicationServicer) GetByName(_ context.Context, name string) (*models.Application, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(name)
	}
	return nil, nil
}

func (m *mockApplicationServicer) UpdateApplication(_ context.Context, id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error) {
	if m.UpdateApplicationFunc != nil {
		return m.UpdateApplicationFunc(id, req)
	}
	return nil, nil
}

func (m *mockApplicationServicer) DeleteApplication(_ context.Context, id uuid.UUID) error {
	if m.DeleteApplicationFunc != nil {
		return m.DeleteApplicationFunc(id)
	}
	return nil
}

func (m *mockApplicationServicer) ListApplications(_ context.Context, page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
	if m.ListApplicationsFunc != nil {
		return m.ListApplicationsFunc(page, perPage, isActive)
	}
	return nil, nil
}

func (m *mockApplicationServicer) GetBranding(_ context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	if m.GetBrandingFunc != nil {
		return m.GetBrandingFunc(applicationID)
	}
	return nil, nil
}

func (m *mockApplicationServicer) UpdateBranding(_ context.Context, applicationID uuid.UUID, req *models.UpdateApplicationBrandingRequest) (*models.ApplicationBranding, error) {
	if m.UpdateBrandingFunc != nil {
		return m.UpdateBrandingFunc(applicationID, req)
	}
	return nil, nil
}

func (m *mockApplicationServicer) GetOrCreateUserProfile(_ context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	return nil, nil
}

func (m *mockApplicationServicer) GetUserProfile(_ context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(userID, applicationID)
	}
	return nil, nil
}

func (m *mockApplicationServicer) UpdateUserProfile(_ context.Context, userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error) {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(userID, applicationID, req)
	}
	return nil, nil
}

func (m *mockApplicationServicer) ListUserProfiles(_ context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	if m.ListUserProfilesFunc != nil {
		return m.ListUserProfilesFunc(userID)
	}
	return nil, nil
}

func (m *mockApplicationServicer) ListApplicationUsers(_ context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error) {
	if m.ListApplicationUsersFunc != nil {
		return m.ListApplicationUsersFunc(applicationID, page, perPage)
	}
	return nil, nil
}

func (m *mockApplicationServicer) BanUser(_ context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	if m.BanUserFunc != nil {
		return m.BanUserFunc(userID, applicationID, bannedBy, reason)
	}
	return nil
}

func (m *mockApplicationServicer) UnbanUser(_ context.Context, userID, applicationID uuid.UUID) error {
	if m.UnbanUserFunc != nil {
		return m.UnbanUserFunc(userID, applicationID)
	}
	return nil
}

func (m *mockApplicationServicer) DeleteUserProfile(_ context.Context, userID, applicationID uuid.UUID) error {
	if m.DeleteUserProfileFunc != nil {
		return m.DeleteUserProfileFunc(userID, applicationID)
	}
	return nil
}

func (m *mockApplicationServicer) CheckUserAccess(_ context.Context, userID, applicationID uuid.UUID) error {
	if m.CheckUserAccessFunc != nil {
		return m.CheckUserAccessFunc(userID, applicationID)
	}
	return nil
}

func (m *mockApplicationServicer) IsAuthMethodAllowed(_ context.Context, appID uuid.UUID, method string) error {
	return nil
}

func (m *mockApplicationServicer) GenerateSecret(_ context.Context, appID uuid.UUID) (string, error) {
	if m.GenerateSecretFunc != nil {
		return m.GenerateSecretFunc(appID)
	}
	return "", nil
}

func (m *mockApplicationServicer) RotateSecret(_ context.Context, appID uuid.UUID) (string, error) {
	if m.RotateSecretFunc != nil {
		return m.RotateSecretFunc(appID)
	}
	return "", nil
}

func (m *mockApplicationServicer) ValidateSecret(_ context.Context, secret string) (*models.Application, error) {
	if m.ValidateSecretFunc != nil {
		return m.ValidateSecretFunc(secret)
	}
	return nil, nil
}

func (m *mockApplicationServicer) GetAuthConfig(_ context.Context, app *models.Application) (*models.AuthConfigResponse, error) {
	if m.GetAuthConfigFunc != nil {
		return m.GetAuthConfigFunc(app)
	}
	return nil, nil
}
