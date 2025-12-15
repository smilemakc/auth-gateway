package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SessionCreationParams contains parameters for creating a new session
type SessionCreationParams struct {
	UserID      uuid.UUID
	TokenHash   string // Hash of refresh token (use utils.HashToken)
	IPAddress   string
	UserAgent   string
	ExpiresAt   time.Time
	SessionName string // Optional: custom session name
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

// SessionCreationService provides unified session creation across all authentication flows.
// This service follows SOLID principles:
// - Single Responsibility: Only handles session creation
// - Open/Closed: New auth flows can use this without modification
// - Dependency Inversion: Depends on SessionStore interface
type SessionCreationService struct {
	sessionRepo SessionStore
	logger      *logger.Logger
}

// NewSessionCreationService creates a new session creation service
func NewSessionCreationService(sessionRepo SessionStore, log *logger.Logger) *SessionCreationService {
	return &SessionCreationService{
		sessionRepo: sessionRepo,
		logger:      log,
	}
}

// CreateSession creates a new session for any authentication flow.
// Returns the created session and an error if session creation failed.
// The caller decides how to handle the error (fatal vs non-fatal).
func (s *SessionCreationService) CreateSession(ctx context.Context, params SessionCreationParams) (*models.Session, error) {
	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Parse user agent using centralized utility (DRY principle)
	deviceInfo := utils.ParseUserAgent(params.UserAgent)

	// Generate session name if not provided
	sessionName := params.SessionName
	if sessionName == "" {
		sessionName = utils.GenerateSessionName(deviceInfo)
	}

	// Create session model
	session := &models.Session{
		UserID:       params.UserID,
		TokenHash:    params.TokenHash,
		DeviceType:   deviceInfo.DeviceType,
		OS:           deviceInfo.OS,
		Browser:      deviceInfo.Browser,
		IPAddress:    params.IPAddress,
		UserAgent:    params.UserAgent,
		SessionName:  sessionName,
		LastActiveAt: time.Now(),
		ExpiresAt:    params.ExpiresAt,
	}

	// Save session
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
// Errors are logged but not returned to the caller.
func (s *SessionCreationService) CreateSessionNonFatal(ctx context.Context, params SessionCreationParams) *models.Session {
	session, err := s.CreateSession(ctx, params)
	if err != nil {
		// Error already logged in CreateSession
		return nil
	}
	return session
}

// CreateSessionFromRequest is a convenience method that creates a session
// from common authentication response parameters.
// Uses refresh token hash for session tracking (correct lifecycle match).
func (s *SessionCreationService) CreateSessionFromRequest(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	ipAddress string,
	userAgent string,
	tokenExpiration time.Duration,
) (*models.Session, error) {
	params := SessionCreationParams{
		UserID:    userID,
		TokenHash: utils.HashToken(refreshToken),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(tokenExpiration),
	}
	return s.CreateSession(ctx, params)
}

// CreateSessionFromRequestNonFatal is the non-fatal version of CreateSessionFromRequest.
// Returns nil on failure instead of error.
func (s *SessionCreationService) CreateSessionFromRequestNonFatal(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	ipAddress string,
	userAgent string,
	tokenExpiration time.Duration,
) *models.Session {
	session, _ := s.CreateSessionFromRequest(ctx, userID, refreshToken, ipAddress, userAgent, tokenExpiration)
	return session
}

// RevokeSessionByTokenHash revokes a session by its token hash.
// Used when a token is revoked to also revoke the associated session.
// Returns nil error if session not found (idempotent operation).
func (s *SessionCreationService) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
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
// Convenience method that hashes the token before looking up the session.
func (s *SessionCreationService) RevokeSessionByToken(ctx context.Context, token string) error {
	tokenHash := utils.HashToken(token)
	return s.RevokeSessionByTokenHash(ctx, tokenHash)
}
