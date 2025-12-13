package repository

import (
	"context"
	"database/sql"
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

// GetEmailTemplateByID retrieves a template by ID
func (r *TemplateRepository) GetEmailTemplateByID(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error) {
	template := new(models.EmailTemplate)

	err := r.db.NewSelect().
		Model(template).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}

	return template, err
}

// GetEmailTemplateByType retrieves a template by type
func (r *TemplateRepository) GetEmailTemplateByType(ctx context.Context, templateType string) (*models.EmailTemplate, error) {
	template := new(models.EmailTemplate)

	err := r.db.NewSelect().
		Model(template).
		Where("type = ?", templateType).
		Where("is_active = ?", true).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found for type: %s", templateType)
	}

	return template, err
}

// ListEmailTemplates retrieves all templates
func (r *TemplateRepository) ListEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	templates := make([]models.EmailTemplate, 0)

	err := r.db.NewSelect().
		Model(&templates).
		Order("type").
		Scan(ctx)

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
		Set("updated_at = ?", bun.Ident("CURRENT_TIMESTAMP")).
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
