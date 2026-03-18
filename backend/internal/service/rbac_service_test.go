package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Permission CRUD Tests
// ============================================================

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
			assert.Equal(t, req.Resource, permission.Resource)
			assert.Equal(t, req.Action, permission.Action)
			assert.Equal(t, req.Description, permission.Description)
			permission.ID = uuid.New()
			return nil
		}

		p, err := svc.CreatePermission(ctx, req)
		assert.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, req.Name, p.Name)
		assert.Equal(t, req.Resource, p.Resource)
		assert.Equal(t, req.Action, p.Action)
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

	t.Run("CreateFails", func(t *testing.T) {
		req := &models.CreatePermissionRequest{
			Name:     "new-perm",
			Resource: "resource",
			Action:   "write",
		}

		mockRBAC.GetPermissionByNameFunc = func(ctx context.Context, name string) (*models.Permission, error) {
			return nil, errors.New("not found")
		}
		mockRBAC.CreatePermissionFunc = func(ctx context.Context, permission *models.Permission) error {
			return errors.New("db error")
		}

		p, err := svc.CreatePermission(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, "db error", err.Error())
	})
}

func TestRBACService_GetPermission(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		permID := uuid.New()
		mockRBAC.GetPermissionByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
			assert.Equal(t, permID, id)
			return &models.Permission{ID: permID, Name: "users.read", Resource: "users", Action: "read"}, nil
		}

		p, err := svc.GetPermission(ctx, permID)
		assert.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, permID, p.ID)
		assert.Equal(t, "users.read", p.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRBAC.GetPermissionByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
			return nil, errors.New("not found")
		}

		p, err := svc.GetPermission(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestRBACService_ListPermissions(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Permission{
			{ID: uuid.New(), Name: "users.read"},
			{ID: uuid.New(), Name: "users.write"},
		}
		mockRBAC.ListPermissionsFunc = func(ctx context.Context) ([]models.Permission, error) {
			return expected, nil
		}

		perms, err := svc.ListPermissions(ctx)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
		assert.Equal(t, "users.read", perms[0].Name)
		assert.Equal(t, "users.write", perms[1].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		mockRBAC.ListPermissionsFunc = func(ctx context.Context) ([]models.Permission, error) {
			return []models.Permission{}, nil
		}

		perms, err := svc.ListPermissions(ctx)
		assert.NoError(t, err)
		assert.Empty(t, perms)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.ListPermissionsFunc = func(ctx context.Context) ([]models.Permission, error) {
			return nil, errors.New("db error")
		}

		perms, err := svc.ListPermissions(ctx)
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

func TestRBACService_UpdatePermission(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	permID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		req := &models.UpdatePermissionRequest{Description: "updated desc"}
		mockRBAC.UpdatePermissionFunc = func(ctx context.Context, id uuid.UUID, description string) error {
			assert.Equal(t, permID, id)
			assert.Equal(t, "updated desc", description)
			return nil
		}

		err := svc.UpdatePermission(ctx, permID, req)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		req := &models.UpdatePermissionRequest{Description: "updated"}
		mockRBAC.UpdatePermissionFunc = func(ctx context.Context, id uuid.UUID, description string) error {
			return errors.New("not found")
		}

		err := svc.UpdatePermission(ctx, permID, req)
		assert.Error(t, err)
	})
}

func TestRBACService_DeletePermission(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		permID := uuid.New()
		mockRBAC.DeletePermissionFunc = func(ctx context.Context, id uuid.UUID) error {
			assert.Equal(t, permID, id)
			return nil
		}

		err := svc.DeletePermission(ctx, permID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.DeletePermissionFunc = func(ctx context.Context, id uuid.UUID) error {
			return errors.New("permission in use")
		}

		err := svc.DeletePermission(ctx, uuid.New())
		assert.Error(t, err)
	})
}

// ============================================================
// Role CRUD Tests
// ============================================================

