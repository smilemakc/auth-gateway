package models

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorSetupRequest represents a request to enable 2FA
type TwoFactorSetupRequest struct {
	// User's current password for verification
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// TwoFactorSetupResponse contains the TOTP secret and QR code
type TwoFactorSetupResponse struct {
	// TOTP secret for manual entry
	Secret string `json:"secret" example:"JBSWY3DPEHPK3PXP"`
	// QR code URL for scanning with authenticator app
	QRCodeURL string `json:"qr_code_url" example:"otpauth://totp/AuthGateway:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=AuthGateway"`
	// List of backup codes for account recovery
	BackupCodes []string `json:"backup_codes" example:"123456,234567,345678,456789,567890"`
}

// TwoFactorVerifyRequest represents a request to verify 2FA code
type TwoFactorVerifyRequest struct {
	// 6-digit TOTP code from authenticator app
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// TwoFactorDisableRequest represents a request to disable 2FA
type TwoFactorDisableRequest struct {
	// User's current password for verification
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
	// 6-digit TOTP code from authenticator app
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// TwoFactorStatusResponse contains the user's 2FA status
type TwoFactorStatusResponse struct {
	// Whether 2FA is enabled
	Enabled bool `json:"enabled" example:"true"`
	// Timestamp when 2FA was enabled
	EnabledAt *time.Time `json:"enabled_at,omitempty" example:"2024-01-15T10:30:00Z"`
	// Number of remaining backup codes
	BackupCodes int `json:"backup_codes_remaining" example:"3"`
}

// TwoFactorLoginRequest represents a 2FA code submission during login
type TwoFactorLoginRequest struct {
	// User ID requesting 2FA verification
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 6-digit TOTP code from authenticator app
	Code string `json:"code" binding:"required" example:"123456"`
}

// BackupCode represents a backup code for 2FA
type BackupCode struct {
	ID        uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" bun:"user_id,type:uuid"`
	CodeHash  string     `json:"-" bun:"code_hash"` // Never expose the hash
	Used      bool       `json:"used" bun:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" bun:"used_at"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
