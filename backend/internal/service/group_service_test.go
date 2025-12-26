package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGroupStore is a simple mock implementation of GroupRepository
type mockGroupStore struct {
	groups       map[uuid.UUID]*models.Group
	groupsByName map[string]*models.Group
	users        map[uuid.UUID][]uuid.UUID // groupID -> userIDs
	createFunc   func(ctx context.Context, group *models.Group) error
	getByIDFunc  func(ctx context.Context, id uuid.UUID) (*models.Group, error)
}

func newMockGroupStore() *mockGroupStore {
	return &mockGroupStore{
		groups:       make(map[uuid.UUID]*models.Group),
		groupsByName: make(map[string]*models.Group),
		users:        make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockGroupStore) Create(ctx context.Context, group *models.Group) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, group)
	}
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	m.groups[group.ID] = group
	m.groupsByName[group.Name] = group
	return nil
}

func (m *mockGroupStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	group, ok := m.groups[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return group, nil
}

func (m *mockGroupStore) GetByName(ctx context.Context, name string) (*models.Group, error) {
	group, ok := m.groupsByName[name]
	if !ok {
		return nil, models.ErrNotFound
	}
	return group, nil
}

func (m *mockGroupStore) List(ctx context.Context, page, pageSize int) ([]*models.Group, int, error) {
	groups := make([]*models.Group, 0, len(m.groups))
	for _, g := range m.groups {
		groups = append(groups, g)
	}
	total := len(groups)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*models.Group{}, total, nil
	}
	if end > total {
		end = total
	}
	return groups[start:end], total, nil
}

func (m *mockGroupStore) Update(ctx context.Context, group *models.Group) error {
	if _, ok := m.groups[group.ID]; !ok {
		return models.ErrNotFound
	}
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupStore) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.groups, id)
	return nil
}

func (m *mockGroupStore) AddUser(ctx context.Context, groupID, userID uuid.UUID) error {
	m.users[groupID] = append(m.users[groupID], userID)
	return nil
}

func (m *mockGroupStore) RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error {
	users := m.users[groupID]
	for i, uid := range users {
		if uid == userID {
			m.users[groupID] = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockGroupStore) GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
	userIDs := m.users[groupID]
	users := make([]*models.User, len(userIDs))
	for i, uid := range userIDs {
		users[i] = &models.User{ID: uid}
	}
	return users, len(users), nil
}

func (m *mockGroupStore) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*models.Group, error) {
	groups := make([]*models.Group, 0)
	for gid, userIDs := range m.users {
		for _, uid := range userIDs {
			if uid == userID {
				if group, ok := m.groups[gid]; ok {
					groups = append(groups, group)
				}
			}
		}
	}
	return groups, nil
}

func (m *mockGroupStore) GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	return len(m.users[groupID]), nil
}

func setupGroupService() (*GroupService, *mockGroupStore, *mockUserStore) {
	mGroup := &mockGroupStore{}
	mUser := &mockUserStore{}
	log := logger.New("test", logger.InfoLevel, false)
	svc := NewGroupService(mGroup, mUser, log)
	return svc, mGroup, mUser
}

func TestGroupService_CreateGroup(t *testing.T) {
	svc, _, _ := setupGroupService()
	ctx := context.Background()

	t.Run("create group successfully", func(t *testing.T) {
		req := &models.CreateGroupRequest{
			Name:        "test-group",
			DisplayName: "Test Group",
			Description: "Test description",
		}

		group, err := svc.CreateGroup(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, req.Name, group.Name)
		assert.Equal(t, req.DisplayName, group.DisplayName)
	})

	t.Run("fail on duplicate name", func(t *testing.T) {
		req := &models.CreateGroupRequest{
			Name:        "duplicate",
			DisplayName: "Duplicate",
		}

		// Create first group
		_, err := svc.CreateGroup(ctx, req)
		assert.NoError(t, err)

		// Try to create duplicate
		group, err := svc.CreateGroup(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, group)
	})

	t.Run("create group with parent", func(t *testing.T) {
		// Create parent first
		parentReq := &models.CreateGroupRequest{
			Name:        "parent-group",
			DisplayName: "Parent Group",
		}
		parent, err := svc.CreateGroup(ctx, parentReq)
		require.NoError(t, err)

		req := &models.CreateGroupRequest{
			Name:          "child-group",
			DisplayName:   "Child Group",
			ParentGroupID: &parent.ID,
		}

		group, err := svc.CreateGroup(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, parent.ID, *group.ParentGroupID)
	})
}

func TestGroupService_UpdateGroup(t *testing.T) {
	svc, _, _ := setupGroupService()
	ctx := context.Background()

	t.Run("update group successfully", func(t *testing.T) {
		// Create group first
		createReq := &models.CreateGroupRequest{
			Name:        "update-test",
			DisplayName: "Original",
		}
		group, err := svc.CreateGroup(ctx, createReq)
		require.NoError(t, err)

		req := &models.UpdateGroupRequest{
			DisplayName: stringPtr("Updated Name"),
			Description: stringPtr("Updated description"),
		}

		updated, err := svc.UpdateGroup(ctx, group.ID, req)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "Updated Name", updated.DisplayName)
		assert.Equal(t, "Updated description", updated.Description)
	})
}

func TestGroupService_DeleteGroup(t *testing.T) {
	svc, _, _ := setupGroupService()
	ctx := context.Background()

	t.Run("delete group successfully", func(t *testing.T) {
		// Create group first
		createReq := &models.CreateGroupRequest{
			Name:        "delete-test",
			DisplayName: "Delete Test",
		}
		group, err := svc.CreateGroup(ctx, createReq)
		require.NoError(t, err)

		err = svc.DeleteGroup(ctx, group.ID)
		assert.NoError(t, err)

		// Verify deleted
		_, err = svc.GetGroup(ctx, group.ID)
		assert.Error(t, err)
	})
}

func TestGroupService_AddGroupMembers(t *testing.T) {
	svc, _, _ := setupGroupService()
	ctx := context.Background()

	t.Run("add members successfully", func(t *testing.T) {
		// Create group first
		createReq := &models.CreateGroupRequest{
			Name:        "add-members",
			DisplayName: "Add Members",
		}
		group, err := svc.CreateGroup(ctx, createReq)
		require.NoError(t, err)

		userIDs := []uuid.UUID{uuid.New(), uuid.New()}

		err = svc.AddGroupMembers(ctx, group.ID, userIDs)
		assert.NoError(t, err)
	})
}

func TestGroupService_RemoveGroupMember(t *testing.T) {
	svc, _, _ := setupGroupService()
	ctx := context.Background()

	t.Run("remove member successfully", func(t *testing.T) {
		// Create group first
		createReq := &models.CreateGroupRequest{
			Name:        "remove-member",
			DisplayName: "Remove Member",
		}
		group, err := svc.CreateGroup(ctx, createReq)
		require.NoError(t, err)

		userID := uuid.New()
		err = svc.AddGroupMembers(ctx, group.ID, []uuid.UUID{userID})
		require.NoError(t, err)

		err = svc.RemoveGroupMember(ctx, group.ID, userID)
		assert.NoError(t, err)
	})
}

func stringPtr(s string) *string {
	return &s
}