func TestRBACService_CreateRole(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &models.CreateRoleRequest{
			Name:        "test-role",
			DisplayName: "Test Role",
			Description: "A test role",
			Permissions: []uuid.UUID{uuid.New()},
		}

		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, nil
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			assert.Equal(t, req.Name, role.Name)
			assert.Equal(t, req.DisplayName, role.DisplayName)
			assert.Equal(t, req.Description, role.Description)
			assert.False(t, role.IsSystemRole)
			role.ID = uuid.New()
			return nil
		}
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
			assert.Len(t, permissionIDs, 1)
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: req.Name, DisplayName: req.DisplayName}, nil
		}

		r, err := svc.CreateRole(ctx, req)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, req.Name, r.Name)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		req := &models.CreateRoleRequest{Name: "existing-role"}
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: name}, nil
		}

		r, err := svc.CreateRole(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, r)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("CreateFails", func(t *testing.T) {
		req := &models.CreateRoleRequest{Name: "fail-role"}
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, errors.New("not found")
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			return errors.New("db error")
		}

		r, err := svc.CreateRole(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("SetPermissionsFails", func(t *testing.T) {
		permID := uuid.New()
		req := &models.CreateRoleRequest{
			Name:        "role-with-perms",
			Permissions: []uuid.UUID{permID},
		}
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, errors.New("not found")
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			role.ID = uuid.New()
			return nil
		}
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
			return errors.New("invalid permission ID")
		}

		r, err := svc.CreateRole(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, r)
		assert.Equal(t, "invalid permission ID", err.Error())
	})

	t.Run("NoPermissions", func(t *testing.T) {
		req := &models.CreateRoleRequest{
			Name:        "role-no-perms",
			DisplayName: "No Perms",
		}
		roleID := uuid.New()

		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, errors.New("not found")
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			role.ID = roleID
			return nil
		}
		setPermsCalled := false
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, permissionIDs []uuid.UUID) error {
			setPermsCalled = true
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: req.Name}, nil
		}

		r, err := svc.CreateRole(ctx, req)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.False(t, setPermsCalled, "SetRolePermissions should not be called when no permissions provided")
	})
}

func TestRBACService_GetRole(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		roleID := uuid.New()
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: roleID, Name: "admin", DisplayName: "Administrator"}, nil
		}

		r, err := svc.GetRole(ctx, roleID)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, roleID, r.ID)
		assert.Equal(t, "admin", r.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		r, err := svc.GetRole(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}

func TestRBACService_GetRoleByName(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: name}, nil
		}

		r, err := svc.GetRoleByName(ctx, "moderator")
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "moderator", r.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		r, err := svc.GetRoleByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}

func TestRBACService_ListRoles(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Role{
			{ID: uuid.New(), Name: "admin"},
			{ID: uuid.New(), Name: "user"},
			{ID: uuid.New(), Name: "moderator"},
		}
		mockRBAC.ListRolesFunc = func(ctx context.Context) ([]models.Role, error) {
			return expected, nil
		}

		roles, err := svc.ListRoles(ctx)
		assert.NoError(t, err)
		assert.Len(t, roles, 3)
	})

	t.Run("Empty", func(t *testing.T) {
		mockRBAC.ListRolesFunc = func(ctx context.Context) ([]models.Role, error) {
			return []models.Role{}, nil
		}

		roles, err := svc.ListRoles(ctx)
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.ListRolesFunc = func(ctx context.Context) ([]models.Role, error) {
			return nil, errors.New("db error")
		}

		roles, err := svc.ListRoles(ctx)
		assert.Error(t, err)
		assert.Nil(t, roles)
	})
}

