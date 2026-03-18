package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type bulkTestFixture struct {
	handler *BulkHandler
	bulkSvc *mockBulkServicer
}

func setupBulkTestFixture() *bulkTestFixture {
	svc := &mockBulkServicer{}
	h := NewBulkHandler(svc, testLogger())
	return &bulkTestFixture{handler: h, bulkSvc: svc}
}

func sampleBulkResult() *models.BulkOperationResult {
	return &models.BulkOperationResult{
		Total:   1,
		Success: 1,
		Failed:  0,
	}
}

// ===========================================================================
// BulkCreateUsers
// ===========================================================================

func TestBulkHandler_BulkCreateUsers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkCreateUsersFunc = func(req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error) {
		assert.Len(t, req.Users, 1)
		return sampleBulkResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-create", fix.handler.BulkCreateUsers)

	body := `{"users":[{"email":"a@test.com","username":"user_a","password":"Test123!"}]}`
	req := httptest.NewRequest(http.MethodPost, "/bulk-create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.BulkOperationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Success)
}

func TestBulkHandler_BulkCreateUsers_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing users", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/bulk-create", fix.handler.BulkCreateUsers)

			req := httptest.NewRequest(http.MethodPost, "/bulk-create", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestBulkHandler_BulkCreateUsers_ShouldReturn400_WhenExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	// Build 101 users
	users := make([]map[string]interface{}, 101)
	for i := range users {
		users[i] = map[string]interface{}{
			"email":    fmt.Sprintf("user%d@test.com", i),
			"username": fmt.Sprintf("user_%d", i),
			"password": "Test123!",
		}
	}
	bodyBytes, _ := json.Marshal(map[string]interface{}{"users": users})

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-create", fix.handler.BulkCreateUsers)

	req := httptest.NewRequest(http.MethodPost, "/bulk-create", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestBulkHandler_BulkCreateUsers_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkCreateUsersFunc = func(req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-create", fix.handler.BulkCreateUsers)

	body := `{"users":[{"email":"a@test.com","username":"user_a","password":"Test123!"}]}`
	req := httptest.NewRequest(http.MethodPost, "/bulk-create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===========================================================================
// BulkUpdateUsers
// ===========================================================================

func TestBulkHandler_BulkUpdateUsers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	userID := uuid.New()
	fix.bulkSvc.BulkUpdateUsersFunc = func(req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error) {
		assert.Len(t, req.Users, 1)
		return sampleBulkResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/bulk-update", fix.handler.BulkUpdateUsers)

	body := fmt.Sprintf(`{"users":[{"id":"%s","is_active":true}]}`, userID.String())
	req := httptest.NewRequest(http.MethodPut, "/bulk-update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.BulkOperationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Success)
}

func TestBulkHandler_BulkUpdateUsers_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing users", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.PUT("/bulk-update", fix.handler.BulkUpdateUsers)

			req := httptest.NewRequest(http.MethodPut, "/bulk-update", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestBulkHandler_BulkUpdateUsers_ShouldReturn400_WhenExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	users := make([]map[string]interface{}, 101)
	for i := range users {
		users[i] = map[string]interface{}{
			"id":        uuid.New().String(),
			"is_active": true,
		}
	}
	bodyBytes, _ := json.Marshal(map[string]interface{}{"users": users})

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/bulk-update", fix.handler.BulkUpdateUsers)

	req := httptest.NewRequest(http.MethodPut, "/bulk-update", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestBulkHandler_BulkUpdateUsers_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkUpdateUsersFunc = func(req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/bulk-update", fix.handler.BulkUpdateUsers)

	body := fmt.Sprintf(`{"users":[{"id":"%s","is_active":true}]}`, uuid.New().String())
	req := httptest.NewRequest(http.MethodPut, "/bulk-update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===========================================================================
// BulkDeleteUsers
// ===========================================================================

func TestBulkHandler_BulkDeleteUsers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkDeleteUsersFunc = func(req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error) {
		assert.Len(t, req.UserIDs, 1)
		return sampleBulkResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-delete", fix.handler.BulkDeleteUsers)

	body := fmt.Sprintf(`{"user_ids":["%s"]}`, uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.BulkOperationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Success)
}

func TestBulkHandler_BulkDeleteUsers_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing user_ids", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/bulk-delete", fix.handler.BulkDeleteUsers)

			req := httptest.NewRequest(http.MethodPost, "/bulk-delete", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestBulkHandler_BulkDeleteUsers_ShouldReturn400_WhenExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf(`"%s"`, uuid.New().String())
	}
	body := fmt.Sprintf(`{"user_ids":[%s]}`, strings.Join(ids, ","))

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-delete", fix.handler.BulkDeleteUsers)

	req := httptest.NewRequest(http.MethodPost, "/bulk-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestBulkHandler_BulkDeleteUsers_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkDeleteUsersFunc = func(req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-delete", fix.handler.BulkDeleteUsers)

	body := fmt.Sprintf(`{"user_ids":["%s"]}`, uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===========================================================================
// BulkAssignRoles
// ===========================================================================

func TestBulkHandler_BulkAssignRoles_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	userID := uuid.New()
	roleID := uuid.New()
	targetUserID := uuid.New()

	fix.bulkSvc.BulkAssignRolesFunc = func(req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error) {
		assert.Equal(t, userID, assignedBy)
		assert.Len(t, req.UserIDs, 1)
		assert.Len(t, req.RoleIDs, 1)
		return sampleBulkResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-assign", func(c *gin.Context) {
		c.Set("user_id", userID.String())
		fix.handler.BulkAssignRoles(c)
	})

	body := fmt.Sprintf(`{"user_ids":["%s"],"role_ids":["%s"]}`, targetUserID.String(), roleID.String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.BulkOperationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Success)
}

func TestBulkHandler_BulkAssignRoles_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-assign", fix.handler.BulkAssignRoles)

	body := fmt.Sprintf(`{"user_ids":["%s"],"role_ids":["%s"]}`, uuid.New().String(), uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBulkHandler_BulkAssignRoles_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing user_ids", fmt.Sprintf(`{"role_ids":["%s"]}`, uuid.New().String())},
		{"missing role_ids", fmt.Sprintf(`{"user_ids":["%s"]}`, uuid.New().String())},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/bulk-assign", func(c *gin.Context) {
				c.Set("user_id", uuid.New().String())
				fix.handler.BulkAssignRoles(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestBulkHandler_BulkAssignRoles_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	fix.bulkSvc.BulkAssignRolesFunc = func(req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error) {
		return nil, fmt.Errorf("role assignment failed")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-assign", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		fix.handler.BulkAssignRoles(c)
	})

	body := fmt.Sprintf(`{"user_ids":["%s"],"role_ids":["%s"]}`, uuid.New().String(), uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBulkHandler_BulkAssignRoles_ShouldReturn400_WhenExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf(`"%s"`, uuid.New().String())
	}
	body := fmt.Sprintf(`{"user_ids":[%s],"role_ids":["%s"]}`, strings.Join(ids, ","), uuid.New().String())

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-assign", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		fix.handler.BulkAssignRoles(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestBulkHandler_BulkAssignRoles_ShouldHandleUUIDUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupBulkTestFixture()

	userID := uuid.New()

	fix.bulkSvc.BulkAssignRolesFunc = func(req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error) {
		assert.Equal(t, userID, assignedBy)
		return sampleBulkResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/bulk-assign", func(c *gin.Context) {
		c.Set("user_id", userID) // Set as uuid.UUID, not string
		fix.handler.BulkAssignRoles(c)
	})

	body := fmt.Sprintf(`{"user_ids":["%s"],"role_ids":["%s"]}`, uuid.New().String(), uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/bulk-assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
