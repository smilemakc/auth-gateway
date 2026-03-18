package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/uptrace/bun"
)

type AdminUserService struct {
	userRepo       UserStore
	rbacRepo       RBACStore
	oauthRepo      OAuthStore
	backupCodeRepo BackupCodeStore
	apiKeyRepo     APIKeyStore
	auditRepo      AuditStore
	appRepo        ApplicationStore
	bcryptCost     int
	db             TransactionDB
}

func (s *AdminUserService) ListUsers(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminUserListResponse, error) {
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

func (s *AdminUserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil, UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	adminUser := s.userToAdminResponse(user)

	apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.APIKeysCount = len(apiKeys)
	}

	oauthAccounts, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.OAuthAccountsCount = len(oauthAccounts)
	}

	return adminUser, nil
}

func (s *AdminUserService) CreateUser(ctx context.Context, req *models.AdminCreateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	user := &models.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      utils.SanitizeUsername(req.Username),
		FullName:      utils.SanitizeHTML(req.FullName),
		AccountType:   req.AccountType,
		IsActive:      true,
		EmailVerified: true,
	}

	if user.AccountType == "" {
		user.AccountType = string(models.AccountTypeHuman)
	}

	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password, s.bcryptCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = hash
	}

	var roleIDs []uuid.UUID
	if len(req.RoleIDs) > 0 {
		roleIDs = req.RoleIDs
	} else {
		defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
		if err == nil {
			roleIDs = []uuid.UUID{defaultRole.ID}
		}
	}

	err := s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		if userRepo, ok := s.userRepo.(*repository.UserRepository); ok {
			if err := userRepo.CreateWithTx(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			if err := s.userRepo.Create(ctx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		}

		if len(roleIDs) > 0 {
			if err := s.rbacRepo.SetUserRoles(ctx, user.ID, roleIDs, adminID); err != nil {
				return fmt.Errorf("failed to assign roles: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetUser(ctx, user.ID)
}

func (s *AdminUserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.AdminUpdateUserRequest, adminID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	if req.RoleIDs != nil {
		if len(*req.RoleIDs) == 0 {
			return nil, models.NewAppError(400, "At least one role must be assigned")
		}
		if err := s.rbacRepo.SetUserRoles(ctx, userID, *req.RoleIDs, adminID); err != nil {
			return nil, fmt.Errorf("failed to update user roles: %w", err)
		}
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}

	if req.Username != nil && *req.Username != "" {
		user.Username = utils.SanitizeUsername(*req.Username)
	}

	if req.FullName != nil {
		user.FullName = utils.SanitizeHTML(*req.FullName)
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(ctx, userID)
}

func (s *AdminUserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return err
	}

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

func (s *AdminUserService) AdminReset2FA(ctx context.Context, userID, adminID uuid.UUID) error {
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

func (s *AdminUserService) GetUserOAuthAccounts(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
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

func (s *AdminUserService) AssignRole(ctx context.Context, userID, roleID, adminID uuid.UUID) (*models.AdminUserResponse, error) {
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

func (s *AdminUserService) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*models.AdminUserResponse, error) {
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

func (s *AdminUserService) userToAdminResponse(user *models.User) *models.AdminUserResponse {
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
