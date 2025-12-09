package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"` // Hashed token
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TokenHash string     `json:"-" db:"token_hash"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
}

// JWTClaims represents custom JWT claims
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the refresh token is revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
