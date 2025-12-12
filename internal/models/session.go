package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an enhanced refresh token with device tracking
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash    string     `json:"-" db:"token_hash"` // Never expose hash
	DeviceType   string     `json:"device_type,omitempty" db:"device_type"` // "mobile", "desktop", "tablet"
	OS           string     `json:"os,omitempty" db:"os"`                   // "iOS 17.2", "Windows 11", "Ubuntu 22.04"
	Browser      string     `json:"browser,omitempty" db:"browser"`         // "Chrome 120", "Safari 17"
	IPAddress    string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" db:"user_agent"`
	SessionName  string     `json:"session_name,omitempty" db:"session_name"`
	LastActiveAt time.Time  `json:"last_active_at" db:"last_active_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// ActiveSessionResponse is returned to the user
type ActiveSessionResponse struct {
	ID           uuid.UUID `json:"id"`
	DeviceType   string    `json:"device_type,omitempty"`
	OS           string    `json:"os,omitempty"`
	Browser      string    `json:"browser,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	SessionName  string    `json:"session_name,omitempty"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsCurrent    bool      `json:"is_current"` // True if this is the current session
}

// RevokeSessionRequest is the request to revoke a specific session
type RevokeSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// UpdateSessionNameRequest is the request to update session name
type UpdateSessionNameRequest struct {
	SessionName string `json:"session_name" binding:"required,max=100"`
}

// SessionListResponse contains paginated session list
type SessionListResponse struct {
	Sessions   []ActiveSessionResponse `json:"sessions"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	PerPage    int                     `json:"per_page"`
	TotalPages int                     `json:"total_pages"`
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
	SessionsByDevice    map[string]int `json:"sessions_by_device"` // device_type -> count
	SessionsByOS        map[string]int `json:"sessions_by_os"`     // os -> count
	SessionsByBrowser   map[string]int `json:"sessions_by_browser"` // browser -> count
}
