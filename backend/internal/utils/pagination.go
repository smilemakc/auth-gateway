package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ParsePagination extracts page and pageSize from query parameters with validation.
// Optional defaultPageSize overrides the default of 20.
func ParsePagination(c *gin.Context, defaultPageSize ...int) (page, pageSize int) {
	defPS := DefaultPageSize
	if len(defaultPageSize) > 0 && defaultPageSize[0] > 0 {
		defPS = defaultPageSize[0]
	}
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defPS)))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		pageSize = defPS
	}
	return
}
