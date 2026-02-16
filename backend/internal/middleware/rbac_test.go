package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

// mockRBACStore implements service.RBACStore for testing the RBAC middleware.
// Only the permission-checking methods are relevant for middleware tests.
type mockRBACStore struct {
	hasPermissionFn     func(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)
	hasAnyPermissionFn  func(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
	hasAllPermissionsFn func(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error)
}

// --- PermissionRepository stubs ---
func (m *mockRBACStore) CreatePermission(ctx context.Context, permission *models.Permission) error {
	return nil
}
func (m *mockRBACStore) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStore) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStore) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return nil, nil
}
func (m *mockRBACStore) UpdatePermission(ctx context.Context, id uuid.UUID, description string) error {
	return nil
}
func (m *mockRBACStore) DeletePermission(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRBACStore) ListPermissionsByApp(ctx context.Context, appID *uuid.UUID) ([]models.Permission, error) {
	return nil, nil
}

// --- RoleRepository stubs ---
func (m *mockRBACStore) CreateRole(ctx context.Context, role *models.Role) error { return nil }
func (m *mockRBACStore) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStore) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStore) ListRoles(ctx context.Context) ([]models.Role, error) { return nil, nil }
func (m *mockRBACStore) UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error {
	return nil
}
func (m *mockRBACStore) DeleteRole(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockRBACStore) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}
func (m *mockRBACStore) GetRoleByNameAndApp(ctx context.Context, name string, appID *uuid.UUID) (*models.Role, error) {
	return nil, nil
}
func (m *mockRBACStore) ListRolesByApp(ctx context.Context, appID *uuid.UUID) ([]models.Role, error) {
	return nil, nil
}

// --- UserRoleRepository stubs ---
func (m *mockRBACStore) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	return nil
}
func (m *mockRBACStore) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *mockRBACStore) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRBACStore) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	return nil
}
func (m *mockRBACStore) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (m *mockRBACStore) GetUserRolesInApp(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRBACStore) AssignRoleToUserInApp(ctx context.Context, userID, roleID, assignedBy uuid.UUID, appID *uuid.UUID) error {
	return nil
}

// --- PermissionChecker implementation ---
func (m *mockRBACStore) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if m.hasPermissionFn != nil {
		return m.hasPermissionFn(ctx, userID, permissionName)
	}
	return false, nil
}

func (m *mockRBACStore) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	if m.hasAnyPermissionFn != nil {
		return m.hasAnyPermissionFn(ctx, userID, permissionNames)
	}
	return false, nil
}

func (m *mockRBACStore) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	if m.hasAllPermissionsFn != nil {
		return m.hasAllPermissionsFn(ctx, userID, permissionNames)
	}
	return false, nil
}

func (m *mockRBACStore) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	return nil, nil
}

func (m *mockRBACStore) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	return nil, nil
}

func (m *mockRBACStore) HasPermissionInApp(ctx context.Context, userID uuid.UUID, permissionName string, appID *uuid.UUID) (bool, error) {
	return false, nil
}

// mockAuditLogger satisfies the service.AuditLogger interface.
type mockAuditLogger struct{}

func (m *mockAuditLogger) LogWithAction(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
}
func (m *mockAuditLogger) Log(params service.AuditLogParams) {}

// newTestRBACMiddleware creates an RBACMiddleware backed by the mock store.
func newTestRBACMiddleware(store *mockRBACStore) *RBACMiddleware {
	rbacService := service.NewRBACService(store, &mockAuditLogger{})
	return NewRBACMiddleware(rbacService)
}

// --- RequirePermission tests ---

func TestRequirePermission_ShouldAllow_WhenUserHasPermission(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasPermissionFn: func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			if uid == userID && perm == "users.delete" {
				return true, nil
			}
			return false, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequirePermission("users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_ShouldReturn403_WhenUserLacksPermission(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasPermissionFn: func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequirePermission("users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

func TestRequirePermission_ShouldReturn401_WhenNoUserInContext(t *testing.T) {
	store := &mockRBACStore{}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(mw.RequirePermission("users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_ShouldReturn500_WhenPermissionCheckErrors(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasPermissionFn: func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, errors.New("database error")
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequirePermission("users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to check permissions")
}

// --- RequireAnyPermission tests ---

func TestRequireAnyPermission_ShouldAllow_WhenUserHasOneOfPermissions(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAnyPermissionFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return true, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAnyPermission("users.read", "users.write"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyPermission_ShouldReturn403_WhenUserHasNoneOfPermissions(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAnyPermissionFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAnyPermission("admin.all", "users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAnyPermission_ShouldReturn401_WhenNoUserInContext(t *testing.T) {
	store := &mockRBACStore{}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(mw.RequireAnyPermission("users.read"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAnyPermission_ShouldReturn500_WhenPermissionCheckErrors(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAnyPermissionFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, errors.New("database error")
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAnyPermission("users.read"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- RequireAllPermissions tests ---

func TestRequireAllPermissions_ShouldAllow_WhenUserHasAllPermissions(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAllPermissionsFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return true, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAllPermissions("users.read", "users.write"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAllPermissions_ShouldReturn403_WhenUserLacksSomePermissions(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAllPermissionsFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, nil
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAllPermissions("users.read", "users.delete"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAllPermissions_ShouldReturn401_WhenNoUserInContext(t *testing.T) {
	store := &mockRBACStore{}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(mw.RequireAllPermissions("users.read", "users.write"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAllPermissions_ShouldReturn500_WhenPermissionCheckErrors(t *testing.T) {
	userID := uuid.New()
	store := &mockRBACStore{
		hasAllPermissionsFn: func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, errors.New("connection refused")
		},
	}
	mw := newTestRBACMiddleware(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		c.Next()
	})
	r.Use(mw.RequireAllPermissions("users.read"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
