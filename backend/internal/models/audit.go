package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Action       string     `json:"action" db:"action"`
	ResourceType string     `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   string     `json:"resource_id,omitempty" db:"resource_id"`
	IPAddress    string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" db:"user_agent"`
	Status       string     `json:"status" db:"status"`
	Details      []byte     `json:"details,omitempty" db:"details"` // JSONB in PostgreSQL
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// AuditAction represents the type of action being logged
type AuditAction string

const (
	ActionSignUp         AuditAction = "signup"
	ActionSignIn         AuditAction = "signin"
	ActionSignInFailed   AuditAction = "signin_failed"
	ActionSignOut        AuditAction = "signout"
	ActionRefreshToken   AuditAction = "refresh_token"
	ActionChangePassword AuditAction = "change_password"
	ActionForgotPassword AuditAction = "forgot_password"
	ActionResetPassword  AuditAction = "reset_password"
	ActionOAuthBegin     AuditAction = "oauth_begin"
	ActionOAuthCallback  AuditAction = "oauth_callback"
	ActionUpdateProfile  AuditAction = "update_profile"
)

// AuditStatus represents the status of an audited action
type AuditStatus string

const (
	StatusSuccess AuditStatus = "success"
	StatusFailed  AuditStatus = "failed"
	StatusBlocked AuditStatus = "blocked"
)

// CreateAuditLog is a helper to create a new audit log entry
func CreateAuditLog(userID *uuid.UUID, action AuditAction, status AuditStatus, ip, userAgent string, details []byte) *AuditLog {
	return &AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    string(action),
		IPAddress: ip,
		UserAgent: userAgent,
		Status:    string(status),
		Details:   details,
		CreatedAt: time.Now(),
	}
}
