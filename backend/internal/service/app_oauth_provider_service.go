package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

var (
	ErrOAuthProviderNotFound      = errors.New("oauth provider configuration not found")
	ErrOAuthProviderExists        = errors.New("oauth provider already configured for this application")
	ErrInvalidOAuthProvider       = errors.New("invalid oauth provider configuration")
	ErrMissingRequiredCredentials = errors.New("missing required oauth credentials")
)

type AppOAuthProviderService struct {
	repo    AppOAuthProviderStore
	appRepo ApplicationStore
	log     *logger.Logger
}

func NewAppOAuthProviderService(repo AppOAuthProviderStore, appRepo ApplicationStore, log *logger.Logger) *AppOAuthProviderService {
	return &AppOAuthProviderService{
		repo:    repo,
		appRepo: appRepo,
		log:     log,
	}
}

func (s *AppOAuthProviderService) Create(ctx context.Context, appID uuid.UUID, req *models.CreateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error) {
	_, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, ErrApplicationNotFound
	}

	provider := strings.TrimSpace(strings.ToLower(req.Provider))
	if provider == "" {
		return nil, ErrInvalidOAuthProvider
	}

	if err := s.validateProviderConfig(provider, req); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByAppAndProvider(ctx, appID, provider)
	if err == nil && existing != nil {
		return nil, ErrOAuthProviderExists
	}

	scopes := req.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	oauthProvider := &models.ApplicationOAuthProvider{
		ID:            uuid.New(),
		ApplicationID: appID,
		Provider:      provider,
		ClientID:      strings.TrimSpace(req.ClientID),
		ClientSecret:  req.ClientSecret,
		CallbackURL:   strings.TrimSpace(req.CallbackURL),
		Scopes:        scopes,
		AuthURL:       strings.TrimSpace(req.AuthURL),
		TokenURL:      strings.TrimSpace(req.TokenURL),
		UserInfoURL:   strings.TrimSpace(req.UserInfoURL),
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, oauthProvider); err != nil {
		s.log.Error("failed to create oauth provider", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"provider":       provider,
		})
		return nil, fmt.Errorf("failed to create oauth provider: %w", err)
	}

	s.log.Info("oauth provider created", map[string]interface{}{
		"id":             oauthProvider.ID.String(),
		"application_id": appID.String(),
		"provider":       provider,
	})

	return oauthProvider, nil
}

func (s *AppOAuthProviderService) GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error) {
	provider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOAuthProviderNotFound
	}
	return provider, nil
}

func (s *AppOAuthProviderService) GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return nil, ErrInvalidOAuthProvider
	}

	oauthProvider, err := s.repo.GetByAppAndProvider(ctx, appID, provider)
	if err != nil {
		return nil, ErrOAuthProviderNotFound
	}
	return oauthProvider, nil
}

func (s *AppOAuthProviderService) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error) {
	providers, err := s.repo.ListByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth providers: %w", err)
	}
	return providers, nil
}

func (s *AppOAuthProviderService) ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error) {
	providers, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all oauth providers: %w", err)
	}
	return providers, nil
}

func (s *AppOAuthProviderService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error) {
	provider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOAuthProviderNotFound
	}

	if req.ClientID != nil {
		provider.ClientID = strings.TrimSpace(*req.ClientID)
	}
	if req.ClientSecret != nil {
		provider.ClientSecret = *req.ClientSecret
	}
	if req.CallbackURL != nil {
		provider.CallbackURL = strings.TrimSpace(*req.CallbackURL)
	}
	if req.Scopes != nil {
		provider.Scopes = req.Scopes
	}
	if req.AuthURL != nil {
		provider.AuthURL = strings.TrimSpace(*req.AuthURL)
	}
	if req.TokenURL != nil {
		provider.TokenURL = strings.TrimSpace(*req.TokenURL)
	}
	if req.UserInfoURL != nil {
		provider.UserInfoURL = strings.TrimSpace(*req.UserInfoURL)
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}

	provider.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, provider); err != nil {
		s.log.Error("failed to update oauth provider", map[string]interface{}{
			"error":      err.Error(),
			"id":         id.String(),
			"provider":   provider.Provider,
			"is_active":  provider.IsActive,
		})
		return nil, fmt.Errorf("failed to update oauth provider: %w", err)
	}

	s.log.Info("oauth provider updated", map[string]interface{}{
		"id":             id.String(),
		"application_id": provider.ApplicationID.String(),
		"provider":       provider.Provider,
		"is_active":      provider.IsActive,
	})

	return s.repo.GetByID(ctx, id)
}

func (s *AppOAuthProviderService) Delete(ctx context.Context, id uuid.UUID) error {
	provider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrOAuthProviderNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete oauth provider", map[string]interface{}{
			"error":    err.Error(),
			"id":       id.String(),
			"provider": provider.Provider,
		})
		return fmt.Errorf("failed to delete oauth provider: %w", err)
	}

	s.log.Info("oauth provider deleted", map[string]interface{}{
		"id":             id.String(),
		"application_id": provider.ApplicationID.String(),
		"provider":       provider.Provider,
	})

	return nil
}

func (s *AppOAuthProviderService) validateProviderConfig(provider string, req *models.CreateAppOAuthProviderRequest) error {
	clientID := strings.TrimSpace(req.ClientID)
	clientSecret := strings.TrimSpace(req.ClientSecret)

	if clientID == "" || clientSecret == "" {
		return ErrMissingRequiredCredentials
	}

	knownProviders := map[string]bool{
		"google":    true,
		"github":    true,
		"yandex":    true,
		"instagram": true,
	}

	if !knownProviders[provider] {
		authURL := strings.TrimSpace(req.AuthURL)
		tokenURL := strings.TrimSpace(req.TokenURL)
		userInfoURL := strings.TrimSpace(req.UserInfoURL)

		if authURL == "" || tokenURL == "" || userInfoURL == "" {
			return fmt.Errorf("custom provider requires auth_url, token_url, and user_info_url")
		}
	}

	return nil
}
