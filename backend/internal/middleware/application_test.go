package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock implementations ---

type mockAppService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*models.Application, error)
}

func (m *mockAppService) GetByID(ctx context.Context, id uuid.UUID) (*models.Application, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

type mockAccessChecker struct {
	checkUserAccessFn func(ctx context.Context, userID, applicationID uuid.UUID) error
}

func (m *mockAccessChecker) CheckUserAccess(ctx context.Context, userID, applicationID uuid.UUID) error {
	if m.checkUserAccessFn != nil {
		return m.checkUserAccessFn(ctx, userID, applicationID)
	}
	return nil
}

func newTestLogger() *logger.Logger {
	return logger.New("test", logger.DebugLevel, false)
}

// --- ExtractApplicationID tests ---

func TestExtractApplicationID_ShouldSetAppIDInContext_WhenValidUUIDHeader(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())
	appID := uuid.New()

	r := gin.New()
	r.Use(mw.ExtractApplicationID())
	var capturedAppID *uuid.UUID
	r.GET("/test", func(c *gin.Context) {
		capturedAppID, _ = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Application-ID", appID.String())

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedAppID)
	assert.Equal(t, appID, *capturedAppID)
}

func TestExtractApplicationID_ShouldSetAppIDFromQueryParam_WhenNoHeader(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())
	appID := uuid.New()

	r := gin.New()
	r.Use(mw.ExtractApplicationID())
	var capturedAppID *uuid.UUID
	r.GET("/test", func(c *gin.Context) {
		capturedAppID, _ = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?app_id="+appID.String(), nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedAppID)
	assert.Equal(t, appID, *capturedAppID)
}

func TestExtractApplicationID_ShouldContinueWithoutAppID_WhenNoHeaderOrQuery(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	r := gin.New()
	r.Use(mw.ExtractApplicationID())
	var hasAppID bool
	r.GET("/test", func(c *gin.Context) {
		_, hasAppID = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, hasAppID)
}

func TestExtractApplicationID_ShouldContinueWithoutAppID_WhenInvalidUUID(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	r := gin.New()
	r.Use(mw.ExtractApplicationID())
	var hasAppID bool
	r.GET("/test", func(c *gin.Context) {
		_, hasAppID = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Application-ID", "not-a-valid-uuid")

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, hasAppID, "should continue without setting app ID for invalid UUID")
}

// --- RequireApplicationID tests ---

func TestRequireApplicationID_ShouldAllow_WhenValidUUIDHeader(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())
	appID := uuid.New()

	r := gin.New()
	r.Use(mw.RequireApplicationID())
	var capturedAppID *uuid.UUID
	r.GET("/test", func(c *gin.Context) {
		capturedAppID, _ = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Application-ID", appID.String())

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedAppID)
	assert.Equal(t, appID, *capturedAppID)
}

func TestRequireApplicationID_ShouldReturn400_WhenMissingHeader(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	r := gin.New()
	r.Use(mw.RequireApplicationID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "X-Application-ID header is required")
}

func TestRequireApplicationID_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	r := gin.New()
	r.Use(mw.RequireApplicationID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Application-ID", "invalid-uuid")

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID format")
}

// --- ValidateApplicationAccess tests ---

func TestValidateApplicationAccess_ShouldAllow_WhenAccessCheckerReturnsNoError(t *testing.T) {
	// Arrange
	checker := &mockAccessChecker{
		checkUserAccessFn: func(ctx context.Context, userID, applicationID uuid.UUID) error {
			return nil
		},
	}
	mw := NewApplicationMiddleware(nil, checker, newTestLogger())

	userID := uuid.New()
	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationAccess())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateApplicationAccess_ShouldReturn403_WhenAccessCheckFails(t *testing.T) {
	// Arrange
	checker := &mockAccessChecker{
		checkUserAccessFn: func(ctx context.Context, userID, applicationID uuid.UUID) error {
			return errors.New("access denied")
		},
	}
	mw := NewApplicationMiddleware(nil, checker, newTestLogger())

	userID := uuid.New()
	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationAccess())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestValidateApplicationAccess_ShouldSkip_WhenNoUserInContext(t *testing.T) {
	// Arrange
	checker := &mockAccessChecker{}
	mw := NewApplicationMiddleware(nil, checker, newTestLogger())

	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationAccess())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "should skip validation if no user")
}

func TestValidateApplicationAccess_ShouldSkip_WhenNoApplicationInContext(t *testing.T) {
	// Arrange
	checker := &mockAccessChecker{}
	mw := NewApplicationMiddleware(nil, checker, newTestLogger())

	userID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationAccess())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "should skip validation if no app ID")
}

func TestValidateApplicationAccess_ShouldSkip_WhenAccessCheckerIsNil(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	userID := uuid.New()
	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationAccess())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "should skip when accessChecker is nil")
}

// --- ValidateApplicationExists tests ---

func TestValidateApplicationExists_ShouldAllow_WhenApplicationExistsAndActive(t *testing.T) {
	// Arrange
	appID := uuid.New()
	appSvc := &mockAppService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "test-app",
				IsActive: true,
			}, nil
		},
	}
	mw := NewApplicationMiddleware(appSvc, nil, newTestLogger())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationExists())
	var capturedApp *models.Application
	r.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("application")
		if exists {
			capturedApp = val.(*models.Application)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedApp)
	assert.Equal(t, "test-app", capturedApp.Name)
}

func TestValidateApplicationExists_ShouldReturn404_WhenApplicationNotFound(t *testing.T) {
	// Arrange
	appSvc := &mockAppService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}
	mw := NewApplicationMiddleware(appSvc, nil, newTestLogger())

	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationExists())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Contains(t, body["message"], "Application not found")
}

func TestValidateApplicationExists_ShouldReturn403_WhenApplicationInactive(t *testing.T) {
	// Arrange
	appID := uuid.New()
	appSvc := &mockAppService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{
				ID:       appID,
				Name:     "test-app",
				IsActive: false,
			}, nil
		},
	}
	mw := NewApplicationMiddleware(appSvc, nil, newTestLogger())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationExists())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Application is inactive")
}

func TestValidateApplicationExists_ShouldReturn403_WhenAppServiceReturnsNilApp(t *testing.T) {
	// Arrange
	appSvc := &mockAppService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, nil
		},
	}
	mw := NewApplicationMiddleware(appSvc, nil, newTestLogger())

	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationExists())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Application is inactive")
}

func TestValidateApplicationExists_ShouldSkip_WhenNoAppIDInContext(t *testing.T) {
	// Arrange
	appSvc := &mockAppService{}
	mw := NewApplicationMiddleware(appSvc, nil, newTestLogger())

	r := gin.New()
	r.Use(mw.ValidateApplicationExists())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "should skip validation when no app ID")
}

func TestValidateApplicationExists_ShouldSkip_WhenAppServiceIsNil(t *testing.T) {
	// Arrange
	mw := NewApplicationMiddleware(nil, nil, newTestLogger())

	appID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	})
	r.Use(mw.ValidateApplicationExists())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "should skip validation when appService is nil")
}
