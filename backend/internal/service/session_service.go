package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// SessionService handles session management business logic
type SessionService struct {
	sessionRepo SessionStore
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo SessionStore) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSession creates a new session with device tracking
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash, ipAddress, userAgent string, expiresAt time.Time) (*models.Session, error) {
	// Parse user agent
	deviceInfo := utils.ParseUserAgent(userAgent)

	session := &models.Session{
		UserID:       userID,
		TokenHash:    tokenHash,
		DeviceType:   deviceInfo.DeviceType,
		OS:           deviceInfo.OS,
		Browser:      deviceInfo.Browser,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActiveAt: time.Now(),
		ExpiresAt:    expiresAt,
	}

	err := s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID, page, perPage int) (*models.SessionListResponse, error) {
	sessions, total, err := s.sessionRepo.GetUserSessionsPaginated(ctx, userID, page, perPage)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	var responseSessions []models.ActiveSessionResponse
	for _, session := range sessions {
		responseSessions = append(responseSessions, models.ActiveSessionResponse{
			ID:           session.ID,
			DeviceType:   session.DeviceType,
			OS:           session.OS,
			Browser:      session.Browser,
			IPAddress:    session.IPAddress,
			SessionName:  session.SessionName,
			LastActiveAt: session.LastActiveAt,
			CreatedAt:    session.CreatedAt,
			ExpiresAt:    session.ExpiresAt,
			IsCurrent:    false, // This should be set by the handler
		})
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

// RevokeSession revokes a specific session
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeUserSession(ctx, userID, sessionID)
}

// RevokeAllUserSessions revokes all sessions for a user except the current one
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error {
	return s.sessionRepo.RevokeAllUserSessions(ctx, userID, exceptSessionID)
}

// UpdateSessionName updates the session name
func (s *SessionService) UpdateSessionName(ctx context.Context, sessionID uuid.UUID, name string) error {
	return s.sessionRepo.UpdateSessionName(ctx, sessionID, name)
}

// GetSessionStats retrieves session statistics
func (s *SessionService) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	return s.sessionRepo.GetSessionStats(ctx)
}

// GetAllSessions retrieves all active sessions (admin only)
func (s *SessionService) GetAllSessions(ctx context.Context, page, perPage int) (*models.SessionListResponse, error) {
	sessions, total, err := s.sessionRepo.GetAllSessionsPaginated(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	var responseSessions []models.ActiveSessionResponse
	for _, session := range sessions {
		responseSessions = append(responseSessions, models.ActiveSessionResponse{
			ID:           session.ID,
			DeviceType:   session.DeviceType,
			OS:           session.OS,
			Browser:      session.Browser,
			IPAddress:    session.IPAddress,
			SessionName:  session.SessionName,
			LastActiveAt: session.LastActiveAt,
			CreatedAt:    session.CreatedAt,
			ExpiresAt:    session.ExpiresAt,
			IsCurrent:    false,
		})
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

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpiredSessions(ctx, 7*24*time.Hour) // Delete sessions older than 7 days
}
