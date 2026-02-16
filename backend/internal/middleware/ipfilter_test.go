package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockIPFilterStore implements service.IPFilterStore for testing
type mockIPFilterStore struct {
	getActiveIPFiltersFn func(ctx context.Context) ([]models.IPFilter, error)
}

func (m *mockIPFilterStore) CreateIPFilter(ctx context.Context, filter *models.IPFilter) error {
	return nil
}
func (m *mockIPFilterStore) GetIPFilterByID(ctx context.Context, id uuid.UUID) (*models.IPFilter, error) {
	return nil, nil
}
func (m *mockIPFilterStore) ListIPFilters(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
	return nil, 0, nil
}
func (m *mockIPFilterStore) UpdateIPFilter(ctx context.Context, id uuid.UUID, reason string, isActive bool) error {
	return nil
}
func (m *mockIPFilterStore) DeleteIPFilter(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockIPFilterStore) GetActiveIPFilters(ctx context.Context) ([]models.IPFilter, error) {
	if m.getActiveIPFiltersFn != nil {
		return m.getActiveIPFiltersFn(ctx)
	}
	return nil, nil
}

func newTestIPFilterMiddleware(store *mockIPFilterStore) *IPFilterMiddleware {
	svc := service.NewIPFilterService(store)
	return NewIPFilterMiddleware(svc)
}

// --- CheckIPFilter tests ---

func TestCheckIPFilter_ShouldAllow_WhenNoFiltersExist(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckIPFilter_ShouldBlock_WhenIPIsBlacklisted(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "192.168.1.100/32",
					FilterType: "blacklist",
					Reason:     "Suspicious activity",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Access denied")
}

func TestCheckIPFilter_ShouldAllow_WhenIPIsNotBlacklisted(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "10.0.0.0/8",
					FilterType: "blacklist",
					Reason:     "Blocked range",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckIPFilter_ShouldBlock_WhenIPNotInWhitelist(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "10.0.0.0/8",
					FilterType: "whitelist",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Access denied")
}

func TestCheckIPFilter_ShouldAllow_WhenIPIsInWhitelist(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "192.168.1.0/24",
					FilterType: "whitelist",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckIPFilter_ShouldFailOpen_WhenErrorCheckingFilters(t *testing.T) {
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return nil, errors.New("database connection error")
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "should fail-open on error for availability")
}

func TestCheckIPFilter_ShouldBlock_WhenIPMatchesBothWhitelistAndBlacklist(t *testing.T) {
	// Blacklist takes precedence: even if IP is whitelisted, blacklist still blocks.
	// The whitelist pass only prevents "not in whitelist" denial; the blacklist
	// check runs independently afterward.
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "192.168.1.0/24",
					FilterType: "whitelist",
					IsActive:   true,
				},
				{
					IPCIDR:     "192.168.1.100/32",
					FilterType: "blacklist",
					Reason:     "Blacklist overrides whitelist",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "blacklisted IP should be blocked even if also whitelisted")
	assert.Contains(t, w.Body.String(), "Access denied")
}

func TestCheckIPFilter_ShouldAllow_WhenWhitelistedAndNotBlacklisted(t *testing.T) {
	// IP matches whitelist range but is NOT in blacklist -> allowed
	store := &mockIPFilterStore{
		getActiveIPFiltersFn: func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{
					IPCIDR:     "192.168.1.0/24",
					FilterType: "whitelist",
					IsActive:   true,
				},
				{
					IPCIDR:     "10.0.0.0/8",
					FilterType: "blacklist",
					Reason:     "Different range",
					IsActive:   true,
				},
			}, nil
		},
	}
	mw := newTestIPFilterMiddleware(store)

	r := gin.New()
	r.Use(mw.CheckIPFilter())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "whitelisted IP not in blacklist should be allowed")
}
