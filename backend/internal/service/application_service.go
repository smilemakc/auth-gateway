package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

var (
	ErrApplicationNotFound    = errors.New("application not found")
	ErrApplicationNameExists  = errors.New("application name already exists")
	ErrUserProfileNotFound    = errors.New("user application profile not found")
	ErrUserProfileExists      = errors.New("user profile for this application already exists")
	ErrUserBannedFromApp      = errors.New("user is banned from this application")
	ErrCannotDeleteSystemApp  = errors.New("cannot delete system application")
	ErrInvalidApplicationName = errors.New("invalid application name: must be lowercase alphanumeric with hyphens only")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type ApplicationService struct {
	appRepo      ApplicationStore
	appOAuthRepo AppOAuthProviderStore
	logger       *logger.Logger
}

func NewApplicationService(appRepo ApplicationStore, appOAuthRepo AppOAuthProviderStore, log *logger.Logger) *ApplicationService {
	return &ApplicationService{
		appRepo:      appRepo,
		appOAuthRepo: appOAuthRepo,
		logger:       log,
	}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
	name := strings.TrimSpace(strings.ToLower(req.Name))
	if !isValidSlug(name) {
		return nil, "", ErrInvalidApplicationName
	}

	existing, err := s.appRepo.GetApplicationByName(ctx, name)
	if err == nil && existing != nil {
		return nil, "", ErrApplicationNameExists
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	allowedAuthMethods := req.AllowedAuthMethods
	if len(allowedAuthMethods) == 0 {
		allowedAuthMethods = []string{"password"}
	}

	app := &models.Application{
		ID:                 uuid.New(),
		Name:               name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		HomepageURL:        req.HomepageURL,
		CallbackURLs:       req.CallbackURLs,
		IsActive:           isActive,
		IsSystem:           false,
		OwnerID:            ownerID,
		AllowedAuthMethods: allowedAuthMethods,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.appRepo.CreateApplication(ctx, app); err != nil {
		s.logger.Error("failed to create application", map[string]interface{}{
			"error": err.Error(),
			"name":  name,
		})
		return nil, "", fmt.Errorf("failed to create application: %w", err)
	}

	branding := &models.ApplicationBranding{
		ID:              uuid.New(),
		ApplicationID:   app.ID,
		PrimaryColor:    "#3B82F6",
		SecondaryColor:  "#8B5CF6",
		BackgroundColor: "#FFFFFF",
		UpdatedAt:       time.Now(),
	}

	if err := s.appRepo.CreateOrUpdateBranding(ctx, branding); err != nil {
		s.logger.Warn("failed to create default branding", map[string]interface{}{
			"error":          err.Error(),
			"application_id": app.ID.String(),
		})
	}

	// Auto-generate application secret
	secret, err := s.GenerateSecret(ctx, app.ID)
	if err != nil {
		s.logger.Warn("failed to auto-generate application secret", map[string]interface{}{
			"error":          err.Error(),
			"application_id": app.ID.String(),
		})
	}

	s.logger.Info("application created", map[string]interface{}{
		"application_id": app.ID.String(),
		"name":           name,
		"display_name":   req.DisplayName,
	})

	app.Branding = branding
	return app, secret, nil
}

func (s *ApplicationService) GetByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	app, err := s.appRepo.GetApplicationByID(ctx, id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	return app, nil
}

func (s *ApplicationService) GetByName(ctx context.Context, name string) (*models.Application, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	app, err := s.appRepo.GetApplicationByName(ctx, name)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	return app, nil
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error) {
	app, err := s.appRepo.GetApplicationByID(ctx, id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}

	if req.DisplayName != "" {
		app.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.HomepageURL != "" {
		app.HomepageURL = req.HomepageURL
	}
	if req.CallbackURLs != nil {
		app.CallbackURLs = req.CallbackURLs
	}
	if req.IsActive != nil {
		app.IsActive = *req.IsActive
	}
	if req.AllowedAuthMethods != nil {
		app.AllowedAuthMethods = req.AllowedAuthMethods
	}

	app.UpdatedAt = time.Now()

	if err := s.appRepo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("failed to update application", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	s.logger.Info("application updated", map[string]interface{}{
		"application_id": id.String(),
		"name":           app.Name,
	})

	return s.appRepo.GetApplicationByID(ctx, id)
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	app, err := s.appRepo.GetApplicationByID(ctx, id)
	if err != nil {
		return ErrApplicationNotFound
	}

	if app.IsSystem {
		return ErrCannotDeleteSystemApp
	}

	app.IsActive = false
	app.UpdatedAt = time.Now()

	if err := s.appRepo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("failed to delete application", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		return fmt.Errorf("failed to delete application: %w", err)
	}

	s.logger.Info("application deleted", map[string]interface{}{
		"application_id": id.String(),
		"name":           app.Name,
	})

	return nil
}

func (s *ApplicationService) ListApplications(ctx context.Context, page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	apps, total, err := s.appRepo.ListApplications(ctx, page, perPage, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	totalPages := (total + perPage - 1) / perPage

	applications := make([]models.Application, len(apps))
	for i, app := range apps {
		applications[i] = *app
	}

	return &models.ApplicationListResponse{
		Applications: applications,
		Total:        total,
		Page:         page,
		PageSize:     perPage,
		TotalPages:   totalPages,
	}, nil
}

func (s *ApplicationService) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	branding, err := s.appRepo.GetBranding(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branding: %w", err)
	}
	return branding, nil
}

func (s *ApplicationService) UpdateBranding(ctx context.Context, applicationID uuid.UUID, req *models.UpdateApplicationBrandingRequest) (*models.ApplicationBranding, error) {
	branding, err := s.appRepo.GetBranding(ctx, applicationID)
	if err != nil {
		branding = &models.ApplicationBranding{
			ID:              uuid.New(),
			ApplicationID:   applicationID,
			PrimaryColor:    "#3B82F6",
			SecondaryColor:  "#8B5CF6",
			BackgroundColor: "#FFFFFF",
		}
	}

	if req.LogoURL != "" {
		branding.LogoURL = req.LogoURL
	}
	if req.FaviconURL != "" {
		branding.FaviconURL = req.FaviconURL
	}
	if req.PrimaryColor != "" {
		branding.PrimaryColor = req.PrimaryColor
	}
	if req.SecondaryColor != "" {
		branding.SecondaryColor = req.SecondaryColor
	}
	if req.BackgroundColor != "" {
		branding.BackgroundColor = req.BackgroundColor
	}
	if req.CustomCSS != "" {
		branding.CustomCSS = req.CustomCSS
	}
	if req.CompanyName != "" {
		branding.CompanyName = req.CompanyName
	}
	if req.SupportEmail != "" {
		branding.SupportEmail = req.SupportEmail
	}
	if req.TermsURL != "" {
		branding.TermsURL = req.TermsURL
	}
	if req.PrivacyURL != "" {
		branding.PrivacyURL = req.PrivacyURL
	}

	branding.UpdatedAt = time.Now()

	if err := s.appRepo.CreateOrUpdateBranding(ctx, branding); err != nil {
		s.logger.Error("failed to update branding", map[string]interface{}{
			"error":          err.Error(),
			"application_id": applicationID.String(),
		})
		return nil, fmt.Errorf("failed to update branding: %w", err)
	}

	s.logger.Info("application branding updated", map[string]interface{}{
		"application_id": applicationID.String(),
	})

	return s.appRepo.GetBranding(ctx, applicationID)
}

func (s *ApplicationService) GetOrCreateUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err == nil {
		now := time.Now()
		if err := s.appRepo.UpdateLastAccess(ctx, userID, applicationID); err != nil {
			s.logger.Warn("failed to update last access", map[string]interface{}{
				"error":          err.Error(),
				"user_id":        userID.String(),
				"application_id": applicationID.String(),
			})
		}
		profile.LastAccessAt = &now
		return profile, nil
	}

	now := time.Now()
	profile = &models.UserApplicationProfile{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: applicationID,
		IsActive:      true,
		IsBanned:      false,
		LastAccessAt:  &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.appRepo.CreateUserProfile(ctx, profile); err != nil {
		s.logger.Error("failed to create user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": applicationID.String(),
		})
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}

	s.logger.Info("user application profile created", map[string]interface{}{
		"user_id":        userID.String(),
		"application_id": applicationID.String(),
	})

	return profile, nil
}

func (s *ApplicationService) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err != nil {
		return nil, ErrUserProfileNotFound
	}
	return profile, nil
}

func (s *ApplicationService) UpdateUserProfile(ctx context.Context, userID, applicationID uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error) {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err != nil {
		return nil, ErrUserProfileNotFound
	}

	if req.DisplayName != nil {
		profile.DisplayName = req.DisplayName
	}
	if req.AvatarURL != nil {
		profile.AvatarURL = req.AvatarURL
	}
	if req.Nickname != nil {
		profile.Nickname = req.Nickname
	}
	if req.Metadata != nil {
		profile.Metadata = req.Metadata
	}
	if req.AppRoles != nil {
		profile.AppRoles = req.AppRoles
	}
	if req.IsActive != nil {
		profile.IsActive = *req.IsActive
	}
	if req.IsBanned != nil {
		profile.IsBanned = *req.IsBanned
	}
	if req.BanReason != nil {
		profile.BanReason = req.BanReason
	}

	profile.UpdatedAt = time.Now()

	if err := s.appRepo.UpdateUserProfile(ctx, profile); err != nil {
		s.logger.Error("failed to update user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": applicationID.String(),
		})
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return s.appRepo.GetUserProfile(ctx, userID, applicationID)
}

func (s *ApplicationService) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	profiles, err := s.appRepo.ListUserProfiles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user profiles: %w", err)
	}
	return profiles, nil
}

func (s *ApplicationService) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	profiles, total, err := s.appRepo.ListApplicationUsers(ctx, applicationID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list application users: %w", err)
	}

	totalPages := (total + perPage - 1) / perPage

	profilesList := make([]models.UserApplicationProfile, len(profiles))
	for i, profile := range profiles {
		profilesList[i] = *profile
	}

	return &models.UserAppProfileListResponse{
		Profiles:   profilesList,
		Total:      total,
		Page:       page,
		PageSize:   perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *ApplicationService) BanUser(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err != nil {
		return ErrUserProfileNotFound
	}

	if profile.IsBanned {
		return nil
	}

	if err := s.appRepo.BanUserFromApplication(ctx, userID, applicationID, bannedBy, reason); err != nil {
		s.logger.Error("failed to ban user from application", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": applicationID.String(),
			"banned_by":      bannedBy.String(),
		})
		return fmt.Errorf("failed to ban user: %w", err)
	}

	s.logger.Info("user banned from application", map[string]interface{}{
		"user_id":        userID.String(),
		"application_id": applicationID.String(),
		"banned_by":      bannedBy.String(),
		"reason":         reason,
	})

	return nil
}

