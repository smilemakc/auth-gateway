package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	Phone             *string    `json:"phone,omitempty" db:"phone"`
	Username          string     `json:"username" db:"username"`
	PasswordHash      string     `json:"-" db:"password_hash"` // Never expose password hash
	FullName          string     `json:"full_name,omitempty" db:"full_name"`
	ProfilePictureURL string     `json:"profile_picture_url,omitempty" db:"profile_picture_url"`
	Role              string     `json:"role" db:"role"` // Deprecated: use RoleID instead
	RoleID            *uuid.UUID `json:"role_id,omitempty" db:"role_id"` // New RBAC role reference
	AccountType       string     `json:"account_type" db:"account_type"` // "human" or "service"
	EmailVerified     bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified     bool       `json:"phone_verified" db:"phone_verified"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	TOTPSecret        *string    `json:"-" db:"totp_secret"` // Never expose TOTP secret
	TOTPEnabled       bool       `json:"totp_enabled" db:"totp_enabled"`
	TOTPEnabledAt     *time.Time `json:"totp_enabled_at,omitempty" db:"totp_enabled_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// UserRole defines available user roles
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
)

// AccountType defines user account types
type AccountType string

const (
	AccountTypeHuman  AccountType = "human"
	AccountTypeService AccountType = "service"
)

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	switch UserRole(role) {
	case RoleUser, RoleModerator, RoleAdmin:
		return true
	default:
		return false
	}
}

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
		Role:              u.Role,
		RoleID:            u.RoleID,
		AccountType:       u.AccountType,
		EmailVerified:     u.EmailVerified,
		PhoneVerified:     u.PhoneVerified,
		IsActive:          u.IsActive,
		TOTPEnabled:       u.TOTPEnabled,
		TOTPEnabledAt:     u.TOTPEnabledAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}
