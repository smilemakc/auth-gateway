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
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers – construct services for AdvancedAdminHandler
// ---------------------------------------------------------------------------

func newTestRBACService(rbacRepo service.RBACStore, auditLogger service.AuditLogger) *service.RBACService {
	return service.NewRBACService(rbacRepo, auditLogger)
}

func newTestIPFilterService(ipFilterRepo service.IPFilterStore) *service.IPFilterService {
	return service.NewIPFilterService(ipFilterRepo)
}

// newTestSessionService constructs a SessionService using struct literals
// because the constructor takes *service.BlacklistService (concrete), but
// the struct field is BlackListStore (interface).
func newTestSessionService(sessionRepo service.SessionStore, blacklist service.BlackListStore) *service.SessionService {
	return service.NewSessionService(sessionRepo, nil, testLogger(), 10)
}

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessExpires:  15 * time.Minute,
			RefreshExpires: 7 * 24 * time.Hour,
		},
		Security: config.SecurityConfig{
			PasswordPolicy: config.PasswordPolicyConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSpecial:   false,
			},
		},
	}
}

// advancedAdminTestFixture wraps handler and mock dependencies.
type advancedAdminTestFixture struct {
	handler      *AdvancedAdminHandler
	rbacStore    *mockRBACStoreHandler
	sessionStore *mockSessionStoreHandler
	ipFilterRepo *mockIPFilterStoreHandler
	auditLog     *mockAuditLoggerHandler
	blacklist    *mockBlackListStoreHandler
	cfg          *config.Config
}

func setupAdvancedAdminTestFixture() *advancedAdminTestFixture {
	rbacStore := &mockRBACStoreHandler{}
	auditLog := &mockAuditLoggerHandler{}
	sessionStore := &mockSessionStoreHandler{}
	ipFilterRepo := &mockIPFilterStoreHandler{}
	blacklist := &mockBlackListStoreHandler{}
	cfg := testConfig()

	rbacSvc := newTestRBACService(rbacStore, auditLog)
	sessionSvc := newTestSessionService(sessionStore, blacklist)
	ipFilterSvc := newTestIPFilterService(ipFilterRepo)

	// brandingRepo, systemRepo, geoRepo are concrete types — pass nil.
	// Tests for those endpoints are skipped.
	h := NewAdvancedAdminHandler(
		rbacSvc, sessionSvc, ipFilterSvc,
		nil, nil, nil, // brandingRepo, systemRepo, geoRepo
		testLogger(), cfg,
	)
	return &advancedAdminTestFixture{
		handler:      h,
		rbacStore:    rbacStore,
		sessionStore: sessionStore,
		ipFilterRepo: ipFilterRepo,
		auditLog:     auditLog,
		blacklist:    blacklist,
		cfg:          cfg,
	}
}

// ============================================================
// RBAC - Permissions Tests
// ============================================================

