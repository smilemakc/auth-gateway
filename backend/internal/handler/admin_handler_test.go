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
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers – construct AdminService wired with mock repositories
// ---------------------------------------------------------------------------

func newTestAdminService(
	userRepo service.UserStore,
	apiKeyRepo service.APIKeyStore,
	auditRepo service.AuditStore,
	oauthRepo service.OAuthStore,
	rbacRepo service.RBACStore,
	backupCodeRepo service.BackupCodeStore,
	appRepo service.ApplicationStore,
	db service.TransactionDB,
) *service.AdminService {
	return service.NewAdminService(
		userRepo,
		apiKeyRepo,
		auditRepo,
		oauthRepo,
		rbacRepo,
		backupCodeRepo,
		appRepo,
		10, // bcryptCost
		db,
	)
}

// adminTestFixture wraps handler and mock dependencies for admin handler tests.
type adminTestFixture struct {
	handler    *AdminHandler
	userStore  *mockUserStoreHandler
	apiKeyRepo *mockAPIKeyStoreHandler
	auditRepo  *mockAuditStoreHandler
	oauthRepo  *mockOAuthStoreHandler
	rbacStore  *mockRBACStoreHandler
	backupRepo *mockBackupCodeStoreHandler
	appRepo    *mockAppStoreHandler
	db         *mockTransactionDBHandler
	otpStore   *mockOTPStoreHandler
	auditLog   *mockAuditLoggerHandler
}

func setupAdminTestFixture() *adminTestFixture {
	userStore := &mockUserStoreHandler{}
	apiKeyRepo := &mockAPIKeyStoreHandler{}
	auditRepo := &mockAuditStoreHandler{}
	oauthRepo := &mockOAuthStoreHandler{}
	rbacStore := &mockRBACStoreHandler{}
	backupRepo := &mockBackupCodeStoreHandler{}
	appRepo := &mockAppStoreHandler{}
	db := &mockTransactionDBHandler{}
	otpStore := &mockOTPStoreHandler{}
	auditLog := &mockAuditLoggerHandler{}

	adminSvc := newTestAdminService(
		userStore, apiKeyRepo, auditRepo, oauthRepo, rbacStore, backupRepo, appRepo, db,
	)
	userSvc := newTestUserService(userStore, auditLog)
	otpSvc := newTestOTPService(otpStore, userStore, auditLog)

	// AuditService uses concrete *repository.AuditRepository — pass nil.
	// SendPasswordReset calls h.auditService.Log() which fires asynchronously;
	// tests that exercise that path set auditService to a non-nil value.
	h := NewAdminHandler(adminSvc, userSvc, otpSvc, nil, testLogger())
	return &adminTestFixture{
		handler:    h,
		userStore:  userStore,
		apiKeyRepo: apiKeyRepo,
		auditRepo:  auditRepo,
		oauthRepo:  oauthRepo,
		rbacStore:  rbacStore,
		backupRepo: backupRepo,
		appRepo:    appRepo,
		db:         db,
		otpStore:   otpStore,
		auditLog:   auditLog,
	}
}

