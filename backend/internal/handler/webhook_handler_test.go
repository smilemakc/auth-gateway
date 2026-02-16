package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type webhookTestFixture struct {
	handler    *WebhookHandler
	webhookSvc *mockWebhookServicer
}

func setupWebhookTestFixture() *webhookTestFixture {
	svc := &mockWebhookServicer{}
	h := NewWebhookHandler(svc, testLogger())
	return &webhookTestFixture{handler: h, webhookSvc: svc}
}

func sampleWebhook() *models.Webhook {
	return &models.Webhook{
		ID:        uuid.New(),
		Name:      "Test Webhook",
		URL:       "https://example.com/hook",
		Events:    json.RawMessage(`["user.created"]`),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ===========================================================================
// ListWebhooks
// ===========================================================================

func TestWebhookHandler_ListWebhooks_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	fix.webhookSvc.ListWebhooksFunc = func(page, perPage int) (*models.WebhookListResponse, error) {
		return &models.WebhookListResponse{
			Webhooks:   []models.WebhookWithCreator{{Webhook: *sampleWebhook()}},
			Total:      1,
			Page:       page,
			PageSize:   perPage,
			TotalPages: 1,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks", fix.handler.ListWebhooks)

	req := httptest.NewRequest(http.MethodGet, "/webhooks?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.WebhookListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Webhooks, 1)
}

func TestWebhookHandler_ListWebhooks_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	fix.webhookSvc.ListWebhooksFunc = func(page, perPage int) (*models.WebhookListResponse, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks", fix.handler.ListWebhooks)

	req := httptest.NewRequest(http.MethodGet, "/webhooks?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebhookHandler_ListWebhooks_ShouldUseDefaultPagination_WhenNoParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	var capturedPage, capturedPageSize int
	fix.webhookSvc.ListWebhooksFunc = func(page, perPage int) (*models.WebhookListResponse, error) {
		capturedPage = page
		capturedPageSize = perPage
		return &models.WebhookListResponse{
			Webhooks:   []models.WebhookWithCreator{},
			Total:      0,
			Page:       page,
			PageSize:   perPage,
			TotalPages: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks", fix.handler.ListWebhooks)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 20, capturedPageSize)
}

func TestWebhookHandler_ListWebhooks_ShouldListByApp_WhenAppIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	appID := uuid.New()
	wh := sampleWebhook()
	wh.ApplicationID = &appID

	fix.webhookSvc.ListWebhooksByAppFunc = func(id uuid.UUID) ([]*models.Webhook, error) {
		return []*models.Webhook{wh}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks", func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, appID)
		fix.handler.ListWebhooks(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.WebhookListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Webhooks, 1)
}

// ===========================================================================
// GetWebhook
// ===========================================================================

func TestWebhookHandler_GetWebhook_ShouldReturn200_WhenFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	wh := sampleWebhook()
	fix.webhookSvc.GetWebhookFunc = func(id uuid.UUID) (*models.Webhook, error) {
		return wh, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id", fix.handler.GetWebhook)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+wh.ID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), wh.ID.String())
}

func TestWebhookHandler_GetWebhook_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	fix.webhookSvc.GetWebhookFunc = func(id uuid.UUID) (*models.Webhook, error) {
		return nil, fmt.Errorf("not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id", fix.handler.GetWebhook)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWebhookHandler_GetWebhook_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id", fix.handler.GetWebhook)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// CreateWebhook
// ===========================================================================

func TestWebhookHandler_CreateWebhook_ShouldReturn201_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	wh := sampleWebhook()

	fix.webhookSvc.CreateWebhookFunc = func(req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error) {
		assert.Equal(t, userID, createdBy)
		return wh, "secret_123", nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.CreateWebhook(c)
	})

	body := `{"name":"Test Hook","url":"https://example.com/hook","events":["user.created"]}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "secret_123")
	assert.Contains(t, w.Body.String(), "webhook")
}

func TestWebhookHandler_CreateWebhook_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks", fix.handler.CreateWebhook)

	body := `{"name":"Test Hook","url":"https://example.com/hook","events":["user.created"]}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebhookHandler_CreateWebhook_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing url", `{"name":"Test","events":["user.created"]}`},
		{"missing events", `{"name":"Test","url":"https://example.com/hook"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/webhooks", func(c *gin.Context) {
				c.Set(utils.UserIDKey, userID)
				fix.handler.CreateWebhook(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestWebhookHandler_CreateWebhook_ShouldReturn400_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	fix.webhookSvc.CreateWebhookFunc = func(req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error) {
		return nil, "", fmt.Errorf("invalid webhook configuration")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.CreateWebhook(c)
	})

	body := `{"name":"Test Hook","url":"https://example.com/hook","events":["user.created"]}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// UpdateWebhook
// ===========================================================================

func TestWebhookHandler_UpdateWebhook_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	webhookID := uuid.New()

	fix.webhookSvc.UpdateWebhookFunc = func(id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error {
		assert.Equal(t, webhookID, id)
		assert.Equal(t, userID, updatedBy)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.UpdateWebhook(c)
	})

	body := `{"url":"https://example.com/hook2"}`
	req := httptest.NewRequest(http.MethodPut, "/webhooks/"+webhookID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Webhook updated successfully", resp.Message)
}

func TestWebhookHandler_UpdateWebhook_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/webhooks/:id", fix.handler.UpdateWebhook)

	body := `{"url":"https://example.com/hook2"}`
	req := httptest.NewRequest(http.MethodPut, "/webhooks/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebhookHandler_UpdateWebhook_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.UpdateWebhook(c)
	})

	body := `{"url":"https://example.com/hook2"}`
	req := httptest.NewRequest(http.MethodPut, "/webhooks/bad-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_UpdateWebhook_ShouldReturn400_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	fix.webhookSvc.UpdateWebhookFunc = func(id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error {
		return fmt.Errorf("webhook not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.UpdateWebhook(c)
	})

	body := `{"url":"https://example.com/hook2"}`
	req := httptest.NewRequest(http.MethodPut, "/webhooks/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// DeleteWebhook
// ===========================================================================

func TestWebhookHandler_DeleteWebhook_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	webhookID := uuid.New()

	fix.webhookSvc.DeleteWebhookFunc = func(id uuid.UUID, deletedBy uuid.UUID) error {
		assert.Equal(t, webhookID, id)
		assert.Equal(t, userID, deletedBy)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.DeleteWebhook(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+webhookID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Webhook deleted successfully", resp.Message)
}

func TestWebhookHandler_DeleteWebhook_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/webhooks/:id", fix.handler.DeleteWebhook)

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebhookHandler_DeleteWebhook_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()
	fix.webhookSvc.DeleteWebhookFunc = func(id uuid.UUID, deletedBy uuid.UUID) error {
		return fmt.Errorf("webhook not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.DeleteWebhook(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWebhookHandler_DeleteWebhook_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	userID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/webhooks/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.DeleteWebhook(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/not-valid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// TestWebhook
// ===========================================================================

func TestWebhookHandler_TestWebhook_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	webhookID := uuid.New()
	fix.webhookSvc.TestWebhookFunc = func(id uuid.UUID, req *models.TestWebhookRequest) error {
		assert.Equal(t, webhookID, id)
		assert.Equal(t, "user.created", req.EventType)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks/:id/test", fix.handler.TestWebhook)

	body := `{"event_type":"user.created"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/"+webhookID.String()+"/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Test webhook sent", resp.Message)
}

func TestWebhookHandler_TestWebhook_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing event_type", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/webhooks/:id/test", fix.handler.TestWebhook)

			req := httptest.NewRequest(http.MethodPost, "/webhooks/"+uuid.New().String()+"/test", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestWebhookHandler_TestWebhook_ShouldReturn404_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	fix.webhookSvc.TestWebhookFunc = func(id uuid.UUID, req *models.TestWebhookRequest) error {
		return fmt.Errorf("webhook not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks/:id/test", fix.handler.TestWebhook)

	body := `{"event_type":"user.created"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/"+uuid.New().String()+"/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWebhookHandler_TestWebhook_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/webhooks/:id/test", fix.handler.TestWebhook)

	body := `{"event_type":"user.created"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/bad-id/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// ListWebhookDeliveries
// ===========================================================================

func TestWebhookHandler_ListWebhookDeliveries_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	webhookID := uuid.New()
	fix.webhookSvc.ListWebhookDeliveriesFunc = func(id uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error) {
		assert.Equal(t, webhookID, id)
		return &models.WebhookDeliveryListResponse{
			Deliveries: []models.WebhookDelivery{},
			Total:      0,
			Page:       page,
			PageSize:   perPage,
			TotalPages: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id/deliveries", fix.handler.ListWebhookDeliveries)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String()+"/deliveries?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.WebhookDeliveryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

func TestWebhookHandler_ListWebhookDeliveries_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	fix.webhookSvc.ListWebhookDeliveriesFunc = func(id uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id/deliveries", fix.handler.ListWebhookDeliveries)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+uuid.New().String()+"/deliveries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebhookHandler_ListWebhookDeliveries_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/:id/deliveries", fix.handler.ListWebhookDeliveries)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/bad-id/deliveries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// GetAvailableEvents
// ===========================================================================

func TestWebhookHandler_GetAvailableEvents_ShouldReturn200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/events", fix.handler.GetAvailableEvents)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/events", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "events")
}

func TestWebhookHandler_GetAvailableEvents_ShouldReturnEmptyList_WhenNoEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupWebhookTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/webhooks/events", fix.handler.GetAvailableEvents)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/events", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["events"])
}
