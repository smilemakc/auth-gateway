package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SessionCreationParams contains parameters for creating a new session
type SessionCreationParams struct {
	UserID          uuid.UUID
	ApplicationID   *uuid.UUID // Optional: application context for this session
	TokenHash       string     // Hash of refresh token (use utils.HashToken)
	AccessTokenHash string     // Hash of access token for immediate revocation
	IPAddress       string
	UserAgent       string
	ExpiresAt       time.Time
	SessionName     string // Optional: custom session name
}

// Validate validates the session creation parameters
func (p *SessionCreationParams) Validate() error {
	if p.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if p.TokenHash == "" {
		return errors.New("token_hash is required")
	}
	if p.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}
	return nil
}

// SessionRefreshParams contains parameters for refreshing session tokens
type SessionRefreshParams struct {
	OldRefreshTokenHash string    // Hash of old refresh token to find the session
	NewRefreshTokenHash string    // Hash of new refresh token
	NewAccessTokenHash  string    // Hash of new access token
	NewExpiresAt        time.Time // New expiration time
}

// Validate validates the session refresh parameters
func (p *SessionRefreshParams) Validate() error {
	if p.OldRefreshTokenHash == "" {
		return errors.New("old_refresh_token_hash is required")
	}
	if p.NewRefreshTokenHash == "" {
		return errors.New("new_refresh_token_hash is required")
	}
	if p.NewExpiresAt.IsZero() {
		return errors.New("new_expires_at is required")
	}
	return nil
}

// SessionService handles all session management business logic
type SessionService struct {
	sessionRepo      SessionStore
	blacklistService BlackListStore
	logger           *logger.Logger
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo SessionStore, blacklistService *BlacklistService, logger *logger.Logger) *SessionService {
	return &SessionService{
		sessionRepo:      sessionRepo,
		blacklistService: blacklistService,
		logger:           logger,
	}
}

// =============================================================================
// Session Creation Methods
// =============================================================================

// CreateSessionWithParams creates a new session for any authentication flow.
// Returns the created session and an error if session creation failed.
func (s *SessionService) CreateSessionWithParams(ctx context.Context, params SessionCreationParams) (*models.Session, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Parse user agent using centralized utility
	deviceInfo := utils.ParseUserAgent(params.UserAgent)

	// Generate session name if not provided
	sessionName := params.SessionName
	if sessionName == "" {
		sessionName = utils.GenerateSessionName(deviceInfo)
	}

	session := &models.Session{
		UserID:          params.UserID,
		ApplicationID:   params.ApplicationID,
		TokenHash:       params.TokenHash,
		AccessTokenHash: params.AccessTokenHash,
		DeviceType:      deviceInfo.DeviceType,
		OS:              deviceInfo.OS,
		Browser:         deviceInfo.Browser,
		IPAddress:       params.IPAddress,
		UserAgent:       params.UserAgent,
		SessionName:     sessionName,
		LastActiveAt:    time.Now(),
		ExpiresAt:       params.ExpiresAt,
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		s.logger.Error("session creation failed", map[string]interface{}{
			"user_id":     params.UserID,
			"ip_address":  params.IPAddress,
			"device_type": deviceInfo.DeviceType,
			"error":       err.Error(),
		})
		return nil, err
	}

	s.logger.Info("session created", map[string]interface{}{
		"session_id":   session.ID,
		"user_id":      params.UserID,
		"device_type":  session.DeviceType,
		"browser":      session.Browser,
		"os":           session.OS,
		"session_name": session.SessionName,
	})

	return session, nil
}

// CreateSessionNonFatal creates a session but returns nil instead of error on failure.
// Use this when session creation should not block the authentication flow.
func (s *SessionService) CreateSessionNonFatal(ctx context.Context, params SessionCreationParams) *models.Session {
	session, err := s.CreateSessionWithParams(ctx, params)
	if err != nil {
		return nil
	}
	return session
}

// CreateSessionFromRequest is a convenience method that creates a session
// from common authentication response parameters.
func (s *SessionService) CreateSessionFromRequest(
	ctx context.Context,
	userID uuid.UUID,
	accessToken string,
	refreshToken string,
	ipAddress string,
	userAgent string,
	tokenExpiration time.Duration,
) (*models.Session, error) {
	params := SessionCreationParams{
		UserID:          userID,
		TokenHash:       utils.HashToken(refreshToken),
		AccessTokenHash: utils.HashToken(accessToken),
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		ExpiresAt:       time.Now().Add(tokenExpiration),
	}
	return s.CreateSessionWithParams(ctx, params)
}

