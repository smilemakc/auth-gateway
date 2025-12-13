package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/smilemakc/auth-gateway/internal/models"
	"net"
	"net/http"
	"strings"
	"time"
)

// GeoIPService provides IP geolocation functionality
type GeoIPService struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeoIPService creates a new geo IP service
func NewGeoIPService(apiKey string) *GeoIPService {
	return &GeoIPService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetLocation retrieves geographic location for an IP address
func (s *GeoIPService) GetLocation(ip string) (*models.GeoLocation, error) {
	// Check for private/local IPs
	if isPrivateIP(ip) {
		return &models.GeoLocation{
			CountryCode: "XX",
			CountryName: "Local Network",
			City:        "Local",
			Latitude:    0,
			Longitude:   0,
		}, nil
	}

	// If no API key, return unknown location
	if s.apiKey == "" {
		return s.getFallbackLocation(ip)
	}

	// Use ip-api.com (free tier)
	return s.getLocationFromIPAPI(ip)
}

// getLocationFromIPAPI uses ip-api.com to get location data
func (s *GeoIPService) getLocationFromIPAPI(ip string) (*models.GeoLocation, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,city,lat,lon", ip)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return s.getFallbackLocation(ip)
	}
	defer resp.Body.Close()

	var result struct {
		Status      string  `json:"status"`
		Message     string  `json:"message"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getFallbackLocation(ip)
	}

	if result.Status != "success" {
		return s.getFallbackLocation(ip)
	}

	return &models.GeoLocation{
		CountryCode: result.CountryCode,
		CountryName: result.Country,
		City:        result.City,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
	}, nil
}

// getFallbackLocation returns a default location for unknown IPs
func (s *GeoIPService) getFallbackLocation(ip string) (*models.GeoLocation, error) {
	return &models.GeoLocation{
		CountryCode: "XX",
		CountryName: "Unknown",
		City:        "Unknown",
		Latitude:    0,
		Longitude:   0,
	}, nil
}

// isPrivateIP checks if an IP address is private/local
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for loopback
	if ip.IsLoopback() {
		return true
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// ValidateIPOrCIDR validates if a string is a valid IP address or CIDR range
func ValidateIPOrCIDR(ipCIDR string) bool {
	// Check if it's a CIDR range
	if strings.Contains(ipCIDR, "/") {
		_, _, err := net.ParseCIDR(ipCIDR)
		return err == nil
	}

	// Check if it's a simple IP address
	ip := net.ParseIP(ipCIDR)
	return ip != nil
}

// IPMatchesCIDR checks if an IP address matches a CIDR range
func IPMatchesCIDR(ipStr, cidr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// If cidr doesn't contain /, treat it as a single IP
	if !strings.Contains(cidr, "/") {
		return ipStr == cidr, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	return ipNet.Contains(ip), nil
}