func TestRBACService_UpdateRole(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	roleID := uuid.New()

	t.Run("Success_WithPermissions", func(t *testing.T) {
		permIDs := []uuid.UUID{uuid.New(), uuid.New()}
		req := &models.UpdateRoleRequest{
			DisplayName: "Updated Name",
			Description: "Updated desc",
			Permissions: permIDs,
		}

		mockRBAC.UpdateRoleFunc = func(ctx context.Context, id uuid.UUID, displayName, description string) error {
			assert.Equal(t, roleID, id)
			assert.Equal(t, "Updated Name", displayName)
			assert.Equal(t, "Updated desc", description)
			return nil
		}
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, pIDs []uuid.UUID) error {
			assert.Equal(t, roleID, rID)
			assert.Len(t, pIDs, 2)
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "role", DisplayName: "Updated Name"}, nil
		}

		r, err := svc.UpdateRole(ctx, roleID, req)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "Updated Name", r.DisplayName)
	})

	t.Run("Success_WithoutPermissions", func(t *testing.T) {
		req := &models.UpdateRoleRequest{
			DisplayName: "Only Name",
			Description: "Only desc",
			Permissions: nil, // nil means don't update permissions
		}

		mockRBAC.UpdateRoleFunc = func(ctx context.Context, id uuid.UUID, displayName, description string) error {
			return nil
		}
		setPermsCalled := false
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, pIDs []uuid.UUID) error {
			setPermsCalled = true
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "role", DisplayName: "Only Name"}, nil
		}

		r, err := svc.UpdateRole(ctx, roleID, req)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.False(t, setPermsCalled)
	})

	t.Run("UpdateFails", func(t *testing.T) {
		req := &models.UpdateRoleRequest{DisplayName: "fail"}
		mockRBAC.UpdateRoleFunc = func(ctx context.Context, id uuid.UUID, displayName, description string) error {
			return errors.New("not found")
		}

		r, err := svc.UpdateRole(ctx, roleID, req)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("SetPermissionsFails", func(t *testing.T) {
		req := &models.UpdateRoleRequest{
			DisplayName: "test",
			Permissions: []uuid.UUID{uuid.New()},
		}
		mockRBAC.UpdateRoleFunc = func(ctx context.Context, id uuid.UUID, displayName, description string) error {
			return nil
		}
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, pIDs []uuid.UUID) error {
			return errors.New("invalid permission")
		}

		r, err := svc.UpdateRole(ctx, roleID, req)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}

func TestRBACService_DeleteRole(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		roleID := uuid.New()
		mockRBAC.DeleteRoleFunc = func(ctx context.Context, id uuid.UUID) error {
			assert.Equal(t, roleID, id)
			return nil
		}

		err := svc.DeleteRole(ctx, roleID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.DeleteRoleFunc = func(ctx context.Context, id uuid.UUID) error {
			return errors.New("role has assigned users")
		}

		err := svc.DeleteRole(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestRBACService_SetRolePermissions(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		roleID := uuid.New()
		permIDs := []uuid.UUID{uuid.New(), uuid.New()}

		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, pIDs []uuid.UUID) error {
			assert.Equal(t, roleID, rID)
			assert.Len(t, pIDs, 2)
			return nil
		}

		err := svc.SetRolePermissions(ctx, roleID, permIDs)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.SetRolePermissionsFunc = func(ctx context.Context, rID uuid.UUID, pIDs []uuid.UUID) error {
			return errors.New("invalid permission")
		}

		err := svc.SetRolePermissions(ctx, uuid.New(), []uuid.UUID{uuid.New()})
		assert.Error(t, err)
	})
}

// ============================================================
// User-Role Assignment Tests
// ============================================================

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
			return &models.Role{ID: id, Name: "editor"}, nil
		}
		mockRBAC.AssignRoleToUserFunc = func(ctx context.Context, uid, rid, ab uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, roleID, rid)
			assert.Equal(t, adminID, ab)
			return nil
		}
		auditLogged := false
		mockAudit.LogFunc = func(params AuditLogParams) {
			auditLogged = true
			assert.Equal(t, models.ActionRoleAssigned, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
			assert.Equal(t, &userID, params.UserID)
			assert.Equal(t, userID.String(), params.Details["user_id"])
			assert.Equal(t, roleID.String(), params.Details["role_id"])
			assert.Equal(t, "editor", params.Details["role_name"])
			assert.Equal(t, adminID.String(), params.Details["assigned_by"])
		}

		err := svc.AssignRoleToUser(ctx, userID, roleID, adminID)
		assert.NoError(t, err)
		assert.True(t, auditLogged)
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		err := svc.AssignRoleToUser(ctx, userID, roleID, adminID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role not found")
	})

	t.Run("AssignFails", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "editor"}, nil
		}
		mockRBAC.AssignRoleToUserFunc = func(ctx context.Context, uid, rid, ab uuid.UUID) error {
			return errors.New("already assigned")
		}

		err := svc.AssignRoleToUser(ctx, userID, roleID, adminID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to assign role")
	})
}

