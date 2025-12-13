package models

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorSetupRequest represents a request to enable 2FA
type TwoFactorSetupRequest struct {
	Password string `json:"password" binding:"required"`
}

// TwoFactorSetupResponse contains the TOTP secret and QR code
type TwoFactorSetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// TwoFactorVerifyRequest represents a request to verify 2FA code
type TwoFactorVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// TwoFactorDisableRequest represents a request to disable 2FA
type TwoFactorDisableRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}

// TwoFactorStatusResponse contains the user's 2FA status
type TwoFactorStatusResponse struct {
	Enabled     bool       `json:"enabled"`
	EnabledAt   *time.Time `json:"enabled_at,omitempty"`
	BackupCodes int        `json:"backup_codes_remaining"`
}

// TwoFactorLoginRequest represents a 2FA code submission during login
type TwoFactorLoginRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Code   string    `json:"code" binding:"required"`
}

// BackupCode represents a backup code for 2FA
type BackupCode struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	CodeHash  string     `json:"-" db:"code_hash"` // Never expose the hash
	Used      bool       `json:"used" db:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
