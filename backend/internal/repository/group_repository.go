package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// GroupRepository handles group-related database operations
type GroupRepository struct {
	db *Database
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *Database) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create creates a new group
func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	_, err := r.db.NewInsert().
		Model(group).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	group := new(models.Group)
	err := r.db.NewSelect().
		Model(group).
		Where("g.id = ?", id).
		Relation("ParentGroup").
		Relation("ChildGroups").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// GetByName retrieves a group by name
func (r *GroupRepository) GetByName(ctx context.Context, name string) (*models.Group, error) {
	group := new(models.Group)
	err := r.db.NewSelect().
		Model(group).
		Where("name = ?", name).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group by name: %w", err)
	}

	return group, nil
}

// List retrieves all groups with pagination
func (r *GroupRepository) List(ctx context.Context, page, pageSize int) ([]*models.Group, int, error) {
	var groups []*models.Group

	// Get total count
	count, err := r.db.NewSelect().
		Model((*models.Group)(nil)).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err = r.db.NewSelect().
		Model(&groups).
		Relation("ParentGroup").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}

	return groups, count, nil
}

// Update updates a group
func (r *GroupRepository) Update(ctx context.Context, group *models.Group) error {
	_, err := r.db.NewUpdate().
		Model(group).
		Where("g.id = ?", group.ID).
		Where("is_system_group = ?", false). // Prevent updating system groups
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

// Delete deletes a group (only if not a system group and has no members)
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if group is system group
	group, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if group.IsSystemGroup {
		return models.NewAppError(400, "Cannot delete system group")
	}

	// Check if group has members
	memberCount, err := r.db.NewSelect().
		Model((*models.UserGroup)(nil)).
		Where("group_id = ?", id).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check group members: %w", err)
	}

	if memberCount > 0 {
		return models.NewAppError(400, "Cannot delete group with members")
	}

	// Check if group has child groups
	childCount, err := r.db.NewSelect().
		Model((*models.Group)(nil)).
		Where("parent_group_id = ?", id).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check child groups: %w", err)
	}

	if childCount > 0 {
		return models.NewAppError(400, "Cannot delete group with child groups")
	}

	_, err = r.db.NewDelete().
		Model((*models.Group)(nil)).
		Where("g.id = ?", id).
		Exec(ctx)

	return handlePgError(err)
}

// AddUser adds a user to a group
func (r *GroupRepository) AddUser(ctx context.Context, groupID, userID uuid.UUID) error {
	userGroup := &models.UserGroup{
		UserID:  userID,
		GroupID: groupID,
	}

	_, err := r.db.NewInsert().
		Model(userGroup).
		On("CONFLICT (user_id, group_id) DO NOTHING").
		Exec(ctx)

	return handlePgError(err)
}

// RemoveUser removes a user from a group
func (r *GroupRepository) RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.UserGroup)(nil)).
		Where("group_id = ?", groupID).
		Where("user_id = ?", userID).
		Exec(ctx)

	return handlePgError(err)
}

// GetGroupMembers retrieves all users in a group
func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
	var users []*models.User

	// Get total count
	count, err := r.db.NewSelect().
		Model((*models.UserGroup)(nil)).
		Where("group_id = ?", groupID).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count group members: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err = r.db.NewSelect().
		Model(&users).
		Join("JOIN user_groups ON \"user\".id = user_groups.user_id").
		Where("user_groups.group_id = ?", groupID).
		Order("\"user\".created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get group members: %w", err)
	}

	return users, count, nil
}

// GetUserGroups retrieves all groups for a user
func (r *GroupRepository) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*models.Group, error) {
	var groups []*models.Group

	err := r.db.NewSelect().
		Model(&groups).
		Join("JOIN user_groups ON \"group\".id = user_groups.group_id").
		Where("user_groups.user_id = ?", userID).
		Relation("ParentGroup").
		Order("\"group\".name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	return groups, nil
}

// GetGroupMemberCount returns the number of members in a group
func (r *GroupRepository) GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.UserGroup)(nil)).
		Where("group_id = ?", groupID).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get group member count: %w", err)
	}

	return count, nil
}
