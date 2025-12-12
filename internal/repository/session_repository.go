package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
)

// SessionRepository handles session database operations
type SessionRepository struct {
	db *Database
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new session (refresh token) with device tracking
func (r *SessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO refresh_tokens (
			user_id, token_hash, device_type, os, browser, ip_address,
			user_agent, session_name, last_active_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		session.UserID, session.TokenHash, session.DeviceType, session.OS,
		session.Browser, session.IPAddress, session.UserAgent, session.SessionName,
		session.LastActiveAt, session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

// GetSessionByID retrieves a session by ID
func (r *SessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	query := `SELECT * FROM refresh_tokens WHERE id = $1`
	err := r.db.GetContext(ctx, &session, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	return &session, err
}

// GetSessionByTokenHash retrieves a session by token hash
func (r *SessionRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL`
	err := r.db.GetContext(ctx, &session, query, tokenHash)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or revoked")
	}
	return &session, err
}

// GetUserSessions retrieves all active sessions for a user
func (r *SessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	query := `
		SELECT * FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY last_active_at DESC
	`
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	return sessions, err
}

// GetUserSessionsPaginated retrieves paginated active sessions for a user
func (r *SessionRepository) GetUserSessionsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated sessions
	var sessions []models.Session
	query := `
		SELECT * FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY last_active_at DESC
		LIMIT $2 OFFSET $3
	`
	err = r.db.SelectContext(ctx, &sessions, query, userID, perPage, offset)
	return sessions, total, err
}

// GetAllActiveSessionsPaginated retrieves all active sessions with pagination (admin)
func (r *SessionRepository) GetAllActiveSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM refresh_tokens
		WHERE revoked_at IS NULL AND expires_at > NOW()
	`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated sessions
	var sessions []models.Session
	query := `
		SELECT * FROM refresh_tokens
		WHERE revoked_at IS NULL AND expires_at > NOW()
		ORDER BY last_active_at DESC
		LIMIT $1 OFFSET $2
	`
	err = r.db.SelectContext(ctx, &sessions, query, perPage, offset)
	return sessions, total, err
}

// UpdateSessionActivity updates the last active timestamp
func (r *SessionRepository) UpdateSessionActivity(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET last_active_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateSessionName updates the session name
func (r *SessionRepository) UpdateSessionName(ctx context.Context, id uuid.UUID, name string) error {
	query := `UPDATE refresh_tokens SET session_name = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// RevokeSession revokes a specific session
func (r *SessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found or already revoked")
	}
	return nil
}

// RevokeUserSession revokes a session only if it belongs to the user
func (r *SessionRepository) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found, already revoked, or does not belong to user")
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user except the current one
func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error {
	var query string
	var err error

	if exceptSessionID != nil {
		query = `
			UPDATE refresh_tokens
			SET revoked_at = NOW()
			WHERE user_id = $1 AND id != $2 AND revoked_at IS NULL
		`
		_, err = r.db.ExecContext(ctx, query, userID, *exceptSessionID)
	} else {
		query = `
			UPDATE refresh_tokens
			SET revoked_at = NOW()
			WHERE user_id = $1 AND revoked_at IS NULL
		`
		_, err = r.db.ExecContext(ctx, query, userID)
	}

	return err
}

// GetSessionStats retrieves session statistics
func (r *SessionRepository) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	var total int
	query := `SELECT COUNT(*) FROM refresh_tokens WHERE revoked_at IS NULL AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &total, query)
	if err != nil {
		return nil, err
	}

	stats := &models.SessionStats{
		TotalActiveSessions: total,
		SessionsByDevice:    make(map[string]int),
		SessionsByOS:        make(map[string]int),
		SessionsByBrowser:   make(map[string]int),
	}

	// Get sessions by device type
	type DeviceCount struct {
		DeviceType string `db:"device_type"`
		Count      int    `db:"count"`
	}
	var deviceCounts []DeviceCount
	deviceQuery := `
		SELECT COALESCE(device_type, 'unknown') as device_type, COUNT(*) as count
		FROM refresh_tokens
		WHERE revoked_at IS NULL AND expires_at > NOW()
		GROUP BY device_type
	`
	err = r.db.SelectContext(ctx, &deviceCounts, deviceQuery)
	if err != nil {
		return nil, err
	}
	for _, dc := range deviceCounts {
		stats.SessionsByDevice[dc.DeviceType] = dc.Count
	}

	// Get sessions by OS
	type OSCount struct {
		OS    string `db:"os"`
		Count int    `db:"count"`
	}
	var osCounts []OSCount
	osQuery := `
		SELECT COALESCE(os, 'unknown') as os, COUNT(*) as count
		FROM refresh_tokens
		WHERE revoked_at IS NULL AND expires_at > NOW()
		GROUP BY os
		ORDER BY count DESC
		LIMIT 10
	`
	err = r.db.SelectContext(ctx, &osCounts, osQuery)
	if err != nil {
		return nil, err
	}
	for _, oc := range osCounts {
		stats.SessionsByOS[oc.OS] = oc.Count
	}

	// Get sessions by browser
	type BrowserCount struct {
		Browser string `db:"browser"`
		Count   int    `db:"count"`
	}
	var browserCounts []BrowserCount
	browserQuery := `
		SELECT COALESCE(browser, 'unknown') as browser, COUNT(*) as count
		FROM refresh_tokens
		WHERE revoked_at IS NULL AND expires_at > NOW()
		GROUP BY browser
		ORDER BY count DESC
		LIMIT 10
	`
	err = r.db.SelectContext(ctx, &browserCounts, browserQuery)
	if err != nil {
		return nil, err
	}
	for _, bc := range browserCounts {
		stats.SessionsByBrowser[bc.Browser] = bc.Count
	}

	return stats, nil
}

// CountUserActiveSessions counts active sessions for a user
func (r *SessionRepository) CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

// DeleteExpiredSessions deletes expired and old revoked sessions
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
		   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - $1::interval)
	`
	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}
