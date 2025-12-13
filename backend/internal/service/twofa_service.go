package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	BackupCodeLength = 8
	BackupCodeCount  = 10
)

// TwoFactorService provides 2FA operations
type TwoFactorService struct {
	userRepo       *repository.UserRepository
	backupCodeRepo *repository.BackupCodeRepository
	auditRepo      *repository.AuditRepository
	issuer         string
}

// NewTwoFactorService creates a new 2FA service
func NewTwoFactorService(
	userRepo *repository.UserRepository,
	backupCodeRepo *repository.BackupCodeRepository,
	auditRepo *repository.AuditRepository,
	issuer string,
) *TwoFactorService {
	return &TwoFactorService{
		userRepo:       userRepo,
		backupCodeRepo: backupCodeRepo,
		auditRepo:      auditRepo,
		issuer:         issuer,
	}
}

// SetupTOTP generates a new TOTP secret and backup codes for a user
func (s *TwoFactorService) SetupTOTP(ctx context.Context, userID uuid.UUID, password string) (*models.TwoFactorSetupResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Verify password
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, models.NewAppError(401, "Invalid password")
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: user.Email,
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA256,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.generateBackupCodes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Store TOTP secret (but don't enable yet - requires verification)
	if err := s.userRepo.UpdateTOTPSecret(userID, key.Secret()); err != nil {
		return nil, err
	}

	// Get plain text backup codes before they're hashed
	plainBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		plainBackupCodes[i] = code.CodeHash // Temporarily stored as plain text
	}

	// Hash backup codes before storing
	for i := range backupCodes {
		hash, err := utils.HashPassword(backupCodes[i].CodeHash, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		backupCodes[i].CodeHash = hash
	}

	// Store backup codes
	if err := s.backupCodeRepo.CreateBatch(backupCodes); err != nil {
		return nil, err
	}

	return &models.TwoFactorSetupResponse{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: plainBackupCodes,
	}, nil
}

// VerifyTOTPSetup verifies the initial TOTP code and enables 2FA
func (s *TwoFactorService) VerifyTOTPSetup(ctx context.Context, userID uuid.UUID, code string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		return models.NewAppError(400, "2FA setup not initiated")
	}

	// Verify code
	valid := totp.Validate(code, *user.TOTPSecret)
	if !valid {
		return models.NewAppError(401, "Invalid verification code")
	}

	// Enable 2FA
	if err := s.userRepo.EnableTOTP(userID); err != nil {
		return err
	}

	return nil
}

// VerifyTOTP verifies a TOTP code for login
func (s *TwoFactorService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}

	if !user.TOTPEnabled || user.TOTPSecret == nil {
		return false, models.NewAppError(400, "2FA not enabled")
	}

	// Try TOTP code first
	valid := totp.Validate(code, *user.TOTPSecret)
	if valid {
		return true, nil
	}

	// Try backup code if TOTP failed
	backupCodeValid, err := s.verifyBackupCode(userID, code)
	if err != nil {
		return false, err
	}

	return backupCodeValid, nil
}

// verifyBackupCode verifies and marks a backup code as used
func (s *TwoFactorService) verifyBackupCode(userID uuid.UUID, code string) (bool, error) {
	// Get all unused backup codes
	codes, err := s.backupCodeRepo.GetUnusedByUserID(userID)
	if err != nil {
		return false, err
	}

	// Check each code
	for _, backupCode := range codes {
		if err := utils.CheckPassword(backupCode.CodeHash, code); err == nil {
			// Mark as used
			if err := s.backupCodeRepo.MarkAsUsed(backupCode.ID); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

// DisableTOTP disables 2FA for a user
func (s *TwoFactorService) DisableTOTP(ctx context.Context, userID uuid.UUID, password, code string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify password
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return models.NewAppError(401, "Invalid password")
	}

	// Verify 2FA code
	valid, err := s.VerifyTOTP(ctx, userID, code)
	if err != nil {
		return err
	}

	if !valid {
		return models.NewAppError(401, "Invalid 2FA code")
	}

	// Disable 2FA
	if err := s.userRepo.DisableTOTP(userID); err != nil {
		return err
	}

	// Delete all backup codes
	if err := s.backupCodeRepo.DeleteAllByUserID(userID); err != nil {
		return err
	}

	return nil
}

// GetStatus returns the user's 2FA status
func (s *TwoFactorService) GetStatus(ctx context.Context, userID uuid.UUID) (*models.TwoFactorStatusResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Count unused backup codes
	backupCodeCount, err := s.backupCodeRepo.CountUnusedByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &models.TwoFactorStatusResponse{
		Enabled:     user.TOTPEnabled,
		EnabledAt:   user.TOTPEnabledAt,
		BackupCodes: backupCodeCount,
	}, nil
}

// RegenerateBackupCodes generates new backup codes for a user
func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, password string) ([]string, error) {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Verify password
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, models.NewAppError(401, "Invalid password")
	}

	// Verify 2FA is enabled
	if !user.TOTPEnabled {
		return nil, models.NewAppError(400, "2FA not enabled")
	}

	// Delete old backup codes
	if err := s.backupCodeRepo.DeleteAllByUserID(userID); err != nil {
		return nil, err
	}

	// Generate new backup codes
	backupCodes, err := s.generateBackupCodes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Get plain text backup codes
	plainBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		plainBackupCodes[i] = code.CodeHash
	}

	// Hash backup codes
	for i := range backupCodes {
		hash, err := utils.HashPassword(backupCodes[i].CodeHash, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		backupCodes[i].CodeHash = hash
	}

	// Store backup codes
	if err := s.backupCodeRepo.CreateBatch(backupCodes); err != nil {
		return nil, err
	}

	return plainBackupCodes, nil
}

// generateBackupCodes generates random backup codes
func (s *TwoFactorService) generateBackupCodes(userID uuid.UUID) ([]*models.BackupCode, error) {
	codes := make([]*models.BackupCode, BackupCodeCount)

	for i := 0; i < BackupCodeCount; i++ {
		// Generate random bytes
		b := make([]byte, BackupCodeLength)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Encode as alphanumeric
		code := base64.RawURLEncoding.EncodeToString(b)[:BackupCodeLength]

		codes[i] = &models.BackupCode{
			ID:        uuid.New(),
			UserID:    userID,
			CodeHash:  code, // Will be hashed before storage
			Used:      false,
			CreatedAt: time.Now(),
		}
	}

	return codes, nil
}
