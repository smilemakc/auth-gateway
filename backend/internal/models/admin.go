package models

import (
	"time"

	"github.com/google/uuid"
)

// RoleInfo represents basic role information for API responses
type RoleInfo struct {
	// Role's unique identifier
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Role's system name
	Name string `json:"name" example:"admin"`
	// Role's display name
	DisplayName string `json:"display_name" example:"Administrator"`
}

// AdminUserResponse represents a user in admin panel
type AdminUserResponse struct {
	// User's unique identifier
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User's email address
	Email string `json:"email" example:"user@example.com"`
	// User's phone number
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// User's username
	Username string `json:"username" example:"johndoe"`
	// User's full name
	FullName string `json:"full_name,omitempty" example:"John Doe"`
	// URL to user's profile picture
	ProfilePictureURL string `json:"profile_picture_url,omitempty" example:"https://example.com/avatars/user.jpg"`
	// User's assigned roles
	Roles []RoleInfo `json:"roles"`
	// Account type: "human" or "service"
	AccountType string `json:"account_type" example:"human"`
	// Whether email has been verified
	EmailVerified bool `json:"email_verified" example:"true"`
	// Whether phone has been verified
	PhoneVerified bool `json:"phone_verified" example:"false"`
	// Whether the account is active
	IsActive bool `json:"is_active" example:"true"`
	// Whether TOTP 2FA is enabled
	TOTPEnabled bool `json:"totp_enabled" example:"false"`
	// Timestamp when TOTP 2FA was enabled
	TOTPEnabledAt *time.Time `json:"totp_enabled_at,omitempty" example:"2024-01-15T10:30:00Z"`
	// Timestamp of last login
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-01-15T10:30:00Z"`
	// Timestamp when user was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when user was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	// Number of API keys owned by user
	APIKeysCount int `json:"api_keys_count" example:"3"`
	// Number of OAuth accounts linked to user
	OAuthAccountsCount int `json:"oauth_accounts_count" example:"2"`
}

// AdminUpdateUserRequest represents admin user update request
type AdminUpdateUserRequest struct {
	// Role IDs to assign to the user
	RoleIDs *[]uuid.UUID `json:"role_ids,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174001"`
	// Whether the account should be active
	IsActive *bool `json:"is_active,omitempty" example:"true"`
	// User's email address
	Email *string `json:"email,omitempty" example:"user@example.com"`
	// Unique username
	Username *string `json:"username,omitempty" example:"johndoe"`
	// User's full name
	FullName *string `json:"full_name,omitempty" example:"John Doe"`
	// User's phone number
	Phone *string `json:"phone,omitempty" example:"+1234567890"`
	// Whether email has been verified
	EmailVerified *bool `json:"email_verified,omitempty" example:"true"`
}

// AdminCreateUserRequest represents admin user creation request
type AdminCreateUserRequest struct {
	// User's email address
	Email string `json:"email" binding:"required,email" example:"newuser@example.com"`
	// Unique username (3-100 characters)
	Username string `json:"username" binding:"required,min=3,max=100" example:"newuser"`
	// User's password (minimum 8 characters)
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	// User's full name
	FullName string `json:"full_name" binding:"required" example:"New User"`
	// Role IDs to assign to the user
	RoleIDs []uuid.UUID `json:"role_ids" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Account type: "human" or "service" (defaults to "human")
	AccountType string `json:"account_type" example:"human"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	// Role ID to assign
	RoleID uuid.UUID `json:"role_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AdminStatsResponse represents system statistics
type AdminStatsResponse struct {
	// Total number of users in the system
	TotalUsers int `json:"total_users" example:"1250"`
	// Number of active users
	ActiveUsers int `json:"active_users" example:"1100"`
	// Number of users with verified email
	VerifiedEmailUsers int `json:"verified_email_users" example:"950"`
	// Number of users with verified phone
	VerifiedPhoneUsers int `json:"verified_phone_users" example:"650"`
	// Number of users with 2FA enabled
	Users2FAEnabled int `json:"users_2fa_enabled" example:"320"`
	// Total number of API keys
	TotalAPIKeys int `json:"total_api_keys" example:"45"`
	// Number of active API keys
	ActiveAPIKeys int `json:"active_api_keys" example:"38"`
	// Total number of OAuth accounts
	TotalOAuthAccounts int `json:"total_oauth_accounts" example:"425"`
	// Users count by role (users with multiple roles are counted in each)
	UsersByRole map[string]int `json:"users_by_role" example:"admin:5,user:1200,moderator:15"`
	// Number of signups in the last 24 hours
	RecentSignups int `json:"recent_signups_24h" example:"12"`
	// Number of logins in the last 24 hours
	RecentLogins int `json:"recent_logins_24h" example:"450"`
}

// AdminUserListResponse represents paginated user list
type AdminUserListResponse struct {
	// List of users
	Users []*AdminUserResponse `json:"users"`
	// Total number of users matching the query
	Total int `json:"total" example:"1250"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PageSize int `json:"page_size" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"63"`
}

