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

type ApplicationRepository struct {
	db *Database
}

func NewApplicationRepository(db *Database) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) CreateApplication(ctx context.Context, app *models.Application) error {
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(app).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *ApplicationRepository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	app := new(models.Application)

	err := r.db.NewSelect().
		Model(app).
		Where("app.id = ?", id).
		Relation("Branding").
		Relation("Owner").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("application not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application by id: %w", err)
	}

	return app, nil
}

func (r *ApplicationRepository) GetApplicationByName(ctx context.Context, name string) (*models.Application, error) {
	app := new(models.Application)

	err := r.db.NewSelect().
		Model(app).
		Where("name = ?", name).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("application not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application by name: %w", err)
	}

	return app, nil
}

func (r *ApplicationRepository) UpdateApplication(ctx context.Context, app *models.Application) error {
	app.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(app).
		Column("display_name", "description", "homepage_url", "callback_urls",
			"is_active", "is_system", "owner_id", "allowed_auth_methods",
			"secret_hash", "secret_prefix", "secret_last_rotated_at", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

func (r *ApplicationRepository) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.Application)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

func (r *ApplicationRepository) HardDeleteApplication(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.Application)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to hard delete application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

func (r *ApplicationRepository) ListApplications(ctx context.Context, page, perPage int, isActive *bool) ([]*models.Application, int, error) {
	apps := make([]*models.Application, 0)

	query := r.db.NewSelect().
		Model(&apps)

	if isActive != nil {
		query = query.Where("app.is_active = ?", *isActive)
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count applications: %w", err)
	}

	offset := (page - 1) * perPage

	err = query.
		Relation("Owner").
		Relation("Branding").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list applications: %w", err)
	}

	return apps, total, nil
}

func (r *ApplicationRepository) GetBranding(ctx context.Context, applicationID uuid.UUID) (*models.ApplicationBranding, error) {
	branding := new(models.ApplicationBranding)

	err := r.db.NewSelect().
		Model(branding).
		Where("application_id = ?", applicationID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application branding: %w", err)
	}

	return branding, nil
}

func (r *ApplicationRepository) CreateOrUpdateBranding(ctx context.Context, branding *models.ApplicationBranding) error {
	branding.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(branding).
		On("CONFLICT (application_id) DO UPDATE").
		Set("logo_url = EXCLUDED.logo_url").
		Set("favicon_url = EXCLUDED.favicon_url").
		Set("primary_color = EXCLUDED.primary_color").
		Set("secondary_color = EXCLUDED.secondary_color").
		Set("background_color = EXCLUDED.background_color").
		Set("custom_css = EXCLUDED.custom_css").
		Set("company_name = EXCLUDED.company_name").
		Set("support_email = EXCLUDED.support_email").
		Set("terms_url = EXCLUDED.terms_url").
		Set("privacy_url = EXCLUDED.privacy_url").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *ApplicationRepository) CreateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(profile).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *ApplicationRepository) GetUserProfile(ctx context.Context, userID, applicationID uuid.UUID) (*models.UserApplicationProfile, error) {
	profile := new(models.UserApplicationProfile)

	err := r.db.NewSelect().
		Model(profile).
		Where("user_id = ?", userID).
		Where("application_id = ?", applicationID).
		Relation("User").
		Relation("Application").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return profile, nil
}

func (r *ApplicationRepository) UpdateUserProfile(ctx context.Context, profile *models.UserApplicationProfile) error {
	profile.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(profile).
		Column("display_name", "avatar_url", "nickname", "metadata",
			"app_roles", "is_active", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *ApplicationRepository) DeleteUserProfile(ctx context.Context, userID, applicationID uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.UserApplicationProfile)(nil)).
		Where("user_id = ?", userID).
		Where("application_id = ?", applicationID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *ApplicationRepository) ListUserProfiles(ctx context.Context, userID uuid.UUID) ([]*models.UserApplicationProfile, error) {
	profiles := make([]*models.UserApplicationProfile, 0)

	err := r.db.NewSelect().
		Model(&profiles).
		Where("user_id = ?", userID).
		Relation("Application").
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list user profiles: %w", err)
	}

	return profiles, nil
}

func (r *ApplicationRepository) ListApplicationUsers(ctx context.Context, applicationID uuid.UUID, page, perPage int) ([]*models.UserApplicationProfile, int, error) {
	profiles := make([]*models.UserApplicationProfile, 0)

	query := r.db.NewSelect().
		Model(&profiles).
		Where("application_id = ?", applicationID)

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count application users: %w", err)
	}

	offset := (page - 1) * perPage

	err = query.
		Relation("User").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list application users: %w", err)
	}

	return profiles, total, nil
}

func (r *ApplicationRepository) UpdateLastAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.UserApplicationProfile)(nil)).
		Set("last_access_at = ?", now).
		Where("user_id = ?", userID).
		Where("application_id = ?", applicationID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update last access: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *ApplicationRepository) BanUserFromApplication(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, banReason string) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.UserApplicationProfile)(nil)).
		Set("is_banned = ?", true).
		Set("ban_reason = ?", banReason).
		Set("banned_at = ?", now).
		Set("banned_by = ?", bannedBy).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("application_id = ?", applicationID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to ban user from application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *ApplicationRepository) UnbanUserFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.UserApplicationProfile)(nil)).
		Set("is_banned = ?", false).
		Set("ban_reason = ?", nil).
		Set("banned_at = ?", nil).
		Set("banned_by = ?", nil).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("application_id = ?", applicationID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to unban user from application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *ApplicationRepository) GetBySecretHash(ctx context.Context, hash string) (*models.Application, error) {
	app := new(models.Application)
	err := r.db.NewSelect().
		Model(app).
		Where("secret_hash = ?", hash).
		Where("secret_hash != ''").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("application not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application by secret hash: %w", err)
	}
	return app, nil
}
