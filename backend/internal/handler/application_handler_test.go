package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func setupApplicationHandler() (*ApplicationHandler, *mockApplicationServicer) {
	svc := &mockApplicationServicer{}
	h := NewApplicationHandler(svc, testLogger())
	return h, svc
}

func newTestApplication() *models.Application {
	now := time.Now()
	ownerID := uuid.New()
	return &models.Application{
		ID:          uuid.New(),
		Name:        "test-app",
		DisplayName: "Test Application",
		Description: "A test application",
		IsActive:    true,
		OwnerID:     &ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ---------------------------------------------------------------------------
// CreateApplication Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_CreateApplication_ShouldReturn201_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	app := newTestApplication()
	secret := "app_testsecret123456789"
	userID := uuid.New()

	svc.CreateApplicationFunc = func(req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
		assert.Equal(t, "my-app", req.Name)
		assert.Equal(t, "My App", req.DisplayName)
		assert.NotNil(t, ownerID)
		assert.Equal(t, userID, *ownerID)
		return app, secret, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.CreateApplication(c)
	})

	body := `{"name":"my-app","display_name":"My App"}`
	req := httptest.NewRequest(http.MethodPost, "/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["application"])
	assert.Equal(t, secret, resp["secret"])
	assert.Equal(t, "Store this secret securely. It will not be shown again.", resp["warning"])
}

func TestApplicationHandler_CreateApplication_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	userID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.CreateApplication(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/applications", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApplicationHandler_CreateApplication_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications", h.CreateApplication)

	body := `{"name":"my-app","display_name":"My App"}`
	req := httptest.NewRequest(http.MethodPost, "/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApplicationHandler_CreateApplication_ShouldReturn400_WhenInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	userID := uuid.New()

	svc.CreateApplicationFunc = func(req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
		return nil, "", service.ErrInvalidApplicationName
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.CreateApplication(c)
	})

	body := `{"name":"my-app","display_name":"My App"}`
	req := httptest.NewRequest(http.MethodPost, "/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application name")
}

func TestApplicationHandler_CreateApplication_ShouldReturn409_WhenNameExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	userID := uuid.New()

	svc.CreateApplicationFunc = func(req *models.CreateApplicationRequest, ownerID *uuid.UUID) (*models.Application, string, error) {
		return nil, "", service.ErrApplicationNameExists
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.CreateApplication(c)
	})

	body := `{"name":"existing-app","display_name":"Existing App"}`
	req := httptest.NewRequest(http.MethodPost, "/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already exists")
}

// ---------------------------------------------------------------------------
// GetApplication Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_GetApplication_ShouldReturn200_WhenApplicationExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	app := newTestApplication()

	svc.GetByIDFunc = func(id uuid.UUID) (*models.Application, error) {
		assert.Equal(t, app.ID, id)
		return app, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications/:id", h.GetApplication)

	req := httptest.NewRequest(http.MethodGet, "/applications/"+app.ID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Application
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, app.ID, resp.ID)
	assert.Equal(t, app.Name, resp.Name)
}

func TestApplicationHandler_GetApplication_ShouldReturn400_WhenInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications/:id", h.GetApplication)

	req := httptest.NewRequest(http.MethodGet, "/applications/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestApplicationHandler_GetApplication_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.GetByIDFunc = func(id uuid.UUID) (*models.Application, error) {
		return nil, service.ErrApplicationNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications/:id", h.GetApplication)

	req := httptest.NewRequest(http.MethodGet, "/applications/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Application not found")
}

