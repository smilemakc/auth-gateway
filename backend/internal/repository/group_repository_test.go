package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGroupTestDB creates a test database for group tests
// Uses the same setupTestDB function pattern as other tests
func setupGroupTestDB(t *testing.T) (*Database, func()) {
	t.Helper()
	cfg := &config.DatabaseConfig{
		Host:         getGroupTestEnv("TEST_DB_HOST", "localhost"),
		Port:         getGroupTestEnv("TEST_DB_PORT", "5432"),
		User:         getGroupTestEnv("TEST_DB_USER", "postgres"),
		Password:     getGroupTestEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:       getGroupTestEnv("TEST_DB_NAME", "auth_gateway_test"),
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	return db, func() { db.Close() }
}

func getGroupTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestGroupRepository_Create(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	t.Run("create group successfully", func(t *testing.T) {
		group := &models.Group{
			Name:        "test-group",
			DisplayName: "Test Group",
			Description: "Test description",
		}

		err := repo.Create(ctx, group)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, group.ID)
		assert.False(t, group.CreatedAt.IsZero())
	})

	t.Run("create group with parent", func(t *testing.T) {
		parent := &models.Group{
			Name:        "parent-group",
			DisplayName: "Parent Group",
		}
		err := repo.Create(ctx, parent)
		require.NoError(t, err)

		child := &models.Group{
			Name:          "child-group",
			DisplayName:   "Child Group",
			ParentGroupID: &parent.ID,
		}

		err = repo.Create(ctx, child)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, *child.ParentGroupID)
	})

	t.Run("fail on duplicate name", func(t *testing.T) {
		group1 := &models.Group{
			Name:        "duplicate-name",
			DisplayName: "First",
		}
		err := repo.Create(ctx, group1)
		require.NoError(t, err)

		group2 := &models.Group{
			Name:        "duplicate-name",
			DisplayName: "Second",
		}
		err = repo.Create(ctx, group2)
		assert.Error(t, err)
	})
}

func TestGroupRepository_GetByID(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	t.Run("get existing group", func(t *testing.T) {
		group := &models.Group{
			Name:        "get-test",
			DisplayName: "Get Test",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, group.ID, retrieved.ID)
		assert.Equal(t, group.Name, retrieved.Name)
		assert.Equal(t, group.DisplayName, retrieved.DisplayName)
	})

	t.Run("get non-existent group", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)
	})

	t.Run("get group with parent", func(t *testing.T) {
		parent := &models.Group{
			Name:        "parent-get",
			DisplayName: "Parent",
		}
		err := repo.Create(ctx, parent)
		require.NoError(t, err)

		child := &models.Group{
			Name:          "child-get",
			DisplayName:   "Child",
			ParentGroupID: &parent.ID,
		}
		err = repo.Create(ctx, child)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, child.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.ParentGroup)
		assert.Equal(t, parent.ID, retrieved.ParentGroup.ID)
	})
}

func TestGroupRepository_GetByName(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	t.Run("get by name", func(t *testing.T) {
		group := &models.Group{
			Name:        "by-name",
			DisplayName: "By Name",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		retrieved, err := repo.GetByName(ctx, "by-name")
		require.NoError(t, err)
		assert.Equal(t, group.ID, retrieved.ID)
	})

	t.Run("get non-existent name", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)
	})
}