func (s *ApplicationService) UnbanUser(ctx context.Context, userID, applicationID uuid.UUID) error {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err != nil {
		return ErrUserProfileNotFound
	}

	if !profile.IsBanned {
		return nil
	}

	if err := s.appRepo.UnbanUserFromApplication(ctx, userID, applicationID); err != nil {
		s.logger.Error("failed to unban user from application", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": applicationID.String(),
		})
		return fmt.Errorf("failed to unban user: %w", err)
	}

	s.logger.Info("user unbanned from application", map[string]interface{}{
		"user_id":        userID.String(),
		"application_id": applicationID.String(),
	})

	return nil
}

func (s *ApplicationService) CheckUserAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	profile, err := s.appRepo.GetUserProfile(ctx, userID, applicationID)
	if err != nil {
		return fmt.Errorf("user has no access to application: %w", err)
	}

	if profile.IsBanned {
		return ErrUserBannedFromApp
	}

	return nil
}

func isValidSlug(name string) bool {
	if len(name) < 3 || len(name) > 100 {
		return false
	}
	return slugRegex.MatchString(name)
}

// IsAuthMethodAllowed checks if an auth method is allowed for a given application
func (s *ApplicationService) IsAuthMethodAllowed(ctx context.Context, appID uuid.UUID, method string) error {
	app, err := s.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if !app.IsActive {
		return models.NewAppError(403, "Application is not active")
	}
	for _, m := range app.AllowedAuthMethods {
		if m == method {
			return nil
		}
	}
	return models.NewAppError(403, fmt.Sprintf("Auth method '%s' is not allowed for this application", method))
}

