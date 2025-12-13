package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] Creating user_roles junction table...")

		// ============================================================
		// 1. Create user_roles junction table
		// ============================================================
		_, err := db.ExecContext(ctx, `
			CREATE TABLE user_roles (
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
				assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				assigned_by UUID REFERENCES users(id),
				PRIMARY KEY (user_id, role_id)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create user_roles table: %w", err)
		}

		// ============================================================
		// 2. Create indexes
		// ============================================================
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id)",
		}

		for _, indexSQL := range indexes {
			if _, err := db.ExecContext(ctx, indexSQL); err != nil {
				return fmt.Errorf("failed to create index: %w (SQL: %s)", err, indexSQL)
			}
		}

		// ============================================================
		// 3. Migrate existing data from users.role_id
		// ============================================================
		_, err = db.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT id, role_id
			FROM users
			WHERE role_id IS NOT NULL
		`)
		if err != nil {
			return fmt.Errorf("failed to migrate existing user roles: %w", err)
		}

		// ============================================================
		// 4. Update user_role_permissions view
		// ============================================================
		_, err = db.ExecContext(ctx, `
			DROP VIEW IF EXISTS user_role_permissions
		`)
		if err != nil {
			return fmt.Errorf("failed to drop existing user_role_permissions view: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			CREATE VIEW user_role_permissions AS
			SELECT
				u.id AS user_id,
				u.username,
				u.email,
				json_agg(DISTINCT jsonb_build_object(
					'role_id', r.id,
					'role_name', r.name,
					'role_display_name', r.display_name
				)) FILTER (WHERE r.id IS NOT NULL) AS roles,
				json_agg(DISTINCT jsonb_build_object(
					'permission_id', p.id,
					'permission_name', p.name,
					'resource', p.resource,
					'action', p.action
				)) FILTER (WHERE p.id IS NOT NULL) AS permissions
			FROM users u
			LEFT JOIN user_roles ur ON ur.user_id = u.id
			LEFT JOIN roles r ON r.id = ur.role_id
			LEFT JOIN role_permissions rp ON rp.role_id = r.id
			LEFT JOIN permissions p ON p.id = rp.permission_id
			GROUP BY u.id, u.username, u.email
		`)
		if err != nil {
			return fmt.Errorf("failed to create updated user_role_permissions view: %w", err)
		}

		fmt.Println(" OK")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] Reverting user_roles junction table...")

		// ============================================================
		// 1. Restore original user_role_permissions view
		// ============================================================
		_, err := db.ExecContext(ctx, `
			DROP VIEW IF EXISTS user_role_permissions
		`)
		if err != nil {
			return fmt.Errorf("failed to drop user_role_permissions view: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			CREATE VIEW user_role_permissions AS
			SELECT
				u.id AS user_id,
				u.username,
				u.email,
				r.id AS role_id,
				r.name AS role_name,
				r.display_name AS role_display_name,
				COALESCE(json_agg(
					json_build_object(
						'permission_id', p.id,
						'permission_name', p.name,
						'resource', p.resource,
						'action', p.action
					) ORDER BY p.name
				) FILTER (WHERE p.id IS NOT NULL), '[]') AS permissions
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			LEFT JOIN role_permissions rp ON r.id = rp.role_id
			LEFT JOIN permissions p ON rp.permission_id = p.id
			GROUP BY u.id, u.username, u.email, r.id, r.name, r.display_name
		`)
		if err != nil {
			return fmt.Errorf("failed to restore original user_role_permissions view: %w", err)
		}

		// ============================================================
		// 2. Restore users.role_id from user_roles
		// ============================================================
		_, err = db.ExecContext(ctx, `
			UPDATE users u
			SET role_id = (
				SELECT ur.role_id
				FROM user_roles ur
				WHERE ur.user_id = u.id
				ORDER BY ur.assigned_at ASC
				LIMIT 1
			)
			WHERE EXISTS (
				SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to restore users.role_id: %w", err)
		}

		// ============================================================
		// 3. Drop user_roles table
		// ============================================================
		_, err = db.ExecContext(ctx, `
			DROP TABLE IF EXISTS user_roles CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop user_roles table: %w", err)
		}

		fmt.Println(" OK")
		return nil
	})
}
