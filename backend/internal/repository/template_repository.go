package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// TemplateRepository handles email template database operations
type TemplateRepository struct {
	db *Database
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *Database) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// CreateEmailTemplate creates a new email template
func (r *TemplateRepository) CreateEmailTemplate(ctx context.Context, template *models.EmailTemplate) error {
	_, err := r.db.NewInsert().
		Model(template).
		Returning("*").
		Exec(ctx)

	return err
}

// UpsertEmailTemplate inserts or updates a template by (type, application_id).
func (r *TemplateRepository) UpsertEmailTemplate(ctx context.Context, template *models.EmailTemplate) error {
	_, err := r.db.NewInsert().
		Model(template).
		On("CONFLICT (type, COALESCE(application_id, '00000000-0000-0000-0000-000000000000')) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("subject = EXCLUDED.subject").
		Set("html_body = EXCLUDED.html_body").
		Set("text_body = EXCLUDED.text_body").
		Set("variables = EXCLUDED.variables").
		Set("is_active = EXCLUDED.is_active").
		Set("updated_at = NOW()").
		Returning("*").
		Exec(ctx)
	return err
}

// GetEmailTemplateByID retrieves a template by ID with optional application relation loading
func (r *TemplateRepository) GetEmailTemplateByID(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error) {
	template := new(models.EmailTemplate)

	err := r.db.NewSelect().
		Model(template).
		Relation("Application").
		Where("email_template.id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("template not found")
	}

	return template, err
}

// GetEmailTemplateByType retrieves a global template by type (no application scope)
func (r *TemplateRepository) GetEmailTemplateByType(ctx context.Context, templateType string) (*models.EmailTemplate, error) {
	template := new(models.EmailTemplate)

	err := r.db.NewSelect().
		Model(template).
		Where("type = ?", templateType).
		Where("is_active = ?", true).
		Where("application_id IS NULL").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("template not found for type: %s", templateType)
	}

	return template, err
}

// ListEmailTemplates retrieves all templates with optional application filter
func (r *TemplateRepository) ListEmailTemplates(ctx context.Context, applicationID *uuid.UUID) ([]models.EmailTemplate, error) {
	templates := make([]models.EmailTemplate, 0)

	query := r.db.NewSelect().
		Model(&templates).
		Relation("Application")

	if applicationID != nil {
		query = query.Where("email_template.application_id = ?", applicationID)
	}

	err := query.Order("email_template.type").Scan(ctx)

	return templates, err
}

// UpdateEmailTemplate updates a template
func (r *TemplateRepository) UpdateEmailTemplate(ctx context.Context, id uuid.UUID, name, subject, htmlBody, textBody string, variables interface{}, isActive bool) error {
	result, err := r.db.NewUpdate().
		Model((*models.EmailTemplate)(nil)).
		Set("name = ?", name).
		Set("subject = ?", subject).
		Set("html_body = ?", htmlBody).
		Set("text_body = ?", textBody).
		Set("variables = ?", variables).
		Set("is_active = ?", isActive).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// DeleteEmailTemplate deletes a template
func (r *TemplateRepository) DeleteEmailTemplate(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.EmailTemplate)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// CreateTemplateVersion creates a version history record
func (r *TemplateRepository) CreateTemplateVersion(ctx context.Context, templateID uuid.UUID, subject, htmlBody, textBody string, createdBy *uuid.UUID) error {
	version := &models.EmailTemplateVersion{
		TemplateID: templateID,
		Subject:    subject,
		HTMLBody:   htmlBody,
		TextBody:   textBody,
		CreatedBy:  createdBy,
	}

	_, err := r.db.NewInsert().
		Model(version).
		Exec(ctx)

	return err
}

// CreateTemplateVersionWithApp creates a version history record with application ID
func (r *TemplateRepository) CreateTemplateVersionWithApp(ctx context.Context, templateID uuid.UUID, applicationID *uuid.UUID, subject, htmlBody, textBody string, createdBy *uuid.UUID) error {
	version := &models.EmailTemplateVersion{
		TemplateID:    templateID,
		ApplicationID: applicationID,
		Subject:       subject,
		HTMLBody:      htmlBody,
		TextBody:      textBody,
		CreatedBy:     createdBy,
	}

	_, err := r.db.NewInsert().
		Model(version).
		Exec(ctx)

	return err
}

// ListEmailTemplatesByApplication retrieves templates for a specific application
func (r *TemplateRepository) ListEmailTemplatesByApplication(ctx context.Context, applicationID uuid.UUID) ([]models.EmailTemplate, error) {
	templates := make([]models.EmailTemplate, 0)

	err := r.db.NewSelect().
		Model(&templates).
		Relation("Application").
		Where("email_template.application_id = ?", applicationID).
		Order("email_template.type").
		Scan(ctx)

	return templates, err
}

// GetEmailTemplateByTypeAndApp retrieves a template by type for a specific application
func (r *TemplateRepository) GetEmailTemplateByTypeAndApp(ctx context.Context, templateType string, applicationID uuid.UUID) (*models.EmailTemplate, error) {
	template := new(models.EmailTemplate)

	err := r.db.NewSelect().
		Model(template).
		Relation("Application").
		Where("email_template.type = ?", templateType).
		Where("email_template.application_id = ?", applicationID).
		Where("email_template.is_active = ?", true).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("template not found for type: %s and application: %s", templateType, applicationID)
	}

	return template, err
}

// CopyTemplatesForApplication copies default templates to a new application
func (r *TemplateRepository) CopyTemplatesForApplication(ctx context.Context, sourceAppID, targetAppID uuid.UUID) error {
	sourceTemplates := make([]models.EmailTemplate, 0)

	err := r.db.NewSelect().
		Model(&sourceTemplates).
		Where("email_template.application_id = ?", sourceAppID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch source templates: %w", err)
	}

	if len(sourceTemplates) == 0 {
		return fmt.Errorf("no templates found for source application: %s", sourceAppID)
	}

	for _, srcTemplate := range sourceTemplates {
		newTemplate := &models.EmailTemplate{
			Type:          srcTemplate.Type,
			Name:          srcTemplate.Name,
			Subject:       srcTemplate.Subject,
			HTMLBody:      srcTemplate.HTMLBody,
			TextBody:      srcTemplate.TextBody,
			Variables:     srcTemplate.Variables,
			IsActive:      srcTemplate.IsActive,
			ApplicationID: &targetAppID,
		}

		_, err := r.db.NewInsert().
			Model(newTemplate).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to copy template %s: %w", srcTemplate.Type, err)
		}
	}

	return nil
}