// ---------------------------------------------------------------------------
// GetStats Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_GetStats_ShouldReturn200_WhenServiceSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.CountFunc = func(isActive *bool) (int, error) {
		return 10, nil
	}
	fix.userStore.ListFunc = func(opts ...service.UserListOption) ([]*models.User, error) {
		return []*models.User{
			{ID: uuid.New(), IsActive: true, EmailVerified: true},
			{ID: uuid.New(), IsActive: true},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/stats", fix.handler.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats models.AdminStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 10, stats.TotalUsers)
	assert.Equal(t, 2, stats.ActiveUsers)
}

func TestAdminHandler_GetStats_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.CountFunc = func(isActive *bool) (int, error) {
		return 0, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/stats", fix.handler.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// ListUsers Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListUsers_ShouldReturn200_WhenUsersExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	fix.userStore.ListFunc = func(opts ...service.UserListOption) ([]*models.User, error) {
		return []*models.User{
			{ID: userID, Email: "admin@test.com", Username: "admin", IsActive: true},
		}, nil
	}
	fix.userStore.CountFunc = func(isActive *bool) (int, error) {
		return 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users", fix.handler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminUserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "admin@test.com", resp.Users[0].Email)
}

func TestAdminHandler_ListUsers_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.ListFunc = func(opts ...service.UserListOption) ([]*models.User, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users", fix.handler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// GetUser Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_GetUser_ShouldReturn200_WhenUserExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       id,
			Email:    "user@test.com",
			Username: "testuser",
			IsActive: true,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id", fix.handler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "user@test.com", resp.Email)
}

func TestAdminHandler_GetUser_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id", fix.handler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_GetUser_ShouldReturn404_WhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, models.NewAppError(http.StatusNotFound, "User not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id", fix.handler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_GetUser_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id", fix.handler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// GetUserOAuthAccounts Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_GetUserOAuthAccounts_ShouldReturn200_WhenAccountsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, Email: "user@test.com"}, nil
	}
	fix.oauthRepo.GetByUserIDFunc = func(uid uuid.UUID) ([]*models.OAuthAccount, error) {
		return []*models.OAuthAccount{
			{ID: uuid.New(), UserID: uid, Provider: "google"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id/oauth-accounts", fix.handler.GetUserOAuthAccounts)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String()+"/oauth-accounts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.OAuthAccountListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "google", resp.Accounts[0].Provider)
}

func TestAdminHandler_GetUserOAuthAccounts_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id/oauth-accounts", fix.handler.GetUserOAuthAccounts)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/bad-id/oauth-accounts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_GetUserOAuthAccounts_ShouldReturn404_WhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, models.NewAppError(http.StatusNotFound, "User not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/:id/oauth-accounts", fix.handler.GetUserOAuthAccounts)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+uuid.New().String()+"/oauth-accounts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// CreateUser Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_CreateUser_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing email", `{"username":"newuser","password":"SecurePass1!","full_name":"New User"}`},
		{"missing username", `{"email":"new@test.com","password":"SecurePass1!","full_name":"New User"}`},
		{"missing password", `{"email":"new@test.com","username":"newuser","full_name":"New User"}`},
		{"missing full_name", `{"email":"new@test.com","username":"newuser","password":"SecurePass1!"}`},
		{"password too short", `{"email":"new@test.com","username":"newuser","password":"abc","full_name":"New User"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAdminTestFixture()
			adminID := uuid.New()

			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/admin/users", func(c *gin.Context) {
				c.Set(utils.UserIDKey, adminID)
				fix.handler.CreateUser(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAdminHandler_CreateUser_ShouldReturn401_WhenNoAdminID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users", fix.handler.CreateUser)

	body := `{"email":"new@test.com","username":"newuser","password":"SecurePass1!","full_name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_CreateUser_ShouldReturn200_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	adminID := uuid.New()
	createdUserID := uuid.New()

	fix.rbacStore.GetRoleByNameFunc = func(name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	fix.userStore.CreateFunc = func(user *models.User) error {
		user.ID = createdUserID
		return nil
	}
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       createdUserID,
			Email:    "new@test.com",
			Username: "newuser",
			FullName: "New User",
			IsActive: true,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.CreateUser(c)
	})

	body := `{"email":"new@test.com","username":"newuser","password":"SecurePass1!","full_name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "new@test.com", resp.Email)
}

func TestAdminHandler_CreateUser_ShouldReturn500_WhenCreateFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	adminID := uuid.New()
	fix.rbacStore.GetRoleByNameFunc = func(name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}
	// The service wraps CreateFunc errors with fmt.Errorf, losing AppError type,
	// so RespondWithError returns 500 for any creation failure.
	fix.userStore.CreateFunc = func(user *models.User) error {
		return fmt.Errorf("unique constraint violation")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.CreateUser(c)
	})

	body := `{"email":"dup@test.com","username":"dupuser","password":"SecurePass1!","full_name":"Dup User"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// UpdateUser Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_UpdateUser_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/users/:id", fix.handler.UpdateUser)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/bad-id", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_UpdateUser_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/users/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.UpdateUser(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+uuid.New().String(), strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_UpdateUser_ShouldReturn401_WhenNoAdminID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/users/:id", fix.handler.UpdateUser)

	body := `{"is_active":false}`
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_UpdateUser_ShouldReturn200_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	adminID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       userID,
			Email:    "user@test.com",
			Username: "testuser",
			IsActive: true,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/users/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.UpdateUser(c)
	})

	body := `{"is_active":false}`
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+userID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_UpdateUser_ShouldReturn404_WhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, models.NewAppError(http.StatusNotFound, "User not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/admin/users/:id", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.UpdateUser(c)
	})

	body := `{"is_active":false}`
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// DeleteUser Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_DeleteUser_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id", fix.handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/not-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_DeleteUser_ShouldReturn200_WhenUserExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, IsActive: true}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id", fix.handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "User deleted successfully", resp.Message)
}

