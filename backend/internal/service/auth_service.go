package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
)

// AuthService provides authentication operations
type AuthService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	auditRepo  *repository.AuditRepository
	jwtService *jwt.Service
	redis      *RedisService
	bcryptCost int
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
	jwtService *jwt.Service,
	redis *RedisService,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		auditRepo:  auditRepo,
		jwtService: jwtService,
		redis:      redis,
		bcryptCost: bcryptCost,
	}
}

// SignUp creates a new user account
func (s *AuthService) SignUp(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string) (*models.AuthResponse, error) {
	// Require either email or phone
	if req.Email == "" && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	// Normalize inputs
	var email string
	var normalizedPhone *string

	if req.Email != "" {
		email = utils.NormalizeEmail(req.Email)
		// Validate email
		if !utils.IsValidEmail(email) {
			return nil, models.NewAppError(400, "Invalid email format")
		}
		// Check if email exists
		if exists, err := s.userRepo.EmailExists(email); err != nil {
			return nil, err
		} else if exists {
			s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "email_exists",
				"email":  email,
			})
			return nil, models.ErrEmailAlreadyExists
		}
	}

	if req.Phone != nil && *req.Phone != "" {
		phone := utils.NormalizePhone(*req.Phone)
		normalizedPhone = &phone
		// Validate phone
		if !utils.IsValidPhone(phone) {
			return nil, models.NewAppError(400, "Invalid phone format")
		}
		// Check if phone exists
		if exists, err := s.userRepo.PhoneExists(phone); err != nil {
			return nil, err
		} else if exists {
			s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "phone_exists",
				"phone":  phone,
			})
			return nil, models.ErrPhoneAlreadyExists
		}
	}

	username := utils.NormalizeUsername(req.Username)

	if !utils.IsValidUsername(username) {
		return nil, models.NewAppError(400, "Invalid username format")
	}

	if !utils.IsPasswordValid(req.Password) {
		return nil, models.NewAppError(400, "Password must be at least 8 characters")
	}

	// Check if username already exists
	if exists, err := s.userRepo.UsernameExists(username); err != nil {
		return nil, err
	} else if exists {
		s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason":   "username_exists",
			"username": username,
		})
		return nil, models.ErrUsernameAlreadyExists
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        normalizedPhone,
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		Role:         string(models.RoleUser),
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "create_failed",
			"error":  err.Error(),
		})
		return nil, err
	}

	// Generate tokens
	authResp, err := s.generateAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Log successful signup
	s.logAudit(&user.ID, models.ActionSignUp, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, req *models.SignInRequest, ip, userAgent string) (*models.AuthResponse, error) {
	// Require either email or phone
	if req.Email == "" && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	var user *models.User
	var err error

	// Get user by email or phone
	if req.Email != "" {
		email := utils.NormalizeEmail(req.Email)
		user, err = s.userRepo.GetByEmail(email)
		if err != nil {
			s.logAudit(nil, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "user_not_found",
				"email":  email,
			})
			return nil, models.ErrInvalidCredentials
		}
	} else if req.Phone != nil && *req.Phone != "" {
		phone := utils.NormalizePhone(*req.Phone)
		user, err = s.userRepo.GetByPhone(phone)
		if err != nil {
			s.logAudit(nil, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "user_not_found",
				"phone":  phone,
			})
			return nil, models.ErrInvalidCredentials
		}
	}

	// Check password
	if err := utils.CheckPassword(user.PasswordHash, req.Password); err != nil {
		s.logAudit(&user.ID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_password",
		})
		return nil, models.ErrInvalidCredentials
	}

	// Check if 2FA is enabled
	if user.TOTPEnabled {
		// Generate temporary 2FA token
		twoFactorToken, err := s.jwtService.GenerateTwoFactorToken(user)
		if err != nil {
			return nil, fmt.Errorf("failed to generate 2FA token: %w", err)
		}

		return &models.AuthResponse{
			Requires2FA:    true,
			TwoFactorToken: twoFactorToken,
			User:           user.PublicUser(),
		}, nil
	}

	// Generate tokens
	authResp, err := s.generateAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Log successful signin
	s.logAudit(&user.ID, models.ActionSignIn, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// Verify2FALogin verifies 2FA code and completes login
func (s *AuthService) Verify2FALogin(ctx context.Context, twoFactorToken, code, ip, userAgent string) (*models.AuthResponse, error) {
	// Validate 2FA token
	claims, err := s.jwtService.ValidateAccessToken(twoFactorToken)
	if err != nil {
		return nil, models.NewAppError(401, "Invalid or expired 2FA token")
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Verify 2FA is enabled
	if !user.TOTPEnabled || user.TOTPSecret == nil {
		return nil, models.NewAppError(400, "2FA not enabled")
	}

	// Verify TOTP code
	if !totp.Validate(code, *user.TOTPSecret) {
		// Try backup code
		backupCodeValid, err := s.verifyBackupCode(user.ID, code)
		if err != nil || !backupCodeValid {
			s.logAudit(&user.ID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "invalid_2fa_code",
			})
			return nil, models.NewAppError(401, "Invalid 2FA code")
		}
	}

	// Generate full auth tokens
	authResp, err := s.generateAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Log successful signin with 2FA
	s.logAudit(&user.ID, models.ActionSignIn, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"2fa": true,
	})

	return authResp, nil
}