func TestRBACService_RemoveRoleFromUser(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "editor"}, nil
		}
		mockRBAC.RemoveRoleFromUserFunc = func(ctx context.Context, uid, rid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, roleID, rid)
			return nil
		}
		auditLogged := false
		mockAudit.LogFunc = func(params AuditLogParams) {
			auditLogged = true
			assert.Equal(t, models.ActionRoleRevoked, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		err := svc.RemoveRoleFromUser(ctx, userID, roleID)
		assert.NoError(t, err)
		assert.True(t, auditLogged)
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		err := svc.RemoveRoleFromUser(ctx, userID, roleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role not found")
	})

	t.Run("RemoveFails", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "editor"}, nil
		}
		mockRBAC.RemoveRoleFromUserFunc = func(ctx context.Context, uid, rid uuid.UUID) error {
			return errors.New("not assigned")
		}

		err := svc.RemoveRoleFromUser(ctx, userID, roleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove role")
	})

	t.Run("PreventLastAdminRemoval", func(t *testing.T) {
		adminRoleID := uuid.New()
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: string(models.RoleAdmin)}, nil
		}
		mockRBAC.GetUsersWithRoleFunc = func(ctx context.Context, rid uuid.UUID) ([]models.User, error) {
			return []models.User{{ID: userID}}, nil // Only one admin
		}

		err := svc.RemoveRoleFromUser(ctx, userID, adminRoleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "last administrator")
	})

	t.Run("AllowAdminRemovalWhenMultipleAdminsExist", func(t *testing.T) {
		adminRoleID := uuid.New()
		otherAdminID := uuid.New()

		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: string(models.RoleAdmin)}, nil
		}
		mockRBAC.GetUsersWithRoleFunc = func(ctx context.Context, rid uuid.UUID) ([]models.User, error) {
			return []models.User{{ID: userID}, {ID: otherAdminID}}, nil // Multiple admins
		}
		mockRBAC.RemoveRoleFromUserFunc = func(ctx context.Context, uid, rid uuid.UUID) error {
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {}

		err := svc.RemoveRoleFromUser(ctx, userID, adminRoleID)
		assert.NoError(t, err)
	})
}

// ============================================================
// Permission Checking Tests
// ============================================================

