package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParsePagination_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	page, pageSize := ParsePagination(c)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, pageSize)
}

func TestParsePagination_InvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?page=0&page_size=-1", nil)
	page, pageSize := ParsePagination(c)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, pageSize)
}

func TestParsePagination_ValidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?page=3&page_size=50", nil)
	page, pageSize := ParsePagination(c)
	assert.Equal(t, 3, page)
	assert.Equal(t, 50, pageSize)
}

func TestParsePagination_MaxPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?page=1&page_size=500", nil)
	page, pageSize := ParsePagination(c)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, pageSize)
}

func TestParsePagination_CustomDefaultPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	page, pageSize := ParsePagination(c, 50)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}

func TestParsePagination_CustomDefaultExceedsMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?page_size=200", nil)
	page, pageSize := ParsePagination(c, 50)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}
