package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	bun.BaseModel  `bun:"table:oauth_accounts"`
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
	ProviderOneC      OAuthProvider = "onec"
)

// IsValidProvider checks if a provider is valid
func IsValidProvider(provider string) bool {
	switch OAuthProvider(provider) {
	case ProviderGoogle, ProviderYandex, ProviderGitHub, ProviderInstagram, ProviderTelegram, ProviderOneC:
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
	OTPTypeRegistration  OTPType = "registration"
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
	// Email address to send OTP to (required if phone not provided)
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	// Phone number to send OTP to (required if email not provided)
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// OTP type (verification, password_reset, 2fa, login)
	Type OTPType `json:"type" binding:"required" example:"verification"`
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	// Email address that received the OTP
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	// Phone number that received the OTP
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// 6-digit OTP code
	Code string `json:"code" binding:"required,len=6" example:"123456"`
	// OTP type (verification, password_reset, 2fa, login)
	Type OTPType `json:"type" binding:"required" example:"verification"`
}

// VerifyOTPResponse represents OTP verification response
type VerifyOTPResponse struct {
	// Whether the OTP is valid
	Valid bool `json:"valid" example:"true"`
	// Access token (if OTP is for login)
	AccessToken string `json:"access_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Refresh token (if OTP is for login)
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// User information (if OTP is for login)
	User *User `json:"user,omitempty"`
}

// OAuthCallbackRequest represents OAuth callback data
type OAuthCallbackRequest struct {
	// Authorization code from OAuth provider
	Code string `json:"code" form:"code" binding:"required" example:"4/0AY0e-g7xxxxxxxxxxxxxxxxxxx"`
	// State parameter for CSRF protection
	State string `json:"state" form:"state" binding:"required" example:"random_state_string_123"`
}

// OAuthLoginResponse represents OAuth login response
type OAuthLoginResponse struct {
	// JWT access token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Refresh token for obtaining new access tokens
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Authenticated user information
	User *User `json:"user"`
	// Whether this is a newly created user
	IsNewUser bool `json:"is_new_user" example:"false"`
}