// AdminAuditLogResponse represents an audit log entry
type AdminAuditLogResponse struct {
	// Audit log entry ID
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User ID who performed the action
	UserID *uuid.UUID `json:"user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User email who performed the action
	UserEmail string `json:"user_email,omitempty" example:"admin@example.com"`
	// Action performed
	Action string `json:"action" example:"signin"`
	// Action status
	Status string `json:"status" example:"success"`
	// IP address of the request
	IP string `json:"ip" example:"192.168.1.1"`
	// User agent string
	UserAgent string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`
	// Additional details about the action
	Details map[string]interface{} `json:"details,omitempty"`
	// Timestamp when action was performed
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// AdminAPIKeyResponse represents an API key in admin panel
type AdminAPIKeyResponse struct {
	// API key ID
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// User ID who owns the key
	UserID uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Username of the key owner
	Username string `json:"username,omitempty" example:"johndoe"`
	// Email of the key owner
	UserEmail string `json:"user_email,omitempty" example:"john@example.com"`
	// Display name of the key owner
	OwnerName string `json:"owner_name,omitempty" example:"John Doe"`
	// API key name
	Name string `json:"name" example:"Production API Key"`
	// API key prefix (first 12 characters)
	KeyPrefix string `json:"key_prefix" example:"agw_abc123de"`
	// API key scopes/permissions
	Scopes []string `json:"scopes" example:"users:read,token:validate"`
	// Expiration timestamp (null if never expires)
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	// Last usage timestamp
	LastUsedAt *time.Time `json:"last_used_at,omitempty" example:"2024-01-15T10:30:00Z"`
	// Whether the key is active
	IsActive bool `json:"is_active" example:"true"`
	// Timestamp when key was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty" example:"2024-01-15T10:30:00Z"`
	// Timestamp when key was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// AdminAPIKeyListResponse represents paginated admin API key list
type AdminAPIKeyListResponse struct {
	// List of API keys
	APIKeys []*AdminAPIKeyResponse `json:"api_keys"`
	// Total number of API keys
	Total int `json:"total" example:"10"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PageSize int `json:"page_size" example:"50"`
	// Total number of pages
	TotalPages int `json:"total_pages"`
}

// AuditLogListResponse represents paginated audit log list
type AuditLogListResponse struct {
	// List of audit logs
	Logs []*AdminAuditLogResponse `json:"logs"`
	// Total number of audit logs
	Total int `json:"total" example:"150"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PageSize int `json:"page_size" example:"50"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"3"`
}

// OAuthAccountListResponse represents user OAuth accounts list
type OAuthAccountListResponse struct {
	// List of OAuth accounts
	Accounts []*OAuthAccount `json:"accounts"`
	// Total number of accounts
	Total int `json:"total" example:"3"`
}
