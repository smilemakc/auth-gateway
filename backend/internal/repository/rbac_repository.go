package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// RBACRepository handles RBAC database operations
type RBACRepository struct {
	db *Database
}

// NewRBACRepository creates a new RBAC repository
func NewRBACRepository(db *Database) *RBACRepository {
	return &RBACRepository{db: db}
}

// ============================================================
// Permission Methods
// ============================================================

// CreatePermission creates a new permission
func (r *RBACRepository) CreatePermission(ctx context.Context, permission *models.Permission) error {
	_, err := r.db.NewInsert().
		Model(permission).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetPermissionByID retrieves a permission by ID
func (r *RBACRepository) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	permission := new(models.Permission)

	err := r.db.NewSelect().
		Model(permission).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by id: %w", err)
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (r *RBACRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	permission := new(models.Permission)

	err := r.db.NewSelect().
		Model(permission).
		Where("name = ?", name).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}

	return permission, nil
}

// ListPermissions retrieves all permissions
func (r *RBACRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	permissions := make([]models.Permission, 0)

	err := r.db.NewSelect().
		Model(&permissions).
		Order("resource", "action").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

// ListPermissionsByResource retrieves permissions for a specific resource
func (r *RBACRepository) ListPermissionsByResource(ctx context.Context, resource string) ([]models.Permission, error) {
	permissions := make([]models.Permission, 0)

	err := r.db.NewSelect().
		Model(&permissions).
		Where("resource = ?", resource).
		Order("action").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by resource: %w", err)
	}

	return permissions, nil
}

// UpdatePermission updates a permission
func (r *RBACRepository) UpdatePermission(ctx context.Context, id uuid.UUID, description string) error {
	result, err := r.db.NewUpdate().
		Model((*models.Permission)(nil)).
		Set("description = ?", description).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

// DeletePermission deletes a permission
func (r *RBACRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.Permission)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

// ============================================================
// Role Methods
// ============================================================

// CreateRole creates a new role
func (r *RBACRepository) CreateRole(ctx context.Context, role *models.Role) error {
	_, err := r.db.NewInsert().
		Model(role).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetRoleByID retrieves a role by ID with its permissions
func (r *RBACRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	role := new(models.Role)

	err := r.db.NewSelect().
		Model(role).
		Where("id = ?", id).
		Relation("Permissions").
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *RBACRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	role := new(models.Role)

	err := r.db.NewSelect().
		Model(role).
		Where("name = ?", name).
		Relation("Permissions").
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return role, nil
}

// ListRoles retrieves all roles
func (r *RBACRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles := make([]models.Role, 0)

	err := r.db.NewSelect().
		Model(&roles).
		Relation("Permissions").
		Order("is_system_role DESC", "name").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// UpdateRole updates a role
func (r *RBACRepository) UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error {
	result, err := r.db.NewUpdate().
		Model((*models.Role)(nil)).
		Set("display_name = ?", displayName).
		Set("description = ?", description).
		Set("updated_at = ?", bun.Ident("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// DeleteRole deletes a role (only if not a system role)
func (r *RBACRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.Role)(nil)).
		Where("id = ?", id).
		Where("is_system_role = ?", false).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("role not found or is a system role")
	}

	return nil
}

// ============================================================
// Role-Permission Methods
// ============================================================

// GetRolePermissions retrieves all permissions for a role
func (r *RBACRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	permissions := make([]models.Permission, 0)

	err := r.db.NewSelect().
		Model(&permissions).
		Join("INNER JOIN role_permissions AS rp ON rp.permission_id = permission.id").
		Where("rp.role_id = ?", roleID).
		Order("permission.resource", "permission.action").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// AddPermissionToRole adds a permission to a role
func (r *RBACRepository) AddPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	rolePermission := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	_, err := r.db.NewInsert().
		Model(rolePermission).
		On("CONFLICT (role_id, permission_id) DO NOTHING").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *RBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.RolePermission)(nil)).
		Where("role_id = ?", roleID).
		Where("permission_id = ?", permissionID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
}

// SetRolePermissions sets all permissions for a role (replaces existing)
func (r *RBACRepository) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete existing permissions
		_, err := tx.NewDelete().
			Model((*models.RolePermission)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete existing role permissions: %w", err)
		}

		// Add new permissions
		if len(permissionIDs) > 0 {
			rolePermissions := make([]*models.RolePermission, len(permissionIDs))
			for i, permID := range permissionIDs {
				rolePermissions[i] = &models.RolePermission{
					RoleID:       roleID,
					PermissionID: permID,
				}
			}

			_, err = tx.NewInsert().
				Model(&rolePermissions).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to insert new role permissions: %w", err)
			}
		}

		return nil
	})
}

// ============================================================
// User Permission Checking
// ============================================================