func TestAdvancedAdmin_ListPermissions_ShouldReturn200_WhenPermissionsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.ListPermissionsFunc = func() ([]models.Permission, error) {
		return []models.Permission{
			{ID: uuid.New(), Name: "users.read", Resource: "users", Action: "read"},
			{ID: uuid.New(), Name: "users.write", Resource: "users", Action: "write"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permissions", fix.handler.ListPermissions)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.PermissionListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Permissions, 2)
}

func TestAdvancedAdmin_ListPermissions_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.ListPermissionsFunc = func() ([]models.Permission, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permissions", fix.handler.ListPermissions)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdvancedAdmin_CreatePermission_ShouldReturn201_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.CreatePermissionFunc = func(p *models.Permission) error {
		p.ID = uuid.New()
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/rbac/permissions", fix.handler.CreatePermission)

	body := `{"name":"users.delete","resource":"users","action":"delete","description":"Can delete users"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/rbac/permissions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdvancedAdmin_CreatePermission_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing name", `{"resource":"users","action":"read"}`},
		{"missing resource", `{"name":"users.read","action":"read"}`},
		{"missing action", `{"name":"users.read","resource":"users"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAdvancedAdminTestFixture()

			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/admin/rbac/permissions", fix.handler.CreatePermission)

			req := httptest.NewRequest(http.MethodPost, "/admin/rbac/permissions", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAdvancedAdmin_GetPermission_ShouldReturn200_WhenExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	permID := uuid.New()
	fix.rbacStore.GetPermissionByIDFunc = func(id uuid.UUID) (*models.Permission, error) {
		return &models.Permission{ID: id, Name: "users.read", Resource: "users", Action: "read"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permissions/:id", fix.handler.GetPermission)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permissions/"+permID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_GetPermission_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permissions/:id", fix.handler.GetPermission)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permissions/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_GetPermission_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.GetPermissionByIDFunc = func(id uuid.UUID) (*models.Permission, error) {
		return nil, fmt.Errorf("not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permissions/:id", fix.handler.GetPermission)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permissions/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdvancedAdmin_DeletePermission_ShouldReturn204_WhenExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.DeletePermissionFunc = func(id uuid.UUID) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/rbac/permissions/:id", fix.handler.DeletePermission)

	req := httptest.NewRequest(http.MethodDelete, "/admin/rbac/permissions/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAdvancedAdmin_DeletePermission_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/rbac/permissions/:id", fix.handler.DeletePermission)

	req := httptest.NewRequest(http.MethodDelete, "/admin/rbac/permissions/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_UpdatePermission_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/rbac/permissions/:id", fix.handler.UpdatePermission)

	req := httptest.NewRequest(http.MethodPut, "/admin/rbac/permissions/bad-id", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_UpdatePermission_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/rbac/permissions/:id", fix.handler.UpdatePermission)

	req := httptest.NewRequest(http.MethodPut, "/admin/rbac/permissions/"+uuid.New().String(), strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================
// RBAC - Roles Tests
// ============================================================

func TestAdvancedAdmin_ListRoles_ShouldReturn200_WhenRolesExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.ListRolesFunc = func() ([]models.Role, error) {
		return []models.Role{
			{ID: uuid.New(), Name: "admin", DisplayName: "Administrator"},
			{ID: uuid.New(), Name: "user", DisplayName: "User"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/roles", fix.handler.ListRoles)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/roles", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.RoleListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
}

func TestAdvancedAdmin_ListRoles_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.ListRolesFunc = func() ([]models.Role, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/roles", fix.handler.ListRoles)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/roles", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdvancedAdmin_CreateRole_ShouldReturn201_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.CreateRoleFunc = func(r *models.Role) error {
		r.ID = uuid.New()
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/rbac/roles", fix.handler.CreateRole)

	body := `{"name":"moderator","display_name":"Moderator","description":"Can moderate content"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/rbac/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdvancedAdmin_CreateRole_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing name", `{"display_name":"Mod"}`},
		{"missing display_name", `{"name":"mod"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAdvancedAdminTestFixture()

			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/admin/rbac/roles", fix.handler.CreateRole)

			req := httptest.NewRequest(http.MethodPost, "/admin/rbac/roles", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAdvancedAdmin_GetRole_ShouldReturn200_WhenExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	roleID := uuid.New()
	fix.rbacStore.GetRoleByIDFunc = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: id, Name: "admin", DisplayName: "Administrator"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/roles/:id", fix.handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/roles/"+roleID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_GetRole_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/roles/:id", fix.handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/roles/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_GetRole_ShouldReturn404_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.GetRoleByIDFunc = func(id uuid.UUID) (*models.Role, error) {
		return nil, fmt.Errorf("not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/roles/:id", fix.handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/roles/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdvancedAdmin_UpdateRole_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/rbac/roles/:id", fix.handler.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/rbac/roles/bad-id", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_UpdateRole_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/rbac/roles/:id", fix.handler.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/rbac/roles/"+uuid.New().String(), strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_UpdateRole_ShouldReturn200_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	roleID := uuid.New()
	fix.rbacStore.UpdateRoleFunc = func(id uuid.UUID, displayName, description string) error {
		return nil
	}
	fix.rbacStore.GetRoleByIDFunc = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: id, Name: "admin", DisplayName: "Updated Admin"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/rbac/roles/:id", fix.handler.UpdateRole)

	body := `{"display_name":"Updated Admin","description":"Updated description"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/rbac/roles/"+roleID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_DeleteRole_ShouldReturn204_WhenExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.DeleteRoleFunc = func(id uuid.UUID) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/rbac/roles/:id", fix.handler.DeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/rbac/roles/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAdvancedAdmin_DeleteRole_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/rbac/roles/:id", fix.handler.DeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/rbac/roles/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_DeleteRole_ShouldReturn400_WhenDeleteFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.DeleteRoleFunc = func(id uuid.UUID) error {
		return fmt.Errorf("cannot delete system role")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/rbac/roles/:id", fix.handler.DeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/rbac/roles/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================
// RBAC - Permission Matrix Tests
// ============================================================

func TestAdvancedAdmin_GetPermissionMatrix_ShouldReturn200_WhenSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.GetPermissionMatrixFunc = func() (*models.PermissionMatrix, error) {
		return &models.PermissionMatrix{
			Resources: []models.ResourcePermissions{
				{Resource: "users", Permissions: []models.PermissionWithRoles{}},
			},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permission-matrix", fix.handler.GetPermissionMatrix)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permission-matrix", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_GetPermissionMatrix_ShouldReturn500_WhenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.rbacStore.GetPermissionMatrixFunc = func() (*models.PermissionMatrix, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/rbac/permission-matrix", fix.handler.GetPermissionMatrix)

	req := httptest.NewRequest(http.MethodGet, "/admin/rbac/permission-matrix", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================
// Session Management Tests
// ============================================================

func TestAdvancedAdmin_ListUserSessions_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sessions", fix.handler.ListUserSessions)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdvancedAdmin_ListUserSessions_ShouldReturn200_WhenAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	userID := uuid.New()
	fix.sessionStore.GetUserSessionsPaginatedFunc = func(uid uuid.UUID, page, perPage int) ([]models.Session, int, error) {
		return []models.Session{
			{ID: uuid.New(), UserID: uid, DeviceType: "desktop", Browser: "Chrome"},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sessions", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.ListUserSessions(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SessionListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
}

func TestAdvancedAdmin_ListUserSessions_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	userID := uuid.New()
	fix.sessionStore.GetUserSessionsPaginatedFunc = func(uid uuid.UUID, page, perPage int) ([]models.Session, int, error) {
		return nil, 0, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sessions", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.ListUserSessions(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdvancedAdmin_RevokeSession_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/sessions/:id", fix.handler.RevokeSession)

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdvancedAdmin_RevokeSession_ShouldReturn400_WhenSessionIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/sessions/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.RevokeSession(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_RevokeSession_ShouldReturn400_WhenSessionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetSessionByIDFunc = func(id uuid.UUID) (*models.Session, error) {
		return nil, fmt.Errorf("session not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/sessions/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.RevokeSession(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_AdminRevokeSession_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/sessions/:id", fix.handler.AdminRevokeSession)

	req := httptest.NewRequest(http.MethodDelete, "/admin/sessions/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_AdminRevokeSession_ShouldReturn400_WhenSessionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetSessionByIDFunc = func(id uuid.UUID) (*models.Session, error) {
		return nil, fmt.Errorf("session not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/sessions/:id", fix.handler.AdminRevokeSession)

	req := httptest.NewRequest(http.MethodDelete, "/admin/sessions/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_GetSessionStats_ShouldReturn200_WhenSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetSessionStatsFunc = func() (*models.SessionStats, error) {
		return &models.SessionStats{
			TotalActiveSessions: 42,
			SessionsByDevice:    map[string]int{"desktop": 30, "mobile": 12},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/sessions/stats", fix.handler.GetSessionStats)

	req := httptest.NewRequest(http.MethodGet, "/admin/sessions/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats models.SessionStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 42, stats.TotalActiveSessions)
}

func TestAdvancedAdmin_GetSessionStats_ShouldReturn500_WhenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetSessionStatsFunc = func() (*models.SessionStats, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/sessions/stats", fix.handler.GetSessionStats)

	req := httptest.NewRequest(http.MethodGet, "/admin/sessions/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdvancedAdmin_ListUserSessionsAdmin_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id/sessions", fix.handler.ListUserSessionsAdmin)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/bad-id/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_ListUserSessionsAdmin_ShouldReturn200_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	userID := uuid.New()
	fix.sessionStore.GetUserSessionsPaginatedFunc = func(uid uuid.UUID, page, perPage int) ([]models.Session, int, error) {
		return []models.Session{
			{ID: uuid.New(), UserID: uid, Browser: "Firefox"},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id/sessions", fix.handler.ListUserSessionsAdmin)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String()+"/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_ListAllSessions_ShouldReturn200_WhenSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetAllSessionsPaginatedFunc = func(page, perPage int) ([]models.Session, int, error) {
		return []models.Session{
			{ID: uuid.New(), UserID: uuid.New(), Browser: "Chrome"},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/sessions", fix.handler.ListAllSessions)

	req := httptest.NewRequest(http.MethodGet, "/admin/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_ListAllSessions_ShouldReturn500_WhenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.sessionStore.GetAllSessionsPaginatedFunc = func(page, perPage int) ([]models.Session, int, error) {
		return nil, 0, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/sessions", fix.handler.ListAllSessions)

	req := httptest.NewRequest(http.MethodGet, "/admin/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================
// IP Filter Tests
// ============================================================

func TestAdvancedAdmin_ListIPFilters_ShouldReturn200_WhenFiltersExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.ipFilterRepo.ListIPFiltersFunc = func(page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
		return []models.IPFilterWithCreator{
			{IPFilter: models.IPFilter{ID: uuid.New(), IPCIDR: "192.168.1.0/24", FilterType: "blacklist"}},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/ip-filters", fix.handler.ListIPFilters)

	req := httptest.NewRequest(http.MethodGet, "/admin/ip-filters", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdvancedAdmin_ListIPFilters_ShouldReturn500_WhenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.ipFilterRepo.ListIPFiltersFunc = func(page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
		return nil, 0, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/ip-filters", fix.handler.ListIPFilters)

	req := httptest.NewRequest(http.MethodGet, "/admin/ip-filters", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdvancedAdmin_CreateIPFilter_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/ip-filters", fix.handler.CreateIPFilter)

	body := `{"ip_cidr":"192.168.1.0/24","filter_type":"blacklist"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/ip-filters", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdvancedAdmin_CreateIPFilter_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing ip_cidr", `{"filter_type":"blacklist"}`},
		{"missing filter_type", `{"ip_cidr":"192.168.1.0/24"}`},
		{"invalid filter_type", `{"ip_cidr":"192.168.1.0/24","filter_type":"invalid"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAdvancedAdminTestFixture()

			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/admin/ip-filters", func(c *gin.Context) {
				c.Set(utils.UserIDKey, uuid.New())
				fix.handler.CreateIPFilter(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/admin/ip-filters", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAdvancedAdmin_CreateIPFilter_ShouldReturn201_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.ipFilterRepo.CreateIPFilterFunc = func(filter *models.IPFilter) error {
		filter.ID = uuid.New()
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/ip-filters", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.CreateIPFilter(c)
	})

	body := `{"ip_cidr":"192.168.1.0/24","filter_type":"blacklist","reason":"Suspicious"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/ip-filters", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdvancedAdmin_DeleteIPFilter_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/ip-filters/:id", fix.handler.DeleteIPFilter)

	req := httptest.NewRequest(http.MethodDelete, "/admin/ip-filters/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvancedAdmin_DeleteIPFilter_ShouldReturn204_WhenExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.ipFilterRepo.DeleteIPFilterFunc = func(id uuid.UUID) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/ip-filters/:id", fix.handler.DeleteIPFilter)

	req := httptest.NewRequest(http.MethodDelete, "/admin/ip-filters/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAdvancedAdmin_DeleteIPFilter_ShouldReturn400_WhenDeleteFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	fix.ipFilterRepo.DeleteIPFilterFunc = func(id uuid.UUID) error {
		return fmt.Errorf("filter not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/ip-filters/:id", fix.handler.DeleteIPFilter)

	req := httptest.NewRequest(http.MethodDelete, "/admin/ip-filters/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================
// System Health Tests
// ============================================================

func TestAdvancedAdmin_GetSystemHealth_ShouldReturn200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/system/health", fix.handler.GetSystemHealth)

	req := httptest.NewRequest(http.MethodGet, "/admin/system/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SystemHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "healthy", resp.DatabaseStatus)
	assert.Equal(t, "healthy", resp.RedisStatus)
}

// ============================================================
// Password Policy Tests
// ============================================================

func TestAdvancedAdmin_GetPasswordPolicy_ShouldReturn200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/system/password-policy", fix.handler.GetPasswordPolicy)

	req := httptest.NewRequest(http.MethodGet, "/admin/system/password-policy", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(8), resp["minLength"])
	assert.Equal(t, true, resp["requireUppercase"])
	assert.Equal(t, true, resp["requireLowercase"])
	assert.Equal(t, true, resp["requireNumbers"])
	assert.Equal(t, false, resp["requireSpecial"])
	assert.Equal(t, float64(15), resp["jwtTtlMinutes"])
	assert.Equal(t, float64(7), resp["refreshTtlDays"])
}

func TestAdvancedAdmin_UpdatePasswordPolicy_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/system/password-policy", fix.handler.UpdatePasswordPolicy)

	body := `{"minLength":10}`
	req := httptest.NewRequest(http.MethodPut, "/admin/system/password-policy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdvancedAdmin_UpdatePasswordPolicy_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/system/password-policy", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.UpdatePasswordPolicy(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/system/password-policy", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================
// RevokeAllSessions Tests
// ============================================================

func TestAdvancedAdmin_RevokeAllSessions_ShouldReturn400_WhenUserIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdvancedAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/sessions/revoke-all", fix.handler.RevokeAllSessions)

	req := httptest.NewRequest(http.MethodPost, "/sessions/revoke-all?user_id=bad-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