func TestApplicationHandler_GetApplication_ShouldReturn500_WhenServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.GetByIDFunc = func(id uuid.UUID) (*models.Application, error) {
		return nil, errors.New("database connection error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications/:id", h.GetApplication)

	req := httptest.NewRequest(http.MethodGet, "/applications/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// ListApplications Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_ListApplications_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	listResp := &models.ApplicationListResponse{
		Applications: []models.Application{*newTestApplication()},
		Total:        1,
		Page:         1,
		PageSize:     20,
		TotalPages:   1,
	}

	svc.ListApplicationsFunc = func(page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, perPage)
		assert.Nil(t, isActive)
		return listResp, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications", h.ListApplications)

	req := httptest.NewRequest(http.MethodGet, "/applications?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ApplicationListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Applications, 1)
}

func TestApplicationHandler_ListApplications_ShouldReturn200_WhenFilterByIsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	listResp := &models.ApplicationListResponse{
		Applications: []models.Application{*newTestApplication()},
		Total:        1,
		Page:         1,
		PageSize:     20,
		TotalPages:   1,
	}

	svc.ListApplicationsFunc = func(page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
		require.NotNil(t, isActive)
		assert.True(t, *isActive)
		return listResp, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications", h.ListApplications)

	req := httptest.NewRequest(http.MethodGet, "/applications?page=1&page_size=20&is_active=true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestApplicationHandler_ListApplications_ShouldReturn500_WhenServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.ListApplicationsFunc = func(page, perPage int, isActive *bool) (*models.ApplicationListResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/applications", h.ListApplications)

	req := httptest.NewRequest(http.MethodGet, "/applications?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// UpdateApplication Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_UpdateApplication_ShouldReturn200_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	app := newTestApplication()
	app.DisplayName = "Updated App"

	svc.UpdateApplicationFunc = func(id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error) {
		assert.Equal(t, app.ID, id)
		assert.Equal(t, "Updated App", req.DisplayName)
		return app, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/applications/:id", h.UpdateApplication)

	body := `{"display_name":"Updated App","is_active":true}`
	req := httptest.NewRequest(http.MethodPut, "/applications/"+app.ID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Application
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Updated App", resp.DisplayName)
}

func TestApplicationHandler_UpdateApplication_ShouldReturn400_WhenInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/applications/:id", h.UpdateApplication)

	body := `{"display_name":"Updated App"}`
	req := httptest.NewRequest(http.MethodPut, "/applications/not-a-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApplicationHandler_UpdateApplication_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	appID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/applications/:id", h.UpdateApplication)

	req := httptest.NewRequest(http.MethodPut, "/applications/"+appID.String(), strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApplicationHandler_UpdateApplication_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.UpdateApplicationFunc = func(id uuid.UUID, req *models.UpdateApplicationRequest) (*models.Application, error) {
		return nil, service.ErrApplicationNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/applications/:id", h.UpdateApplication)

	body := `{"display_name":"Updated App"}`
	req := httptest.NewRequest(http.MethodPut, "/applications/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Application not found")
}

// ---------------------------------------------------------------------------
// DeleteApplication Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_DeleteApplication_ShouldReturn204_WhenDeleted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()

	svc.DeleteApplicationFunc = func(id uuid.UUID) error {
		assert.Equal(t, appID, id)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/applications/:id", h.DeleteApplication)

	req := httptest.NewRequest(http.MethodDelete, "/applications/"+appID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestApplicationHandler_DeleteApplication_ShouldReturn400_WhenInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/applications/:id", h.DeleteApplication)

	req := httptest.NewRequest(http.MethodDelete, "/applications/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApplicationHandler_DeleteApplication_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.DeleteApplicationFunc = func(id uuid.UUID) error {
		return service.ErrApplicationNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/applications/:id", h.DeleteApplication)

	req := httptest.NewRequest(http.MethodDelete, "/applications/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Application not found")
}

func TestApplicationHandler_DeleteApplication_ShouldReturn403_WhenSystemApp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.DeleteApplicationFunc = func(id uuid.UUID) error {
		return service.ErrCannotDeleteSystemApp
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/applications/:id", h.DeleteApplication)

	req := httptest.NewRequest(http.MethodDelete, "/applications/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Cannot delete system application")
}

// ---------------------------------------------------------------------------
// RotateSecret Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_RotateSecret_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()
	secret := "app_newsecret1234567890abcdef"

	svc.RotateSecretFunc = func(id uuid.UUID) (string, error) {
		assert.Equal(t, appID, id)
		return secret, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/rotate-secret", h.RotateSecret)

	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/rotate-secret", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, secret, resp["secret"])
	assert.Equal(t, secret[:12], resp["prefix"])
	assert.NotEmpty(t, resp["rotated_at"])
	assert.Equal(t, "Store this secret securely. It will not be shown again.", resp["warning"])
}

func TestApplicationHandler_RotateSecret_ShouldReturn400_WhenInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/rotate-secret", h.RotateSecret)

	req := httptest.NewRequest(http.MethodPost, "/applications/not-a-uuid/rotate-secret", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestApplicationHandler_RotateSecret_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.RotateSecretFunc = func(id uuid.UUID) (string, error) {
		return "", service.ErrApplicationNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/rotate-secret", h.RotateSecret)

	req := httptest.NewRequest(http.MethodPost, "/applications/"+uuid.New().String()+"/rotate-secret", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Application not found")
}

func TestApplicationHandler_RotateSecret_ShouldReturn500_WhenServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	svc.RotateSecretFunc = func(id uuid.UUID) (string, error) {
		return "", errors.New("unexpected failure")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/rotate-secret", h.RotateSecret)

	req := httptest.NewRequest(http.MethodPost, "/applications/"+uuid.New().String()+"/rotate-secret", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to rotate secret")
}

// ---------------------------------------------------------------------------
// BanUser Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_BanUser_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()
	targetUserID := uuid.New()
	adminID := uuid.New()

	svc.BanUserFunc = func(userID, applicationID, bannedBy uuid.UUID, reason string) error {
		assert.Equal(t, targetUserID, userID)
		assert.Equal(t, appID, applicationID)
		assert.Equal(t, adminID, bannedBy)
		assert.Equal(t, "Violation of ToS", reason)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/ban", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		h.BanUser(c)
	})

	body := `{"reason":"Violation of ToS"}`
	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/users/"+targetUserID.String()+"/ban", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "User banned successfully", resp.Message)
}

func TestApplicationHandler_BanUser_ShouldReturn400_WhenInvalidAppID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	adminID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/ban", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		h.BanUser(c)
	})

	body := `{"reason":"Violation of ToS"}`
	req := httptest.NewRequest(http.MethodPost, "/applications/not-a-uuid/users/"+uuid.New().String()+"/ban", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestApplicationHandler_BanUser_ShouldReturn400_WhenInvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	appID := uuid.New()
	adminID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/ban", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		h.BanUser(c)
	})

	body := `{"reason":"Violation of ToS"}`
	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/users/not-a-uuid/ban", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid user ID")
}

func TestApplicationHandler_BanUser_ShouldReturn404_WhenProfileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()
	targetUserID := uuid.New()
	adminID := uuid.New()

	svc.BanUserFunc = func(userID, applicationID, bannedBy uuid.UUID, reason string) error {
		return service.ErrUserProfileNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/ban", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		h.BanUser(c)
	})

	body := `{"reason":"Violation of ToS"}`
	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/users/"+targetUserID.String()+"/ban", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User profile not found")
}

