package authgateway

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// AppOAuthProviders - Application OAuth Provider management

// ListAppOAuthProviders lists OAuth providers for an application
func (s *AdminService) ListAppOAuthProviders(ctx context.Context, appID string) ([]models.ApplicationOAuthProvider, error) {
	var result []models.ApplicationOAuthProvider
	err := s.client.get(ctx, fmt.Sprintf("/api/admin/applications/%s/oauth-providers", appID), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateAppOAuthProvider creates an OAuth provider for an application
func (s *AdminService) CreateAppOAuthProvider(ctx context.Context, appID string, req *models.CreateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error) {
	var result models.ApplicationOAuthProvider
	err := s.client.post(ctx, fmt.Sprintf("/api/admin/applications/%s/oauth-providers", appID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAppOAuthProvider gets an OAuth provider by ID
func (s *AdminService) GetAppOAuthProvider(ctx context.Context, appID, id string) (*models.ApplicationOAuthProvider, error) {
	var result models.ApplicationOAuthProvider
	err := s.client.get(ctx, fmt.Sprintf("/api/admin/applications/%s/oauth-providers/%s", appID, id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateAppOAuthProvider updates an OAuth provider
func (s *AdminService) UpdateAppOAuthProvider(ctx context.Context, appID, id string, req *models.UpdateAppOAuthProviderRequest) (*models.ApplicationOAuthProvider, error) {
	var result models.ApplicationOAuthProvider
	err := s.client.put(ctx, fmt.Sprintf("/api/admin/applications/%s/oauth-providers/%s", appID, id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAppOAuthProvider deletes an OAuth provider
func (s *AdminService) DeleteAppOAuthProvider(ctx context.Context, appID, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/applications/%s/oauth-providers/%s", appID, id), nil)
}

// TelegramBots - Telegram Bot management

// ListTelegramBots lists Telegram bots for an application
func (s *AdminService) ListTelegramBots(ctx context.Context, appID string) ([]models.TelegramBot, error) {
	var result []models.TelegramBot
	err := s.client.get(ctx, fmt.Sprintf("/api/admin/applications/%s/telegram-bots", appID), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateTelegramBot creates a Telegram bot for an application
func (s *AdminService) CreateTelegramBot(ctx context.Context, appID string, req *models.CreateTelegramBotRequest) (*models.TelegramBot, error) {
	var result models.TelegramBot
	err := s.client.post(ctx, fmt.Sprintf("/api/admin/applications/%s/telegram-bots", appID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTelegramBot gets a Telegram bot by ID
func (s *AdminService) GetTelegramBot(ctx context.Context, appID, id string) (*models.TelegramBot, error) {
	var result models.TelegramBot
	err := s.client.get(ctx, fmt.Sprintf("/api/admin/applications/%s/telegram-bots/%s", appID, id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTelegramBot updates a Telegram bot
func (s *AdminService) UpdateTelegramBot(ctx context.Context, appID, id string, req *models.UpdateTelegramBotRequest) (*models.TelegramBot, error) {
	var result models.TelegramBot
	err := s.client.put(ctx, fmt.Sprintf("/api/admin/applications/%s/telegram-bots/%s", appID, id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTelegramBot deletes a Telegram bot
func (s *AdminService) DeleteTelegramBot(ctx context.Context, appID, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/applications/%s/telegram-bots/%s", appID, id), nil)
}

// UserTelegram - User Telegram account/access management

// ListUserTelegramAccounts lists Telegram accounts for a user
func (s *AdminService) ListUserTelegramAccounts(ctx context.Context, userID string) ([]models.UserTelegramAccount, error) {
	var result []models.UserTelegramAccount
	err := s.client.get(ctx, fmt.Sprintf("/api/admin/users/%s/telegram-accounts", userID), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListUserTelegramBotAccess lists Telegram bot access for a user
func (s *AdminService) ListUserTelegramBotAccess(ctx context.Context, userID string, appID string) ([]models.UserTelegramBotAccess, error) {
	var result []models.UserTelegramBotAccess
	path := fmt.Sprintf("/api/admin/users/%s/telegram-bot-access", userID)
	if appID != "" {
		path += "?app_id=" + appID
	}
	err := s.client.get(ctx, path, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
