package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseUUIDParam_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}

	id, ok := ParseUUIDParam(c, "id")
	assert.True(t, ok)
	assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), id)
	assert.Equal(t, 200, w.Code)
}

func TestParseUUIDParam_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	id, ok := ParseUUIDParam(c, "id")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, id)
	assert.Equal(t, 400, w.Code)
	assert.True(t, c.IsAborted())
}

func TestParseUUIDParam_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: ""}}

	id, ok := ParseUUIDParam(c, "id")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, id)
	assert.Equal(t, 400, w.Code)
}

func TestParseUUIDQuery_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/?user_id=550e8400-e29b-41d4-a716-446655440000", nil)

	id, ok := ParseUUIDQuery(c, "user_id")
	assert.True(t, ok)
	assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), id)
}

func TestParseUUIDQuery_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	id, ok := ParseUUIDQuery(c, "user_id")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, id)
	assert.Equal(t, 200, w.Code) // should NOT set error for empty query param
}

func TestParseUUIDQuery_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/?user_id=bad", nil)

	id, ok := ParseUUIDQuery(c, "user_id")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, id)
	assert.Equal(t, 400, w.Code)
}
