package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/uptrace/bun"
)

// AdminService provides admin operations
type AdminService struct {
	userRepo       UserStore
	apiKeyRepo     APIKeyStore
	auditRepo      AuditStore
	oauthRepo      OAuthStore
	rbacRepo       RBACStore
	backupCodeRepo BackupCodeStore
	appRepo        ApplicationStore
	bcryptCost     int
	db             TransactionDB
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo UserStore,
	apiKeyRepo APIKeyStore,
	auditRepo AuditStore,
	oauthRepo OAuthStore,
	rbacRepo RBACStore,
	backupCodeRepo BackupCodeStore,
	appRepo ApplicationStore,
	bcryptCost int,
	db TransactionDB,
) *AdminService {
	return &AdminService{
		userRepo:       userRepo,
		apiKeyRepo:     apiKeyRepo,
		auditRepo:      auditRepo,
		oauthRepo:      oauthRepo,
		rbacRepo:       rbacRepo,
		backupCodeRepo: backupCodeRepo,
		appRepo:        appRepo,
		bcryptCost:     bcryptCost,
		db:             db,
	}
}

// GetStats returns system statistics
func (s *AdminService) GetStats(ctx context.Context) (*models.AdminStatsResponse, error) {
	stats := &models.AdminStatsResponse{
		UsersByRole: make(map[string]int),
	}

	// Total users
	totalUsers, err := s.userRepo.Count(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats.TotalUsers = totalUsers

	// Get all users for detailed stats
	users, err := s.userRepo.List(ctx, UserListLimit(10000), UserListOffset(0))
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	for _, user := range users {
		if user.IsActive {
			stats.ActiveUsers++
		}
		if user.EmailVerified {
			stats.VerifiedEmailUsers++
		}
		if user.PhoneVerified {
			stats.VerifiedPhoneUsers++
		}
		if user.TOTPEnabled {
			stats.Users2FAEnabled++
		}

		// Recent signups
		if user.CreatedAt.After(yesterday) {
			stats.RecentSignups++
		}
	}

	// Aggregate users by role (note: users with multiple roles counted in each)
	stats.UsersByRole = make(map[string]int)
	for _, user := range users {
		roles, err := s.rbacRepo.GetUserRoles(ctx, user.ID)
		if err == nil {
			for _, role := range roles {
				stats.UsersByRole[role.Name]++
			}
		}
	}

	// API keys stats
	allAPIKeys, err := s.apiKeyRepo.ListAll(ctx)
	if err == nil {
		stats.TotalAPIKeys = len(allAPIKeys)
		for _, key := range allAPIKeys {
			if key.IsActive {
				stats.ActiveAPIKeys++
			}
		}
	}

	// OAuth accounts stats
	oauthAccounts, err := s.oauthRepo.ListAll(ctx)
	if err == nil {
		stats.TotalOAuthAccounts = len(oauthAccounts)
	}

	// Recent logins from audit logs
	recentLogins, err := s.auditRepo.CountByActionSince(ctx, models.ActionSignIn, yesterday)
	if err == nil {
		stats.RecentLogins = recentLogins
	}

	return stats, nil
}

func (s *AdminService) ListUsers(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminUserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var users []*models.User
	var total int
	var err error

	if appID != nil {
		var profiles []*models.UserApplicationProfile
		profiles, total, err = s.appRepo.ListApplicationUsers(ctx, *appID, page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to list users by app: %w", err)
		}
		users = make([]*models.User, 0, len(profiles))
		for _, p := range profiles {
			if p.User != nil {
				users = append(users, p.User)
			}
		}
	} else {
		offset := (page - 1) * pageSize
		users, err = s.userRepo.List(ctx, UserListLimit(pageSize), UserListOffset(offset), UserListWithRoles())
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}
		total, err = s.userRepo.Count(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to count users: %w", err)
		}
	}

	adminUsers := make([]*models.AdminUserResponse, 0, len(users))
	for _, user := range users {
		adminUser := s.userToAdminResponse(user)

		apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, user.ID)
		if err == nil {
			adminUser.APIKeysCount = len(apiKeys)
		}

		oauthAccounts, err := s.oauthRepo.GetByUserID(ctx, user.ID)
		if err == nil {
			adminUser.OAuthAccountsCount = len(oauthAccounts)
		}

		adminUsers = append(adminUsers, adminUser)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.AdminUserListResponse{
		Users:      adminUsers,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUser returns detailed user information
func (s *AdminService) GetUser(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil, UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	adminUser := s.userToAdminResponse(user)

	// Count API keys
	apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.APIKeysCount = len(apiKeys)
	}

	// Count OAuth accounts
	oauthAccounts, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.OAuthAccountsCount = len(oauthAccounts)
	}

	return adminUser, nil
}