// ---------------------------------------------------------------------------
// UnbanUser Tests
// ---------------------------------------------------------------------------

func TestApplicationHandler_UnbanUser_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()
	targetUserID := uuid.New()

	svc.UnbanUserFunc = func(userID, applicationID uuid.UUID) error {
		assert.Equal(t, targetUserID, userID)
		assert.Equal(t, appID, applicationID)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/unban", h.UnbanUser)

	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/users/"+targetUserID.String()+"/unban", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "User unbanned successfully", resp.Message)
}

func TestApplicationHandler_UnbanUser_ShouldReturn400_WhenInvalidAppID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupApplicationHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/unban", h.UnbanUser)

	req := httptest.NewRequest(http.MethodPost, "/applications/not-a-uuid/users/"+uuid.New().String()+"/unban", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestApplicationHandler_UnbanUser_ShouldReturn404_WhenProfileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupApplicationHandler()

	appID := uuid.New()
	targetUserID := uuid.New()

	svc.UnbanUserFunc = func(userID, applicationID uuid.UUID) error {
		return service.ErrUserProfileNotFound
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/applications/:id/users/:user_id/unban", h.UnbanUser)

	req := httptest.NewRequest(http.MethodPost, "/applications/"+appID.String()+"/users/"+targetUserID.String()+"/unban", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User profile not found")
}
