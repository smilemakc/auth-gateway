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

type EmailProfileRepository struct {
	db *Database
}

func NewEmailProfileRepository(db *Database) *EmailProfileRepository {
	return &EmailProfileRepository{db: db}
}

func (r *EmailProfileRepository) Create(ctx context.Context, profile *models.EmailProfile) error {
	_, err := r.db.NewInsert().
		Model(profile).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create email profile: %w", err)
	}

	return nil
}

func (r *EmailProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EmailProfile, error) {
	profile := new(models.EmailProfile)

	err := r.db.NewSelect().
		Model(profile).
		Relation("Provider").
		Where("email_profile.id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get email profile: %w", err)
	}

	return profile, nil
}

func (r *EmailProfileRepository) GetDefault(ctx context.Context) (*models.EmailProfile, error) {
	profile := new(models.EmailProfile)

	err := r.db.NewSelect().
		Model(profile).
		Relation("Provider").
		Where("email_profile.is_default = ?", true).
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get default email profile: %w", err)
	}

	return profile, nil
}

func (r *EmailProfileRepository) GetAll(ctx context.Context) ([]*models.EmailProfile, error) {
	profiles := make([]*models.EmailProfile, 0)

	err := r.db.NewSelect().
		Model(&profiles).
		Relation("Provider").
		Order("email_profile.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get all email profiles: %w", err)
	}

	return profiles, nil
}

func (r *EmailProfileRepository) GetAllActive(ctx context.Context) ([]*models.EmailProfile, error) {
	profiles := make([]*models.EmailProfile, 0)

	err := r.db.NewSelect().
		Model(&profiles).
		Relation("Provider").
		Where("email_profile.is_active = ?", true).
		Order("email_profile.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active email profiles: %w", err)
	}

	return profiles, nil
}

func (r *EmailProfileRepository) Update(ctx context.Context, id uuid.UUID, profile *models.EmailProfile) error {
	result, err := r.db.NewUpdate().
		Model((*models.EmailProfile)(nil)).
		Set("name = ?", profile.Name).
		Set("provider_id = ?", profile.ProviderID).
		Set("from_email = ?", profile.FromEmail).
		Set("from_name = ?", profile.FromName).
		Set("reply_to = ?", profile.ReplyTo).
		Set("is_default = ?", profile.IsDefault).
		Set("is_active = ?", profile.IsActive).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update email profile: %w", err)
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

func (r *EmailProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.EmailProfile)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete email profile: %w", err)
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

func (r *EmailProfileRepository) SetDefault(ctx context.Context, id uuid.UUID) error {
	return r.db.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Model((*models.EmailProfile)(nil)).
			Set("is_default = ?", false).
			Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to clear default flags: %w", err)
		}

		result, err := tx.NewUpdate().
			Model((*models.EmailProfile)(nil)).
			Set("is_default = ?", true).
			Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
			Where("id = ?", id).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to set default flag: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return models.ErrNotFound
		}

		return nil
	})
}

func (r *EmailProfileRepository) ClearDefault(ctx context.Context) error {
	_, err := r.db.NewUpdate().
		Model((*models.EmailProfile)(nil)).
		Set("is_default = ?", false).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to clear default flags: %w", err)
	}

	return nil
}

func (r *EmailProfileRepository) GetTemplatesForProfile(ctx context.Context, profileID uuid.UUID) ([]*models.EmailProfileTemplate, error) {
	templates := make([]*models.EmailProfileTemplate, 0)

	err := r.db.NewSelect().
		Model(&templates).
		Relation("Template").
		Where("email_profile_template.profile_id = ?", profileID).
		Order("email_profile_template.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get templates for profile: %w", err)
	}

	return templates, nil
}

func (r *EmailProfileRepository) GetTemplateForOTPType(ctx context.Context, profileID uuid.UUID, otpType string) (*models.EmailProfileTemplate, error) {
	template := new(models.EmailProfileTemplate)

	err := r.db.NewSelect().
		Model(template).
		Relation("Template").
		Where("email_profile_template.profile_id = ?", profileID).
		Where("email_profile_template.otp_type = ?", otpType).
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get template for OTP type: %w", err)
	}

	return template, nil
}

func (r *EmailProfileRepository) SetTemplateForOTPType(ctx context.Context, profileID uuid.UUID, otpType string, templateID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewInsert().
		Model(&models.EmailProfileTemplate{
			ProfileID:  profileID,
			OTPType:    models.OTPType(otpType),
			TemplateID: templateID,
			CreatedAt:  now,
			UpdatedAt:  now,
		}).
		On("CONFLICT (profile_id, otp_type) DO UPDATE").
		Set("template_id = EXCLUDED.template_id").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to set template for OTP type: %w", err)
	}

	return nil
}