func TestRBACService_PermissionChecks(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("CheckUserPermission_HasPermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, "users.read", perm)
			return true, nil
		}

		allowed, err := svc.CheckUserPermission(ctx, userID, "users.read")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("CheckUserPermission_DoesNotHavePermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, nil
		}

		allowed, err := svc.CheckUserPermission(ctx, userID, "users.delete")
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("CheckUserPermission_Error", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, errors.New("db error")
		}

		allowed, err := svc.CheckUserPermission(ctx, userID, "users.read")
		assert.Error(t, err)
		assert.False(t, allowed)
	})

	t.Run("CheckUserAnyPermission_HasOne", func(t *testing.T) {
		mockRBAC.HasAnyPermissionFunc = func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			assert.Equal(t, []string{"users.read", "users.write"}, perms)
			return true, nil
		}

		allowed, err := svc.CheckUserAnyPermission(ctx, userID, []string{"users.read", "users.write"})
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("CheckUserAnyPermission_HasNone", func(t *testing.T) {
		mockRBAC.HasAnyPermissionFunc = func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, nil
		}

		allowed, err := svc.CheckUserAnyPermission(ctx, userID, []string{"admin.read", "admin.write"})
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("CheckUserAllPermissions_HasAll", func(t *testing.T) {
		mockRBAC.HasAllPermissionsFunc = func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			assert.Equal(t, []string{"users.read", "users.write"}, perms)
			return true, nil
		}

		allowed, err := svc.CheckUserAllPermissions(ctx, userID, []string{"users.read", "users.write"})
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("CheckUserAllPermissions_MissingSome", func(t *testing.T) {
		mockRBAC.HasAllPermissionsFunc = func(ctx context.Context, uid uuid.UUID, perms []string) (bool, error) {
			return false, nil
		}

		allowed, err := svc.CheckUserAllPermissions(ctx, userID, []string{"users.read", "admin.delete"})
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRBACService_GetUserPermissions(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Permission{
			{ID: uuid.New(), Name: "users.read"},
			{ID: uuid.New(), Name: "users.write"},
		}
		mockRBAC.GetUserPermissionsFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Permission, error) {
			assert.Equal(t, userID, uid)
			return expected, nil
		}

		perms, err := svc.GetUserPermissions(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})

	t.Run("Empty", func(t *testing.T) {
		mockRBAC.GetUserPermissionsFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Permission, error) {
			return []models.Permission{}, nil
		}

		perms, err := svc.GetUserPermissions(ctx, userID)
		assert.NoError(t, err)
		assert.Empty(t, perms)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.GetUserPermissionsFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Permission, error) {
			return nil, errors.New("db error")
		}

		perms, err := svc.GetUserPermissions(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

func TestRBACService_GetPermissionMatrix(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expected := &models.PermissionMatrix{
			Resources: []models.ResourcePermissions{
				{Resource: "users"},
			},
		}
		mockRBAC.GetPermissionMatrixFunc = func(ctx context.Context) (*models.PermissionMatrix, error) {
			return expected, nil
		}

		matrix, err := svc.GetPermissionMatrix(ctx)
		assert.NoError(t, err)
		require.NotNil(t, matrix)
		assert.Len(t, matrix.Resources, 1)
		assert.Equal(t, "users", matrix.Resources[0].Resource)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.GetPermissionMatrixFunc = func(ctx context.Context) (*models.PermissionMatrix, error) {
			return nil, errors.New("db error")
		}

		matrix, err := svc.GetPermissionMatrix(ctx)
		assert.Error(t, err)
		assert.Nil(t, matrix)
	})
}

// ============================================================
// GetUserRoles Tests
// ============================================================

func TestRBACService_GetUserRoles(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Role{
			{ID: uuid.New(), Name: "admin"},
			{ID: uuid.New(), Name: "editor"},
		}
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			assert.Equal(t, userID, uid)
			return expected, nil
		}

		roles, err := svc.GetUserRoles(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
		assert.Equal(t, "admin", roles[0].Name)
		assert.Equal(t, "editor", roles[1].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{}, nil
		}

		roles, err := svc.GetUserRoles(ctx, userID)
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("Error", func(t *testing.T) {
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return nil, errors.New("db error")
		}

		roles, err := svc.GetUserRoles(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, roles)
		assert.Contains(t, err.Error(), "failed to get user roles")
	})
}

// ============================================================
// SetUserRoles Tests
// ============================================================

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
			assert.Equal(t, userID, uid)
			assert.Equal(t, adminID, ab)
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRolesUpdated, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
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

	t.Run("RoleNotFound", func(t *testing.T) {
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{}, nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		err := svc.SetUserRoles(ctx, userID, []uuid.UUID{uuid.New()}, adminID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("SetFails", func(t *testing.T) {
		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{}, nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "user"}, nil
		}
		mockRBAC.SetUserRolesFunc = func(ctx context.Context, uid uuid.UUID, rIDs []uuid.UUID, ab uuid.UUID) error {
			return errors.New("db error")
		}

		err := svc.SetUserRoles(ctx, userID, []uuid.UUID{roleID}, adminID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set user roles")
	})

	t.Run("AllowAdminRemovalWhenMultipleAdminsExist", func(t *testing.T) {
		otherAdminID := uuid.New()
		nonAdminRoleID := uuid.New()

		mockRBAC.GetUserRolesFunc = func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{{ID: adminRoleID, Name: string(models.RoleAdmin)}}, nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "user"}, nil
		}
		mockRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: adminRoleID, Name: string(models.RoleAdmin)}, nil
		}
		mockRBAC.GetUsersWithRoleFunc = func(ctx context.Context, rid uuid.UUID) ([]models.User, error) {
			return []models.User{{ID: userID}, {ID: otherAdminID}}, nil
		}
		mockRBAC.SetUserRolesFunc = func(ctx context.Context, uid uuid.UUID, rIDs []uuid.UUID, ab uuid.UUID) error {
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {}

		err := svc.SetUserRoles(ctx, userID, []uuid.UUID{nonAdminRoleID}, adminID)
		assert.NoError(t, err)
	})
}

