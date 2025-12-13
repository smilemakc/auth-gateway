package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// SMSLogRepository handles SMS log database operations
type SMSLogRepository struct {
	db *Database
}

// NewSMSLogRepository creates a new SMS log repository
func NewSMSLogRepository(db *Database) *SMSLogRepository {
	return &SMSLogRepository{db: db}
}

// Create creates a new SMS log entry
func (r *SMSLogRepository) Create(ctx context.Context, log *models.SMSLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create SMS log: %w", err)
	}

	return nil
}

// GetByID retrieves an SMS log by ID
func (r *SMSLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SMSLog, error) {
	log := new(models.SMSLog)

	err := r.db.NewSelect().
		Model(log).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get SMS log: %w", err)
	}

	return log, nil
}

// UpdateStatus updates the status of an SMS log
func (r *SMSLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
	sentAt := time.Now()

	query := r.db.NewUpdate().
		Model((*models.SMSLog)(nil)).
		Set("status = ?", status).
		Set("error_message = ?", errorMsg).
		Where("id = ?", id)

	if status == "sent" {
		query = query.Set("sent_at = ?", sentAt)
	}

	result, err := query.Exec(ctx)
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
	logs := make([]*models.SMSLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Where("phone = ?", phone).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get SMS logs by phone: %w", err)
	}

	return logs, nil
}

// CountByPhoneAndTimeRange counts SMS logs for a phone number within a time range
func (r *SMSLogRepository) CountByPhoneAndTimeRange(ctx context.Context, phone string, start, end time.Time) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("phone = ?", phone).
		Where("created_at BETWEEN ? AND ?", start, end).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count SMS logs: %w", err)
	}

	return int64(count), nil
}

// CountByPhoneAndType counts SMS logs for a phone number and type within a time range
func (r *SMSLogRepository) CountByPhoneAndType(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int64, error) {
	since := time.Now().Add(-duration)

	count, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("phone = ?", phone).
		Where("type = ?", otpType).
		Where("created_at > ?", since).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count SMS logs by type: %w", err)
	}

	return int64(count), nil
}

// GetRecent retrieves recent SMS logs
func (r *SMSLogRepository) GetRecent(ctx context.Context, limit int) ([]*models.SMSLog, error) {
	logs := make([]*models.SMSLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

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
	totalSent, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("status = ?", "sent").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sent: %w", err)
	}
	stats.TotalSent = int64(totalSent)

	// Total failed
	totalFailed, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("status = ?", "failed").
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total failed: %w", err)
	}
	stats.TotalFailed = int64(totalFailed)

	// Sent today
	today := time.Now().Truncate(24 * time.Hour)
	sentToday, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("status = ?", "sent").
		Where("created_at >= ?", today).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent today: %w", err)
	}
	stats.SentToday = int64(sentToday)

	// Sent this hour
	thisHour := time.Now().Truncate(time.Hour)
	sentThisHour, err := r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Where("status = ?", "sent").
		Where("created_at >= ?", thisHour).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent this hour: %w", err)
	}
	stats.SentThisHour = int64(sentThisHour)

	// By type
	type TypeCount struct {
		Type  string `bun:"type"`
		Count int64  `bun:"count"`
	}
	var typeCounts []TypeCount

	err = r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Column("type").
		ColumnExpr("COUNT(*) as count").
		Group("type").
		Scan(ctx, &typeCounts)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by type: %w", err)
	}

	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	// By status
	type StatusCount struct {
		Status string `bun:"status"`
		Count  int64  `bun:"count"`
	}
	var statusCounts []StatusCount

	err = r.db.NewSelect().
		Model((*models.SMSLog)(nil)).
		Column("status").
		ColumnExpr("COUNT(*) as count").
		Group("status").
		Scan(ctx, &statusCounts)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by status: %w", err)
	}

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
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

	result, err := r.db.NewDelete().
		Model((*models.SMSLog)(nil)).
		Where("created_at < ?", cutoff).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old SMS logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
