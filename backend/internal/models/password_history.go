package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// PasswordHistory stores password history to prevent reuse
type PasswordHistory struct {
	bun.BaseModel `bun:"table:password_history"`

	ID           uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" bun:"user_id,type:uuid,notnull"`
	PasswordHash string    `json:"-" bun:"password_hash,notnull"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`

	User *User `json:"-" bun:"rel:belongs-to,join:user_id=id"`
}

// BeforeAppendModel hook for automatic timestamp management
func (p *PasswordHistory) BeforeAppendModel(ctx context.Context, query bun.QueryHook) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	return nil
}

// AccountLockout stores account lockout information
type AccountLockout struct {
	bun.BaseModel `bun:"table:account_lockouts"`

	ID                uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" bun:"user_id,type:uuid,notnull,unique"`
	FailedAttempts    int        `json:"failed_attempts" bun:"failed_attempts,notnull,default:0"`
	LockedUntil       *time.Time `json:"locked_until,omitempty" bun:"locked_until"`
	LastFailedAttempt *time.Time `json:"last_failed_attempt,omitempty" bun:"last_failed_attempt"`
	CreatedAt         time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt         time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	User *User `json:"-" bun:"rel:belongs-to,join:user_id=id"`
}

// BeforeAppendModel hook for automatic timestamp management
func (a *AccountLockout) BeforeAppendModel(ctx context.Context, query bun.QueryHook) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	a.UpdatedAt = time.Now()
	return nil
}

// IsLocked checks if the account is currently locked
func (a *AccountLockout) IsLocked() bool {
	if a.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*a.LockedUntil)
}

// AccountLockoutPolicy defines account lockout policy settings
type AccountLockoutPolicy struct {
	MaxFailedAttempts int           `json:"max_failed_attempts" example:"5"`
	LockoutDuration   time.Duration `json:"lockout_duration" example:"30m"`
	ResetAfter        time.Duration `json:"reset_after" example:"15m"` // Reset failed attempts after this duration
}