// ============================================================
// Application-Scoped RBAC Tests
// ============================================================

func TestRBACService_CreateRoleInApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	appID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetRoleByNameAndAppFunc = func(ctx context.Context, name string, aID *uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}
		mockRBAC.CreateRoleFunc = func(ctx context.Context, role *models.Role) error {
			assert.Equal(t, "app-role", role.Name)
			assert.Equal(t, &appID, role.ApplicationID)
			assert.False(t, role.IsSystemRole)
			role.ID = uuid.New()
			return nil
		}
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "app-role", ApplicationID: &appID}, nil
		}

		r, err := svc.CreateRoleInApp(ctx, "app-role", "App Role", "desc", &appID)
		assert.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "app-role", r.Name)
		assert.Equal(t, &appID, r.ApplicationID)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		mockRBAC.GetRoleByNameAndAppFunc = func(ctx context.Context, name string, aID *uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: name}, nil
		}

		r, err := svc.CreateRoleInApp(ctx, "existing", "Existing", "desc", &appID)
		assert.Error(t, err)
		assert.Nil(t, r)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestRBACService_HasPermissionInApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	userID := uuid.New()
	appID := uuid.New()

	t.Run("HasGlobalPermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return true, nil
		}

		allowed, err := svc.HasPermissionInApp(ctx, userID, "users.read", &appID)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("HasAppPermission_NotGlobal", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, nil
		}
		mockRBAC.HasPermissionInAppFunc = func(ctx context.Context, uid uuid.UUID, perm string, aID *uuid.UUID) (bool, error) {
			assert.Equal(t, &appID, aID)
			return true, nil
		}

		allowed, err := svc.HasPermissionInApp(ctx, userID, "users.read", &appID)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("NoPermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, nil
		}
		mockRBAC.HasPermissionInAppFunc = func(ctx context.Context, uid uuid.UUID, perm string, aID *uuid.UUID) (bool, error) {
			return false, nil
		}

		allowed, err := svc.HasPermissionInApp(ctx, userID, "admin.delete", &appID)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("NilAppID_NoGlobalPermission", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, nil
		}

		allowed, err := svc.HasPermissionInApp(ctx, userID, "users.read", nil)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("GlobalCheckError", func(t *testing.T) {
		mockRBAC.HasPermissionFunc = func(ctx context.Context, uid uuid.UUID, perm string) (bool, error) {
			return false, errors.New("db error")
		}

		allowed, err := svc.HasPermissionInApp(ctx, userID, "users.read", &appID)
		assert.Error(t, err)
		assert.False(t, allowed)
	})
}

