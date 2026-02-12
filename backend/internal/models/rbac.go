package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a system permission
type Permission struct {
	// Permission unique identifier
	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Permission name (e.g., "users.delete", "api_keys.view")
	Name          string     `json:"name" bun:"name,notnull" binding:"required" example:"users.delete"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	// Resource the permission applies to (e.g., "users", "api_keys")
	Resource string `json:"resource" bun:"resource,notnull" binding:"required" example:"users"`
	// Action allowed on the resource (e.g., "create", "read", "update", "delete")
	Action string `json:"action" bun:"action,notnull" binding:"required" example:"delete"`
	// Human-readable description of the permission
	Description string `json:"description,omitempty" bun:"description" example:"Allows deleting users from the system"`
	// Timestamp when permission was created
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
}

// Role represents a user role
type Role struct {
	// Role unique identifier
	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	// System name for the role
	Name          string     `json:"name" bun:"name,notnull" binding:"required" example:"admin"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	// Human-readable display name
	DisplayName string `json:"display_name" bun:"display_name,notnull" binding:"required" example:"Administrator"`
	// Role description
	Description string `json:"description,omitempty" bun:"description" example:"Full system access with all permissions"`
	// Whether this is a system-defined role (cannot be deleted)
	IsSystemRole bool `json:"is_system_role" bun:"is_system_role" example:"true"`
	// Timestamp when role was created
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	// Timestamp when role was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

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
	UserID        uuid.UUID  `json:"user_id" bun:"user_id,pk,type:uuid"`
	RoleID        uuid.UUID  `json:"role_id" bun:"role_id,pk,type:uuid"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	AssignedAt    time.Time  `json:"assigned_at" bun:"assigned_at,nullzero,notnull,default:current_timestamp"`
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
	// Permission name (3-100 characters)
	Name string `json:"name" binding:"required,min=3,max=100" example:"users.delete"`
	// Resource name (2-50 characters)
	Resource string `json:"resource" binding:"required,min=2,max=50" example:"users"`
	// Action name (2-50 characters)
	Action string `json:"action" binding:"required,min=2,max=50" example:"delete"`
	// Permission description
	Description string `json:"description" example:"Allows deleting users from the system"`
}

// UpdatePermissionRequest is the request body for updating a permission
type UpdatePermissionRequest struct {
	// Updated permission description
	Description string `json:"description" example:"Updated permission description"`
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	// Role system name (2-50 characters)
	Name string `json:"name" binding:"required,min=2,max=50" example:"moderator"`
	// Role display name (2-100 characters)
	DisplayName string `json:"display_name" binding:"required,min=2,max=100" example:"Moderator"`
	// Role description
	Description string `json:"description" example:"Can manage users and content"`
	// List of permission IDs to assign to the role
	Permissions []uuid.UUID `json:"permissions" example:"123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174001"`
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	// Role display name (2-100 characters)
	DisplayName string `json:"display_name" binding:"min=2,max=100" example:"Updated Moderator"`
	// Role description
	Description string `json:"description" example:"Updated role description"`
	// List of permission IDs to assign to the role
	Permissions []uuid.UUID `json:"permissions" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// RolePermissionsRequest is the request to set role permissions
type RolePermissionsRequest struct {
	// List of permission IDs to assign to the role
	Permissions []uuid.UUID `json:"permissions" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174001"`
}

// RoleDetailResponse includes role with its permissions
type RoleDetailResponse struct {
	// Role unique identifier
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Role system name
	Name string `json:"name" example:"admin"`
	// Role display name
	DisplayName string `json:"display_name" example:"Administrator"`
	// Role description
	Description string `json:"description,omitempty" example:"Full system access"`
	// Whether this is a system role
	IsSystemRole bool `json:"is_system_role" example:"true"`
	// List of permissions assigned to this role
	Permissions []Permission `json:"permissions"`
	// Timestamp when role was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when role was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
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
	// User ID to check permissions for
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Permission to check (e.g., "users.delete")
	Permission string `json:"permission" binding:"required" example:"users.delete"`
}

// CheckPermissionResponse returns whether the user has the permission
type CheckPermissionResponse struct {
	// Whether the user has the requested permission
	HasPermission bool `json:"has_permission" example:"true"`
	// Name of the role that grants this permission
	RoleName string `json:"role_name,omitempty" example:"admin"`
}
