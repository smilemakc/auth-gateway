package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/uptrace/bun"
)

// AuthService provides authentication operations
type AuthService struct {
	userRepo           UserStore
	tokenRepo          TokenStore
	rbacRepo           RBACStore
	auditService       AuditLogger
	jwtService         TokenService
	blacklistService   *BlacklistService
	redis              CacheService
	sessionService     *SessionService
	twoFAService       *TwoFactorService
	loginAlertService  *LoginAlertService
	webhookService     *WebhookService
	bcryptCost         int
	passwordPolicy     utils.PasswordPolicy
	db                 TransactionDB
	appRepo            ApplicationStore
	strictTokenBinding bool
}

// TransactionDB defines the interface for database transactions
type TransactionDB interface {
	RunInTx(ctx context.Context, fn func(context.Context, bun.Tx) error) error
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo UserStore,
	tokenRepo TokenStore,
	rbacRepo RBACStore,
	auditService AuditLogger,
	jwtService TokenService,
	blacklistService *BlacklistService,
	redis CacheService,
	sessionService *SessionService,
	twoFAService *TwoFactorService,
	bcryptCost int,
	passwordPolicy utils.PasswordPolicy,
	db TransactionDB,
	appRepo ApplicationStore,
	loginAlertService *LoginAlertService,
	webhookService *WebhookService,
	strictTokenBinding bool,
) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		rbacRepo:           rbacRepo,
		auditService:       auditService,
		jwtService:         jwtService,
		blacklistService:   blacklistService,
		redis:              redis,
		sessionService:     sessionService,
		twoFAService:       twoFAService,
		loginAlertService:  loginAlertService,
		webhookService:     webhookService,
		bcryptCost:         bcryptCost,
		passwordPolicy:     passwordPolicy,
		db:                 db,
		appRepo:            appRepo,
		strictTokenBinding: strictTokenBinding,
	}
}

