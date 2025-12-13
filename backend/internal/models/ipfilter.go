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
	IPCIDR     string     `json:"ip_cidr" binding:"required"`
	FilterType string     `json:"filter_type" binding:"required,oneof=whitelist blacklist"`
	Reason     string     `json:"reason"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// UpdateIPFilterRequest is the request to update an IP filter
type UpdateIPFilterRequest struct {
	Reason    string     `json:"reason"`
	IsActive  *bool      `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// IPFilterListResponse contains paginated IP filter list
type IPFilterListResponse struct {
	Filters    []IPFilterWithCreator `json:"filters"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	TotalPages int                   `json:"total_pages"`
}

// CheckIPRequest is used to check if an IP is allowed
type CheckIPRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
}

// CheckIPResponse returns whether the IP is allowed
type CheckIPResponse struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason,omitempty"`
	FilterType string `json:"filter_type,omitempty"` // "whitelist", "blacklist", or ""
}
