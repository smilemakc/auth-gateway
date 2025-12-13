package repository

import (
	"context"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// GeoRepository handles geo-location database operations
type GeoRepository struct {
	db *Database
}

// NewGeoRepository creates a new geo repository
func NewGeoRepository(db *Database) *GeoRepository {
	return &GeoRepository{db: db}
}

// GetLoginGeoDistribution retrieves login distribution for map visualization
func (r *GeoRepository) GetLoginGeoDistribution(ctx context.Context, days int) ([]models.LoginLocation, error) {
	locations := make([]models.LoginLocation, 0)

	err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Column("country_code", "country_name", "city", "latitude", "longitude").
		ColumnExpr("COUNT(*) as login_count").
		ColumnExpr("MAX(created_at) as last_login_at").
		Where("action = ?", "login").
		Where("status = ?", "success").
		Where("country_code IS NOT NULL").
		Where("created_at >= ?", bun.Safe("NOW() - INTERVAL '1 day' * ?"), days).
		Group("country_code", "country_name", "city", "latitude", "longitude").
		Order("login_count DESC").
		Scan(ctx, &locations)

	return locations, err
}

// GetTopCountries retrieves top countries by login count
func (r *GeoRepository) GetTopCountries(ctx context.Context, limit, days int) ([]models.CountryStats, error) {
	stats := make([]models.CountryStats, 0)

	err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Column("country_code", "country_name").
		ColumnExpr("COUNT(*) as login_count").
		ColumnExpr("COUNT(DISTINCT user_id) as user_count").
		Where("action = ?", "login").
		Where("status = ?", "success").
		Where("country_code IS NOT NULL").
		Where("created_at >= ?", bun.Safe("NOW() - INTERVAL '1 day' * ?"), days).
		Group("country_code", "country_name").
		Order("login_count DESC").
		Limit(limit).
		Scan(ctx, &stats)

	return stats, err
}

// GetTopCities retrieves top cities by login count
func (r *GeoRepository) GetTopCities(ctx context.Context, limit, days int) ([]models.CityStats, error) {
	stats := make([]models.CityStats, 0)

	err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Column("country_code", "country_name", "city").
		ColumnExpr("COUNT(*) as login_count").
		ColumnExpr("COUNT(DISTINCT user_id) as user_count").
		Where("action = ?", "login").
		Where("status = ?", "success").
		Where("city IS NOT NULL").
		Where("created_at >= ?", bun.Safe("NOW() - INTERVAL '1 day' * ?"), days).
		Group("country_code", "country_name", "city").
		Order("login_count DESC").
		Limit(limit).
		Scan(ctx, &stats)

	return stats, err
}

// UpdateOrCreateLoginLocation updates or creates login location aggregation
func (r *GeoRepository) UpdateOrCreateLoginLocation(ctx context.Context, location *models.LoginLocation) error {
	_, err := r.db.NewInsert().
		Model(location).
		On("CONFLICT (country_code, city) DO UPDATE").
		Set("login_count = login_locations.login_count + 1").
		Set("last_login_at = EXCLUDED.last_login_at").
		Exec(ctx)

	return err
}
