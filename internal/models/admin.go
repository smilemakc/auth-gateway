package models

import (
	"time"

	"github.com/google/uuid"
)

// AdminUserResponse represents a user in admin panel
type AdminUserResponse struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Phone             *string    `json:"phone,omitempty"`
	Username          string     `json:"username"`
	FullName          string     `json:"full_name,omitempty"`
	Role              string     `json:"role"`
	EmailVerified     bool       `json:"email_verified"`
	PhoneVerified     bool       `json:"phone_verified"`
	IsActive          bool       `json:"is_active"`
	TOTPEnabled       bool       `json:"totp_enabled"`
	TOTPEnabledAt     *time.Time `json:"totp_enabled_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	APIKeysCount      int        `json:"api_keys_count"`
	OAuthAccountsCount int       `json:"oauth_accounts_count"`
}

// AdminUpdateUserRequest represents admin user update request
type AdminUpdateUserRequest struct {
	Role     *string `json:"role,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// AdminStatsResponse represents system statistics
type AdminStatsResponse struct {
	TotalUsers           int            `json:"total_users"`
	ActiveUsers          int            `json:"active_users"`
	VerifiedEmailUsers   int            `json:"verified_email_users"`
	VerifiedPhoneUsers   int            `json:"verified_phone_users"`
	Users2FAEnabled      int            `json:"users_2fa_enabled"`
	TotalAPIKeys         int            `json:"total_api_keys"`
	ActiveAPIKeys        int            `json:"active_api_keys"`
	TotalOAuthAccounts   int            `json:"total_oauth_accounts"`
	UsersByRole          map[string]int `json:"users_by_role"`
	RecentSignups        int            `json:"recent_signups_24h"`
	RecentLogins         int            `json:"recent_logins_24h"`
}

// AdminUserListResponse represents paginated user list
type AdminUserListResponse struct {
	Users      []*AdminUserResponse `json:"users"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// AdminAuditLogResponse represents an audit log entry
type AdminAuditLogResponse struct {
	ID        uuid.UUID              `json:"id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Status    string                 `json:"status"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AdminAPIKeyResponse represents an API key in admin panel
type AdminAPIKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Username    string     `json:"username"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"` // First 12 characters
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsRevoked   bool       `json:"is_revoked"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