// GetUserPermissions retrieves all permissions for a user based on their roles
func (r *RBACRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	permissions := make([]models.Permission, 0)

	err := r.db.NewSelect().
		Model(&permissions).
		Distinct().
		Join("INNER JOIN role_permissions AS rp ON rp.permission_id = permission.id").
		Join("INNER JOIN roles AS r ON r.id = rp.role_id").
		Join("INNER JOIN user_roles AS ur ON ur.role_id = r.id").
		Join("INNER JOIN users AS u ON u.id = ur.user_id").
		Where("u.id = ?", userID).
		Order("permission.resource", "permission.action").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *RBACRepository) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.Permission)(nil)).
		Join("INNER JOIN role_permissions AS rp ON rp.permission_id = permission.id").
		Join("INNER JOIN roles AS r ON r.id = rp.role_id").
		Join("INNER JOIN user_roles AS ur ON ur.role_id = r.id").
		Join("INNER JOIN users AS u ON u.id = ur.user_id").
		Where("u.id = ?", userID).
		Where("permission.name = ?", permissionName).
		Count(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return count > 0, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (r *RBACRepository) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.Permission)(nil)).
		Join("INNER JOIN role_permissions AS rp ON rp.permission_id = permission.id").
		Join("INNER JOIN roles AS r ON r.id = rp.role_id").
		Join("INNER JOIN user_roles AS ur ON ur.role_id = r.id").
		Join("INNER JOIN users AS u ON u.id = ur.user_id").
		Where("u.id = ?", userID).
		Where("permission.name IN (?)", bun.In(permissionNames)).
		Count(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check any permission: %w", err)
	}

	return count > 0, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (r *RBACRepository) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.Permission)(nil)).
		ColumnExpr("COUNT(DISTINCT permission.name)").
		Join("INNER JOIN role_permissions AS rp ON rp.permission_id = permission.id").
		Join("INNER JOIN roles AS r ON r.id = rp.role_id").
		Join("INNER JOIN user_roles AS ur ON ur.role_id = r.id").
		Join("INNER JOIN users AS u ON u.id = ur.user_id").
		Where("u.id = ?", userID).
		Where("permission.name IN (?)", bun.In(permissionNames)).
		Count(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check all permissions: %w", err)
	}

	return count == len(permissionNames), nil
}

// ============================================================
// User-Role Management Methods
// ============================================================

// GetUserRoles returns all roles assigned to a user with their permissions
func (r *RBACRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.NewSelect().
		Model(&roles).
		Join("INNER JOIN user_roles AS ur ON ur.role_id = role.id").
		Where("ur.user_id = ?", userID).
		Relation("Permissions").
		Order("role.name").
		Scan(ctx)

	if err != nil {
		return nil, handlePgError(err)
	}
	return roles, nil
}

// AssignRoleToUser assigns a role to a user (idempotent - won't fail if already assigned)
func (r *RBACRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	userRole := &models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: &assignedBy,
	}

	_, err := r.db.NewInsert().
		Model(userRole).
		On("CONFLICT (user_id, role_id) DO NOTHING").
		Exec(ctx)

	return handlePgError(err)
}

// RemoveRoleFromUser removes a role from a user
func (r *RBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.UserRole)(nil)).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Exec(ctx)

	return handlePgError(err)
}

// SetUserRoles atomically replaces all user roles (transaction)
func (r *RBACRepository) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete all existing roles for this user
		_, err := tx.NewDelete().
			Model((*models.UserRole)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return handlePgError(err)
		}

		// Insert new roles if any
		if len(roleIDs) > 0 {
			userRoles := make([]models.UserRole, len(roleIDs))
			for i, roleID := range roleIDs {
				userRoles[i] = models.UserRole{
					UserID:     userID,
					RoleID:     roleID,
					AssignedBy: &assignedBy,
				}
			}

			_, err = tx.NewInsert().
				Model(&userRoles).
				Exec(ctx)
			if err != nil {
				return handlePgError(err)
			}
		}

		return nil
	})
}

// GetUsersWithRole returns all users with a specific role
func (r *RBACRepository) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.NewSelect().
		Model(&users).
		Join("INNER JOIN user_roles AS ur ON ur.user_id = users.id").
		Where("ur.role_id = ? AND users.is_active = ?", roleID, true).
		Scan(ctx)

	if err != nil {
		return nil, handlePgError(err)
	}
	return users, nil
}

// GetPermissionMatrix retrieves the complete permission matrix for all roles
func (r *RBACRepository) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	// Get all permissions grouped by resource
	permissions := make([]models.Permission, 0)

	err := r.db.NewSelect().
		Model(&permissions).
		Order("resource", "action").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	// Group permissions by resource
	resourceMap := make(map[string][]models.Permission)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm)
	}

	// For each permission, get the roles that have it
	var resources []models.ResourcePermissions
	for resource, perms := range resourceMap {
		var permWithRoles []models.PermissionWithRoles
		for _, perm := range perms {
			var roleIDs []uuid.UUID

			err := r.db.NewSelect().
				Model((*models.RolePermission)(nil)).
				Column("role_id").
				Where("permission_id = ?", perm.ID).
				Scan(ctx, &roleIDs)

			if err != nil {
				return nil, fmt.Errorf("failed to get role IDs for permission: %w", err)
			}

			permWithRoles = append(permWithRoles, models.PermissionWithRoles{
				PermissionID: perm.ID,
				Name:         perm.Name,
				Action:       perm.Action,
				Description:  perm.Description,
				Roles:        roleIDs,
			})
		}

		resources = append(resources, models.ResourcePermissions{
			Resource:    resource,
			Permissions: permWithRoles,
		})
	}

	return &models.PermissionMatrix{
		Resources: resources,
	}, nil
}