func TestAdminHandler_DeleteUser_ShouldReturn404_WhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, models.NewAppError(http.StatusNotFound, "User not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id", fix.handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_DeleteUser_ShouldReturn400_WhenLastAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	adminRoleID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, IsActive: true}, nil
	}
	fix.rbacStore.GetUserRolesFunc = func() ([]models.Role, error) {
		return []models.Role{{ID: adminRoleID, Name: "admin"}}, nil
	}
	fix.rbacStore.GetUsersWithRoleFunc = func(roleID uuid.UUID) ([]models.User, error) {
		return []models.User{{ID: userID}}, nil // only one admin
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id", fix.handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// ListAPIKeys Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListAPIKeys_ShouldReturn200_WhenKeysExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	keyID := uuid.New()
	ownerID := uuid.New()
	fix.apiKeyRepo.ListAllFunc = func() ([]*models.APIKey, error) {
		return []*models.APIKey{
			{ID: keyID, UserID: ownerID, Name: "Test Key", KeyPrefix: "agw_test", Scopes: []byte(`["users:read"]`), IsActive: true},
		}, nil
	}
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, Email: "owner@test.com", Username: "owner"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/api-keys", fix.handler.ListAPIKeys)

	req := httptest.NewRequest(http.MethodGet, "/admin/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminAPIKeyListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.APIKeys, 1)
	assert.Equal(t, "Test Key", resp.APIKeys[0].Name)
}

func TestAdminHandler_ListAPIKeys_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.apiKeyRepo.ListAllFunc = func() ([]*models.APIKey, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/api-keys", fix.handler.ListAPIKeys)

	req := httptest.NewRequest(http.MethodGet, "/admin/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// RevokeAPIKey Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_RevokeAPIKey_ShouldReturn200_WhenKeyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	keyID := uuid.New()
	fix.apiKeyRepo.RevokeFunc = func(id uuid.UUID) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/api-keys/:id/revoke", fix.handler.RevokeAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/admin/api-keys/"+keyID.String()+"/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "API key revoked successfully", resp.Message)
}

func TestAdminHandler_RevokeAPIKey_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/api-keys/:id/revoke", fix.handler.RevokeAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/admin/api-keys/invalid-uuid/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_RevokeAPIKey_ShouldReturnError_WhenKeyNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.apiKeyRepo.RevokeFunc = func(id uuid.UUID) error {
		return models.NewAppError(http.StatusNotFound, "API key not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/api-keys/:id/revoke", fix.handler.RevokeAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/admin/api-keys/"+uuid.New().String()+"/revoke", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// ListAuditLogs Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListAuditLogs_ShouldReturn200_WhenLogsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	logID := uuid.New()
	fix.auditRepo.ListFunc = func(limit, offset int) ([]*models.AuditLog, error) {
		return []*models.AuditLog{
			{ID: logID, Action: "signin", Status: "success", IPAddress: "127.0.0.1"},
		}, nil
	}
	fix.auditRepo.CountFunc = func() (int, error) {
		return 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/audit-logs", fix.handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AuditLogListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Logs, 1)
	assert.Equal(t, "signin", resp.Logs[0].Action)
}

func TestAdminHandler_ListAuditLogs_ShouldReturn400_WhenUserIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/audit-logs", fix.handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-logs?user_id=not-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_ListAuditLogs_ShouldReturn200_WhenFilteredByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	fix.auditRepo.GetByUserIDFunc = func(uid uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
		return []*models.AuditLog{
			{ID: uuid.New(), UserID: &uid, Action: "signup", Status: "success"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/audit-logs", fix.handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-logs?user_id="+userID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_ListAuditLogs_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.auditRepo.ListFunc = func(limit, offset int) ([]*models.AuditLog, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/audit-logs", fix.handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// AssignRole Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_AssignRole_ShouldReturn400_WhenUserIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/roles", fix.handler.AssignRole)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/bad-id/roles", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_AssignRole_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/roles", func(c *gin.Context) {
		c.Set(utils.UserIDKey, uuid.New())
		fix.handler.AssignRole(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/roles", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_AssignRole_ShouldReturn401_WhenNoAdminID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	roleID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/roles", fix.handler.AssignRole)

	body := fmt.Sprintf(`{"role_id":"%s"}`, roleID)
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_AssignRole_ShouldReturn200_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	roleID := uuid.New()
	adminID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       id,
			Email:    "user@test.com",
			Username: "testuser",
			IsActive: true,
			Roles:    []models.Role{{ID: uuid.New(), Name: "user"}},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/roles", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.AssignRole(c)
	})

	body := fmt.Sprintf(`{"role_id":"%s"}`, roleID)
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+userID.String()+"/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_AssignRole_ShouldReturn400_WhenRoleAlreadyAssigned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	roleID := uuid.New()
	adminID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       id,
			Email:    "user@test.com",
			IsActive: true,
			Roles:    []models.Role{{ID: roleID, Name: "admin"}},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/roles", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.AssignRole(c)
	})

	body := fmt.Sprintf(`{"role_id":"%s"}`, roleID)
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+userID.String()+"/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// RemoveRole Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_RemoveRole_ShouldReturn400_WhenUserIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id/roles/:roleId", fix.handler.RemoveRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/bad-id/roles/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_RemoveRole_ShouldReturn400_WhenRoleIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id/roles/:roleId", fix.handler.RemoveRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+uuid.New().String()+"/roles/bad-id", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_RemoveRole_ShouldReturn200_WhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	roleID := uuid.New()
	otherRoleID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       id,
			Email:    "user@test.com",
			IsActive: true,
			Roles: []models.Role{
				{ID: roleID, Name: "admin"},
				{ID: otherRoleID, Name: "user"},
			},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id/roles/:roleId", fix.handler.RemoveRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String()+"/roles/"+roleID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_RemoveRole_ShouldReturn400_WhenOnlyOneRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	roleID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:    id,
			Roles: []models.Role{{ID: roleID, Name: "user"}},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id/roles/:roleId", fix.handler.RemoveRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String()+"/roles/"+roleID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_RemoveRole_ShouldReturn404_WhenRoleNotAssigned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	roleID := uuid.New()
	anotherRoleID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID: id,
			Roles: []models.Role{
				{ID: anotherRoleID, Name: "user"},
				{ID: uuid.New(), Name: "moderator"},
			},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/admin/users/:id/roles/:roleId", fix.handler.RemoveRole)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String()+"/roles/"+roleID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Reset2FA Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_Reset2FA_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/reset-2fa", fix.handler.Reset2FA)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/bad-uuid/reset-2fa", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_Reset2FA_ShouldReturn401_WhenNoAdminID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/reset-2fa", fix.handler.Reset2FA)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/reset-2fa", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_Reset2FA_ShouldReturn200_WhenTOTPEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	userID := uuid.New()
	adminID := uuid.New()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:          id,
			TOTPEnabled: true,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/reset-2fa", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.Reset2FA(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+userID.String()+"/reset-2fa", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "2FA has been disabled for user", resp["message"])
	assert.Equal(t, userID.String(), resp["user_id"])
}