// CreateSessionFromRequestNonFatal is the non-fatal version of CreateSessionFromRequest.
func (s *SessionService) CreateSessionFromRequestNonFatal(
	ctx context.Context,
	userID uuid.UUID,
	accessToken string,
	refreshToken string,
	ipAddress string,
	userAgent string,
	tokenExpiration time.Duration,
) *models.Session {
	session, _ := s.CreateSessionFromRequest(ctx, userID, accessToken, refreshToken, ipAddress, userAgent, tokenExpiration)
	return session
}

// =============================================================================
// Session Refresh Methods
// =============================================================================

// RefreshSession updates an existing session with new token hashes and expiration.
// This should be used instead of creating a new session when refreshing tokens.
func (s *SessionService) RefreshSession(ctx context.Context, params SessionRefreshParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	if err := s.sessionRepo.RefreshSessionTokens(
		ctx,
		params.OldRefreshTokenHash,
		params.NewRefreshTokenHash,
		params.NewAccessTokenHash,
		params.NewExpiresAt,
	); err != nil {
		s.logger.Warn("failed to refresh session tokens", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.logger.Debug("session tokens refreshed", map[string]interface{}{
		"new_expires_at": params.NewExpiresAt,
	})

	return nil
}

// RefreshSessionNonFatal refreshes session tokens but ignores errors.
// Returns true if refresh succeeded, false otherwise.
func (s *SessionService) RefreshSessionNonFatal(ctx context.Context, params SessionRefreshParams) bool {
	err := s.RefreshSession(ctx, params)
	return err == nil
}

// RefreshSessionFromTokens is a convenience method that refreshes a session
// from plain tokens (will hash them internally).
func (s *SessionService) RefreshSessionFromTokens(
	ctx context.Context,
	oldRefreshToken string,
	newRefreshToken string,
	newAccessToken string,
	newExpiresAt time.Time,
) error {
	params := SessionRefreshParams{
		OldRefreshTokenHash: utils.HashToken(oldRefreshToken),
		NewRefreshTokenHash: utils.HashToken(newRefreshToken),
		NewAccessTokenHash:  utils.HashToken(newAccessToken),
		NewExpiresAt:        newExpiresAt,
	}
	return s.RefreshSession(ctx, params)
}

// RefreshSessionFromTokensNonFatal is the non-fatal version of RefreshSessionFromTokens.
func (s *SessionService) RefreshSessionFromTokensNonFatal(
	ctx context.Context,
	oldRefreshToken string,
	newRefreshToken string,
	newAccessToken string,
	newExpiresAt time.Time,
) bool {
	err := s.RefreshSessionFromTokens(ctx, oldRefreshToken, newRefreshToken, newAccessToken, newExpiresAt)
	return err == nil
}

// =============================================================================
// Session Query Methods
// =============================================================================

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID, page, perPage int) (*models.SessionListResponse, error) {
	sessions, total, err := s.sessionRepo.GetUserSessionsPaginated(ctx, userID, page, perPage)
	if err != nil {
		return nil, err
	}

	var responseSessions []models.ActiveSessionResponse
	for _, session := range sessions {
		resp := models.ActiveSessionResponse{
			ID:           session.ID,
			UserID:       session.UserID,
			DeviceType:   session.DeviceType,
			OS:           session.OS,
			Browser:      session.Browser,
			UserAgent:    session.UserAgent,
			IPAddress:    session.IPAddress,
			SessionName:  session.SessionName,
			LastActiveAt: session.LastActiveAt,
			CreatedAt:    session.CreatedAt,
			ExpiresAt:    session.ExpiresAt,
			IsCurrent:    false, // This should be set by the handler
		}
		if session.User != nil {
			resp.UserEmail = session.User.Email
			resp.UserName = session.User.Username
		}
		responseSessions = append(responseSessions, resp)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.SessionListResponse{
		Sessions:   responseSessions,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetAllSessions retrieves all active sessions (admin only)
func (s *SessionService) GetAllSessions(ctx context.Context, page, perPage int) (*models.SessionListResponse, error) {
	sessions, total, err := s.sessionRepo.GetAllSessionsPaginated(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	var responseSessions []models.ActiveSessionResponse
	for _, session := range sessions {
		resp := models.ActiveSessionResponse{
			ID:           session.ID,
			UserID:       session.UserID,
			DeviceType:   session.DeviceType,
			OS:           session.OS,
			Browser:      session.Browser,
			UserAgent:    session.UserAgent,
			IPAddress:    session.IPAddress,
			SessionName:  session.SessionName,
			LastActiveAt: session.LastActiveAt,
			CreatedAt:    session.CreatedAt,
			ExpiresAt:    session.ExpiresAt,
			IsCurrent:    false,
		}
		if session.User != nil {
			resp.UserEmail = session.User.Email
			resp.UserName = session.User.Username
		}
		responseSessions = append(responseSessions, resp)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.SessionListResponse{
		Sessions:   responseSessions,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetSessionStats retrieves session statistics
func (s *SessionService) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	return s.sessionRepo.GetSessionStats(ctx)
}

// =============================================================================
// Session Revocation Methods
// =============================================================================

// RevokeSession revokes a specific session (user can only revoke their own session)
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verify ownership
	if session.UserID != userID {
		return fmt.Errorf("session does not belong to user")
	}

	// Blacklist both access and refresh tokens
	if err := s.blacklistService.BlacklistSessionTokens(ctx, session); err != nil {
		s.logger.Error("Failed to blacklist session tokens", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Continue with revocation even if blacklisting fails
	}

	return s.sessionRepo.RevokeUserSession(ctx, userID, sessionID)
}

// AdminRevokeSession revokes any session by ID (admin only, no user ownership check)
func (s *SessionService) AdminRevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Blacklist both access and refresh tokens
	if err := s.blacklistService.BlacklistSessionTokens(ctx, session); err != nil {
		s.logger.Error("Failed to blacklist session tokens", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Continue with revocation even if blacklisting fails
	}

	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

// RevokeAllUserSessions revokes all sessions for a user except the current one
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error {
	sessions, _ := s.sessionRepo.GetUserSessions(ctx, userID)
	for _, session := range sessions {
		if exceptSessionID != nil && *exceptSessionID == session.ID {
			continue
		}
		if err := s.RevokeSession(ctx, userID, session.ID); err != nil {
			return err
		}
	}
	return s.sessionRepo.RevokeAllUserSessions(ctx, userID, exceptSessionID)
}

// RevokeSessionByTokenHash revokes a session by its token hash.
// Returns nil error if session not found (idempotent operation).
func (s *SessionService) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
	session, err := s.sessionRepo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		// Session not found or already revoked - not an error
		return nil
	}

	if err := s.sessionRepo.RevokeSession(ctx, session.ID); err != nil {
		s.logger.Warn("failed to revoke session", map[string]interface{}{
			"session_id": session.ID,
			"error":      err.Error(),
		})
		return err
	}

	s.logger.Info("session revoked", map[string]interface{}{
		"session_id": session.ID,
		"user_id":    session.UserID,
	})

	return nil
}

// RevokeSessionByToken revokes a session by the plain token.
func (s *SessionService) RevokeSessionByToken(ctx context.Context, token string) error {
	tokenHash := utils.HashToken(token)
	return s.RevokeSessionByTokenHash(ctx, tokenHash)
}

// =============================================================================
// Session Maintenance Methods
// =============================================================================

// UpdateSessionName updates the session name
func (s *SessionService) UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error {
	return s.sessionRepo.UpdateSessionName(ctx, sessionID, name)
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpiredSessions(ctx, 7*24*time.Hour) // Delete sessions older than 7 days
}

// =============================================================================
// App-Aware Session Methods
// =============================================================================

// GetUserSessionsByApp retrieves all sessions for a specific user in a specific application
func (s *SessionService) GetUserSessionsByApp(ctx context.Context, userID, appID uuid.UUID) ([]models.Session, error) {
	return s.sessionRepo.GetUserSessionsByApp(ctx, userID, appID)
}

// GetAppSessionsPaginated retrieves paginated sessions for a specific application
func (s *SessionService) GetAppSessionsPaginated(ctx context.Context, appID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	return s.sessionRepo.GetAppSessionsPaginated(ctx, appID, page, perPage)
}