// SignUp creates a new user account
func (s *AuthService) SignUp(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	// Require either email or phone
	if req.Email == "" && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	// Normalize inputs
	var email string
	var phone string

	if req.Email != "" {
		email = utils.NormalizeEmail(req.Email)
		// Validate email
		if !utils.IsValidEmail(email) {
			return nil, models.NewAppError(400, "Invalid email format")
		}
	}

	if req.Phone != nil && *req.Phone != "" {
		phone = utils.NormalizePhone(*req.Phone)
		// Validate phone
		if !utils.IsValidPhone(phone) {
			return nil, models.NewAppError(400, "Invalid phone format")
		}
	}

	username := utils.NormalizeUsername(req.Username)

	if username == "" {
		username = utils.Default(email, strings.ReplaceAll(phone, "+", ""))
	}

	if !utils.IsValidUsername(username) {
		return nil, models.NewAppError(400, "Invalid username format")
	}

	if err := utils.ValidatePassword(req.Password, s.passwordPolicy); err != nil {
		return nil, models.NewAppError(400, err.Error())
	}

	// Check auth method is allowed for this application
	if err := s.checkAuthMethodAllowed(ctx, appID, "password"); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Get default "user" role
	defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	// Create user and assign role in a transaction
	// Rely on database unique constraints instead of pre-checking
	var user *models.User
	err = s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		// Create user
		user = &models.User{
			ID:           uuid.New(),
			Email:        email,
			Phone:        utils.Ptr(phone),
			Username:     username,
			PasswordHash: passwordHash,
			FullName:     req.FullName,
			IsActive:     true,
		}

		// Use transaction-aware repository methods
		if userRepo, ok := s.userRepo.(*repository.UserRepository); ok {
			if err := userRepo.CreateWithTx(ctx, tx, user); err != nil {
				// Database will return unique_violation error if email/username/phone already exists
				// handlePgError will convert it to appropriate error
				s.logAudit(nil, appID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "create_failed",
					"error":  err.Error(),
				})
				return err
			}
		} else {
			// Fallback to non-transactional method if type assertion fails
			if err := s.userRepo.Create(ctx, user); err != nil {
				s.logAudit(nil, appID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "create_failed",
					"error":  err.Error(),
				})
				return err
			}
		}

		// Assign default "user" role to the new user
		if rbacRepo, ok := s.rbacRepo.(*repository.RBACRepository); ok {
			if err := rbacRepo.AssignRoleToUserWithTx(ctx, tx, user.ID, defaultRole.ID, user.ID); err != nil {
				s.logAudit(&user.ID, appID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "role_assignment_failed",
					"error":  err.Error(),
				})
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		} else {
			// Fallback to non-transactional method if type assertion fails
			if err := s.rbacRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, user.ID); err != nil {
				s.logAudit(&user.ID, appID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "role_assignment_failed",
					"error":  err.Error(),
				})
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		// Check if it's a unique constraint violation (already handled by handlePgError)
		return nil, err
	}

	// Reload user with roles for token generation
	user, err = s.userRepo.GetByID(ctx, user.ID, utils.Ptr(true), UserGetWithRoles())
	if err != nil {
		return nil, fmt.Errorf("failed to reload user with roles: %w", err)
	}

	// Generate tokens with device info (isNewUser=true suppresses login alert)
	authResp, err := s.finalizeAuth(ctx, user, ip, userAgent, deviceInfo, appID, true, "password")
	if err != nil {
		return nil, err
	}

	// Log successful signup
	s.logAudit(&user.ID, appID, models.ActionSignUp, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// SignIn authenticates a user and returns tokens
// Implements timing attack protection by always performing password check
func (s *AuthService) SignIn(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
	// Require either email or phone
	if req.Email == "" && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	// Check auth method is allowed for this application
	if err := s.checkAuthMethodAllowed(ctx, appID, "password"); err != nil {
		return nil, err
	}

	var user *models.User
	var err error
	var passwordHash string

	// Get user by email or phone
	if req.Email != "" {
		email := utils.NormalizeEmail(req.Email)
		user, err = s.userRepo.GetByEmail(ctx, email, nil, UserGetWithRoles())
		if err != nil {
			// User not found - use dummy hash to prevent timing attacks
			passwordHash = utils.GetDummyPasswordHash()
		} else {
			passwordHash = user.PasswordHash
		}
	} else if req.Phone != nil && *req.Phone != "" {
		phone := utils.NormalizePhone(*req.Phone)
		user, err = s.userRepo.GetByPhone(ctx, phone, nil, UserGetWithRoles())
		if err != nil {
			passwordHash = utils.GetDummyPasswordHash()
		} else {
			passwordHash = user.PasswordHash
		}
	}

	// Always perform password check to prevent timing attacks
	// This ensures consistent response time regardless of whether user exists
	if err := utils.CheckPassword(passwordHash, req.Password); err != nil {
		// Log failed attempt (but don't reveal if user exists)
		var userID *uuid.UUID
		if user != nil {
			userID = &user.ID
		}
		s.logAudit(userID, appID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_credentials",
		})
		return nil, models.ErrInvalidCredentials
	}

	// If we reach here but user is nil, it means password matched dummy hash
	// This should be extremely rare, but handle it for safety
	if user == nil {
		s.logAudit(nil, appID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_credentials",
		})
		return nil, models.ErrInvalidCredentials
	}

	// Check if 2FA is enabled
	if user.TOTPEnabled {
		// Generate temporary 2FA token
		twoFactorToken, err := s.jwtService.GenerateTwoFactorToken(user, appID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate 2FA token: %w", err)
		}

		return &models.AuthResponse{
			Requires2FA:    true,
			TwoFactorToken: twoFactorToken,
			User:           user.PublicUser(),
		}, nil
	}

	// Generate tokens with device info
	authResp, err := s.finalizeAuth(ctx, user, ip, userAgent, deviceInfo, appID, false, "password")
	if err != nil {
		return nil, err
	}

	// Log successful signin
	s.logAudit(&user.ID, appID, models.ActionSignIn, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// Verify2FALogin verifies 2FA code and completes login
func (s *AuthService) Verify2FALogin(ctx context.Context, twoFactorToken, code, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	// Validate 2FA token
	claims, err := s.jwtService.ValidateAccessToken(twoFactorToken)
	if err != nil {
		return nil, models.NewAppError(401, "Invalid or expired 2FA token")
	}

	// Get user with roles
	user, err := s.userRepo.GetByID(ctx, claims.UserID, utils.Ptr(true), UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	// Verify 2FA is enabled
	if !user.TOTPEnabled || user.TOTPSecret == nil {
		return nil, models.NewAppError(400, "2FA not enabled")
	}

	// Verify TOTP code
	if !totp.Validate(code, *user.TOTPSecret) {
		// Try backup code using TwoFactorService
		if s.twoFAService != nil {
			valid, err := s.twoFAService.VerifyTOTP(ctx, user.ID, code)
			if err != nil || !valid {
				s.logAudit(&user.ID, claims.ApplicationID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "invalid_2fa_code",
				})
				return nil, models.NewAppError(401, "Invalid 2FA code")
			}
		} else {
			s.logAudit(&user.ID, claims.ApplicationID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "invalid_2fa_code",
			})
			return nil, models.NewAppError(401, "Invalid 2FA code")
		}
	}

	// Generate full auth tokens with device info
	authResp, err := s.finalizeAuth(ctx, user, ip, userAgent, deviceInfo, nil, false, "totp")
	if err != nil {
		return nil, err
	}

	// Log successful signin with 2FA
	s.logAudit(&user.ID, claims.ApplicationID, models.ActionSignIn, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"2fa": true,
	})

	return authResp, nil
}

