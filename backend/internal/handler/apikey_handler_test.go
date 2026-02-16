package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// APIKeyHandler test helpers
// ---------------------------------------------------------------------------

func setupAPIKeyHandler() (*APIKeyHandler, *mockAPIKeyServicer) {
	svc := &mockAPIKeyServicer{}
	h := NewAPIKeyHandler(svc, testLogger())
	return h, svc
}

// ---------------------------------------------------------------------------
// Create Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_Create_ShouldReturn201_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.CreateFunc = func(uid uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error) {
		return &models.CreateAPIKeyResponse{
			APIKey:   &models.APIKey{ID: apiKeyID, UserID: uid, Name: req.Name},
			PlainKey: "agw_test_plain_key_12345",
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Create(c)
	})

	body := `{"name":"test-key","scopes":["users:read"]}`
	req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.CreateAPIKeyResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "agw_test_plain_key_12345", resp.PlainKey)
	assert.Equal(t, "test-key", resp.APIKey.Name)
}

func TestAPIKeyHandler_Create_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys", h.Create)

	body := `{"name":"test-key","scopes":["users:read"]}`
	req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_Create_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIKeyHandler_Create_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	svc.CreateFunc = func(uid uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Create(c)
	})

	body := `{"name":"test-key","scopes":["users:read"]}`
	req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// List Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_List_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	svc.ListFunc = func(uid uuid.UUID) ([]*models.APIKey, error) {
		return []*models.APIKey{
			{ID: uuid.New(), UserID: uid, Name: "key-1"},
			{ID: uuid.New(), UserID: uid, Name: "key-2"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ListAPIKeysResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.APIKeys, 2)
}

func TestAPIKeyHandler_List_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_List_ShouldReturn200Empty_WhenNoKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	svc.ListFunc = func(uid uuid.UUID) ([]*models.APIKey, error) {
		return []*models.APIKey{}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ListAPIKeysResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

func TestAPIKeyHandler_List_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	svc.ListFunc = func(uid uuid.UUID) ([]*models.APIKey, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Get Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_Get_ShouldReturn200_WhenKeyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.GetByIDFunc = func(uid, keyID uuid.UUID) (*models.APIKey, error) {
		return &models.APIKey{ID: keyID, UserID: uid, Name: "my-key"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys/"+apiKeyID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var apiKey models.APIKey
	err := json.Unmarshal(w.Body.Bytes(), &apiKey)
	require.NoError(t, err)
	assert.Equal(t, "my-key", apiKey.Name)
}

func TestAPIKeyHandler_Get_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api-keys/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_Get_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIKeyHandler_Get_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.GetByIDFunc = func(uid, keyID uuid.UUID) (*models.APIKey, error) {
		return nil, models.NewAppError(http.StatusNotFound, "API key not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api-keys/"+apiKeyID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Update Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_Update_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.UpdateFunc = func(uid, keyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error) {
		return &models.APIKey{ID: keyID, UserID: uid, Name: "updated-key"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Update(c)
	})

	body := `{"name":"updated-key"}`
	req := httptest.NewRequest(http.MethodPut, "/api-keys/"+apiKeyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var apiKey models.APIKey
	err := json.Unmarshal(w.Body.Bytes(), &apiKey)
	require.NoError(t, err)
	assert.Equal(t, "updated-key", apiKey.Name)
}

func TestAPIKeyHandler_Update_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/api-keys/:id", h.Update)

	body := `{"name":"updated-key"}`
	req := httptest.NewRequest(http.MethodPut, "/api-keys/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_Update_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Update(c)
	})

	body := `{"name":"updated-key"}`
	req := httptest.NewRequest(http.MethodPut, "/api-keys/not-a-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIKeyHandler_Update_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.UpdateFunc = func(uid, keyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Update(c)
	})

	body := `{"name":"updated-key"}`
	req := httptest.NewRequest(http.MethodPut, "/api-keys/"+apiKeyID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Revoke Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_Revoke_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.RevokeFunc = func(uid, keyID uuid.UUID, ip, userAgent string) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys/:id/revoke", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Revoke(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api-keys/"+apiKeyID.String()+"/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "API key revoked successfully", resp.Message)
}

func TestAPIKeyHandler_Revoke_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys/:id/revoke", h.Revoke)

	req := httptest.NewRequest(http.MethodPost, "/api-keys/"+uuid.New().String()+"/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_Revoke_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys/:id/revoke", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Revoke(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api-keys/not-a-uuid/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIKeyHandler_Revoke_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.RevokeFunc = func(uid, keyID uuid.UUID, ip, userAgent string) error {
		return models.NewAppError(http.StatusNotFound, "API key not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api-keys/:id/revoke", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Revoke(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api-keys/"+apiKeyID.String()+"/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Delete Tests
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_Delete_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.DeleteFunc = func(uid, keyID uuid.UUID, ip, userAgent string) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+apiKeyID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "API key deleted successfully", resp.Message)
}

func TestAPIKeyHandler_Delete_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/api-keys/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyHandler_Delete_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupAPIKeyHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIKeyHandler_Delete_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupAPIKeyHandler()

	userID := uuid.New()
	apiKeyID := uuid.New()
	svc.DeleteFunc = func(uid, keyID uuid.UUID, ip, userAgent string) error {
		return errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/api-keys/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+apiKeyID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
