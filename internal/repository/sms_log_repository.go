package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// SMSLogRepository handles SMS log database operations
type SMSLogRepository struct {
	db *sqlx.DB
}

// NewSMSLogRepository creates a new SMS log repository
func NewSMSLogRepository(db *sqlx.DB) *SMSLogRepository {
	return &SMSLogRepository{db: db}
}

// Create creates a new SMS log entry
func (r *SMSLogRepository) Create(ctx context.Context, log *models.SMSLog) error {
	query := `
		INSERT INTO sms_logs (
			id, phone, message, type, provider, message_id, status,
			error_message, sent_at, user_id, ip_address, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.Phone,
		log.Message,
		log.Type,
		log.Provider,
		log.MessageID,
		log.Status,
		log.ErrorMessage,
		log.SentAt,
		log.UserID,
		log.IPAddress,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create SMS log: %w", err)
	}

	return nil
}

// GetByID retrieves an SMS log by ID
func (r *SMSLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SMSLog, error) {
	var log models.SMSLog
	query := `SELECT * FROM sms_logs WHERE id = $1`

	err := r.db.GetContext(ctx, &log, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get SMS log: %w", err)
	}

	return &log, nil
}

// UpdateStatus updates the status of an SMS log
func (r *SMSLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
	query := `
		UPDATE sms_logs SET
			status = $2,
			error_message = $3,
			sent_at = CASE WHEN $2 = 'sent' THEN $4 ELSE sent_at END
		WHERE id = $1
	`

	sentAt := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, status, errorMsg, sentAt)
	if err != nil {
		return fmt.Errorf("failed to update SMS log status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	return nil
}

// GetByPhone retrieves SMS logs for a phone number
func (r *SMSLogRepository) GetByPhone(ctx context.Context, phone string, limit int) ([]*models.SMSLog, error) {
	var logs []*models.SMSLog
	query := `
		SELECT * FROM sms_logs
		WHERE phone = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	err := r.db.SelectContext(ctx, &logs, query, phone, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMS logs by phone: %w", err)
	}

	return logs, nil
}

// CountByPhoneAndTimeRange counts SMS logs for a phone number within a time range
func (r *SMSLogRepository) CountByPhoneAndTimeRange(ctx context.Context, phone string, start, end time.Time) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM sms_logs
		WHERE phone = $1 AND created_at BETWEEN $2 AND $3
	`

	err := r.db.GetContext(ctx, &count, query, phone, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to count SMS logs: %w", err)
	}

	return count, nil
}

// CountByPhoneAndType counts SMS logs for a phone number and type within a time range
func (r *SMSLogRepository) CountByPhoneAndType(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM sms_logs
		WHERE phone = $1 AND type = $2 AND created_at > $3
	`

	since := time.Now().Add(-duration)
	err := r.db.GetContext(ctx, &count, query, phone, otpType, since)
	if err != nil {
		return 0, fmt.Errorf("failed to count SMS logs by type: %w", err)
	}

	return count, nil
}

// GetRecent retrieves recent SMS logs
func (r *SMSLogRepository) GetRecent(ctx context.Context, limit int) ([]*models.SMSLog, error) {
	var logs []*models.SMSLog
	query := `
		SELECT * FROM sms_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	err := r.db.SelectContext(ctx, &logs, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent SMS logs: %w", err)
	}

	return logs, nil
}

// GetStats retrieves SMS statistics
func (r *SMSLogRepository) GetStats(ctx context.Context) (*models.SMSStatsResponse, error) {
	stats := &models.SMSStatsResponse{
		ByType:   make(map[string]int64),
		ByStatus: make(map[string]int64),
	}

	// Total sent
	err := r.db.GetContext(ctx, &stats.TotalSent, `SELECT COUNT(*) FROM sms_logs WHERE status = 'sent'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sent: %w", err)
	}

	// Total failed
	err = r.db.GetContext(ctx, &stats.TotalFailed, `SELECT COUNT(*) FROM sms_logs WHERE status = 'failed'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total failed: %w", err)
	}

	// Sent today
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.GetContext(ctx, &stats.SentToday, `
		SELECT COUNT(*) FROM sms_logs
		WHERE status = 'sent' AND created_at >= $1
	`, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent today: %w", err)
	}

	// Sent this hour
	thisHour := time.Now().Truncate(time.Hour)
	err = r.db.GetContext(ctx, &stats.SentThisHour, `
		SELECT COUNT(*) FROM sms_logs
		WHERE status = 'sent' AND created_at >= $1
	`, thisHour)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent this hour: %w", err)
	}

	// By type
	typeRows, err := r.db.QueryContext(ctx, `
		SELECT type, COUNT(*) as count FROM sms_logs
		GROUP BY type
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var otpType string
		var count int64
		if err := typeRows.Scan(&otpType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		stats.ByType[otpType] = count
	}

	// By status
	statusRows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) as count FROM sms_logs
		GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by status: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.ByStatus[status] = count
	}

	// Recent messages
	recentLogs, err := r.GetRecent(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}

	stats.RecentMessages = make([]models.SMSLog, len(recentLogs))
	for i, log := range recentLogs {
		stats.RecentMessages[i] = *log
	}

	return stats, nil
}

// DeleteOlderThan deletes SMS logs older than the specified duration
func (r *SMSLogRepository) DeleteOlderThan(ctx context.Context, duration time.Duration) (int64, error) {
	cutoff := time.Now().Add(-duration)
	query := `DELETE FROM sms_logs WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old SMS logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
