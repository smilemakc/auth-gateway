package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// ParseUUIDParam extracts and validates a UUID path parameter.
// On error, responds with 400 and aborts the request.
// Returns uuid.Nil and false if parsing fails.
func ParseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, fmt.Sprintf("Invalid %s format", paramName)),
		))
		c.Abort()
		return uuid.Nil, false
	}
	return id, true
}

// ParseUUIDQuery extracts and validates a UUID from query parameters.
// Returns uuid.Nil and false if the parameter is empty or invalid.
func ParseUUIDQuery(c *gin.Context, paramName string) (uuid.UUID, bool) {
	val := c.Query(paramName)
	if val == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, fmt.Sprintf("Invalid %s format", paramName)),
		))
		c.Abort()
		return uuid.Nil, false
	}
	return id, true
}
