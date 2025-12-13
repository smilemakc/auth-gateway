package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
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
	_, err := r.db.NewInsert().
		Model(session).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetSessionByID retrieves a session by ID
func (r *SessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	session := new(models.Session)

	err := r.db.NewSelect().
		Model(session).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by id: %w", err)
	}

	return session, nil
}

// GetSessionByTokenHash retrieves a session by token hash
func (r *SessionRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	session := new(models.Session)

	err := r.db.NewSelect().
		Model(session).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("session not found or revoked")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *SessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	sessions := make([]models.Session, 0)

	err := r.db.NewSelect().
		Model(&sessions).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Order("last_active_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}

// GetUserSessionsPaginated retrieves paginated active sessions for a user
func (r *SessionRepository) GetUserSessionsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Session, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	total, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Count(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user sessions: %w", err)
	}

	// Get paginated sessions
	sessions := make([]models.Session, 0)

	err = r.db.NewSelect().
		Model(&sessions).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Order("last_active_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get paginated user sessions: %w", err)
	}

	return sessions, total, nil
}

// GetAllActiveSessionsPaginated retrieves all active sessions with pagination (admin)
func (r *SessionRepository) GetAllActiveSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	total, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Count(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count all sessions: %w", err)
	}

	// Get paginated sessions
	sessions := make([]models.Session, 0)

	err = r.db.NewSelect().
		Model(&sessions).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Order("last_active_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get paginated sessions: %w", err)
	}

	return sessions, total, nil
}

// GetAllSessionsPaginated is an alias for GetAllActiveSessionsPaginated (implements interface)
func (r *SessionRepository) GetAllSessionsPaginated(ctx context.Context, page, perPage int) ([]models.Session, int, error) {
	return r.GetAllActiveSessionsPaginated(ctx, page, perPage)
}

// UpdateSessionActivity updates the last active timestamp
func (r *SessionRepository) UpdateSessionActivity(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("last_active_at = ?", bun.Ident("NOW()")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return nil
}

// UpdateSessionName updates the session name
func (r *SessionRepository) UpdateSessionName(ctx context.Context, id uuid.UUID, name string) error {
	result, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("session_name = ?", name).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update session name: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// RevokeSession revokes a specific session
func (r *SessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked_at = ?", bun.Ident("NOW()")).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return nil
}

// RevokeUserSession revokes a session only if it belongs to the user
func (r *SessionRepository) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked_at = ?", bun.Ident("NOW()")).
		Where("id = ?", sessionID).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke user session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found, already revoked, or does not belong to user")
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user except the current one
func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, exceptSessionID *uuid.UUID) error {
	query := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked_at = ?", bun.Ident("NOW()")).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL")

	if exceptSessionID != nil {
		query = query.Where("id != ?", *exceptSessionID)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	return nil
}

// GetSessionStats retrieves session statistics
func (r *SessionRepository) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	// Get total count
	total, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count total sessions: %w", err)
	}

	stats := &models.SessionStats{
		TotalActiveSessions: total,
		SessionsByDevice:    make(map[string]int),
		SessionsByOS:        make(map[string]int),
		SessionsByBrowser:   make(map[string]int),
	}

	// Get sessions by device type
	type DeviceCount struct {
		DeviceType string `bun:"device_type"`
		Count      int    `bun:"count"`
	}
	var deviceCounts []DeviceCount

	err = r.db.NewSelect().
		Model((*models.Session)(nil)).
		ColumnExpr("COALESCE(device_type, 'unknown') as device_type").
		ColumnExpr("COUNT(*) as count").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Group("device_type").
		Scan(ctx, &deviceCounts)

	if err != nil {
		return nil, fmt.Errorf("failed to get device stats: %w", err)
	}

	for _, dc := range deviceCounts {
		stats.SessionsByDevice[dc.DeviceType] = dc.Count
	}

	// Get sessions by OS
	type OSCount struct {
		OS    string `bun:"os"`
		Count int    `bun:"count"`
	}
	var osCounts []OSCount

	err = r.db.NewSelect().
		Model((*models.Session)(nil)).
		ColumnExpr("COALESCE(os, 'unknown') as os").
		ColumnExpr("COUNT(*) as count").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Group("os").
		Order("count DESC").
		Limit(10).
		Scan(ctx, &osCounts)

	if err != nil {
		return nil, fmt.Errorf("failed to get OS stats: %w", err)
	}

	for _, oc := range osCounts {
		stats.SessionsByOS[oc.OS] = oc.Count
	}

	// Get sessions by browser
	type BrowserCount struct {
		Browser string `bun:"browser"`
		Count   int    `bun:"count"`
	}
	var browserCounts []BrowserCount

	err = r.db.NewSelect().
		Model((*models.Session)(nil)).
		ColumnExpr("COALESCE(browser, 'unknown') as browser").
		ColumnExpr("COUNT(*) as count").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Group("browser").
		Order("count DESC").
		Limit(10).
		Scan(ctx, &browserCounts)

	if err != nil {
		return nil, fmt.Errorf("failed to get browser stats: %w", err)
	}

	for _, bc := range browserCounts {
		stats.SessionsByBrowser[bc.Browser] = bc.Count
	}

	return stats, nil
}

// CountUserActiveSessions counts active sessions for a user
func (r *SessionRepository) CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", bun.Ident("NOW()")).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count user active sessions: %w", err)
	}

	return count, nil
}

// DeleteExpiredSessions deletes expired and old revoked sessions
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context, olderThan time.Duration) error {
	_, err := r.db.NewDelete().
		Model((*models.Session)(nil)).
		WhereOr("expires_at < ?", bun.Ident("NOW()")).
		WhereOr("(revoked_at IS NOT NULL AND revoked_at < ? - ?::interval)", bun.Ident("NOW()"), olderThan).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}
