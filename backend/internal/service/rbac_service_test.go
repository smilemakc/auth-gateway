package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRBACService_CreatePermission(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &models.CreatePermissionRequest{
			Name:        "test-perm",
			Resource:    "resource",
			Action:      "read",
			Description: "desc",
		}

		mockRBAC.GetPermissionByNameFunc = func(ctx context.Context, name string) (*models.Permission, error) {
			return nil, nil // Not found
		}
		mockRBAC.CreatePermissionFunc = func(ctx context.Context, permission *models.Permission) error {
			assert.Equal(t, req.Name, permission.Name)
			permission.ID = uuid.New()
			return nil
		}

		p, err := svc.CreatePermission(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, req.Name, p.Name)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		req := &models.CreatePermissionRequest{Name: "existing"}
		mockRBAC.GetPermissionByNameFunc = func(ctx context.Context, name string) (*models.Permission, error) {
			return &models.Permission{ID: uuid.New()}, nil
		}

		p, err := svc.CreatePermission(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestRBACService_CreateRole(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &models.CreateRoleRequest{
			Name:        "test-role",
			Permissions: []uuid.UUID{uuid.New()},
		}

		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, nil
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			role.ID = uuid.New()
			return nil
		}
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: req.Name}, nil
		}

		r, err := svc.CreateRole(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, req.Name, r.Name)
	})
}

func TestRBACService_AssignRoleToUser(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()
	adminID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "role"}, nil
		}
		mockRBAC.AssignRoleToUserFunc = func(ctx context.Context, uid, rid, ab uuid.UUID) error {
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRoleAssigned, params.Action)
		}

		err := svc.AssignRoleToUser(ctx, userID, roleID, adminID)
		assert.NoError(t, err)
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}
		err := svc.AssignRoleToUser(ctx, userID, roleID, adminID)
		assert.Error(t, err)
	})
}

func TestRBACService_PermissionChecks(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("CheckUserPermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return true, nil
		}
		allowed, err := svc.CheckUserPermission(ctx, userID, "read")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestRBACService_SetUserRoles(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()
	adminID := uuid.New()
	adminRoleID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{}, nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "user"}, nil
		}
		mockRBAC.SetUserRolesFunc = func(ctx context.Context, uid uuid.UUID, rIDs []uuid.UUID, ab uuid.UUID) error {
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRolesUpdated, params.Action)
		}

		err := svc.SetUserRoles(ctx, userID, []uuid.UUID{roleID}, adminID)
		assert.NoError(t, err)
	})

	t.Run("PreventLastAdminRemoval", func(t *testing.T) {
		// User is currently admin
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{{ID: adminRoleID, Name: string(models.RoleAdmin)}}, nil
		}
		// New role is NOT admin
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "user"}, nil
		}
		// Get Admin Role to check ID
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: adminRoleID, Name: string(models.RoleAdmin)}, nil
		}
		// Only 1 admin exists (this user)
		mockRBAC.GetUsersWithRoleFunc = func(ctx context.Context, rid uuid.UUID) ([]models.User, error) {
			return []models.User{{ID: userID}}, nil
		}

		err := svc.SetUserRoles(ctx, userID, []uuid.UUID{roleID}, adminID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "last administrator")
	})
}
