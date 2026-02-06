package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID        *uuid.UUID `json:"user_id,omitempty" bun:"user_id,type:uuid"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	Action        string     `json:"action" bun:"action"`
	ResourceType string     `json:"resource_type,omitempty" bun:"resource_type"`
	ResourceID   string     `json:"resource_id,omitempty" bun:"resource_id"`
	IPAddress    string     `json:"ip_address,omitempty" bun:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" bun:"user_agent"`
	Status       string     `json:"status" bun:"status"`
	Details      []byte     `json:"details,omitempty" bun:"details,type:jsonb"` // JSONB in PostgreSQL
	CountryCode  string     `json:"country_code,omitempty" bun:"country_code"`
	CountryName  string     `json:"country_name,omitempty" bun:"country_name"`
	City         string     `json:"city,omitempty" bun:"city"`
	Latitude     float64    `json:"latitude,omitempty" bun:"latitude"`
	Longitude    float64    `json:"longitude,omitempty" bun:"longitude"`
	CreatedAt    time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	User         *User      `bun:"rel:belongs-to,join:user_id=id" json:"-"`
}

// AuditAction represents the type of action being logged
type AuditAction string

const (
	ActionSignUp                     AuditAction = "signup"
	ActionSignIn                     AuditAction = "signin"
	ActionSignInFailed               AuditAction = "signin_failed"
	ActionSignOut                    AuditAction = "signout"
	ActionRefreshToken               AuditAction = "refresh_token"
	ActionChangePassword             AuditAction = "change_password"
	ActionForgotPassword             AuditAction = "forgot_password"
	ActionResetPassword              AuditAction = "reset_password"
	ActionOAuthBegin                 AuditAction = "oauth_begin"
	ActionOAuthCallback              AuditAction = "oauth_callback"
	ActionUpdateProfile              AuditAction = "update_profile"
	ActionRoleAssigned               AuditAction = "role_assigned"
	ActionRoleRevoked                AuditAction = "role_revoked"
	ActionRolesUpdated               AuditAction = "roles_updated"
	ActionCreate                     AuditAction = "create"
	ActionUpdate                     AuditAction = "update"
	ActionDelete                     AuditAction = "delete"
	ActionSessionRevoked             AuditAction = "session_revoked"
	Action2FAReset                   AuditAction = "2fa_reset"
	ActionAdminPasswordResetInitiate AuditAction = "admin_password_reset_initiated"
	ActionTest                       AuditAction = "test"
	ActionSend                       AuditAction = "send"
)

// AuditResource represents the type of resource being audited
type AuditResource string

const (
	ResourceWebhook              AuditResource = "webhook"
	ResourceEmailTemplate        AuditResource = "email_template"
	ResourceEmailProvider        AuditResource = "email_provider"
	ResourceEmailProfile         AuditResource = "email_profile"
	ResourceEmailProfileTemplate AuditResource = "email_profile_template"
	ResourceEmail                AuditResource = "email"
	ResourceEmailLog             AuditResource = "email_log"
)

// AuditStatus represents the status of an audited action
type AuditStatus string

const (
	StatusSuccess AuditStatus = "success"
	StatusFailed  AuditStatus = "failed"
	StatusFailure AuditStatus = "failure"
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
