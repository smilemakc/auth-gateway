package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/metrics"
)

// MetricsMiddleware collects HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		metrics.RecordHTTPRequest(method, endpoint, statusCode, duration)
	}
}

// MetricsErrorMiddleware records error metrics
func MetricsErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			severity := "warning"
			if statusCode >= 500 {
				severity = "error"
			}

			errorType := "http"
			if statusCode == 401 || statusCode == 403 {
				errorType = "auth"
			} else if statusCode >= 500 {
				errorType = "server"
			}

			metrics.RecordError(errorType, severity)
		}
	}
}

// MetricsDBMiddleware records database operation metrics (to be used in repository layer)
func MetricsDBMiddleware(operation string, duration time.Duration, err error) {
	metrics.RecordDBQuery(operation, duration, err)
}

// MetricsRedisMiddleware records Redis operation metrics (to be used in service layer)
func MetricsRedisMiddleware(operation string, duration time.Duration, err error) {
	metrics.RecordRedisOperation(operation, duration, err)
}

// Helper function to extract status code from context
func getStatusCode(c *gin.Context) int {
	status, exists := c.Get("status_code")
	if !exists {
		return c.Writer.Status()
	}
	if code, ok := status.(int); ok {
		return code
	}
	if codeStr, ok := status.(string); ok {
		if code, err := strconv.Atoi(codeStr); err == nil {
			return code
		}
	}
	return c.Writer.Status()
}
