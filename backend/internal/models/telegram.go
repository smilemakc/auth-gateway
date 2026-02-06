package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// TelegramBot represents a Telegram bot tied to an application.
// Can be used for authorization, notifications, or as a service bot.
type TelegramBot struct {
	bun.BaseModel `bun:"table:telegram_bots,alias:tb"`

	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	BotToken      string    `json:"-" bun:"bot_token,notnull"`
	BotUsername   string    `json:"bot_username" bun:"bot_username,notnull" example:"my_auth_bot"`
	DisplayName   string    `json:"display_name" bun:"display_name,notnull" example:"My Auth Bot"`
	IsAuthBot     bool      `json:"is_auth_bot" bun:"is_auth_bot,default:false" example:"false"`
	IsActive      bool      `json:"is_active" bun:"is_active,default:true" example:"true"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	Application *Application `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

// UserTelegramAccount represents global link between a User and their Telegram identity.
type UserTelegramAccount struct {
	bun.BaseModel `bun:"table:user_telegram_accounts,alias:uta"`

	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	TelegramUserID int64      `json:"telegram_user_id" bun:"telegram_user_id,notnull" example:"123456789"`
	Username       *string    `json:"username,omitempty" bun:"username" example:"johndoe"`
	FirstName      string     `json:"first_name" bun:"first_name,notnull" example:"John"`
	LastName       *string    `json:"last_name,omitempty" bun:"last_name" example:"Doe"`
	PhotoURL       *string    `json:"photo_url,omitempty" bun:"photo_url" example:"https://t.me/i/userpic/320/johndoe.jpg"`
	AuthDate       time.Time  `json:"auth_date" bun:"auth_date,notnull" example:"2024-01-15T10:30:00Z"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// UserTelegramBotAccess tracks which bots can message a user
// and through which bot the user authorized.
type UserTelegramBotAccess struct {
	bun.BaseModel `bun:"table:user_telegram_bot_access,alias:utba"`

	ID                uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID            uuid.UUID `json:"user_id" bun:"user_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	TelegramBotID     uuid.UUID `json:"telegram_bot_id" bun:"telegram_bot_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	TelegramAccountID uuid.UUID `json:"telegram_account_id" bun:"telegram_account_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	CanSendMessages   bool      `json:"can_send_messages" bun:"can_send_messages,default:true" example:"true"`
	AuthorizedVia     bool      `json:"authorized_via" bun:"authorized_via,default:false" example:"false"`
	CreatedAt         time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	User            *User                `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	TelegramBot     *TelegramBot         `json:"telegram_bot,omitempty" bun:"rel:belongs-to,join:telegram_bot_id=id"`
	TelegramAccount *UserTelegramAccount `json:"telegram_account,omitempty" bun:"rel:belongs-to,join:telegram_account_id=id"`
}

// CreateTelegramBotRequest represents request to create Telegram bot
type CreateTelegramBotRequest struct {
	BotToken    string `json:"bot_token" binding:"required" example:"123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"`
	BotUsername string `json:"bot_username" binding:"required,max=100" example:"my_auth_bot"`
	DisplayName string `json:"display_name" binding:"required,max=100" example:"My Auth Bot"`
	IsAuthBot   bool   `json:"is_auth_bot,omitempty" example:"false"`
}

// UpdateTelegramBotRequest represents request to update Telegram bot
type UpdateTelegramBotRequest struct {
	BotToken    *string `json:"bot_token,omitempty" example:"123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"`
	BotUsername *string `json:"bot_username,omitempty" binding:"omitempty,max=100" example:"my_auth_bot"`
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=100" example:"My Auth Bot"`
	IsAuthBot   *bool   `json:"is_auth_bot,omitempty" example:"false"`
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`
}