// verifyBackupCode is deprecated - use TwoFactorService.VerifyCode instead
// This method is kept for backward compatibility but should not be used
func (s *AuthService) verifyBackupCode(userID uuid.UUID, code string) (bool, error) {
	// Use TwoFactorService if available
	if s.twoFAService != nil {
		// Note: This requires TOTPSecret, so we can't use it directly here
		// The method is deprecated and should not be called
		return false, fmt.Errorf("verifyBackupCode is deprecated, use TwoFactorService.VerifyCode")
	}
	return false, nil
}

// RefreshToken generates new tokens using a refresh token
// This operation is atomic - old token is revoked and new token is created in a single transaction
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logAudit(nil, nil, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_token",
		})
		return nil, models.ErrInvalidToken
	}

	// Check if token is blacklisted using unified blacklist service
	oldTokenHash := utils.HashToken(refreshToken)
	if s.blacklistService.IsBlacklisted(ctx, oldTokenHash) {
		s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "token_blacklisted",
		})
		return nil, models.ErrTokenRevoked
	}

	// Get user with roles (before transaction to avoid deadlocks)
	user, err := s.userRepo.GetByID(ctx, claims.UserID, utils.Ptr(true), UserGetWithRoles())
	if err != nil {
		return nil, err
	}

	// Perform atomic token refresh in a transaction
	var dbToken *models.RefreshToken
	var newAccessToken, newRefreshToken string
	var newDBToken *models.RefreshToken
	var newTokenHash string
	var refreshExpiration time.Duration

	err = s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		// Get token with row lock (SELECT FOR UPDATE)
		tokenRepo, ok := s.tokenRepo.(*repository.TokenRepository)
		if !ok {
			return fmt.Errorf("token repository does not support transactions")
		}

		// Check if token exists and is not revoked (with lock)
		dbToken, err = tokenRepo.GetRefreshTokenForUpdate(ctx, tx, oldTokenHash)
		if err != nil {
			s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "token_not_found",
			})
			return models.ErrInvalidToken
		}

		if dbToken.IsRevoked() {
			s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "token_revoked",
			})
			return models.ErrTokenRevoked
		}

		if dbToken.IsExpired() {
			s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "token_expired",
			})
			return models.ErrTokenExpired
		}

		// Device binding: reject if IP changed (when strict mode enabled)
		if s.strictTokenBinding && dbToken.IPAddress != "" && dbToken.IPAddress != ip {
			s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionRefreshToken, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason":      "device_mismatch",
				"original_ip": dbToken.IPAddress,
				"current_ip":  ip,
			})
			return models.ErrTokenCompromised
		}

		// Revoke old refresh token
		if err := tokenRepo.RevokeRefreshTokenWithTx(ctx, tx, oldTokenHash); err != nil {
			return fmt.Errorf("failed to revoke old token: %w", err)
		}

		// Generate new tokens with app context from original token
		newAccessToken, err = s.jwtService.GenerateAccessToken(user, claims.ApplicationID)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}

		newRefreshToken, err = s.jwtService.GenerateRefreshToken(user, claims.ApplicationID)
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}

		// Save new refresh token to database
		newTokenHash = utils.HashToken(newRefreshToken)
		refreshExpiration = s.jwtService.GetRefreshTokenExpiration()
		newDBToken = &models.RefreshToken{
			ID:          uuid.New(),
			UserID:      user.ID,
			TokenHash:   newTokenHash,
			ExpiresAt:   time.Now().Add(refreshExpiration),
			DeviceType:  deviceInfo.DeviceType,
			OS:          deviceInfo.OS,
			Browser:     deviceInfo.Browser,
			SessionName: dbToken.SessionName, // Keep original session name
			IPAddress:   ip,
			UserAgent:   userAgent,
		}

		if err := tokenRepo.CreateRefreshTokenWithTx(ctx, tx, newDBToken); err != nil {
			return fmt.Errorf("failed to create new refresh token: %w", err)
		}

		return nil
	})

	if err != nil {
		// Check if it's already a models error
		if appErr, ok := err.(*models.AppError); ok {
			return nil, appErr
		}
		return nil, err
	}

	// Token was already created in transaction, no need to create again

	// Update existing session with new token hashes instead of creating a new one
	if s.sessionService != nil {
		s.sessionService.RefreshSessionNonFatal(ctx, SessionRefreshParams{
			OldRefreshTokenHash: oldTokenHash,
			NewRefreshTokenHash: newTokenHash,
			NewAccessTokenHash:  utils.HashToken(newAccessToken),
			NewExpiresAt:        time.Now().Add(refreshExpiration),
		})
	}

	// Log successful refresh
	s.logAudit(&user.ID, claims.ApplicationID, models.ActionRefreshToken, models.StatusSuccess, ip, userAgent, nil)

	return &models.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user.PublicUser(),
		ExpiresIn:    int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
	}, nil
}

