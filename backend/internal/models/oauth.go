package models

import (
	"time"

	"github.com/google/uuid"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Provider       string     `json:"provider" db:"provider"`
	ProviderUserID string     `json:"provider_user_id" db:"provider_user_id"`
	AccessToken    string     `json:"-" db:"access_token"`
	RefreshToken   string     `json:"-" db:"refresh_token"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty" db:"token_expires_at"`
	ProfileData    []byte     `json:"profile_data,omitempty" db:"profile_data"` // JSONB in PostgreSQL
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
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
	ID        uuid.UUID `json:"id" db:"id"`
	Email     *string   `json:"email,omitempty" db:"email"` // Either email or phone is required
	Phone     *string   `json:"phone,omitempty" db:"phone"` // Either email or phone is required
	Code      string    `json:"-" db:"code"`                // Hashed OTP code
	Type      OTPType   `json:"type" db:"type"`             // verification, password_reset, 2fa
	Used      bool      `json:"used" db:"used"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
