package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoIPService_GetLocation(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path or params if needed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success",
			"country":     "United States",
			"countryCode": "US",
			"city":        "Ashburn",
			"lat":         39.0438,
			"lon":         -77.4874,
		})
	}))
	defer server.Close()

	// Since NewGeoIPService sets a default generic client,
	// we need to create a service that uses our mock server logic.
	// But `getLocationFromIPAPI` uses a hardcoded URL "http://ip-api.com...".
	// This makes it hard to test without refactoring or using a specific transport.
	//
	// Ideally we refactor `getLocationFromIPAPI` to use a configurable base URL,
	// OR we rely on `http.Transport` to intercept.
	//
	// However, looking at `geo.go`, `getLocationFromIPAPI` constructs the URL inside.
	// Let's test `isPrivateIP` and validation logic first.
	//
	// To test GetLocation with mock, we would need to mock the `http.Client`.
	// The struct `GeoIPService` has `httpClient *http.Client`.
	// Since we can't easily change the URL in the code without changing the code,
	// checking `GetLocation` for private IPs is straightforward.

	service := NewGeoIPService("test-key")

	t.Run("Private IP", func(t *testing.T) {
		loc, err := service.GetLocation("127.0.0.1")
		assert.NoError(t, err)
		assert.Equal(t, "Local Network", loc.CountryName)
	})

	t.Run("Fallback on empty API Key (for non-private)", func(t *testing.T) {
		s := NewGeoIPService("")
		loc, err := s.GetLocation("8.8.8.8")
		assert.NoError(t, err)
		assert.Equal(t, "Unknown", loc.CountryName)
	})
}

func TestValidateIPOrCIDR(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.0/8", true},
		{"invalid", false},
		{"256.256.256.256", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ValidateIPOrCIDR(tt.input))
		})
	}
}

func TestIPMatchesCIDR(t *testing.T) {
	tests := []struct {
		ip      string
		cidr    string
		match   bool
		wantErr bool
	}{
		{"192.168.1.5", "192.168.1.0/24", true, false},
		{"10.0.0.1", "192.168.1.0/24", false, false},
		{"192.168.1.5", "192.168.1.5", true, false}, // Single IP match
		{"invalid", "10.0.0.0/8", false, true},
		{"10.0.0.1", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.ip+"_"+tt.cidr, func(t *testing.T) {
			match, err := IPMatchesCIDR(tt.ip, tt.cidr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.match, match)
			}
		})
	}
}