// Logout logs out a user by revoking their tokens
func (s *AuthService) Logout(ctx context.Context, accessToken, ip, userAgent string) error {
	// Extract claims without full validation (we just need the user ID)
	claims, err := s.jwtService.ExtractClaims(accessToken)
	if err != nil {
		return models.ErrInvalidToken
	}

	// Add access token to blacklist using unified service
	tokenHash := utils.HashToken(accessToken)
	if err := s.blacklistService.AddAccessToken(ctx, tokenHash, &claims.UserID); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Revoke all refresh tokens for this user
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, claims.UserID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful logout
	s.logAudit(&claims.UserID, claims.ApplicationID, models.ActionSignOut, models.StatusSuccess, ip, userAgent, nil)

	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID, utils.Ptr(true))
	if err != nil {
		return err
	}

	// Verify old password
	if err := utils.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		s.logAudit(&userID, nil, models.ActionChangePassword, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "invalid_old_password",
		})
		return models.ErrInvalidCredentials
	}

	// Validate new password
	if err := utils.ValidatePassword(newPassword, s.passwordPolicy); err != nil {
		return models.NewAppError(400, err.Error())
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Blacklist all active session tokens so access tokens become invalid immediately
	if err := s.blacklistService.BlacklistAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to blacklist session tokens: %w", err)
	}

	// Log successful password change
	s.logAudit(&userID, nil, models.ActionChangePassword, models.StatusSuccess, ip, userAgent, nil)

	return nil
}

