package models

import (
	"time"

	"github.com/google/uuid"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	Provider        string          `json:"provider" db:"provider"`
	ProviderUserID  string          `json:"provider_user_id" db:"provider_user_id"`
	AccessToken     string          `json:"-" db:"access_token"`
	RefreshToken    string          `json:"-" db:"refresh_token"`
	TokenExpiresAt  *time.Time      `json:"token_expires_at,omitempty" db:"token_expires_at"`
	ProfileData     []byte          `json:"profile_data,omitempty" db:"profile_data"` // JSONB in PostgreSQL
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// OAuthProvider represents the available OAuth providers
type OAuthProvider string

const (
	ProviderGoogle    OAuthProvider = "google"
	ProviderYandex    OAuthProvider = "yandex"
	ProviderGitHub    OAuthProvider = "github"
	ProviderInstagram OAuthProvider = "instagram"
)

// IsValidProvider checks if a provider is valid
func IsValidProvider(provider string) bool {
	switch OAuthProvider(provider) {
	case ProviderGoogle, ProviderYandex, ProviderGitHub, ProviderInstagram:
		return true
	default:
		return false
	}
}

// OAuthProviderInfo represents information about an OAuth provider
type OAuthProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IconURL     string `json:"icon_url,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	Username       string
	ProfilePicture string
	Provider       string
}
