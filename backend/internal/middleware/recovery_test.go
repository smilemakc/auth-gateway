package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_ShouldReturn500_WhenHandlerPanics(t *testing.T) {
	// Arrange
	log := logger.New("test", logger.ErrorLevel, false)

	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/test", func(c *gin.Context) {
		panic("something went terribly wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestRecovery_ShouldReturn200_WhenHandlerDoesNotPanic(t *testing.T) {
	// Arrange
	log := logger.New("test", logger.ErrorLevel, false)

	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRecovery_ShouldRecover_WhenPanicWithErrorValue(t *testing.T) {
	// Arrange
	log := logger.New("test", logger.ErrorLevel, false)

	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/test", func(c *gin.Context) {
		panic(42) // panic with non-string value
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestRecovery_ShouldRecover_WhenPanicWithNilValue(t *testing.T) {
	// Arrange
	log := logger.New("test", logger.ErrorLevel, false)

	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/test", func(c *gin.Context) {
		var p *int
		_ = *p // nil pointer dereference
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_ShouldAbortRequest_WhenPanicOccurs(t *testing.T) {
	// Arrange: verify that a subsequent middleware handler is not reached after panic recovery
	log := logger.New("test", logger.ErrorLevel, false)
	handlerReached := false

	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/test", func(c *gin.Context) {
		panic("boom")
	})
	// Register a second route to confirm abort behavior
	r.GET("/after", func(c *gin.Context) {
		handlerReached = true
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, handlerReached, "handler after panic should not be reached")
}
