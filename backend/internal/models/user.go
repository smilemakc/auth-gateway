package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	// Unique user identifier
	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User's email address
	Email string `json:"email" bun:"email,notnull,unique" example:"user@example.com"`
	// User's phone number (optional)
	Phone *string `json:"phone,omitempty" bun:"phone" example:"+1234567890"`
	// Unique username for the user
	Username string `json:"username" bun:"username,notnull,unique" example:"johndoe"`
	// Password hash (never exposed in responses)
	PasswordHash string `json:"-" bun:"password_hash,notnull"`
	// User's full name
	FullName string `json:"full_name,omitempty" bun:"full_name" example:"John Doe"`
	// URL to user's profile picture
	ProfilePictureURL string `json:"profile_picture_url,omitempty" bun:"profile_picture_url" example:"https://example.com/avatars/user.jpg"`
	// Account type: "human" or "service"
	AccountType string `json:"account_type" bun:"account_type" example:"human"`
	// Whether email has been verified
	EmailVerified bool `json:"email_verified" bun:"email_verified" example:"true"`
	// Timestamp when email was verified
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" bun:"email_verified_at" example:"2024-01-15T10:30:00Z"`
	// Whether phone number has been verified
	PhoneVerified bool `json:"phone_verified" bun:"phone_verified" example:"false"`
	// Whether the account is active
	IsActive bool `json:"is_active" bun:"is_active" example:"true"`
	// TOTP secret for 2FA (never exposed in responses)
	TOTPSecret *string `json:"-" bun:"totp_secret"`
	// Whether TOTP 2FA is enabled
	TOTPEnabled bool `json:"totp_enabled" bun:"totp_enabled" example:"false"`
	// Timestamp when TOTP 2FA was enabled
	TOTPEnabledAt *time.Time `json:"totp_enabled_at,omitempty" bun:"totp_enabled_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when password expires (optional, for password expiry policy)
	PasswordExpiresAt *time.Time `json:"password_expires_at,omitempty" bun:"password_expires_at" example:"2024-02-15T10:30:00Z"`
	// Timestamp when password was last changed
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty" bun:"password_changed_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when user was created
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	// Timestamp when user was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	// Relations
	Roles               []Role                     `json:"roles,omitempty" bun:"m2m:user_roles,join:User=Role"`
	ApplicationProfiles []*UserApplicationProfile `json:"-" bun:"rel:has-many,join:id=user_id"`
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
	// User's email address (optional if phone is provided)
	Email string `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	// User's phone number (optional)
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// Unique username (3-100 characters)
	Username string `json:"username,omitempty" binding:"required,min=3,max=100" example:"johndoe"`
	// User's password (minimum 8 characters)
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	// User's full name (optional)
	FullName string `json:"full_name,omitempty" example:"John Doe"`
	// Account type: "human" or "service" (defaults to "human")
	AccountType string `json:"account_type,omitempty" example:"human"`
}

// UpdateUserRequest represents a request to update user profile
type UpdateUserRequest struct {
	// User's full name
	FullName string `json:"full_name,omitempty" example:"John Doe"`
	// URL to user's profile picture
	ProfilePictureURL string `json:"profile_picture_url,omitempty" example:"https://example.com/avatars/user.jpg"`
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	// Current password
	OldPassword string `json:"old_password" binding:"required" example:"OldPass123!"`
	// New password (minimum 8 characters)
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewSecurePass123!"`
}

// SignInRequest represents a sign-in request
type SignInRequest struct {
	// User's email address (required if phone not provided)
	Email string `json:"email" binding:"omitempty" example:"user@example.com"`
	// User's phone number (optional)
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// User's password
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	// Refresh token obtained during authentication
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	// Email address to send password reset code
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest represents a reset password request with OTP
type ResetPasswordRequest struct {
	// Email address associated with the account
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 6-digit verification code sent to email
	Code string `json:"code" binding:"required,len=6" example:"123456"`
	// New password (minimum 8 characters)
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewSecurePass123!"`
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

// InitPasswordlessRegistrationRequest represents a request to initiate passwordless registration
type InitPasswordlessRegistrationRequest struct {
	// User's email address (optional if phone is provided)
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	// User's phone number (optional if email is provided)
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// Optional username (will be auto-generated if not provided)
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=100" example:"johndoe"`
	// User's full name (optional)
	FullName string `json:"full_name,omitempty" example:"John Doe"`
}

// CompletePasswordlessRegistrationRequest represents a request to complete passwordless registration with OTP
type CompletePasswordlessRegistrationRequest struct {
	// User's email address (must match init request)
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	// User's phone number (must match init request)
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// 6-digit OTP code received via email or SMS
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// PendingRegistration represents temporary registration data stored in Redis
type PendingRegistration struct {
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Username  string `json:"username"`
	FullName  string `json:"full_name,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// SyncUserResponse represents a user in sync response
type SyncUserResponse struct {
	ID            uuid.UUID           `json:"id"`
	Email         string              `json:"email"`
	Username      string              `json:"username"`
	FullName      string              `json:"full_name,omitempty"`
	IsActive      bool                `json:"is_active"`
	EmailVerified bool                `json:"email_verified"`
	UpdatedAt     time.Time           `json:"updated_at"`
	AppProfile    *SyncUserAppProfile `json:"app_profile,omitempty"`
}

// SyncUserAppProfile represents the user's app profile in sync response
type SyncUserAppProfile struct {
	DisplayName string   `json:"display_name,omitempty"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	AppRoles    []string `json:"app_roles,omitempty"`
	IsActive    bool     `json:"is_active"`
	IsBanned    bool     `json:"is_banned"`
}

// SyncUsersResponse represents the response for users sync
type SyncUsersResponse struct {
	Users         []SyncUserResponse `json:"users"`
	Total         int                `json:"total"`
	HasMore       bool               `json:"has_more"`
	SyncTimestamp string             `json:"sync_timestamp"`
}

// BulkImportUserEntry represents a single user in the bulk import request
type BulkImportUserEntry struct {
	ID                    *uuid.UUID `json:"id,omitempty"`
	Email                 string     `json:"email" binding:"required,email"`
	Username              string     `json:"username,omitempty"`
	PasswordHashImport    string     `json:"password_hash_import,omitempty"`
	FullName              string     `json:"full_name,omitempty"`
	IsActive              *bool      `json:"is_active,omitempty"`
	SkipEmailVerification bool       `json:"skip_email_verification,omitempty"`
	AppRoles              []string   `json:"app_roles,omitempty"`
}

// BulkImportUsersRequest represents a bulk import request with conflict resolution
type BulkImportUsersRequest struct {
	Users      []BulkImportUserEntry `json:"users" binding:"required,min=1,max=1000"`
	OnConflict string                `json:"on_conflict" binding:"required,oneof=skip update error"`
}

// ImportDetail represents the result for a single user import
type ImportDetail struct {
	Email  string `json:"email"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// ImportUsersResponse represents the bulk import response
type ImportUsersResponse struct {
	Imported int            `json:"imported"`
	Skipped  int            `json:"skipped"`
	Updated  int            `json:"updated"`
	Errors   int            `json:"errors"`
	Details  []ImportDetail `json:"details,omitempty"`
}
