package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
)

// IPFilterRepository handles IP filter database operations
type IPFilterRepository struct {
	db *Database
}

// NewIPFilterRepository creates a new IP filter repository
func NewIPFilterRepository(db *Database) *IPFilterRepository {
	return &IPFilterRepository{db: db}
}

// CreateIPFilter creates a new IP filter
func (r *IPFilterRepository) CreateIPFilter(ctx context.Context, filter *models.IPFilter) error {
	query := `
		INSERT INTO ip_filters (ip_cidr, filter_type, reason, created_by, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		filter.IPCIDR, filter.FilterType, filter.Reason, filter.CreatedBy,
		filter.IsActive, filter.ExpiresAt,
	).Scan(&filter.ID, &filter.CreatedAt, &filter.UpdatedAt)
}

// GetIPFilterByID retrieves an IP filter by ID
func (r *IPFilterRepository) GetIPFilterByID(ctx context.Context, id uuid.UUID) (*models.IPFilter, error) {
	var filter models.IPFilter
	query := `SELECT * FROM ip_filters WHERE id = $1`
	err := r.db.GetContext(ctx, &filter, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("IP filter not found")
	}
	return &filter, err
}

// ListIPFilters retrieves all IP filters with pagination
func (r *IPFilterRepository) ListIPFilters(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
	offset := (page - 1) * perPage

	// Build query based on filter type
	var countQuery, selectQuery string
	var args []interface{}

	if filterType != "" {
		countQuery = `SELECT COUNT(*) FROM ip_filters WHERE filter_type = $1`
		selectQuery = `
			SELECT
				f.*,
				u.username as creator_username,
				u.email as creator_email
			FROM ip_filters f
			LEFT JOIN users u ON f.created_by = u.id
			WHERE f.filter_type = $1
			ORDER BY f.created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{filterType, perPage, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM ip_filters`
		selectQuery = `
			SELECT
				f.*,
				u.username as creator_username,
				u.email as creator_email
			FROM ip_filters f
			LEFT JOIN users u ON f.created_by = u.id
			ORDER BY f.created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{perPage, offset}
	}

	// Get total count
	var total int
	var err error
	if filterType != "" {
		err = r.db.GetContext(ctx, &total, countQuery, filterType)
	} else {
		err = r.db.GetContext(ctx, &total, countQuery)
	}
	if err != nil {
		return nil, 0, err
	}

	// Get filters
	var filters []models.IPFilterWithCreator
	err = r.db.SelectContext(ctx, &filters, selectQuery, args...)
	return filters, total, err
}

// GetActiveIPFilters retrieves all active IP filters
func (r *IPFilterRepository) GetActiveIPFilters(ctx context.Context) ([]models.IPFilter, error) {
	var filters []models.IPFilter
	query := `
		SELECT * FROM ip_filters
		WHERE is_active = true
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY filter_type, created_at DESC
	`
	err := r.db.SelectContext(ctx, &filters, query)
	return filters, err
}

// GetActiveFiltersByType retrieves active filters by type
func (r *IPFilterRepository) GetActiveFiltersByType(ctx context.Context, filterType string) ([]models.IPFilter, error) {
	var filters []models.IPFilter
	query := `
		SELECT * FROM ip_filters
		WHERE filter_type = $1
		  AND is_active = true
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &filters, query, filterType)
	return filters, err
}

// UpdateIPFilter updates an IP filter
func (r *IPFilterRepository) UpdateIPFilter(ctx context.Context, id uuid.UUID, reason string, isActive bool) error {
	query := `
		UPDATE ip_filters
		SET reason = $1, is_active = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, reason, isActive, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("IP filter not found")
	}
	return nil
}

// DeleteIPFilter deletes an IP filter
func (r *IPFilterRepository) DeleteIPFilter(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ip_filters WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("IP filter not found")
	}
	return nil
}

// DeleteExpiredFilters deletes expired IP filters
func (r *IPFilterRepository) DeleteExpiredFilters(ctx context.Context) error {
	query := `DELETE FROM ip_filters WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
