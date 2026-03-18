package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ApplicationOAuthProvider represents per-application OAuth provider configuration.
// Each application can configure its own Google/Yandex/GitHub/etc. keys.
type ApplicationOAuthProvider struct {
	bun.BaseModel `bun:"table:application_oauth_providers,alias:aop"`

	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Provider      string    `json:"provider" bun:"provider,notnull" example:"google"`
	ClientID      string    `json:"client_id" bun:"client_id,notnull" example:"123456789.apps.googleusercontent.com"`
	ClientSecret  string    `json:"-" bun:"client_secret,notnull"`
	CallbackURL   string    `json:"callback_url" bun:"callback_url,notnull" example:"https://example.com/oauth/callback"`
	Scopes        []string  `json:"scopes,omitempty" bun:"scopes,type:jsonb,default:'[]'" example:"openid,email,profile"`
	AuthURL       string    `json:"auth_url,omitempty" bun:"auth_url" example:"https://accounts.google.com/o/oauth2/auth"`
	TokenURL      string    `json:"token_url,omitempty" bun:"token_url" example:"https://oauth2.googleapis.com/token"`
	UserInfoURL   string    `json:"user_info_url,omitempty" bun:"user_info_url" example:"https://www.googleapis.com/oauth2/v2/userinfo"`
	IsActive      bool      `json:"is_active" bun:"is_active,default:true" example:"true"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	Application *Application `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

// CreateAppOAuthProviderRequest represents request to create application OAuth provider
type CreateAppOAuthProviderRequest struct {
	Provider     string   `json:"provider" binding:"required" example:"google"`
	ClientID     string   `json:"client_id" binding:"required" example:"123456789.apps.googleusercontent.com"`
	ClientSecret string   `json:"client_secret" binding:"required" example:"GOCSPX-xxxxxxxxxxxxxxxxxxxxx"`
	CallbackURL  string   `json:"callback_url" binding:"required,url,max=500" example:"https://example.com/oauth/callback"`
	Scopes       []string `json:"scopes,omitempty" example:"openid,email,profile"`
	AuthURL      string   `json:"auth_url,omitempty" binding:"omitempty,url,max=500" example:"https://accounts.google.com/o/oauth2/auth"`
	TokenURL     string   `json:"token_url,omitempty" binding:"omitempty,url,max=500" example:"https://oauth2.googleapis.com/token"`
	UserInfoURL  string   `json:"user_info_url,omitempty" binding:"omitempty,url,max=500" example:"https://www.googleapis.com/oauth2/v2/userinfo"`
}

// UpdateAppOAuthProviderRequest represents request to update application OAuth provider
type UpdateAppOAuthProviderRequest struct {
	ClientID     *string  `json:"client_id,omitempty" binding:"omitempty,max=500" example:"123456789.apps.googleusercontent.com"`
	ClientSecret *string  `json:"client_secret,omitempty" example:"GOCSPX-xxxxxxxxxxxxxxxxxxxxx"`
	CallbackURL  *string  `json:"callback_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/oauth/callback"`
	Scopes       []string `json:"scopes,omitempty" example:"openid,email,profile"`
	AuthURL      *string  `json:"auth_url,omitempty" binding:"omitempty,url,max=500" example:"https://accounts.google.com/o/oauth2/auth"`
	TokenURL     *string  `json:"token_url,omitempty" binding:"omitempty,url,max=500" example:"https://oauth2.googleapis.com/token"`
	UserInfoURL  *string  `json:"user_info_url,omitempty" binding:"omitempty,url,max=500" example:"https://www.googleapis.com/oauth2/v2/userinfo"`
	IsActive     *bool    `json:"is_active,omitempty" example:"true"`
}

// AppOAuthProviderListResponse represents application OAuth providers list
type AppOAuthProviderListResponse struct {
	// List of OAuth providers
	Providers []*ApplicationOAuthProvider `json:"providers"`
	// Total number of providers
	Total int `json:"total" example:"3"`
}
