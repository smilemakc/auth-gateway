package service

import (
	"context"
	"sync"
	"time"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

type GeoLocationProvider interface {
	GetLocation(ip string) (*models.GeoLocation, error)
}

type GeoService struct {
	provider GeoLocationProvider
	cache    *geoCache
}

type geoCache struct {
	mu      sync.RWMutex
	entries map[string]geoCacheEntry
	maxSize int
	ttl     time.Duration
}

type geoCacheEntry struct {
	location  *models.GeoLocation
	expiresAt time.Time
}

func NewGeoService(apiKey string) *GeoService {
	return &GeoService{
		provider: utils.NewGeoIPService(apiKey),
		cache:    newGeoCache(10000, 24*time.Hour),
	}
}

func NewGeoServiceWithProvider(provider GeoLocationProvider) *GeoService {
	return &GeoService{
		provider: provider,
		cache:    newGeoCache(10000, 24*time.Hour),
	}
}

func newGeoCache(maxSize int, ttl time.Duration) *geoCache {
	cache := &geoCache{
		entries: make(map[string]geoCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
	go cache.startCleanupRoutine()
	return cache
}

func (c *geoCache) get(ip string) (*models.GeoLocation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[ip]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.location, true
}

func (c *geoCache) set(ip string, location *models.GeoLocation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[ip] = geoCacheEntry{
		location:  location,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *geoCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *geoCache) startCleanupRoutine() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		c.cleanup()
	}
}

func (c *geoCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

func (s *GeoService) GetLocation(ctx context.Context, ip string) *models.GeoLocation {
	if ip == "" {
		return nil
	}

	if cached, found := s.cache.get(ip); found {
		return cached
	}

	location, err := s.provider.GetLocation(ip)
	if err != nil {
		return nil
	}

	s.cache.set(ip, location)
	return location
}

func (s *GeoService) GetLocationAsync(ip string, callback func(*models.GeoLocation)) {
	go func() {
		location := s.GetLocation(context.Background(), ip)
		if callback != nil {
			callback(location)
		}
	}()
}