func (r *EmailProfileRepository) RemoveTemplateForOTPType(ctx context.Context, profileID uuid.UUID, otpType string) error {
	result, err := r.db.NewDelete().
		Model((*models.EmailProfileTemplate)(nil)).
		Where("profile_id = ?", profileID).
		Where("otp_type = ?", otpType).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove template for OTP type: %w", err)
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

func (r *EmailProfileRepository) CreateLog(ctx context.Context, log *models.EmailLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create email log: %w", err)
	}

	return nil
}

func (r *EmailProfileRepository) GetLogs(ctx context.Context, limit, offset int) ([]*models.EmailLog, int, error) {
	logs := make([]*models.EmailLog, 0)

	count, err := r.db.NewSelect().
		Model(&logs).
		Relation("Profile").
		Order("email_log.created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get email logs: %w", err)
	}

	return logs, count, nil
}

func (r *EmailProfileRepository) GetLogsByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailLog, error) {
	logs := make([]*models.EmailLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Relation("Profile").
		Where("email_log.recipient_email = ?", email).
		Order("email_log.created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get email logs by recipient: %w", err)
	}

	return logs, nil
}

func (r *EmailProfileRepository) GetLogsByProfile(ctx context.Context, profileID uuid.UUID, limit int) ([]*models.EmailLog, error) {
	logs := make([]*models.EmailLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Relation("Profile").
		Where("email_log.profile_id = ?", profileID).
		Order("email_log.created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get email logs by profile: %w", err)
	}

	return logs, nil
}

func (r *EmailProfileRepository) GetStats(ctx context.Context) (*models.EmailStatsResponse, error) {
	stats := &models.EmailStatsResponse{
		ByTemplateType: make(map[string]int64),
		ByStatus:       make(map[string]int64),
		ByProvider:     make(map[string]int64),
		RecentMessages: make([]models.EmailLog, 0),
	}

	err := r.db.NewSelect().
		Model((*models.EmailLog)(nil)).
		ColumnExpr("COUNT(*) FILTER (WHERE status = 'sent') as total_sent").
		ColumnExpr("COUNT(*) FILTER (WHERE status = 'failed') as total_failed").
		ColumnExpr("COUNT(*) FILTER (WHERE status = 'delivered') as total_delivered").
		ColumnExpr("COUNT(*) FILTER (WHERE status = 'bounced') as total_bounced").
		ColumnExpr("COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as sent_today").
		ColumnExpr("COUNT(*) FILTER (WHERE created_at >= DATE_TRUNC('hour', CURRENT_TIMESTAMP)) as sent_this_hour").
		Scan(ctx, &stats.TotalSent, &stats.TotalFailed, &stats.TotalDelivered, &stats.TotalBounced, &stats.SentToday, &stats.SentThisHour)

	if err != nil {
		return nil, fmt.Errorf("failed to get email statistics: %w", err)
	}

	var templateTypeStats []struct {
		TemplateType string `bun:"template_type"`
		Count        int64  `bun:"count"`
	}
	err = r.db.NewSelect().
		Model((*models.EmailLog)(nil)).
		Column("template_type").
		ColumnExpr("COUNT(*) as count").
		Group("template_type").
		Scan(ctx, &templateTypeStats)

	if err != nil {
		return nil, fmt.Errorf("failed to get template type statistics: %w", err)
	}

	for _, stat := range templateTypeStats {
		stats.ByTemplateType[stat.TemplateType] = stat.Count
	}

	var statusStats []struct {
		Status string `bun:"status"`
		Count  int64  `bun:"count"`
	}
	err = r.db.NewSelect().
		Model((*models.EmailLog)(nil)).
		Column("status").
		ColumnExpr("COUNT(*) as count").
		Group("status").
		Scan(ctx, &statusStats)

	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	for _, stat := range statusStats {
		stats.ByStatus[stat.Status] = stat.Count
	}

	var providerStats []struct {
		ProviderType string `bun:"provider_type"`
		Count        int64  `bun:"count"`
	}
	err = r.db.NewSelect().
		Model((*models.EmailLog)(nil)).
		Column("provider_type").
		ColumnExpr("COUNT(*) as count").
		Group("provider_type").
		Scan(ctx, &providerStats)

	if err != nil {
		return nil, fmt.Errorf("failed to get provider statistics: %w", err)
	}

	for _, stat := range providerStats {
		stats.ByProvider[stat.ProviderType] = stat.Count
	}

	err = r.db.NewSelect().
		Model(&stats.RecentMessages).
		Relation("Profile").
		Order("email_log.created_at DESC").
		Limit(10).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}

	return stats, nil
}