// ResetPassword resets a user's password (used for password reset flow)
func (s *AuthService) ResetPassword(ctx context.Context, userID uuid.UUID, newPassword, ip, userAgent string) error {
	// Validate new password
	if err := utils.ValidatePassword(newPassword, s.passwordPolicy); err != nil {
		return models.NewAppError(400, err.Error())
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Blacklist all active session tokens so access tokens become invalid immediately
	if err := s.blacklistService.BlacklistAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to blacklist session tokens: %w", err)
	}

	// Log successful password reset
	s.logAudit(&userID, nil, models.ActionChangePassword, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"reset": true,
	})

	return nil
}

// PendingRegistrationExpiration is the TTL for pending registration data in Redis
const PendingRegistrationExpiration = 10 * time.Minute

// InitPasswordlessRegistration initiates a passwordless registration by storing pending data and sending OTP
func (s *AuthService) InitPasswordlessRegistration(ctx context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error {
	// Require either email or phone
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return models.NewAppError(400, "Either email or phone is required")
	}

	var email string
	var phone string
	var identifier string

	// Normalize and validate email
	if req.Email != nil && *req.Email != "" {
		email = utils.NormalizeEmail(*req.Email)
		if !utils.IsValidEmail(email) {
			return models.NewAppError(400, "Invalid email format")
		}
		identifier = email
	}

	// Normalize and validate phone
	if req.Phone != nil && *req.Phone != "" {
		phone = utils.NormalizePhone(*req.Phone)
		if !utils.IsValidPhone(phone) {
			return models.NewAppError(400, "Invalid phone format")
		}
		if identifier == "" {
			identifier = phone
		}
	}

	// Generate username if not provided
	username := utils.NormalizeUsername(req.Username)
	if username == "" {
		if email != "" {
			// Extract local part from email
			parts := strings.Split(email, "@")
			username = parts[0]
		} else {
			// Use phone number without +
			username = strings.ReplaceAll(phone, "+", "")
		}
	}

	// Validate username
	if !utils.IsValidUsername(username) {
		return models.NewAppError(400, "Invalid username format")
	}

	// Note: We don't pre-check if username exists. If it conflicts during creation,
	// the database unique constraint will catch it and we'll handle the error.
	// For username conflicts, we'll append a random suffix during CompletePasswordlessRegistration if needed.

	// Store pending registration in Redis
	pending := &models.PendingRegistration{
		Email:     email,
		Phone:     phone,
		Username:  username,
		FullName:  req.FullName,
		CreatedAt: time.Now().Unix(),
	}

	if err := s.redis.StorePendingRegistration(ctx, identifier, pending, PendingRegistrationExpiration); err != nil {
		return fmt.Errorf("failed to store pending registration: %w", err)
	}

	// Log init registration
	s.logAudit(nil, nil, models.ActionSignUp, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"step":       "init",
		"identifier": identifier,
	})

	return nil
}

