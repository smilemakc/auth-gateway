package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Group represents an organizational group/department
type Group struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	// Unique group identifier
	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Group name (unique identifier)
	Name string `json:"name" bun:"name,notnull,unique" example:"engineering"`

	// Display name for UI
	DisplayName string `json:"display_name" bun:"display_name,notnull" example:"Engineering Department"`

	// Group description
	Description string `json:"description,omitempty" bun:"description" example:"Engineering team responsible for product development"`

	// Parent group ID for hierarchy (optional)
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty" bun:"parent_group_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Whether this is a system group (cannot be deleted)
	IsSystemGroup bool `json:"is_system_group" bun:"is_system_group,notnull,default:false" example:"false"`

	// Whether this is a dynamic group (membership based on rules)
	IsDynamic bool `json:"is_dynamic" bun:"is_dynamic,notnull,default:false" example:"false"`

	// Dynamic group membership rules (JSON, used when IsDynamic is true)
	// Example: {"department": "Engineering", "role": "developer"}
	MembershipRules map[string]interface{} `json:"membership_rules,omitempty" bun:"membership_rules,type:jsonb"`

	// Group-based permissions (permissions inherited by group members)
	// This is a JSON array of permission IDs
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty" bun:"permission_ids,type:uuid[]"`

	// Timestamp when group was created
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	// Timestamp when group was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	// Relations
	// Note: For self-referential relations, bun may auto-detect based on ParentGroupID field name
	// We explicitly define the relation to avoid conflicts
	ParentGroup *Group   `json:"parent_group,omitempty" bun:"rel:belongs-to,join:parent_group_id=id"`
	ChildGroups []*Group `json:"child_groups,omitempty" bun:"rel:has-many,join:id=parent_group_id"`
	Users       []User   `json:"users,omitempty" bun:"m2m:user_groups,join:Group=User"`
}

// UserGroup represents the many-to-many relationship between users and groups
type UserGroup struct {
	bun.BaseModel `bun:"table:user_groups"`

	// Unique identifier
	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`

	// User ID
	UserID uuid.UUID `json:"user_id" bun:"user_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Group ID
	GroupID uuid.UUID `json:"group_id" bun:"group_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Timestamp when user was added to group
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	// Relations
	User  *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Group *Group `json:"group,omitempty" bun:"rel:belongs-to,join:group_id=id"`
}

// CreateGroupRequest represents a request to create a new group
type CreateGroupRequest struct {
	// Group name (unique identifier)
	Name string `json:"name" binding:"required" example:"engineering"`

	// Display name for UI
	DisplayName string `json:"display_name" binding:"required" example:"Engineering Department"`

	// Group description
	Description string `json:"description,omitempty" example:"Engineering team responsible for product development"`

	// Parent group ID for hierarchy (optional)
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	// Display name for UI
	DisplayName *string `json:"display_name,omitempty" example:"Engineering Department"`

	// Group description
	Description *string `json:"description,omitempty" example:"Engineering team responsible for product development"`

	// Parent group ID for hierarchy (optional)
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AddGroupMembersRequest represents a request to add users to a group
type AddGroupMembersRequest struct {
	// List of user IDs to add
	UserIDs []uuid.UUID `json:"user_ids" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	ID            uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name          string     `json:"name" example:"engineering"`
	DisplayName   string     `json:"display_name" example:"Engineering Department"`
	Description   string     `json:"description,omitempty" example:"Engineering team responsible for product development"`
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	IsSystemGroup bool       `json:"is_system_group" example:"false"`
	MemberCount   int        `json:"member_count" example:"15"`
	CreatedAt     time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// GroupListResponse represents a paginated list of groups
type GroupListResponse struct {
	Groups []GroupResponse `json:"groups"`
	Total  int             `json:"total" example:"50"`
	Page   int             `json:"page" example:"1"`
	Size   int             `json:"size" example:"20"`
}