func TestAdminHandler_Reset2FA_ShouldReturn400_WhenTOTPNotEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	adminID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:          id,
			TOTPEnabled: false,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/reset-2fa", func(c *gin.Context) {
		c.Set(utils.UserIDKey, adminID)
		fix.handler.Reset2FA(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/reset-2fa", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// SyncUsers Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_SyncUsers_ShouldReturn400_WhenUpdatedAfterMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/sync", fix.handler.SyncUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/sync", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_SyncUsers_ShouldReturn400_WhenUpdatedAfterInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/sync", fix.handler.SyncUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/sync?updated_after=not-a-date", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_SyncUsers_ShouldReturn400_WhenApplicationIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/sync", fix.handler.SyncUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/sync?updated_after=2024-01-15T10:30:00Z&application_id=bad-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_SyncUsers_ShouldReturn200_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetUsersUpdatedAfterFunc = func(after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
		return []*models.User{
			{ID: uuid.New(), Email: "synced@test.com", Username: "synced", IsActive: true},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/sync", fix.handler.SyncUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/sync?updated_after=2024-01-15T10:30:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SyncUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Users, 1)
}

func TestAdminHandler_SyncUsers_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetUsersUpdatedAfterFunc = func(after time.Time, appID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
		return nil, 0, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/admin/users/sync", fix.handler.SyncUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/sync?updated_after=2024-01-15T10:30:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// ImportUsers Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ImportUsers_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing users", `{"on_conflict":"skip"}`},
		{"missing on_conflict", `{"users":[{"email":"a@b.com"}]}`},
		{"invalid on_conflict", `{"users":[{"email":"a@b.com"}],"on_conflict":"invalid"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAdminTestFixture()

			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/admin/users/import", fix.handler.ImportUsers)

			req := httptest.NewRequest(http.MethodPost, "/admin/users/import", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAdminHandler_ImportUsers_ShouldReturn200_WhenValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByEmailFunc = func(email string) (*models.User, error) {
		return nil, fmt.Errorf("not found")
	}
	fix.userStore.GetByUsernameFunc = func(username string) (*models.User, error) {
		return nil, fmt.Errorf("not found")
	}
	fix.userStore.CreateFunc = func(user *models.User) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/import", fix.handler.ImportUsers)

	body := `{"users":[{"email":"import@test.com","username":"importuser","full_name":"Import User"}],"on_conflict":"skip"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImportUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Imported)
	assert.Equal(t, 0, resp.Errors)
}

// ---------------------------------------------------------------------------
// SendPasswordReset Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_SendPasswordReset_ShouldReturn400_WhenIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/send-password-reset", fix.handler.SendPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/bad-id/send-password-reset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_SendPasswordReset_ShouldReturn404_WhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, models.NewAppError(http.StatusNotFound, "User not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/send-password-reset", fix.handler.SendPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/send-password-reset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_SendPasswordReset_ShouldReturn400_WhenNoEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, Email: ""}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/send-password-reset", fix.handler.SendPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/send-password-reset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_SendPasswordReset_ShouldReturn400_WhenEmailNotVerified(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAdminTestFixture()

	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, Email: "user@test.com", EmailVerified: false}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/admin/users/:id/send-password-reset", fix.handler.SendPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+uuid.New().String()+"/send-password-reset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
