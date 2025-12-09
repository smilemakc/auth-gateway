package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// Logger middleware logs HTTP requests
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := utils.GetClientIP(c)

		// Log request
		fields := map[string]interface{}{
			"method":      method,
			"path":        path,
			"status":      statusCode,
			"latency_ms":  latency.Milliseconds(),
			"ip":          clientIP,
			"user_agent":  c.Request.UserAgent(),
		}

		// Add user ID if available
		if userID, exists := utils.GetUserIDFromContext(c); exists {
			fields["user_id"] = userID.String()
		}

		// Log based on status code
		if statusCode >= 500 {
			log.Error("HTTP request failed", fields)
		} else if statusCode >= 400 {
			log.Warn("HTTP request error", fields)
		} else {
			log.Info("HTTP request", fields)
		}
	}
}