func TestRBACService_ListRolesByApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	appID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Role{{ID: uuid.New(), Name: "app-admin"}}
		mockRBAC.ListRolesByAppFunc = func(ctx context.Context, aID *uuid.UUID) ([]models.Role, error) {
			assert.Equal(t, &appID, aID)
			return expected, nil
		}

		roles, err := svc.ListRolesByApp(ctx, &appID)
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
	})
}

func TestRBACService_ListPermissionsByApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	appID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Permission{{ID: uuid.New(), Name: "app.read"}}
		mockRBAC.ListPermissionsByAppFunc = func(ctx context.Context, aID *uuid.UUID) ([]models.Permission, error) {
			assert.Equal(t, &appID, aID)
			return expected, nil
		}

		perms, err := svc.ListPermissionsByApp(ctx, &appID)
		assert.NoError(t, err)
		assert.Len(t, perms, 1)
	})
}

func TestRBACService_GetUserRolesInApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := []models.Role{{ID: uuid.New(), Name: "app-user"}}
		mockRBAC.GetUserRolesInAppFunc = func(ctx context.Context, uid uuid.UUID, aID *uuid.UUID) ([]models.Role, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, &appID, aID)
			return expected, nil
		}

		roles, err := svc.GetUserRolesInApp(ctx, userID, &appID)
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
		assert.Equal(t, "app-user", roles[0].Name)
	})
}

func TestRBACService_AssignRoleToUserInApp(t *testing.T) {
	mockRBAC := &mockRBACStore{}
	mockAudit := &mockAuditLogger{}
	svc := NewRBACService(mockRBAC, mockAudit)
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()
	adminID := uuid.New()
	appID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "app-editor"}, nil
		}
		mockRBAC.AssignRoleToUserInAppFunc = func(ctx context.Context, uid, rid, ab uuid.UUID, aID *uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, roleID, rid)
			assert.Equal(t, adminID, ab)
			assert.Equal(t, &appID, aID)
			return nil
		}
		auditLogged := false
		mockAudit.LogFunc = func(params AuditLogParams) {
			auditLogged = true
			assert.Equal(t, models.ActionRoleAssigned, params.Action)
			assert.Equal(t, appID.String(), params.Details["application_id"])
		}

		err := svc.AssignRoleToUserInApp(ctx, userID, roleID, adminID, &appID)
		assert.NoError(t, err)
		assert.True(t, auditLogged)
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return nil, errors.New("not found")
		}

		err := svc.AssignRoleToUserInApp(ctx, userID, roleID, adminID, &appID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role not found")
	})

	t.Run("AssignFails", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "role"}, nil
		}
		mockRBAC.AssignRoleToUserInAppFunc = func(ctx context.Context, uid, rid, ab uuid.UUID, aID *uuid.UUID) error {
			return errors.New("already assigned")
		}

		err := svc.AssignRoleToUserInApp(ctx, userID, roleID, adminID, &appID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to assign role")
	})

	t.Run("NilAppID_NoAppInAuditDetails", func(t *testing.T) {
		mockRBAC.GetRoleByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Role, error) {
			return &models.Role{ID: id, Name: "global-role"}, nil
		}
		mockRBAC.AssignRoleToUserInAppFunc = func(ctx context.Context, uid, rid, ab uuid.UUID, aID *uuid.UUID) error {
			return nil
		}
		mockAudit.LogFunc = func(params AuditLogParams) {
			_, hasAppID := params.Details["application_id"]
			assert.False(t, hasAppID, "audit log should not contain application_id when nil")
		}

		err := svc.AssignRoleToUserInApp(ctx, userID, roleID, adminID, nil)
		assert.NoError(t, err)
	})
}