// verifyBackupCode verifies a backup code (simplified version, full implementation in TwoFactorService)
func (s *AuthService) verifyBackupCode(userID uuid.UUID, code string) (bool, error) {
	// This is a simplified implementation
	// In production, you'd want to use the TwoFactorService
	return false, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logAudit(nil, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_token",
		})
		return nil, models.ErrInvalidToken
	}

	// Check if token is blacklisted in Redis
	tokenHash := utils.HashToken(refreshToken)
	if blacklisted, err := s.redis.IsBlacklisted(ctx, tokenHash); err == nil && blacklisted {
		s.logAudit(&claims.UserID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "token_blacklisted",
		})
		return nil, models.ErrTokenRevoked
	}

	// Check if token exists and is not revoked in database
	dbToken, err := s.tokenRepo.GetRefreshToken(tokenHash)
	if err != nil {
		s.logAudit(&claims.UserID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "token_not_found",
		})
		return nil, models.ErrInvalidToken
	}

	if dbToken.IsRevoked() {
		s.logAudit(&claims.UserID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "token_revoked",
		})
		return nil, models.ErrTokenRevoked
	}

	if dbToken.IsExpired() {
		s.logAudit(&claims.UserID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "token_expired",
		})
		return nil, models.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	if err := s.tokenRepo.RevokeRefreshToken(tokenHash); err != nil {
		return nil, err
	}

	// Generate new tokens
	authResp, err := s.generateAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Log successful refresh
	s.logAudit(&user.ID, models.ActionRefreshToken, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// Logout logs out a user by revoking their tokens
func (s *AuthService) Logout(ctx context.Context, accessToken, ip, userAgent string) error {
	// Extract claims without full validation (we just need the user ID)
	claims, err := s.jwtService.ExtractClaims(accessToken)
	if err != nil {
		return models.ErrInvalidToken
	}

	// Add access token to blacklist in Redis
	tokenHash := utils.HashToken(accessToken)
	expiration := s.jwtService.GetAccessTokenExpiration()
	if err := s.redis.AddToBlacklist(ctx, tokenHash, expiration); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Also add to database for persistence
	blacklistEntry := &models.TokenBlacklist{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserID:    &claims.UserID,
		ExpiresAt: time.Now().Add(expiration),
	}

	if err := s.tokenRepo.AddToBlacklist(blacklistEntry); err != nil {
		// Log but don't fail - Redis blacklist is primary
		fmt.Printf("Failed to add token to DB blacklist: %v\n", err)
	}

	// Revoke all refresh tokens for this user
	if err := s.tokenRepo.RevokeAllUserTokens(claims.UserID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful logout
	s.logAudit(&claims.UserID, models.ActionSignOut, models.StatusSuccess, ip, userAgent, nil)

	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := utils.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		s.logAudit(&userID, models.ActionChangePassword, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_old_password",
		})
		return models.ErrInvalidCredentials
	}

	// Validate new password
	if !utils.IsPasswordValid(newPassword) {
		return models.NewAppError(400, "New password must be at least 8 characters")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful password change
	s.logAudit(&userID, models.ActionChangePassword, models.StatusSuccess, ip, userAgent, nil)

	return nil
}

// ResetPassword resets a user's password (used for password reset flow)
func (s *AuthService) ResetPassword(ctx context.Context, userID uuid.UUID, newPassword, ip, userAgent string) error {
	// Validate new password
	if !utils.IsPasswordValid(newPassword) {
		return models.NewAppError(400, "New password must be at least 8 characters")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful password reset
	s.logAudit(&userID, models.ActionChangePassword, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"reset": true,
	})

	return nil
}

// generateAuthResponse generates access and refresh tokens and saves refresh token
func (s *AuthService) generateAuthResponse(ctx context.Context, user *models.User) (*models.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Save refresh token to database
	tokenHash := utils.HashToken(refreshToken)
	dbToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenExpiration()),
	}

	if err := s.tokenRepo.CreateRefreshToken(dbToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.PublicUser(),
		ExpiresIn:    int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
	}, nil
}

// logAudit logs an audit entry
func (s *AuthService) logAudit(userID *uuid.UUID, action models.AuditAction, status models.AuditStatus, ip, userAgent string, details map[string]interface{}) {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	auditLog := models.CreateAuditLog(userID, action, status, ip, userAgent, detailsJSON)

	// Log asynchronously to avoid blocking
	go func() {
		if err := s.auditRepo.Create(auditLog); err != nil {
			fmt.Printf("Failed to create audit log: %v\n", err)
		}
	}()
}
