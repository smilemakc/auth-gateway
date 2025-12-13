package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required"`         // e.g., "users.delete", "api_keys.view"
	Resource    string    `json:"resource" db:"resource" binding:"required"` // e.g., "users", "api_keys"
	Action      string    `json:"action" db:"action" binding:"required"`     // e.g., "create", "read", "update", "delete"
	Description string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Role represents a user role
type Role struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	Name         string       `json:"name" db:"name" binding:"required"`
	DisplayName  string       `json:"display_name" db:"display_name" binding:"required"`
	Description  string       `json:"description,omitempty" db:"description"`
	IsSystemRole bool         `json:"is_system_role" db:"is_system_role"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	Permissions  []Permission `json:"permissions,omitempty" db:"-"` // Populated via join
}

// RolePermission represents the many-to-many relationship
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
	GrantedAt    time.Time `json:"granted_at" db:"granted_at"`
}

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
	UserID          uuid.UUID    `json:"user_id" db:"user_id"`
	Username        string       `json:"username" db:"username"`
	Email           string       `json:"email" db:"email"`
	RoleID          uuid.UUID    `json:"role_id" db:"role_id"`
	RoleName        string       `json:"role_name" db:"role_name"`
	RoleDisplayName string       `json:"role_display_name" db:"role_display_name"`
	Permissions     []Permission `json:"permissions" db:"permissions"`
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
