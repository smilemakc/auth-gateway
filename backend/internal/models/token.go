package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID           uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"user_id" bun:"user_id,type:uuid,notnull"`
	TokenHash    string     `json:"-" bun:"token_hash,notnull"` // Hashed token
	DeviceType   string     `json:"device_type,omitempty" bun:"device_type"`
	OS           string     `json:"os,omitempty" bun:"os"`
	Browser      string     `json:"browser,omitempty" bun:"browser"`
	IPAddress    string     `json:"ip_address,omitempty" bun:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" bun:"user_agent"`
	LastActiveAt time.Time  `json:"last_active_at" bun:"last_active_at,nullzero,notnull,default:current_timestamp"`
	SessionName  string     `json:"session_name,omitempty" bun:"session_name"`
	ExpiresAt    time.Time  `json:"expires_at" bun:"expires_at,nullzero,notnull"`
	CreatedAt    time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" bun:"revoked_at"`

	// Relation
	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	bun.BaseModel `bun:"table:token_blacklist"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	TokenHash     string     `json:"-" bun:"token_hash,notnull"`
	UserID        *uuid.UUID `json:"user_id,omitempty" bun:"user_id,type:uuid"`
	ExpiresAt     time.Time  `json:"expires_at" bun:"expires_at,nullzero,notnull"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`

	// Relation
	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	// JWT access token for API authentication
	AccessToken string `json:"access_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Refresh token for obtaining new access tokens
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Authenticated user information
	User *User `json:"user,omitempty"`
	// Access token expiration time in seconds
	ExpiresIn int64 `json:"expires_in,omitempty" example:"900"`
	// Whether 2FA verification is required
	Requires2FA bool `json:"requires_2fa,omitempty" example:"false"`
	// Temporary token for 2FA verification (if 2FA is required)
	TwoFactorToken string `json:"two_factor_token,omitempty" example:"temp_2fa_token_xyz"`
}

// TwoFactorLoginVerifyRequest represents 2FA verification during login
type TwoFactorLoginVerifyRequest struct {
	// Temporary 2FA token from initial login response
	TwoFactorToken string `json:"two_factor_token" binding:"required" example:"temp_2fa_token_xyz"`
	// 6-digit TOTP code from authenticator app
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// JWTClaims represents custom JWT claims
type JWTClaims struct {
	// User's unique identifier
	UserID uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User's email address
	Email string `json:"email" example:"user@example.com"`
	// User's username
	Username string `json:"username" example:"johndoe"`
	// User's role names
	Roles []string `json:"roles" example:"user,admin"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the refresh token is revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
