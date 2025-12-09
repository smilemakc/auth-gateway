package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *repository.Database
	redis *service.RedisService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *repository.Database, redis *service.RedisService) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Health checks the health of the service and its dependencies
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /auth/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:   "healthy",
		Services: make(map[string]string),
	}

	// Check database
	if err := h.db.Health(); err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "unhealthy: " + err.Error()
	} else {
		response.Services["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["redis"] = "unhealthy: " + err.Error()
	} else {
		response.Services["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Readiness checks if the service is ready to handle requests
// @Summary Readiness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /auth/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	// Simple readiness check
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Liveness checks if the service is alive
// @Summary Liveness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
