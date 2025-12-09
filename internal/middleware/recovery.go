package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// Recovery middleware recovers from panics and logs them
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Error("Panic recovered", map[string]interface{}{
					"error":  fmt.Sprintf("%v", err),
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				})

				// Return internal server error
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
				c.Abort()
			}
		}()

		c.Next()
	}
}
