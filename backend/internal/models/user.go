package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Email             string     `json:"email" bun:"email,notnull,unique"`
	Phone             *string    `json:"phone,omitempty" bun:"phone"`
	Username          string     `json:"username" bun:"username,notnull,unique"`
	PasswordHash      string     `json:"-" bun:"password_hash,notnull"` // Never expose password hash
	FullName          string     `json:"full_name,omitempty" bun:"full_name"`
	ProfilePictureURL string     `json:"profile_picture_url,omitempty" bun:"profile_picture_url"`
	AccountType       string     `json:"account_type" bun:"account_type"` // "human" or "service"
	EmailVerified     bool       `json:"email_verified" bun:"email_verified"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty" bun:"email_verified_at"`
	PhoneVerified     bool       `json:"phone_verified" bun:"phone_verified"`
	IsActive          bool       `json:"is_active" bun:"is_active"`
	TOTPSecret        *string    `json:"-" bun:"totp_secret"` // Never expose TOTP secret
	TOTPEnabled       bool       `json:"totp_enabled" bun:"totp_enabled"`
	TOTPEnabledAt     *time.Time `json:"totp_enabled_at,omitempty" bun:"totp_enabled_at"`
	CreatedAt         time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt         time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	// Relations
	Roles []Role `json:"roles,omitempty" bun:"m2m:user_roles,join:User=Role"`
}

// BeforeInsert hook for automatic timestamp management
func (u *User) BeforeInsert(ctx context.Context) error {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate hook for automatic timestamp management
func (u *User) BeforeUpdate(ctx context.Context) error {
	u.UpdatedAt = time.Now()
	return nil
}

// AccountType defines user account types
type AccountType string

const (
	AccountTypeHuman   AccountType = "human"
	AccountTypeService AccountType = "service"
)

// IsValidAccountType checks if an account type is valid
func IsValidAccountType(accountType string) bool {
	switch AccountType(accountType) {
	case AccountTypeHuman, AccountTypeService:
		return true
	default:
		return false
	}
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email       string  `json:"email" binding:"omitempty,email"`
	Phone       *string `json:"phone,omitempty"`
	Username    string  `json:"username" binding:"required,min=3,max=100"`
	Password    string  `json:"password" binding:"required,min=8"`
	FullName    string  `json:"full_name,omitempty"`
	AccountType string  `json:"account_type,omitempty"` // "human" or "service"
}

// UpdateUserRequest represents a request to update user profile
type UpdateUserRequest struct {
	FullName          string `json:"full_name,omitempty"`
	ProfilePictureURL string `json:"profile_picture_url,omitempty"`
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// SignInRequest represents a sign-in request
type SignInRequest struct {
	Email    string  `json:"email" binding:"omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password" binding:"required"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a reset password request with OTP
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PublicUser returns a user without sensitive information
func (u *User) PublicUser() *User {
	return &User{
		ID:                u.ID,
		Email:             u.Email,
		Phone:             u.Phone,
		Username:          u.Username,
		FullName:          u.FullName,
		ProfilePictureURL: u.ProfilePictureURL,
		AccountType:       u.AccountType,
		EmailVerified:     u.EmailVerified,
		PhoneVerified:     u.PhoneVerified,
		IsActive:          u.IsActive,
		TOTPEnabled:       u.TOTPEnabled,
		TOTPEnabledAt:     u.TOTPEnabledAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		Roles:             u.Roles,
	}
}
