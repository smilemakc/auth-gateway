package models

// ApplicationOAuthProvider represents a per-app OAuth provider configuration
type ApplicationOAuthProvider struct {
	ID            string   `json:"id"`
	ApplicationID string   `json:"application_id"`
	Provider      string   `json:"provider"`
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret,omitempty"`
	CallbackURL   string   `json:"callback_url"`
	Scopes        []string `json:"scopes"`
	AuthURL       string   `json:"auth_url"`
	TokenURL      string   `json:"token_url"`
	UserInfoURL   string   `json:"user_info_url"`
	IsActive      bool     `json:"is_active"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// TelegramBot represents a Telegram bot for an application
type TelegramBot struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	BotToken      string `json:"bot_token,omitempty"`
	BotUsername   string `json:"bot_username"`
	DisplayName   string `json:"display_name"`
	IsAuthBot     bool   `json:"is_auth_bot"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// UserTelegramAccount represents a user's Telegram account
type UserTelegramAccount struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	TelegramUserID   string `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	PhotoURL         string `json:"photo_url"`
	IsActive         bool   `json:"is_active"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// UserTelegramBotAccess represents user's access to a Telegram bot
type UserTelegramBotAccess struct {
	ID                 string               `json:"id"`
	UserID             string               `json:"user_id"`
	TelegramBotID      string               `json:"telegram_bot_id"`
	TelegramAccountID  string               `json:"telegram_account_id"`
	IsActive           bool                 `json:"is_active"`
	FirstInteractionAt string               `json:"first_interaction_at"`
	LastInteractionAt  string               `json:"last_interaction_at"`
	Bot                *TelegramBot         `json:"bot,omitempty"`
	Account            *UserTelegramAccount `json:"account,omitempty"`
}
