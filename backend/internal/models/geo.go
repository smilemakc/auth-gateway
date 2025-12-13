package models

import (
	"time"
)

// GeoLocation represents geographic location information
type GeoLocation struct {
	CountryCode string  `json:"country_code,omitempty" db:"country_code"` // ISO 3166-1 alpha-2
	CountryName string  `json:"country_name,omitempty" db:"country_name"`
	City        string  `json:"city,omitempty" db:"city"`
	Latitude    float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude   float64 `json:"longitude,omitempty" db:"longitude"`
}

// LoginLocation represents aggregated login location data
type LoginLocation struct {
	CountryCode string    `json:"country_code" db:"country_code"`
	CountryName string    `json:"country_name" db:"country_name"`
	City        string    `json:"city,omitempty" db:"city"`
	Latitude    float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude   float64   `json:"longitude,omitempty" db:"longitude"`
	LoginCount  int       `json:"login_count" db:"login_count"`
	LastLoginAt time.Time `json:"last_login_at" db:"last_login_at"`
}

// GeoDistributionResponse contains login distribution for map visualization
type GeoDistributionResponse struct {
	Locations []LoginLocation `json:"locations"`
	Total     int             `json:"total_logins"`
	Countries int             `json:"unique_countries"`
	Cities    int             `json:"unique_cities"`
}

// GeoStatsResponse contains geographic statistics
type GeoStatsResponse struct {
	TopCountries []CountryStats `json:"top_countries"`
	TopCities    []CityStats    `json:"top_cities"`
	RecentLogins []RecentLogin  `json:"recent_logins"`
}

// CountryStats contains login statistics by country
type CountryStats struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	LoginCount  int    `json:"login_count"`
	UserCount   int    `json:"user_count"`
}

// CityStats contains login statistics by city
type CityStats struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	City        string `json:"city"`
	LoginCount  int    `json:"login_count"`
	UserCount   int    `json:"user_count"`
}

// RecentLogin represents a recent login with geo data
type RecentLogin struct {
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	IPAddress  string      `json:"ip_address"`
	Location   GeoLocation `json:"location"`
	LoginTime  time.Time   `json:"login_time"`
	DeviceType string      `json:"device_type,omitempty"`
	Browser    string      `json:"browser,omitempty"`
	OS         string      `json:"os,omitempty"`
}

// IPGeolocationService interface for geo-location providers
type IPGeolocationService interface {
	GetLocation(ip string) (*GeoLocation, error)
}
