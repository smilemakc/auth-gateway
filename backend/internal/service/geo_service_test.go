package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGeoService_GetLocation(t *testing.T) {
	mockProvider := &mockGeoLocationProvider{}
	svc := NewGeoServiceWithProvider(mockProvider)

	ip := "1.2.3.4"
	expectedLoc := &models.GeoLocation{
		CountryCode: "US",
		City:        "New York",
	}

	t.Run("CacheMiss_ProviderHit", func(t *testing.T) {
		mockProvider.GetLocationFunc = func(ip string) (*models.GeoLocation, error) {
			return expectedLoc, nil
		}

		loc := svc.GetLocation(context.Background(), ip)
		assert.Equal(t, expectedLoc, loc)
	})

	t.Run("CacheHit", func(t *testing.T) {
		// Reset provider mock to ensure it's not called
		mockProvider.GetLocationFunc = func(ip string) (*models.GeoLocation, error) {
			t.Fatal("Provider should not be called")
			return nil, nil
		}

		loc := svc.GetLocation(context.Background(), ip)
		assert.Equal(t, expectedLoc, loc)
	})

	t.Run("ProviderError", func(t *testing.T) {
		mockProvider.GetLocationFunc = func(ip string) (*models.GeoLocation, error) {
			return nil, errors.New("provider error")
		}

		loc := svc.GetLocation(context.Background(), "8.8.8.8")
		assert.Nil(t, loc)
	})

	t.Run("EmptyIP", func(t *testing.T) {
		loc := svc.GetLocation(context.Background(), "")
		assert.Nil(t, loc)
	})
}

func TestGeoService_GetLocationAsync(t *testing.T) {
	mockProvider := &mockGeoLocationProvider{}
	svc := NewGeoServiceWithProvider(mockProvider)

	ip := "1.2.3.4"
	expectedLoc := &models.GeoLocation{CountryCode: "US"}

	mockProvider.GetLocationFunc = func(ip string) (*models.GeoLocation, error) {
		return expectedLoc, nil
	}

	done := make(chan bool)
	callback := func(loc *models.GeoLocation) {
		assert.Equal(t, expectedLoc, loc)
		done <- true
	}

	svc.GetLocationAsync(ip, callback)

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for async callback")
	}
}
