package models

import (
	"github.com/google/uuid"
)

// BulkCreateUsersRequest represents a request to create multiple users
type BulkCreateUsersRequest struct {
	Users []BulkUserCreate `json:"users" binding:"required"`
}

// BulkUserCreate represents a single user to create in bulk operation
type BulkUserCreate struct {
	Email         string `json:"email" binding:"required"`
	Username      string `json:"username" binding:"required"`
	FullName      string `json:"full_name,omitempty"`
	Password      string `json:"password" binding:"required"`
	IsActive      bool   `json:"is_active,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

// BulkUpdateUsersRequest represents a request to update multiple users
type BulkUpdateUsersRequest struct {
	Users []BulkUserUpdate `json:"users" binding:"required"`
}

// BulkUserUpdate represents a single user to update in bulk operation
type BulkUserUpdate struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	Email    *string   `json:"email,omitempty"`
	Username *string   `json:"username,omitempty"`
	FullName *string   `json:"full_name,omitempty"`
	IsActive *bool     `json:"is_active,omitempty"`
}

// BulkDeleteUsersRequest represents a request to delete multiple users
type BulkDeleteUsersRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required"`
}

// BulkAssignRolesRequest represents a request to assign roles to multiple users
type BulkAssignRolesRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required"`
	RoleIDs []uuid.UUID `json:"role_ids" binding:"required"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Total   int                       `json:"total" example:"100"`
	Success int                       `json:"success" example:"95"`
	Failed  int                       `json:"failed" example:"5"`
	Errors  []BulkOperationError      `json:"errors,omitempty"`
	Results []BulkOperationItemResult `json:"results,omitempty"`
}

// BulkOperationError represents an error for a specific item in bulk operation
type BulkOperationError struct {
	Index   int    `json:"index" example:"5"`
	ID      string `json:"id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email   string `json:"email,omitempty" example:"user@example.com"`
	Message string `json:"message" example:"User already exists"`
}

// BulkOperationItemResult represents the result for a single item
type BulkOperationItemResult struct {
	Index   int       `json:"index" example:"0"`
	ID      uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email   string    `json:"email" example:"user@example.com"`
	Success bool      `json:"success" example:"true"`
	Message string    `json:"message,omitempty" example:"User created successfully"`
}
