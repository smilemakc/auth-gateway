package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// PrometheusMetrics exposes Prometheus metrics endpoint
// @Summary Prometheus Metrics
// @Description Exposes Prometheus metrics in standard format
// @Tags Metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