// CompletePasswordlessRegistration completes the registration after OTP verification
func (s *AuthService) CompletePasswordlessRegistration(ctx context.Context, req *models.CompletePasswordlessRegistrationRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	// Require either email or phone
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	var identifier string
	var email string
	var phone string

	if req.Email != nil && *req.Email != "" {
		email = utils.NormalizeEmail(*req.Email)
		identifier = email
	}

	if req.Phone != nil && *req.Phone != "" {
		phone = utils.NormalizePhone(*req.Phone)
		if identifier == "" {
			identifier = phone
		}
	}

	// Retrieve pending registration from Redis
	pending, err := s.redis.GetPendingRegistration(ctx, identifier)
	if err != nil {
		s.logAudit(nil, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason":     "pending_not_found",
			"identifier": identifier,
		})
		return nil, models.NewAppError(400, "Registration not initiated or expired. Please start again.")
	}

	// Verify pending data matches
	if (email != "" && pending.Email != email) || (phone != "" && pending.Phone != phone) {
		return nil, models.NewAppError(400, "Registration data mismatch")
	}

	// Get default "user" role
	defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	// Create user and assign role in a transaction
	var user *models.User
	err = s.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		// Create user without password (passwordless registration)
		user = &models.User{
			ID:            uuid.New(),
			Email:         pending.Email,
			Phone:         utils.Ptr(pending.Phone),
			Username:      pending.Username,
			PasswordHash:  "", // No password for passwordless registration
			FullName:      pending.FullName,
			IsActive:      true,
			EmailVerified: email != "", // Mark email as verified if registering via email
			PhoneVerified: phone != "", // Mark phone as verified if registering via phone
		}

		// Use transaction-aware repository methods
		if userRepo, ok := s.userRepo.(*repository.UserRepository); ok {
			if err := userRepo.CreateWithTx(ctx, tx, user); err != nil {
				// If username conflict, try with a random suffix
				if err == models.ErrUsernameAlreadyExists {
					user.Username = fmt.Sprintf("%s_%d", pending.Username, time.Now().UnixNano()%10000)
					if err := userRepo.CreateWithTx(ctx, tx, user); err != nil {
						s.logAudit(nil, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
							"reason": "create_failed",
							"error":  err.Error(),
						})
						return err
					}
				} else {
					s.logAudit(nil, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
						"reason": "create_failed",
						"error":  err.Error(),
					})
					return err
				}
			}
		} else {
			// Fallback to non-transactional method if type assertion fails
			if err := s.userRepo.Create(ctx, user); err != nil {
				// If username conflict, try with a random suffix
				if err == models.ErrUsernameAlreadyExists {
					user.Username = fmt.Sprintf("%s_%d", pending.Username, time.Now().UnixNano()%10000)
					if err := s.userRepo.Create(ctx, user); err != nil {
						s.logAudit(nil, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
							"reason": "create_failed",
							"error":  err.Error(),
						})
						return err
					}
				} else {
					s.logAudit(nil, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
						"reason": "create_failed",
						"error":  err.Error(),
					})
					return err
				}
			}
		}

		// Assign default "user" role
		if rbacRepo, ok := s.rbacRepo.(*repository.RBACRepository); ok {
			if err := rbacRepo.AssignRoleToUserWithTx(ctx, tx, user.ID, defaultRole.ID, user.ID); err != nil {
				s.logAudit(&user.ID, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "role_assignment_failed",
					"error":  err.Error(),
				})
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		} else {
			// Fallback to non-transactional method if type assertion fails
			if err := s.rbacRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, user.ID); err != nil {
				s.logAudit(&user.ID, nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
					"reason": "role_assignment_failed",
					"error":  err.Error(),
				})
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload user with roles for token generation
	user, err = s.userRepo.GetByID(ctx, user.ID, utils.Ptr(true), UserGetWithRoles())
	if err != nil {
		return nil, fmt.Errorf("failed to reload user with roles: %w", err)
	}

	// Delete pending registration from Redis
	if err := s.redis.DeletePendingRegistration(ctx, identifier); err != nil {
		// Log but don't fail - user is already created
		fmt.Printf("Failed to delete pending registration: %v\n", err)
	}

	// Generate tokens with device info (isNewUser=true — passwordless signup)
	authResp, err := s.finalizeAuth(ctx, user, ip, userAgent, deviceInfo, nil, true, "otp_email")
	if err != nil {
		return nil, err
	}

	// Log successful signup
	s.logAudit(&user.ID, nil, models.ActionSignUp, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"passwordless": true,
	})

	return authResp, nil
}

