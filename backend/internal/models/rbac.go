package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name        string    `json:"name" bun:"name,notnull,unique" binding:"required"`  // e.g., "users.delete", "api_keys.view"
	Resource    string    `json:"resource" bun:"resource,notnull" binding:"required"` // e.g., "users", "api_keys"
	Action      string    `json:"action" bun:"action,notnull" binding:"required"`     // e.g., "create", "read", "update", "delete"
	Description string    `json:"description,omitempty" bun:"description"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// Role represents a user role
type Role struct {
	ID           uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name         string    `json:"name" bun:"name,notnull,unique" binding:"required"`
	DisplayName  string    `json:"display_name" bun:"display_name,notnull" binding:"required"`
	Description  string    `json:"description,omitempty" bun:"description"`
	IsSystemRole bool      `json:"is_system_role" bun:"is_system_role"`
	CreatedAt    time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	// Many-to-many relation with Permission
	Permissions []Permission `json:"permissions,omitempty" bun:"m2m:role_permissions,join:Role=Permission"`
}

// RolePermission represents the many-to-many relationship join table
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" bun:"role_id,pk,type:uuid"`
	PermissionID uuid.UUID `json:"permission_id" bun:"permission_id,pk,type:uuid"`
	GrantedAt    time.Time `json:"granted_at" bun:"granted_at,nullzero,notnull,default:current_timestamp"`

	// Belongs-to relations
	Role       *Role       `bun:"rel:belongs-to,join:role_id=id"`
	Permission *Permission `bun:"rel:belongs-to,join:permission_id=id"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID     uuid.UUID  `json:"user_id" bun:"user_id,pk,type:uuid"`
	RoleID     uuid.UUID  `json:"role_id" bun:"role_id,pk,type:uuid"`
	AssignedAt time.Time  `json:"assigned_at" bun:"assigned_at,nullzero,notnull,default:current_timestamp"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty" bun:"assigned_by,type:uuid"`

	// Belongs-to relations
	User *User `bun:"rel:belongs-to,join:user_id=id"`
	Role *Role `bun:"rel:belongs-to,join:role_id=id"`
}

// RoleType represents the type of role
type RoleType string

const (
	RoleAdmin     RoleType = "admin"
	RoleModerator RoleType = "moderator"
	RoleUser      RoleType = "user"
)

// ============================================================
// Request/Response Models
// ============================================================

// CreatePermissionRequest is the request body for creating a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Resource    string `json:"resource" binding:"required,min=2,max=50"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description"`
}

// UpdatePermissionRequest is the request body for updating a permission
type UpdatePermissionRequest struct {
	Description string `json:"description"`
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Name        string      `json:"name" binding:"required,min=2,max=50"`
	DisplayName string      `json:"display_name" binding:"required,min=2,max=100"`
	Description string      `json:"description"`
	Permissions []uuid.UUID `json:"permissions"` // List of permission IDs
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	DisplayName string      `json:"display_name" binding:"min=2,max=100"`
	Description string      `json:"description"`
	Permissions []uuid.UUID `json:"permissions"` // List of permission IDs to assign
}

// RolePermissionsRequest is the request to set role permissions
type RolePermissionsRequest struct {
	Permissions []uuid.UUID `json:"permissions" binding:"required"` // List of permission IDs
}

// RoleDetailResponse includes role with its permissions
type RoleDetailResponse struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	DisplayName  string       `json:"display_name"`
	Description  string       `json:"description,omitempty"`
	IsSystemRole bool         `json:"is_system_role"`
	Permissions  []Permission `json:"permissions"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// PermissionMatrix represents the permission assignment matrix for UI
type PermissionMatrix struct {
	Resources []ResourcePermissions `json:"resources"`
}

// ResourcePermissions groups permissions by resource
type ResourcePermissions struct {
	Resource    string                `json:"resource"`
	Permissions []PermissionWithRoles `json:"permissions"`
}

// PermissionWithRoles shows which roles have this permission
type PermissionWithRoles struct {
	PermissionID uuid.UUID   `json:"permission_id"`
	Name         string      `json:"name"`
	Action       string      `json:"action"`
	Description  string      `json:"description,omitempty"`
	Roles        []uuid.UUID `json:"roles"` // Role IDs that have this permission
}

// UserRolePermissions combines user, role, and permission information
type UserRolePermissions struct {
	UserID          uuid.UUID    `json:"user_id" bun:"user_id,type:uuid"`
	Username        string       `json:"username" bun:"username"`
	Email           string       `json:"email" bun:"email"`
	RoleID          uuid.UUID    `json:"role_id" bun:"role_id,type:uuid"`
	RoleName        string       `json:"role_name" bun:"role_name"`
	RoleDisplayName string       `json:"role_display_name" bun:"role_display_name"`
	Permissions     []Permission `json:"permissions" bun:"permissions"`
}

// CheckPermissionRequest is used to check if a user has a specific permission
type CheckPermissionRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	Permission string    `json:"permission" binding:"required"` // e.g., "users.delete"
}

// CheckPermissionResponse returns whether the user has the permission
type CheckPermissionResponse struct {
	HasPermission bool   `json:"has_permission"`
	RoleName      string `json:"role_name,omitempty"`
}
