package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	bun.BaseModel  `table:"oauth_accounts"`
	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,type:uuid"`
	Provider       string     `json:"provider" bun:"provider"`
	ProviderUserID string     `json:"provider_user_id" bun:"provider_user_id"`
	AccessToken    string     `json:"-" bun:"access_token"`
	RefreshToken   string     `json:"-" bun:"refresh_token"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty" bun:"token_expires_at"`
	ProfileData    []byte     `json:"profile_data,omitempty" bun:"profile_data,type:jsonb"` // JSONB in PostgreSQL
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// OAuthProvider represents the available OAuth providers
type OAuthProvider string

const (
	ProviderGoogle    OAuthProvider = "google"
	ProviderYandex    OAuthProvider = "yandex"
	ProviderGitHub    OAuthProvider = "github"
	ProviderInstagram OAuthProvider = "instagram"
	ProviderTelegram  OAuthProvider = "telegram"
)

// IsValidProvider checks if a provider is valid
func IsValidProvider(provider string) bool {
	switch OAuthProvider(provider) {
	case ProviderGoogle, ProviderYandex, ProviderGitHub, ProviderInstagram, ProviderTelegram:
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

// OTP represents a one-time password for email or phone verification
type OTP struct {
	ID        uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Email     *string   `json:"email,omitempty" bun:"email"` // Either email or phone is required
	Phone     *string   `json:"phone,omitempty" bun:"phone"` // Either email or phone is required
	Code      string    `json:"-" bun:"code"`                // Hashed OTP code
	Type      OTPType   `json:"type" bun:"type"`             // verification, password_reset, 2fa
	Used      bool      `json:"used" bun:"used"`
	ExpiresAt time.Time `json:"expires_at" bun:"expires_at"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// OTPType represents the type of OTP
type OTPType string

const (
	OTPTypeVerification  OTPType = "verification"
	OTPTypePasswordReset OTPType = "password_reset"
	OTPType2FA           OTPType = "2fa"
	OTPTypeLogin         OTPType = "login"
)

// IsExpired checks if OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if OTP is valid (not used and not expired)
func (o *OTP) IsValid() bool {
	return !o.Used && !o.IsExpired()
}

// SendOTPRequest represents OTP send request
type SendOTPRequest struct {
	Email *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone *string `json:"phone,omitempty"`
	Type  OTPType `json:"type" binding:"required"`
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	Email *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone *string `json:"phone,omitempty"`
	Code  string  `json:"code" binding:"required,len=6"`
	Type  OTPType `json:"type" binding:"required"`
}

// VerifyOTPResponse represents OTP verification response
type VerifyOTPResponse struct {
	Valid        bool   `json:"valid"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         *User  `json:"user,omitempty"`
}

// OAuthCallbackRequest represents OAuth callback data
type OAuthCallbackRequest struct {
	Code  string `json:"code" form:"code" binding:"required"`
	State string `json:"state" form:"state" binding:"required"`
}

// OAuthLoginResponse represents OAuth login response
type OAuthLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
	IsNewUser    bool   `json:"is_new_user"`
}