// finalizeAuth generates access and refresh tokens, creates app profile, triggers webhook, and saves refresh token with device info
func (s *AuthService) finalizeAuth(ctx context.Context, user *models.User, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID, isNewUser bool, authMethod string) (*models.AuthResponse, error) {
	// Auto-create/update app profile on login
	if appID != nil && s.appRepo != nil {
		profile, _ := s.appRepo.GetUserProfile(ctx, user.ID, *appID)
		if profile == nil {
			// First login to this app - create profile
			newProfile := &models.UserApplicationProfile{
				UserID:        user.ID,
				ApplicationID: *appID,
				IsActive:      true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := s.appRepo.CreateUserProfile(ctx, newProfile); err != nil {
				// Non-fatal, don't block auth
			}
		} else {
			// Update last access
			s.appRepo.UpdateLastAccess(ctx, user.ID, *appID)
		}
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Generate session name from device info
	sessionName := utils.GenerateSessionName(deviceInfo)

	// Save refresh token to database with device info
	tokenHash := utils.HashToken(refreshToken)
	refreshExpiration := s.jwtService.GetRefreshTokenExpiration()
	dbToken := &models.RefreshToken{
		ID:          uuid.New(),
		UserID:      user.ID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(refreshExpiration),
		DeviceType:  deviceInfo.DeviceType,
		OS:          deviceInfo.OS,
		Browser:     deviceInfo.Browser,
		SessionName: sessionName,
		IPAddress:   ip,
		UserAgent:   userAgent,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, dbToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Create session using SessionService (non-fatal to not block auth)
	if s.sessionService != nil {
		s.sessionService.CreateSessionNonFatal(ctx, SessionCreationParams{
			UserID:          user.ID,
			ApplicationID:   appID,
			TokenHash:       tokenHash,
			AccessTokenHash: utils.HashToken(accessToken),
			IPAddress:       ip,
			UserAgent:       userAgent,
			ExpiresAt:       time.Now().Add(refreshExpiration),
			SessionName:     sessionName,
		})
	}

	// Check for new device and send login alert email (async, non-blocking)
	if s.loginAlertService != nil {
		go func() {
			alertCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			s.loginAlertService.CheckAndAlert(alertCtx, LoginAlertParams{
				UserID:    user.ID,
				Username:  user.Username,
				Email:     user.Email,
				IP:        ip,
				UserAgent: userAgent,
				Device:    deviceInfo,
				AppID:     appID,
				IsNewUser: isNewUser,
			})
		}()
	}

	// Trigger webhook for user.login event (async, non-blocking)
	if s.webhookService != nil {
		go func() {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			s.webhookService.TriggerWebhook(webhookCtx, "user.login", map[string]interface{}{
				"user_id":        user.ID.String(),
				"email":          user.Email,
				"auth_method":    authMethod,
				"application_id": uuidPtrToString(appID),
				"timestamp":      time.Now().UTC().Format(time.RFC3339),
			})
		}()
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.PublicUser(),
		ExpiresIn:    int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
	}, nil
}

// GenerateTokensForUser generates tokens for a given user (used for OTP/SMS passwordless login).
// This method properly saves the refresh token and creates a session.
func (s *AuthService) GenerateTokensForUser(ctx context.Context, user *models.User, ip, userAgent string) (*models.AuthResponse, error) {
	// Parse device info from user agent
	deviceInfo := utils.ParseUserAgent(userAgent)

	// Use the centralized finalizeAuth method
	return s.finalizeAuth(ctx, user, ip, userAgent, deviceInfo, nil, false, "password")
}

func (s *AuthService) logAudit(userID *uuid.UUID, appID *uuid.UUID, action models.AuditAction, status models.AuditStatus, ip, userAgent string, details map[string]interface{}) {
	s.auditService.Log(AuditLogParams{
		UserID:        userID,
		ApplicationID: appID,
		Action:        action,
		Status:        status,
		IP:            ip,
		UserAgent:     userAgent,
		Details:       details,
	})
}

// checkAuthMethodAllowed checks if the auth method is allowed for the app
func (s *AuthService) checkAuthMethodAllowed(ctx context.Context, appID *uuid.UUID, method string) error {
	if appID == nil || s.appRepo == nil {
		return nil
	}
	app, err := s.appRepo.GetApplicationByID(ctx, *appID)
	if err != nil {
		return nil
	}
	if !app.IsActive {
		return models.NewAppError(403, "Application is not active")
	}
	if len(app.AllowedAuthMethods) == 0 {
		return nil
	}
	for _, m := range app.AllowedAuthMethods {
		if m == method {
			return nil
		}
	}
	return models.NewAppError(403, fmt.Sprintf("Auth method '%s' is not allowed for this application", method))
}

// uuidPtrToString converts a UUID pointer to string, returning empty string if nil
func uuidPtrToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
