package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *Database
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *Database) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, ip_address, user_agent, status, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		query,
		log.ID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.IPAddress,
		log.UserAgent,
		log.Status,
		log.Details,
	).Scan(&log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *AuditRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.Select(&logs, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user id: %w", err)
	}

	return logs, nil
}

// GetByAction retrieves audit logs for a specific action
func (r *AuditRepository) GetByAction(action string, limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.Select(&logs, query, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}

	return logs, nil
}

// GetFailedLoginAttempts retrieves failed login attempts for an IP or user
func (r *AuditRepository) GetFailedLoginAttempts(ipAddress string, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE action = 'signin_failed'
		AND ip_address = $1
		AND created_at > NOW() - INTERVAL '15 minutes'
		ORDER BY created_at DESC
		LIMIT $2
	`

	err := r.db.Select(&logs, query, ipAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed login attempts: %w", err)
	}

	return logs, nil
}

// List retrieves all audit logs with pagination
func (r *AuditRepository) List(limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := `
		SELECT * FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.Select(&logs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, nil
}

// Count returns the total number of audit logs
func (r *AuditRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM audit_logs`

	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes audit logs older than a specified duration (for cleanup)
func (r *AuditRepository) DeleteOlderThan(days int) error {
	query := `DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '$1 days'`

	_, err := r.db.Exec(query, days)
	if err != nil {
		return fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return nil
}