const appSecretPrefix = "app_"

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSecret creates a new application secret
// Returns the raw secret ONE TIME (only the hash is stored)
func (s *ApplicationService) GenerateSecret(ctx context.Context, appID uuid.UUID) (string, error) {
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return "", ErrApplicationNotFound
	}
	if !app.IsActive {
		return "", models.NewAppError(403, "Application is not active")
	}

	token, err := generateSecureToken(40)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	rawSecret := appSecretPrefix + token

	hash := utils.HashToken(rawSecret)
	prefix := rawSecret[:12]

	now := time.Now()
	app.SecretHash = hash
	app.SecretPrefix = prefix
	app.SecretLastRotatedAt = &now
	app.UpdatedAt = now

	if err := s.appRepo.UpdateApplication(ctx, app); err != nil {
		return "", fmt.Errorf("failed to save application secret: %w", err)
	}

	s.logger.Info("application secret generated", map[string]interface{}{
		"application_id": appID.String(),
		"prefix":         prefix,
	})

	return rawSecret, nil
}

// RotateSecret generates a new secret, invalidating the old one
func (s *ApplicationService) RotateSecret(ctx context.Context, appID uuid.UUID) (string, error) {
	return s.GenerateSecret(ctx, appID)
}

// ValidateSecret validates an app_ token and returns the Application
func (s *ApplicationService) ValidateSecret(ctx context.Context, secret string) (*models.Application, error) {
	hash := utils.HashToken(secret)
	app, err := s.appRepo.GetBySecretHash(ctx, hash)
	if err != nil {
		return nil, models.NewAppError(401, "Invalid application secret")
	}
	// Defense-in-depth: constant-time verification of hash match
	if !utils.CompareHashConstantTime(hash, app.SecretHash) {
		return nil, models.NewAppError(401, "Invalid application secret")
	}
	if !app.IsActive {
		return nil, models.NewAppError(403, "Application is not active")
	}
	return app, nil
}

func (s *ApplicationService) GetAuthConfig(ctx context.Context, app *models.Application) (*models.AuthConfigResponse, error) {
	config := &models.AuthConfigResponse{
		ApplicationID:      app.ID,
		Name:               app.Name,
		DisplayName:        app.DisplayName,
		AllowedAuthMethods: app.AllowedAuthMethods,
	}

	if s.appOAuthRepo != nil {
		providers, err := s.appOAuthRepo.ListByApp(ctx, app.ID)
		if err == nil {
			oauthProviders := make([]string, 0)
			for _, p := range providers {
				if p.IsActive {
					oauthProviders = append(oauthProviders, p.Provider)
				}
			}
			config.OAuthProviders = oauthProviders
		}
	}

	branding, err := s.appRepo.GetBranding(ctx, app.ID)
	if err == nil && branding != nil {
		pubBranding := branding.ToPublicResponse()
		config.Branding = &pubBranding
	}

	return config, nil
}
