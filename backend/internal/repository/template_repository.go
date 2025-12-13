package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
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
	query := `
		INSERT INTO email_templates (type, name, subject, html_body, text_body, variables, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		template.Type, template.Name, template.Subject, template.HTMLBody,
		template.TextBody, template.Variables, template.IsActive,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
}

// GetEmailTemplateByID retrieves a template by ID
func (r *TemplateRepository) GetEmailTemplateByID(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	query := `SELECT * FROM email_templates WHERE id = $1`
	err := r.db.GetContext(ctx, &template, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	return &template, err
}

// GetEmailTemplateByType retrieves a template by type
func (r *TemplateRepository) GetEmailTemplateByType(ctx context.Context, templateType string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	query := `SELECT * FROM email_templates WHERE type = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &template, query, templateType)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found for type: %s", templateType)
	}
	return &template, err
}

// ListEmailTemplates retrieves all templates
func (r *TemplateRepository) ListEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	var templates []models.EmailTemplate
	query := `SELECT * FROM email_templates ORDER BY type`
	err := r.db.SelectContext(ctx, &templates, query)
	return templates, err
}

// UpdateEmailTemplate updates a template
func (r *TemplateRepository) UpdateEmailTemplate(ctx context.Context, id uuid.UUID, name, subject, htmlBody, textBody string, variables interface{}, isActive bool) error {
	query := `
		UPDATE email_templates
		SET name = $1, subject = $2, html_body = $3, text_body = $4,
		    variables = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`
	result, err := r.db.ExecContext(ctx, query, name, subject, htmlBody, textBody, variables, isActive, id)
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
	query := `DELETE FROM email_templates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
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
	query := `
		INSERT INTO email_template_versions (template_id, subject, html_body, text_body, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, templateID, subject, htmlBody, textBody, createdBy)
	return err
}
