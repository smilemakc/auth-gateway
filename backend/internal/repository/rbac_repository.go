package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
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
	query := `
		INSERT INTO permissions (name, resource, action, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		permission.Name, permission.Resource, permission.Action, permission.Description,
	).Scan(&permission.ID, &permission.CreatedAt)
}

// GetPermissionByID retrieves a permission by ID
func (r *RBACRepository) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	query := `SELECT * FROM permissions WHERE id = $1`
	err := r.db.GetContext(ctx, &permission, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	return &permission, err
}

// GetPermissionByName retrieves a permission by name
func (r *RBACRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	query := `SELECT * FROM permissions WHERE name = $1`
	err := r.db.GetContext(ctx, &permission, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	return &permission, err
}

// ListPermissions retrieves all permissions
func (r *RBACRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `SELECT * FROM permissions ORDER BY resource, action`
	err := r.db.SelectContext(ctx, &permissions, query)
	return permissions, err
}

// ListPermissionsByResource retrieves permissions for a specific resource
func (r *RBACRepository) ListPermissionsByResource(ctx context.Context, resource string) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `SELECT * FROM permissions WHERE resource = $1 ORDER BY action`
	err := r.db.SelectContext(ctx, &permissions, query, resource)
	return permissions, err
}

// UpdatePermission updates a permission
func (r *RBACRepository) UpdatePermission(ctx context.Context, id uuid.UUID, description string) error {
	query := `UPDATE permissions SET description = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, description, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}

// DeletePermission deletes a permission
func (r *RBACRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
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
	query := `
		INSERT INTO roles (name, display_name, description, is_system_role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		role.Name, role.DisplayName, role.Description, role.IsSystemRole,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
}

// GetRoleByID retrieves a role by ID with its permissions
func (r *RBACRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	query := `SELECT * FROM roles WHERE id = $1`
	err := r.db.GetContext(ctx, &role, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, err
	}

	// Fetch permissions for this role
	permissions, err := r.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	return &role, nil
}

// GetRoleByName retrieves a role by name
func (r *RBACRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	query := `SELECT * FROM roles WHERE name = $1`
	err := r.db.GetContext(ctx, &role, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, err
	}

	// Fetch permissions for this role
	permissions, err := r.GetRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	return &role, nil
}

// ListRoles retrieves all roles
func (r *RBACRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	query := `SELECT * FROM roles ORDER BY is_system_role DESC, name`
	err := r.db.SelectContext(ctx, &roles, query)
	if err != nil {
		return nil, err
	}

	// Fetch permissions for each role
	for i := range roles {
		permissions, err := r.GetRolePermissions(ctx, roles[i].ID)
		if err != nil {
			return nil, err
		}
		roles[i].Permissions = permissions
	}

	return roles, nil
}

// UpdateRole updates a role
func (r *RBACRepository) UpdateRole(ctx context.Context, id uuid.UUID, displayName, description string) error {
	query := `
		UPDATE roles
		SET display_name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, displayName, description, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}

// DeleteRole deletes a role (only if not a system role)
func (r *RBACRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system_role = false`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
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
	var permissions []models.Permission
	query := `
		SELECT p.* FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	err := r.db.SelectContext(ctx, &permissions, query, roleID)
	return permissions, err
}

// AddPermissionToRole adds a permission to a role
func (r *RBACRepository) AddPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}

// RemovePermissionFromRole removes a permission from a role
func (r *RBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}

// SetRolePermissions sets all permissions for a role (replaces existing)
func (r *RBACRepository) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing permissions
	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return err
	}

	// Add new permissions
	for _, permissionID := range permissionIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permissionID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ============================================================
// User Permission Checking
// ============================================================

// GetUserPermissions retrieves all permissions for a user based on their role
func (r *RBACRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT DISTINCT p.* FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN roles r ON rp.role_id = r.id
		INNER JOIN users u ON u.role_id = r.id
		WHERE u.id = $1
		ORDER BY p.resource, p.action
	`
	err := r.db.SelectContext(ctx, &permissions, query, userID)
	return permissions, err
}

// HasPermission checks if a user has a specific permission
func (r *RBACRepository) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN roles r ON rp.role_id = r.id
		INNER JOIN users u ON u.role_id = r.id
		WHERE u.id = $1 AND p.name = $2
	`
	err := r.db.GetContext(ctx, &count, query, userID, permissionName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (r *RBACRepository) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN roles r ON rp.role_id = r.id
		INNER JOIN users u ON u.role_id = r.id
		WHERE u.id = $1 AND p.name = ANY($2)
	`
	err := r.db.GetContext(ctx, &count, query, userID, permissionNames)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (r *RBACRepository) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionNames []string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT p.name) FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN roles r ON rp.role_id = r.id
		INNER JOIN users u ON u.role_id = r.id
		WHERE u.id = $1 AND p.name = ANY($2)
	`
	err := r.db.GetContext(ctx, &count, query, userID, permissionNames)
	if err != nil {
		return false, err
	}
	return count == len(permissionNames), nil
}

// GetUserRole retrieves the role for a user
func (r *RBACRepository) GetUserRole(ctx context.Context, userID uuid.UUID) (*models.Role, error) {
	var role models.Role
	query := `
		SELECT r.* FROM roles r
		INNER JOIN users u ON u.role_id = r.id
		WHERE u.id = $1
	`
	err := r.db.GetContext(ctx, &role, query, userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user role not found")
	}
	if err != nil {
		return nil, err
	}

	// Fetch permissions for this role
	permissions, err := r.GetRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	return &role, nil
}

// GetPermissionMatrix retrieves the complete permission matrix for all roles
func (r *RBACRepository) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrix, error) {
	// Get all permissions grouped by resource
	var permissions []models.Permission
	query := `SELECT * FROM permissions ORDER BY resource, action`
	err := r.db.SelectContext(ctx, &permissions, query)
	if err != nil {
		return nil, err
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
			roleQuery := `
				SELECT role_id FROM role_permissions WHERE permission_id = $1
			`
			err := r.db.SelectContext(ctx, &roleIDs, roleQuery, perm.ID)
			if err != nil {
				return nil, err
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
