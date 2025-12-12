package repository

import (
	"context"

	"github.com/smilemakc/auth-gateway/internal/models"

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
	var locations []models.LoginLocation
	query := `
		SELECT
			country_code,
			country_name,
			city,
			latitude,
			longitude,
			COUNT(*) as login_count,
			MAX(created_at) as last_login_at
		FROM audit_logs
		WHERE action = 'login'
		  AND status = 'success'
		  AND country_code IS NOT NULL
		  AND created_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY country_code, country_name, city, latitude, longitude
		ORDER BY login_count DESC
	`
	err := r.db.SelectContext(ctx, &locations, query, days)
	return locations, err
}

// GetTopCountries retrieves top countries by login count
func (r *GeoRepository) GetTopCountries(ctx context.Context, limit, days int) ([]models.CountryStats, error) {
	var stats []models.CountryStats
	query := `
		SELECT
			country_code,
			country_name,
			COUNT(*) as login_count,
			COUNT(DISTINCT user_id) as user_count
		FROM audit_logs
		WHERE action = 'login'
		  AND status = 'success'
		  AND country_code IS NOT NULL
		  AND created_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY country_code, country_name
		ORDER BY login_count DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &stats, query, days, limit)
	return stats, err
}

// GetTopCities retrieves top cities by login count
func (r *GeoRepository) GetTopCities(ctx context.Context, limit, days int) ([]models.CityStats, error) {
	var stats []models.CityStats
	query := `
		SELECT
			country_code,
			country_name,
			city,
			COUNT(*) as login_count,
			COUNT(DISTINCT user_id) as user_count
		FROM audit_logs
		WHERE action = 'login'
		  AND status = 'success'
		  AND city IS NOT NULL
		  AND created_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY country_code, country_name, city
		ORDER BY login_count DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &stats, query, days, limit)
	return stats, err
}

// UpdateOrCreateLoginLocation updates or creates login location aggregation
func (r *GeoRepository) UpdateOrCreateLoginLocation(ctx context.Context, location *models.LoginLocation) error {
	query := `
		INSERT INTO login_locations (country_code, country_name, city, latitude, longitude, login_count, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (country_code, city)
		DO UPDATE SET
			login_count = login_locations.login_count + 1,
			last_login_at = EXCLUDED.last_login_at
	`
	_, err := r.db.ExecContext(
		ctx, query,
		location.CountryCode, location.CountryName, location.City,
		location.Latitude, location.Longitude, location.LoginCount, location.LastLoginAt,
	)
	return err
}
