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
	_, err := r.db.NewInsert().
		Model(filter).
		Returning("*").
		Exec(ctx)

	return err
}

// GetIPFilterByID retrieves an IP filter by ID
func (r *IPFilterRepository) GetIPFilterByID(ctx context.Context, id uuid.UUID) (*models.IPFilter, error) {
	filter := new(models.IPFilter)

	err := r.db.NewSelect().
		Model(filter).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("IP filter not found")
	}

	return filter, err
}

// ListIPFilters retrieves all IP filters with pagination
func (r *IPFilterRepository) ListIPFilters(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := r.db.NewSelect().
		Model((*models.IPFilter)(nil))

	if filterType != "" {
		countQuery = countQuery.Where("filter_type = ?", filterType)
	}

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get filters with creator info
	filters := make([]models.IPFilterWithCreator, 0)

	query := r.db.NewSelect().
		Model((*models.IPFilter)(nil)).
		ColumnExpr("f.*").
		ColumnExpr("u.username as creator_username").
		ColumnExpr("u.email as creator_email").
		TableExpr("ip_filters AS f").
		Join("LEFT JOIN users AS u ON f.created_by = u.id").
		Order("f.created_at DESC").
		Limit(perPage).
		Offset(offset)

	if filterType != "" {
		query = query.Where("f.filter_type = ?", filterType)
	}

	err = query.Scan(ctx, &filters)

	return filters, total, err
}

// GetActiveIPFilters retrieves all active IP filters
func (r *IPFilterRepository) GetActiveIPFilters(ctx context.Context) ([]models.IPFilter, error) {
	filters := make([]models.IPFilter, 0)

	err := r.db.NewSelect().
		Model(&filters).
		Where("is_active = ?", true).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("expires_at IS NULL").
				WhereOr("expires_at > ?", bun.Ident("NOW()"))
		}).
		Order("filter_type", "created_at DESC").
		Scan(ctx)

	return filters, err
}

// GetActiveFiltersByType retrieves active filters by type
func (r *IPFilterRepository) GetActiveFiltersByType(ctx context.Context, filterType string) ([]models.IPFilter, error) {
	filters := make([]models.IPFilter, 0)

	err := r.db.NewSelect().
		Model(&filters).
		Where("filter_type = ?", filterType).
		Where("is_active = ?", true).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("expires_at IS NULL").
				WhereOr("expires_at > ?", bun.Ident("NOW()"))
		}).
		Order("created_at DESC").
		Scan(ctx)

	return filters, err
}

// UpdateIPFilter updates an IP filter
func (r *IPFilterRepository) UpdateIPFilter(ctx context.Context, id uuid.UUID, reason string, isActive bool) error {
	result, err := r.db.NewUpdate().
		Model((*models.IPFilter)(nil)).
		Set("reason = ?", reason).
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
		return fmt.Errorf("IP filter not found")
	}

	return nil
}

// DeleteIPFilter deletes an IP filter
func (r *IPFilterRepository) DeleteIPFilter(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.IPFilter)(nil)).
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
		return fmt.Errorf("IP filter not found")
	}

	return nil
}

// DeleteExpiredFilters deletes expired IP filters
func (r *IPFilterRepository) DeleteExpiredFilters(ctx context.Context) error {
	_, err := r.db.NewDelete().
		Model((*models.IPFilter)(nil)).
		Where("expires_at IS NOT NULL").
		Where("expires_at < ?", bun.Ident("NOW()")).
		Exec(ctx)

	return err
}