func TestGroupRepository_List(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	// Create test groups
	for i := 0; i < 5; i++ {
		group := &models.Group{
			Name:        uuid.New().String(),
			DisplayName: "Test Group " + string(rune(i)),
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)
	}

	t.Run("list all groups", func(t *testing.T) {
		groups, total, err := repo.List(ctx, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 5)
		assert.GreaterOrEqual(t, len(groups), 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		groups, total, err := repo.List(ctx, 1, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 5)
		assert.LessOrEqual(t, len(groups), 2)
	})
}

func TestGroupRepository_Update(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	t.Run("update group", func(t *testing.T) {
		group := &models.Group{
			Name:        "update-test",
			DisplayName: "Original",
			Description: "Original description",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		group.DisplayName = "Updated"
		group.Description = "Updated description"
		err = repo.Update(ctx, group)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", retrieved.DisplayName)
		assert.Equal(t, "Updated description", retrieved.Description)
	})
}

func TestGroupRepository_Delete(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	t.Run("delete group", func(t *testing.T) {
		group := &models.Group{
			Name:        "delete-test",
			DisplayName: "Delete Test",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		err = repo.Delete(ctx, group.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, group.ID)
		assert.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)
	})

	t.Run("fail to delete group with members", func(t *testing.T) {
		group := &models.Group{
			Name:        "with-members",
			DisplayName: "With Members",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		// Add a member
		userRepo := NewUserRepository(db)
		user := &models.User{
			Email:        "test@example.com",
			Username:     "testuser",
			PasswordHash: "hash",
		}
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.AddUser(ctx, group.ID, user.ID)
		require.NoError(t, err)

		err = repo.Delete(ctx, group.ID)
		assert.Error(t, err)
	})

	t.Run("fail to delete group with children", func(t *testing.T) {
		parent := &models.Group{
			Name:        "parent-delete",
			DisplayName: "Parent",
		}
		err := repo.Create(ctx, parent)
		require.NoError(t, err)

		child := &models.Group{
			Name:          "child-delete",
			DisplayName:   "Child",
			ParentGroupID: &parent.ID,
		}
		err = repo.Create(ctx, child)
		require.NoError(t, err)

		err = repo.Delete(ctx, parent.ID)
		assert.Error(t, err)
	})
}

func TestGroupRepository_AddUser(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("add user to group", func(t *testing.T) {
		group := &models.Group{
			Name:        "add-user",
			DisplayName: "Add User",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		user := &models.User{
			Email:        "user@example.com",
			Username:     "user",
			PasswordHash: "hash",
		}
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.AddUser(ctx, group.ID, user.ID)
		require.NoError(t, err)

		members, _, err := repo.GetGroupMembers(ctx, group.ID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, user.ID, members[0].ID)
	})
}

func TestGroupRepository_RemoveUser(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("remove user from group", func(t *testing.T) {
		group := &models.Group{
			Name:        "remove-user",
			DisplayName: "Remove User",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		user := &models.User{
			Email:        "remove@example.com",
			Username:     "remove",
			PasswordHash: "hash",
		}
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.AddUser(ctx, group.ID, user.ID)
		require.NoError(t, err)

		err = repo.RemoveUser(ctx, group.ID, user.ID)
		require.NoError(t, err)

		members, _, err := repo.GetGroupMembers(ctx, group.ID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, members, 0)
	})
}

func TestGroupRepository_GetGroupMembers(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("get group members", func(t *testing.T) {
		group := &models.Group{
			Name:        "members-test",
			DisplayName: "Members Test",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		// Add multiple users
		for i := 0; i < 3; i++ {
			user := &models.User{
				Email:        uuid.New().String() + "@example.com",
				Username:     uuid.New().String(),
				PasswordHash: "hash",
			}
			err = userRepo.Create(ctx, user)
			require.NoError(t, err)

			err = repo.AddUser(ctx, group.ID, user.ID)
			require.NoError(t, err)
		}

		members, total, err := repo.GetGroupMembers(ctx, group.ID, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, members, 3)
	})
}

func TestGroupRepository_GetUserGroups(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("get user groups", func(t *testing.T) {
		user := &models.User{
			Email:        "groups@example.com",
			Username:     "groups",
			PasswordHash: "hash",
		}
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create and add to multiple groups
		for i := 0; i < 2; i++ {
			group := &models.Group{
				Name:        uuid.New().String(),
				DisplayName: "Group " + string(rune(i)),
			}
			err = repo.Create(ctx, group)
			require.NoError(t, err)

			err = repo.AddUser(ctx, group.ID, user.ID)
			require.NoError(t, err)
		}

		groups, err := repo.GetUserGroups(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, groups, 2)
	})
}

func TestGroupRepository_GetGroupMemberCount(t *testing.T) {
	db, cleanup := setupGroupTestDB(t)
	defer cleanup()

	repo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("get member count", func(t *testing.T) {
		group := &models.Group{
			Name:        "count-test",
			DisplayName: "Count Test",
		}
		err := repo.Create(ctx, group)
		require.NoError(t, err)

		// Add users
		for i := 0; i < 5; i++ {
			user := &models.User{
				Email:        uuid.New().String() + "@example.com",
				Username:     uuid.New().String(),
				PasswordHash: "hash",
			}
			err = userRepo.Create(ctx, user)
			require.NoError(t, err)

			err = repo.AddUser(ctx, group.ID, user.ID)
			require.NoError(t, err)
		}

		count, err := repo.GetGroupMemberCount(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})
}
