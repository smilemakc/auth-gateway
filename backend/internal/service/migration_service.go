package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

type MigrationService struct {
	userRepo       UserStore
	oauthRepo      OAuthStore
	rbacRepo       RBACStore
	appProfileRepo ApplicationStore
	applicationRepo ApplicationStore
}

func NewMigrationService(
	userRepo UserStore,
	oauthRepo OAuthStore,
	rbacRepo RBACStore,
	appProfileRepo ApplicationStore,
	appRepo ApplicationStore,
) *MigrationService {
	return &MigrationService{
		userRepo:        userRepo,
		oauthRepo:       oauthRepo,
		rbacRepo:        rbacRepo,
		appProfileRepo:  appProfileRepo,
		applicationRepo: appRepo,
	}
}

func (s *MigrationService) ImportUsers(ctx context.Context, appID uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error) {
	result := &models.ImportResult{
		Total:   len(entries),
		Created: 0,
		Skipped: 0,
		Errors:  []string{},
	}

	app, err := s.applicationRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if !app.IsActive {
		return nil, models.NewAppError(400, "Application is not active")
	}

	for idx, entry := range entries {
		if err := s.importSingleUser(ctx, appID, entry); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("User %d (%s): %v", idx+1, entry.Email, err))
			result.Skipped++
		} else {
			result.Created++
		}
	}

	return result, nil
}

func (s *MigrationService) importSingleUser(ctx context.Context, appID uuid.UUID, entry models.ImportUserEntry) error {
	email := utils.NormalizeEmail(entry.Email)
	if err := utils.ValidateEmail(email); err != nil {
		return err
	}

	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	username := entry.Username
	if username == "" {
		username = email
	}
	username = utils.NormalizeUsername(username)
	if !utils.IsValidUsername(username) {
		return fmt.Errorf("invalid username format")
	}

	usernameExists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("username already exists")
	}

	passwordHash := entry.Password
	if entry.HashFormat == "plain" {
		passwordHash, err = utils.HashPassword(entry.Password, 10)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
	} else if entry.HashFormat != "" && entry.HashFormat != "bcrypt" {
		passwordHash = fmt.Sprintf("%s:%s", entry.HashFormat, entry.Password)
	}

	isActive := true
	if entry.IsActive != nil {
		isActive = *entry.IsActive
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     utils.SanitizeUsername(username),
		FullName:     utils.SanitizeHTML(entry.FullName),
		PasswordHash: passwordHash,
		IsActive:     isActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	profile := &models.UserApplicationProfile{
		UserID:        user.ID,
		ApplicationID: appID,
	}
	if err := s.appProfileRepo.CreateUserProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to create app profile: %w", err)
	}

	for _, roleName := range entry.Roles {
		role, err := s.rbacRepo.GetRoleByNameAndApp(ctx, roleName, &appID)
		if err != nil {
			continue
		}
		if err := s.rbacRepo.AssignRoleToUserInApp(ctx, user.ID, role.ID, user.ID, &appID); err != nil {
			return fmt.Errorf("failed to assign role %s: %w", roleName, err)
		}
	}

	return nil
}

func (s *MigrationService) ImportOAuthAccounts(ctx context.Context, entries []models.ImportOAuthEntry) (*models.ImportResult, error) {
	result := &models.ImportResult{
		Total:   len(entries),
		Created: 0,
		Skipped: 0,
		Errors:  []string{},
	}

	for idx, entry := range entries {
		if err := s.importSingleOAuthAccount(ctx, entry); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("OAuth %d (%s/%s): %v", idx+1, entry.Email, entry.Provider, err))
			result.Skipped++
		} else {
			result.Created++
		}
	}

	return result, nil
}

func (s *MigrationService) importSingleOAuthAccount(ctx context.Context, entry models.ImportOAuthEntry) error {
	email := utils.NormalizeEmail(entry.Email)
	if err := utils.ValidateEmail(email); err != nil {
		return err
	}

	user, err := s.userRepo.GetByEmail(ctx, email, nil)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	existing, err := s.oauthRepo.GetOAuthAccount(ctx, entry.Provider, entry.ProviderUserID)
	if err == nil && existing != nil {
		return fmt.Errorf("OAuth account already exists")
	}

	oauthAccount := &models.OAuthAccount{
		ID:             uuid.New(),
		UserID:         user.ID,
		Provider:       entry.Provider,
		ProviderUserID: entry.ProviderUserID,
		AccessToken:    entry.AccessToken,
		RefreshToken:   entry.RefreshToken,
	}

	if err := s.oauthRepo.CreateOAuthAccount(ctx, oauthAccount); err != nil {
		return fmt.Errorf("failed to create OAuth account: %w", err)
	}

	return nil
}

func (s *MigrationService) ImportRoles(ctx context.Context, appID uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error) {
	result := &models.ImportResult{
		Total:   len(entries),
		Created: 0,
		Skipped: 0,
		Errors:  []string{},
	}

	app, err := s.applicationRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if !app.IsActive {
		return nil, models.NewAppError(400, "Application is not active")
	}

	for idx, entry := range entries {
		if err := s.importSingleRole(ctx, appID, entry); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Role %d (%s): %v", idx+1, entry.Name, err))
			result.Skipped++
		} else {
			result.Created++
		}
	}

	return result, nil
}

func (s *MigrationService) importSingleRole(ctx context.Context, appID uuid.UUID, entry models.ImportRoleEntry) error {
	existingRole, err := s.rbacRepo.GetRoleByNameAndApp(ctx, entry.Name, &appID)
	if err == nil && existingRole != nil {
		return fmt.Errorf("role already exists")
	}

	role := &models.Role{
		ID:            uuid.New(),
		Name:          entry.Name,
		DisplayName:   entry.Name,
		Description:   entry.Description,
		ApplicationID: &appID,
	}

	if err := s.rbacRepo.CreateRole(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	for _, email := range entry.AssignTo {
		normalizedEmail := utils.NormalizeEmail(email)
		user, err := s.userRepo.GetByEmail(ctx, normalizedEmail, nil)
		if err != nil {
			continue
		}

		if err := s.rbacRepo.AssignRoleToUserInApp(ctx, user.ID, role.ID, user.ID, &appID); err != nil {
			return fmt.Errorf("failed to assign role to user %s: %w", email, err)
		}
	}

	return nil
}
