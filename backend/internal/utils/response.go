package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// RespondWithError sends an appropriate JSON error response.
// If the error is an *models.AppError, it uses the error's status code.
// Otherwise, it responds with 500 Internal Server Error.
func RespondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*models.AppError); ok {
		c.JSON(appErr.Code, models.NewErrorResponse(appErr))
	} else {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
	}
}
