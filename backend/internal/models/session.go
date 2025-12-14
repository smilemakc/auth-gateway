package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an enhanced refresh token with device tracking
type Session struct {
	ID           uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"user_id" bun:"user_id,type:uuid,notnull"`
	TokenHash    string     `json:"-" bun:"token_hash,notnull"`              // Never expose hash
	DeviceType   string     `json:"device_type,omitempty" bun:"device_type"` // "mobile", "desktop", "tablet"
	OS           string     `json:"os,omitempty" bun:"os"`                   // "iOS 17.2", "Windows 11", "Ubuntu 22.04"
	Browser      string     `json:"browser,omitempty" bun:"browser"`         // "Chrome 120", "Safari 17"
	IPAddress    string     `json:"ip_address,omitempty" bun:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" bun:"user_agent"`
	SessionName  string     `json:"session_name,omitempty" bun:"session_name"`
	LastActiveAt time.Time  `json:"last_active_at" bun:"last_active_at,nullzero,notnull"`
	ExpiresAt    time.Time  `json:"expires_at" bun:"expires_at,nullzero,notnull"`
	CreatedAt    time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" bun:"revoked_at"`

	// Relation to User
	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// ActiveSessionResponse is returned to the user
type ActiveSessionResponse struct {
	// Session unique identifier
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Device type (mobile, desktop, tablet)
	DeviceType string `json:"device_type,omitempty" example:"desktop"`
	// Operating system with version
	OS string `json:"os,omitempty" example:"Windows 11"`
	// Browser name with version
	Browser string `json:"browser,omitempty" example:"Chrome 120"`
	// IP address of the session
	IPAddress string `json:"ip_address,omitempty" example:"192.168.1.1"`
	// Custom session name set by user
	SessionName string `json:"session_name,omitempty" example:"Home Desktop"`
	// Timestamp of last activity
	LastActiveAt time.Time `json:"last_active_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when session was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when session expires
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-22T10:30:00Z"`
	// Whether this is the current session
	IsCurrent bool `json:"is_current" example:"true"`
}

// RevokeSessionRequest is the request to revoke a specific session
type RevokeSessionRequest struct {
	// Session ID to revoke
	SessionID uuid.UUID `json:"session_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateSessionNameRequest is the request to update session name
type UpdateSessionNameRequest struct {
	// New session name (max 100 characters)
	SessionName string `json:"session_name" binding:"required,max=100" example:"Work Laptop"`
}

// SessionListResponse contains paginated session list
type SessionListResponse struct {
	// List of active sessions
	Sessions []ActiveSessionResponse `json:"sessions"`
	// Total number of sessions
	Total int `json:"total" example:"5"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// DeviceInfo contains parsed device information from user agent
type DeviceInfo struct {
	DeviceType string // "mobile", "desktop", "tablet", "bot", "unknown"
	OS         string // Operating system with version
	Browser    string // Browser name with version
	IsBot      bool   // Whether this is a bot/crawler
}

// SessionStats contains session statistics for admin dashboard
type SessionStats struct {
	TotalActiveSessions int            `json:"total_active_sessions"`
	SessionsByDevice    map[string]int `json:"sessions_by_device"`  // device_type -> count
	SessionsByOS        map[string]int `json:"sessions_by_os"`      // os -> count
	SessionsByBrowser   map[string]int `json:"sessions_by_browser"` // browser -> count
}
