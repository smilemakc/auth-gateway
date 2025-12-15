package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// AuthService provides authentication operations
type AuthService struct {
	userRepo       UserStore
	tokenRepo      TokenStore
	rbacRepo       RBACStore
	auditService   AuditLogger
	jwtService     TokenService
	redis          CacheService
	sessionService *SessionCreationService
	bcryptCost     int
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo UserStore,
	tokenRepo TokenStore,
	rbacRepo RBACStore,
	auditService AuditLogger,
	jwtService TokenService,
	redis CacheService,
	sessionService *SessionCreationService,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		rbacRepo:       rbacRepo,
		auditService:   auditService,
		jwtService:     jwtService,
		redis:          redis,
		sessionService: sessionService,
		bcryptCost:     bcryptCost,
	}
}

// SignUp creates a new user account
func (s *AuthService) SignUp(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
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
		// Check if email exists
		if exists, err := s.userRepo.EmailExists(ctx, email); err != nil {
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
		phone = utils.NormalizePhone(*req.Phone)
		// Validate phone
		if !utils.IsValidPhone(phone) {
			return nil, models.NewAppError(400, "Invalid phone format")
		}
		// Check if phone exists
		if exists, err := s.userRepo.PhoneExists(ctx, phone); err != nil {
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

	if username == "" {
		username = utils.Default(email, strings.ReplaceAll(phone, "+", ""))
	}

	if !utils.IsValidUsername(username) {
		return nil, models.NewAppError(400, "Invalid username format")
	}

	if !utils.IsPasswordValid(req.Password) {
		return nil, models.NewAppError(400, "Password must be at least 8 characters")
	}

	// Check if username already exists
	if exists, err := s.userRepo.UsernameExists(ctx, username); err != nil {
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

	// Get default "user" role
	defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        utils.Ptr(phone),
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "create_failed",
			"error":  err.Error(),
		})
		return nil, err
	}

	// Assign default "user" role to the new user
	if err := s.rbacRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, user.ID); err != nil {
		s.logAudit(&user.ID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "role_assignment_failed",
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	// Reload user with roles for token generation
	user, err = s.userRepo.GetByIDWithRoles(ctx, user.ID, utils.Ptr(true))
	if err != nil {
		return nil, fmt.Errorf("failed to reload user with roles: %w", err)
	}

	// Generate tokens with device info
	authResp, err := s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Log successful signup
	s.logAudit(&user.ID, models.ActionSignUp, models.StatusSuccess, ip, userAgent, nil)

	return authResp, nil
}

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
	// Require either email or phone
	if req.Email == "" && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	var user *models.User
	var err error

	// Get user by email or phone
	if req.Email != "" {
		email := utils.NormalizeEmail(req.Email)
		user, err = s.userRepo.GetByEmailWithRoles(ctx, email, nil)
		if err != nil {
			s.logAudit(nil, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": err,
				"email":  email,
			})
			return nil, models.ErrInvalidCredentials
		}
	} else if req.Phone != nil && *req.Phone != "" {
		phone := utils.NormalizePhone(*req.Phone)
		user, err = s.userRepo.GetByPhone(ctx, phone, nil)
		if err != nil {
			s.logAudit(nil, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": err,
				"phone":  phone,
			})
			return nil, models.ErrInvalidCredentials
		}
		// Load roles for phone-based login
		roles, roleErr := s.rbacRepo.GetUserRoles(ctx, user.ID)
		if roleErr == nil {
			user.Roles = roles
		}
	}

	// Check password
	if err := utils.CheckPassword(user.PasswordHash, req.Password); err != nil {
		s.logAudit(&user.ID, models.ActionSignInFailed, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": err.Error(),
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

	// Generate tokens with device info
	authResp, err := s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Log successful signin
	s.logAudit(&user.ID, models.ActionSignIn, models.StatusSuccess, ip, userAgent, nil)

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
	user, err := s.userRepo.GetByIDWithRoles(ctx, claims.UserID, utils.Ptr(true))
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

	// Generate full auth tokens with device info
	authResp, err := s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
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
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
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
	dbToken, err := s.tokenRepo.GetRefreshToken(ctx, tokenHash)
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

	// Get user with roles
	user, err := s.userRepo.GetByIDWithRoles(ctx, claims.UserID, utils.Ptr(true))
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	if err := s.tokenRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	// Generate new tokens with device info
	authResp, err := s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
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

	if err := s.tokenRepo.AddToBlacklist(ctx, blacklistEntry); err != nil {
		// Log but don't fail - Redis blacklist is primary
		fmt.Printf("Failed to add token to DB blacklist: %v\n", err)
	}

	// Revoke all refresh tokens for this user
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, claims.UserID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful logout
	s.logAudit(&claims.UserID, models.ActionSignOut, models.StatusSuccess, ip, userAgent, nil)

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
	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
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
	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens (force re-login)
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// Log successful password reset
	s.logAudit(&userID, models.ActionChangePassword, models.StatusSuccess, ip, userAgent, map[string]interface{}{
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
		// Check if email exists
		if exists, err := s.userRepo.EmailExists(ctx, email); err != nil {
			return err
		} else if exists {
			s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "email_exists",
				"email":  email,
			})
			return models.ErrEmailAlreadyExists
		}
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
		// Check if phone exists
		if exists, err := s.userRepo.PhoneExists(ctx, phone); err != nil {
			return err
		} else if exists {
			s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
				"reason": "phone_exists",
				"phone":  phone,
			})
			return models.ErrPhoneAlreadyExists
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

	// Check if username exists
	if exists, err := s.userRepo.UsernameExists(ctx, username); err != nil {
		return err
	} else if exists {
		// Append random suffix to make unique
		username = fmt.Sprintf("%s_%d", username, time.Now().UnixNano()%10000)
	}

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
	s.logAudit(nil, models.ActionSignUp, models.StatusSuccess, ip, userAgent, map[string]interface{}{
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
		s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
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

	// Create user without password (passwordless registration)
	user := &models.User{
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logAudit(nil, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "create_failed",
			"error":  err.Error(),
		})
		return nil, err
	}

	// Assign default "user" role
	if err := s.rbacRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, user.ID); err != nil {
		s.logAudit(&user.ID, models.ActionSignUp, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "role_assignment_failed",
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	// Reload user with roles for token generation
	user, err = s.userRepo.GetByIDWithRoles(ctx, user.ID, utils.Ptr(true))
	if err != nil {
		return nil, fmt.Errorf("failed to reload user with roles: %w", err)
	}

	// Delete pending registration from Redis
	if err := s.redis.DeletePendingRegistration(ctx, identifier); err != nil {
		// Log but don't fail - user is already created
		fmt.Printf("Failed to delete pending registration: %v\n", err)
	}

	// Generate tokens with device info
	authResp, err := s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Log successful signup
	s.logAudit(&user.ID, models.ActionSignUp, models.StatusSuccess, ip, userAgent, map[string]interface{}{
		"passwordless": true,
	})

	return authResp, nil
}

// generateAuthResponse generates access and refresh tokens and saves refresh token with device info
func (s *AuthService) generateAuthResponse(ctx context.Context, user *models.User, ip, userAgent string, deviceInfo models.DeviceInfo) (*models.AuthResponse, error) {
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
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, dbToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Create session using universal SessionCreationService (non-fatal to not block auth)
	if s.sessionService != nil {
		s.sessionService.CreateSessionNonFatal(ctx, SessionCreationParams{
			UserID:      user.ID,
			TokenHash:   tokenHash,
			IPAddress:   ip,
			UserAgent:   userAgent,
			ExpiresAt:   time.Now().Add(refreshExpiration),
			SessionName: sessionName,
		})
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

	// Use the centralized generateAuthResponse method
	return s.generateAuthResponse(ctx, user, ip, userAgent, deviceInfo)
}

func (s *AuthService) logAudit(userID *uuid.UUID, action models.AuditAction, status models.AuditStatus, ip, userAgent string, details map[string]interface{}) {
	s.auditService.Log(AuditLogParams{
		UserID:    userID,
		Action:    action,
		Status:    status,
		IP:        ip,
		UserAgent: userAgent,
		Details:   details,
	})
}
