package models

import (
	"time"

	"github.com/google/uuid"
)

// IPFilter represents an IP whitelist or blacklist entry
type IPFilter struct {
	ID         uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	IPCIDR     string     `json:"ip_cidr" bun:"ip_cidr" binding:"required"`                                   // IP address or CIDR range
	FilterType string     `json:"filter_type" bun:"filter_type" binding:"required,oneof=whitelist blacklist"` // "whitelist" or "blacklist"
	Reason     string     `json:"reason,omitempty" bun:"reason"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`
	IsActive   bool       `json:"is_active" bun:"is_active"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" bun:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt  time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// IPFilterWithCreator includes creator information
type IPFilterWithCreator struct {
	IPFilter
	CreatorUsername string `json:"creator_username,omitempty" bun:"creator_username"`
	CreatorEmail    string `json:"creator_email,omitempty" bun:"creator_email"`
}

// CreateIPFilterRequest is the request to create an IP filter
type CreateIPFilterRequest struct {
	// IP address or CIDR range (e.g., "192.168.1.1" or "192.168.1.0/24")
	IPCIDR string `json:"ip_cidr" binding:"required" example:"192.168.1.0/24"`
	// Filter type: "whitelist" or "blacklist"
	FilterType string `json:"filter_type" binding:"required,oneof=whitelist blacklist" example:"blacklist"`
	// Reason for the filter
	Reason string `json:"reason" example:"Suspicious activity detected"`
	// Optional expiration date
	ExpiresAt *time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// UpdateIPFilterRequest is the request to update an IP filter
type UpdateIPFilterRequest struct {
	// Updated reason for the filter
	Reason string `json:"reason" example:"Updated reason"`
	// Whether the filter is active
	IsActive *bool `json:"is_active" example:"true"`
	// Updated expiration date
	ExpiresAt *time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// IPFilterListResponse contains paginated IP filter list
type IPFilterListResponse struct {
	// List of IP filters
	Filters []IPFilterWithCreator `json:"filters"`
	// Total number of filters
	Total int `json:"total" example:"25"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PageSize int `json:"page_size" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"2"`
}

// CheckIPRequest is used to check if an IP is allowed
type CheckIPRequest struct {
	// IP address to check
	IPAddress string `json:"ip_address" binding:"required" example:"192.168.1.100"`
}

// CheckIPResponse returns whether the IP is allowed
type CheckIPResponse struct {
	// Whether the IP is allowed
	Allowed bool `json:"allowed" example:"true"`
	// Reason if IP is blocked
	Reason string `json:"reason,omitempty" example:"IP is blacklisted"`
	// Filter type that matched: "whitelist", "blacklist", or empty
	FilterType string `json:"filter_type,omitempty" example:"blacklist"`
}
