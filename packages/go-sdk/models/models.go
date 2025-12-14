// Package models provides data types for the Auth Gateway SDK.
package models

import "time"

// User represents a user account in the system.
type User struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	Phone             *string    `json:"phone,omitempty"`
	Username          string     `json:"username"`
	FullName          string     `json:"full_name"`
	ProfilePictureURL string     `json:"profile_picture_url"`
	AccountType       string     `json:"account_type"`
	EmailVerified     bool       `json:"email_verified"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerified     bool       `json:"phone_verified"`
	IsActive          bool       `json:"is_active"`
	TOTPEnabled       bool       `json:"totp_enabled"`
	TOTPEnabledAt     *time.Time `json:"totp_enabled_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Roles             []Role     `json:"roles,omitempty"`
}

// Role represents a role in the RBAC system.
type Role struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	DisplayName  string       `json:"display_name"`
	Description  string       `json:"description"`
	IsSystemRole bool         `json:"is_system_role"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Permissions  []Permission `json:"permissions,omitempty"`
}

// Permission represents a permission in the RBAC system.
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIKey represents an API key for authentication.
type APIKey struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	KeyPrefix   string     `json:"key_prefix"`
	Scopes      []string   `json:"scopes"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Session represents an active user session.
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DeviceType   string    `json:"device_type"`
	OS           string    `json:"os"`
	Browser      string    `json:"browser"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	SessionName  string    `json:"session_name"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsCurrent    bool      `json:"is_current"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	IPAddress string    `json:"ip_address"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// OAuthAccount represents a linked OAuth account.
type OAuthAccount struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Picture    string    `json:"picture"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IPFilter represents an IP filter rule.
type IPFilter struct {
	ID          string    `json:"id"`
	IPAddress   string    `json:"ip_address"`
	Type        string    `json:"type"` // "whitelist" or "blacklist"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// OAuthProvider represents an OAuth provider configuration.
type OAuthProvider struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

// SystemStats represents system statistics.
type SystemStats struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	TotalSessions   int64 `json:"total_sessions"`
	TotalAPIKeys    int64 `json:"total_api_keys"`
	TotalRoles      int64 `json:"total_roles"`
	TotalAuditLogs  int64 `json:"total_audit_logs"`
	VerifiedUsers   int64 `json:"verified_users"`
	TwoFAEnabledUsers int64 `json:"two_fa_enabled_users"`
}

// SessionStats represents session statistics.
type SessionStats struct {
	TotalActiveSessions int64            `json:"total_active_sessions"`
	SessionsByDevice    map[string]int64 `json:"sessions_by_device"`
	SessionsByOS        map[string]int64 `json:"sessions_by_os"`
	SessionsByBrowser   map[string]int64 `json:"sessions_by_browser"`
}

// GeoDistribution represents geographic distribution of users.
type GeoDistribution struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// HealthStatus represents the system health status.
type HealthStatus struct {
	Status    string            `json:"status"`
	Database  string            `json:"database"`
	Redis     string            `json:"redis"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details,omitempty"`
}

// MaintenanceStatus represents maintenance mode status.
type MaintenanceStatus struct {
	Enabled   bool       `json:"enabled"`
	Message   string     `json:"message,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedList is a generic paginated list response.
type PaginatedList[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}
