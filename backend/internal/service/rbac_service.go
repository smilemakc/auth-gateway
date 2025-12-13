package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

type RBACService struct {
	rbacRepo     RBACStore
	auditService AuditLogger
}

func NewRBACService(rbacRepo RBACStore, auditService AuditLogger) *RBACService {
	return &RBACService{
		rbacRepo:     rbacRepo,
		auditService: auditService,
	}
}

// ============================================================
// Permission Methods
// ============================================================

// CreatePermission creates a new permission
func (s *RBACService) CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*models.Permission, error) {
	// Check if permission already exists
	existing, err := s.rbacRepo.GetPermissionByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("permission with name %s already exists", req.Name)
	}

	permission := &models.Permission{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	err = s.rbacRepo.CreatePermission(ctx, permission)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (s *RBACService) GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	return s.rbacRepo.GetPermissionByID(ctx, id)
}

// ListPermissions retrieves all permissions
func (s *RBACService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.rbacRepo.ListPermissions(ctx)
}

// UpdatePermission updates a permission
func (s *RBACService) UpdatePermission(ctx context.Context, id uuid.UUID, req *models.UpdatePermissionRequest) error {
	return s.rbacRepo.UpdatePermission(ctx, id, req.Description)
}

// DeletePermission deletes a permission
func (s *RBACService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.rbacRepo.DeletePermission(ctx, id)
}

// ============================================================
// Role Methods
// ============================================================

// CreateRole creates a new role
func (s *RBACService) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	// Check if role already exists
	existing, err := s.rbacRepo.GetRoleByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name %s already exists", req.Name)
	}

	role := &models.Role{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		IsSystemRole: false,
	}

	err = s.rbacRepo.CreateRole(ctx, role)
	if err != nil {
		return nil, err
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		err = s.rbacRepo.SetRolePermissions(ctx, role.ID, req.Permissions)
		if err != nil {
			return nil, err
		}
	}

	// Fetch role with permissions
	return s.rbacRepo.GetRoleByID(ctx, role.ID)
}

// GetRole retrieves a role by ID
func (s *RBACService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return s.rbacRepo.GetRoleByID(ctx, id)
}

// GetRoleByName retrieves a role by name
func (s *RBACService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return s.rbacRepo.GetRoleByName(ctx, name)
}

// ListRoles retrieves all roles
func (s *RBACService) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.rbacRepo.ListRoles(ctx)
}

// UpdateRole updates a role
func (s *RBACService) UpdateRole(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	err := s.rbacRepo.UpdateRole(ctx, id, req.DisplayName, req.Description)
	if err != nil {
		return nil, err
	}

	// Update permissions if provided
	if req.Permissions != nil {
		err = s.rbacRepo.SetRolePermissions(ctx, id, req.Permissions)
		if err != nil {
			return nil, err
		}
	}

	return s.rbacRepo.GetRoleByID(ctx, id)
}

// DeleteRole deletes a role
func (s *RBACService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.rbacRepo.DeleteRole(ctx, id)
}

// SetRolePermissions sets permissions for a role
func (s *RBACService) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return s.rbacRepo.SetRolePermissions(ctx, roleID, permissionIDs)
}

// ============================================================
// Permission Checking
// ============================================================

// CheckUserPermission checks if a user has a specific permission
func (s *RBACService) CheckUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	return s.rbacRepo.HasPermission(ctx, userID, permission)
}

// CheckUserAnyPermission checks if a user has any of the specified permissions
func (s *RBACService) CheckUserAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	return s.rbacRepo.HasAnyPermission(ctx, userID, permissions)
}

// CheckUserAllPermissions checks if a user has all of the specified permissions
func (s *RBACService) CheckUserAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	return s.rbacRepo.HasAllPermissions(ctx, userID, permissions)
}

// GetUserPermissions retrieves all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	return s.rbacRepo.GetUserPermissions(ctx, userID)
}

// GetPermissionMatrix retrieves the permission matrix for all roles
func (s *RBACService) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	return s.rbacRepo.GetPermissionMatrix(ctx)
}

// ============================================================
// User-Role Management
// ============================================================

// AssignRoleToUser assigns a role to a user with validation
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	role, err := s.rbacRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := s.rbacRepo.AssignRoleToUser(ctx, userID, roleID, assignedBy); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	details := map[string]interface{}{
		"user_id":       userID.String(),
		"role_id":       roleID.String(),
		"role_name":     role.Name,
		"assigned_by":   assignedBy.String(),
		"resource_type": "user_role",
		"resource_id":   userID.String(),
	}
	s.auditService.Log(AuditLogParams{
		UserID:  &userID,
		Action:  models.ActionRoleAssigned,
		Status:  models.StatusSuccess,
		Details: details,
	})

	return nil
}

// RemoveRoleFromUser removes a role from a user with validation
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	role, err := s.rbacRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Name == string(models.RoleAdmin) {
		users, err := s.rbacRepo.GetUsersWithRole(ctx, roleID)
		if err == nil && len(users) == 1 && users[0].ID == userID {
			return models.NewAppError(400, "Cannot remove admin role: this is the last administrator")
		}
	}

	if err := s.rbacRepo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	details := map[string]interface{}{
		"user_id":       userID.String(),
		"role_id":       roleID.String(),
		"role_name":     role.Name,
		"resource_type": "user_role",
		"resource_id":   userID.String(),
	}
	s.auditService.Log(AuditLogParams{
		UserID:  &userID,
		Action:  models.ActionRoleRevoked,
		Status:  models.StatusSuccess,
		Details: details,
	})

	return nil
}

// SetUserRoles replaces all user roles atomically with validation
func (s *RBACService) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	previousRoles, _ := s.rbacRepo.GetUserRoles(ctx, userID)
	previousRoleNames := make([]string, len(previousRoles))
	for i, r := range previousRoles {
		previousRoleNames[i] = r.Name
	}

	newRoleNames := make([]string, len(roleIDs))
	hasAdmin := false
	for i, roleID := range roleIDs {
		role, err := s.rbacRepo.GetRoleByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("role %s not found", roleID)
		}
		newRoleNames[i] = role.Name
		if role.Name == string(models.RoleAdmin) {
			hasAdmin = true
		}
	}

	if !hasAdmin {
		userWasAdmin := false
		for _, r := range previousRoles {
			if r.Name == string(models.RoleAdmin) {
				userWasAdmin = true
				break
			}
		}

		if userWasAdmin {
			adminRole, err := s.rbacRepo.GetRoleByName(ctx, string(models.RoleAdmin))
			if err == nil {
				users, err := s.rbacRepo.GetUsersWithRole(ctx, adminRole.ID)
				if err == nil && len(users) == 1 && users[0].ID == userID {
					return models.NewAppError(400, "Cannot remove admin role: this is the last administrator")
				}
			}
		}
	}

	if err := s.rbacRepo.SetUserRoles(ctx, userID, roleIDs, assignedBy); err != nil {
		return fmt.Errorf("failed to set user roles: %w", err)
	}

	details := map[string]interface{}{
		"user_id":        userID.String(),
		"previous_roles": previousRoleNames,
		"new_roles":      newRoleNames,
		"assigned_by":    assignedBy.String(),
		"resource_type":  "user_role",
		"resource_id":    userID.String(),
	}
	s.auditService.Log(AuditLogParams{
		UserID:  &userID,
		Action:  models.ActionRolesUpdated,
		Status:  models.StatusSuccess,
		Details: details,
	})

	return nil
}

// GetUserRoles returns all roles for a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	roles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}