// CreateUser creates a new user (admin only)
func (s *AdminService) CreateUser(ctx context.Context, req *models.AdminCreateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	// 1. Create user entity
	user := &models.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      req.Username,
		FullName:      req.FullName,
		AccountType:   req.AccountType,
		IsActive:      true,
		EmailVerified: true, // Admins create verified users by default
	}

	if user.AccountType == "" {
		user.AccountType = string(models.AccountTypeHuman)
	}

	// 2. Set password
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password, s.bcryptCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = hash
	} else {
		// Provide a dummy hash if no password provided (should be handled by validation, but safe fallback)
		// Or generated password?
		// For now, assume password is required by request validation model
	}

	// 3. Get roles to assign
	var roleIDs []uuid.UUID
	if len(req.RoleIDs) > 0 {
		roleIDs = req.RoleIDs
	} else {
		// Default role 'user'
		defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
		if err == nil {
			roleIDs = []uuid.UUID{defaultRole.ID}
		}
	}

	// 4. Create user and assign roles in a transaction
	err := s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		// Create user in DB
		if userRepo, ok := s.userRepo.(*repository.UserRepository); ok {
			if err := userRepo.CreateWithTx(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			// Fallback to non-transactional method if type assertion fails
			if err := s.userRepo.Create(ctx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		}

		// Assign roles
		if len(roleIDs) > 0 {
			if err := s.rbacRepo.SetUserRoles(ctx, user.ID, roleIDs, adminID); err != nil {
				// SetUserRoles already uses a transaction internally, so we can call it directly
				return fmt.Errorf("failed to assign roles: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Return created user
	return s.GetUser(ctx, user.ID)
}

// UpdateUser updates user information (admin only)
func (s *AdminService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.AdminUpdateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	// Update roles if provided
	if req.RoleIDs != nil {
		if len(*req.RoleIDs) == 0 {
			return nil, models.NewAppError(400, "At least one role must be assigned")
		}
		if err := s.rbacRepo.SetUserRoles(ctx, userID, *req.RoleIDs, adminID); err != nil {
			return nil, fmt.Errorf("failed to update user roles: %w", err)
		}
	}

	// Update active status if provided
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Update email if provided
	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}

	// Update username if provided
	if req.Username != nil && *req.Username != "" {
		user.Username = *req.Username
	}

	// Update full name if provided
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	// Update phone if provided
	if req.Phone != nil {
		user.Phone = req.Phone
	}

	// Update email verified status if provided
	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// DeleteUser deletes a user (soft delete by setting is_active = false)
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return err
	}

	// Check if user is admin and prevent deleting last admin
	userRoles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err == nil {
		isAdmin := false
		var adminRoleID uuid.UUID
		for _, role := range userRoles {
			if role.Name == string(models.RoleAdmin) {
				isAdmin = true
				adminRoleID = role.ID
				break
			}
		}

		if isAdmin {
			admins, err := s.rbacRepo.GetUsersWithRole(ctx, adminRoleID)
			if err == nil && len(admins) <= 1 {
				return models.NewAppError(400, "Cannot delete the last admin user")
			}
		}
	}

	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// ListAPIKeys returns all API keys with user information
func (s *AdminService) ListAPIKeys(ctx context.Context, appID *uuid.UUID, page, pageSize int) ([]*models.AdminAPIKeyResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var apiKeys []*models.APIKey
	var err error

	if appID != nil {
		apiKeys, err = s.apiKeyRepo.ListByApp(ctx, *appID)
	} else {
		apiKeys, err = s.apiKeyRepo.ListAll(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(apiKeys) {
		return []*models.AdminAPIKeyResponse{}, nil
	}
	if end > len(apiKeys) {
		end = len(apiKeys)
	}

	adminAPIKeys := make([]*models.AdminAPIKeyResponse, 0, end-start)
	for i := start; i < end; i++ {
		key := apiKeys[i]
		user, _ := s.userRepo.GetByID(ctx, key.UserID, nil)

		var scopes []string
		if err := json.Unmarshal(key.Scopes, &scopes); err != nil {
			scopes = []string{}
		}

		resp := &models.AdminAPIKeyResponse{
			ID:         key.ID,
			UserID:     key.UserID,
			Name:       key.Name,
			Prefix:     key.KeyPrefix,
			Scopes:     scopes,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			IsRevoked:  !key.IsActive,
			CreatedAt:  key.CreatedAt,
		}
		if user != nil {
			resp.Username = user.Username
		}
		adminAPIKeys = append(adminAPIKeys, resp)
	}

	return adminAPIKeys, nil
}

// RevokeAPIKey revokes an API key
func (s *AdminService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}

// ListAuditLogs returns paginated audit logs
func (s *AdminService) ListAuditLogs(ctx context.Context, page, pageSize int, userID *uuid.UUID) ([]*models.AdminAuditLogResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	var logs []*models.AuditLog
	var err error

	if userID != nil {
		logs, err = s.auditRepo.GetByUserID(ctx, *userID, pageSize, offset)
	} else {
		logs, err = s.auditRepo.List(ctx, pageSize, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	adminLogs := make([]*models.AdminAuditLogResponse, 0, len(logs))
	for _, log := range logs {
		var details map[string]interface{}
		if log.Details != nil {
			json.Unmarshal(log.Details, &details)
		}

		resp := &models.AdminAuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    string(log.Action),
			Status:    string(log.Status),
			IP:        log.IPAddress,
			UserAgent: log.UserAgent,
			Details:   details,
			CreatedAt: log.CreatedAt,
		}

		if log.User != nil {
			resp.UserEmail = log.User.Email
		}

		adminLogs = append(adminLogs, resp)
	}

	return adminLogs, nil
}

// AssignRole assigns a role to a user
func (s *AdminService) AssignRole(ctx context.Context, userID, roleID, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil, UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	existingRoleIDs := make([]uuid.UUID, len(user.Roles))
	for i, role := range user.Roles {
		existingRoleIDs[i] = role.ID
		if role.ID == roleID {
			return nil, models.NewAppError(400, "User already has this role")
		}
	}

	newRoleIDs := append(existingRoleIDs, roleID)
	if err := s.rbacRepo.SetUserRoles(ctx, userID, newRoleIDs, adminID); err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// RemoveRole removes a role from a user
func (s *AdminService) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil, UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	if len(user.Roles) <= 1 {
		return nil, models.NewAppError(400, "User must have at least one role")
	}

	roleFound := false
	newRoleIDs := make([]uuid.UUID, 0, len(user.Roles)-1)
	for _, role := range user.Roles {
		if role.ID == roleID {
			roleFound = true
			continue
		}
		newRoleIDs = append(newRoleIDs, role.ID)
	}

	if !roleFound {
		return nil, models.NewAppError(404, "User does not have this role")
	}

	if err := s.rbacRepo.SetUserRoles(ctx, userID, newRoleIDs, userID); err != nil {
		return nil, fmt.Errorf("failed to remove role: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// AdminReset2FA administratively disables 2FA for a user
func (s *AdminService) AdminReset2FA(ctx context.Context, userID, adminID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return err
	}

	if !user.TOTPEnabled {
		return models.NewAppError(400, "2FA is not enabled for this user")
	}

	err = s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		if err := s.userRepo.DisableTOTP(ctx, userID); err != nil {
			return err
		}

		if err := s.backupCodeRepo.DeleteAllByUserID(ctx, userID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to reset 2FA: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:        uuid.New(),
		UserID:    &adminID,
		Action:    string(models.Action2FAReset),
		Status:    string(models.StatusSuccess),
		CreatedAt: time.Now(),
		Details:   []byte(fmt.Sprintf(`{"target_user_id":"%s","admin_id":"%s"}`, userID, adminID)),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetUserOAuthAccounts returns OAuth accounts linked to a user
func (s *AdminService) GetUserOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	accounts, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth accounts: %w", err)
	}

	return accounts, nil
}

// userToAdminResponse converts User to AdminUserResponse with roles
func (s *AdminService) userToAdminResponse(user *models.User) *models.AdminUserResponse {
	roles := make([]models.RoleInfo, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, models.RoleInfo{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
		})
	}

	return &models.AdminUserResponse{
		ID:                user.ID,
		Email:             user.Email,
		Phone:             user.Phone,
		Username:          user.Username,
		FullName:          user.FullName,
		ProfilePictureURL: user.ProfilePictureURL,
		Roles:             roles,
		AccountType:       user.AccountType,
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		IsActive:          user.IsActive,
		TOTPEnabled:       user.TOTPEnabled,
		TOTPEnabledAt:     user.TOTPEnabledAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}
