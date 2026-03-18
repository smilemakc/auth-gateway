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

// mockLDAPConfigRepository is a simple mock for LDAP config operations
type mockLDAPConfigRepository struct {
	configs  map[uuid.UUID]*models.LDAPConfig
	active   *models.LDAPConfig
	syncLogs []*models.LDAPSyncLog
}

func newMockLDAPConfigRepository() *mockLDAPConfigRepository {
	return &mockLDAPConfigRepository{
		configs:  make(map[uuid.UUID]*models.LDAPConfig),
		syncLogs: make([]*models.LDAPSyncLog, 0),
	}
}

func (m *mockLDAPConfigRepository) Create(ctx context.Context, config *models.LDAPConfig) error {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	m.configs[config.ID] = config
	if config.SyncEnabled {
		m.active = config
	}
	return nil
}

func (m *mockLDAPConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error) {
	config, ok := m.configs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return config, nil
}

func (m *mockLDAPConfigRepository) GetActive(ctx context.Context) (*models.LDAPConfig, error) {
	if m.active == nil {
		return nil, models.ErrNotFound
	}
	return m.active, nil
}

func (m *mockLDAPConfigRepository) List(ctx context.Context) ([]*models.LDAPConfig, error) {
	configs := make([]*models.LDAPConfig, 0, len(m.configs))
	for _, c := range m.configs {
		configs = append(configs, c)
	}
	return configs, nil
}

func (m *mockLDAPConfigRepository) Update(ctx context.Context, config *models.LDAPConfig) error {
	if _, ok := m.configs[config.ID]; !ok {
		return models.ErrNotFound
	}
	m.configs[config.ID] = config
	if config.SyncEnabled {
		m.active = config
	}
	return nil
}

func (m *mockLDAPConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.configs, id)
	if m.active != nil && m.active.ID == id {
		m.active = nil
	}
	return nil
}

func (m *mockLDAPConfigRepository) CreateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	m.syncLogs = append(m.syncLogs, log)
	return nil
}

func (m *mockLDAPConfigRepository) GetSyncLogs(ctx context.Context, configID uuid.UUID, limit, offset int) ([]*models.LDAPSyncLog, int, error) {
	filtered := make([]*models.LDAPSyncLog, 0)
	for _, log := range m.syncLogs {
		if log.LDAPConfigID == configID {
			filtered = append(filtered, log)
		}
	}
	total := len(filtered)
	start := offset
	end := offset + limit
	if start >= total {
		return []*models.LDAPSyncLog{}, total, nil
	}
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (m *mockLDAPConfigRepository) UpdateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	for i, l := range m.syncLogs {
		if l.ID == log.ID {
			m.syncLogs[i] = log
			return nil
		}
	}
	return models.ErrNotFound
}

func setupLDAPService() (*LDAPService, *mockLDAPConfigRepository, *mockUserStore, *mockGroupStore) {
	mLDAP := newMockLDAPConfigRepository()
	mUser := &mockUserStore{}
	mGroup := newMockGroupStore()
	log := logger.New("test", logger.InfoLevel, false)
	svc := NewLDAPService(mLDAP, mUser, mGroup, log, "")
	return svc, mLDAP, mUser, mGroup
}

func TestLDAPService_CreateConfig(t *testing.T) {
	svc, _, _, _ := setupLDAPService()
	ctx := context.Background()

	t.Run("create config successfully", func(t *testing.T) {
		req := &models.CreateLDAPConfigRequest{
			Server:            "ldap.example.com",
			Port:              389,
			BaseDN:            "dc=example,dc=com",
			BindDN:            "cn=admin,dc=example,dc=com",
			BindPassword:      "password",
			UserSearchFilter:  "(objectClass=person)",
			GroupSearchFilter: "(objectClass=group)",
		}

		config, err := svc.CreateConfig(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, req.Server, config.Server)
		assert.Equal(t, req.Port, config.Port)
	})
}

func TestLDAPService_GetConfig(t *testing.T) {
	svc, _, _, _ := setupLDAPService()
	ctx := context.Background()

	t.Run("get existing config", func(t *testing.T) {
		req := &models.CreateLDAPConfigRequest{
			Server:           "ldap.example.com",
			Port:             389,
			BaseDN:           "dc=example,dc=com",
			BindDN:           "cn=admin,dc=example,dc=com",
			BindPassword:     "password",
			UserSearchFilter: "(objectClass=person)",
		}
		config, err := svc.CreateConfig(ctx, req)
		require.NoError(t, err)

		retrieved, err := svc.GetConfig(ctx, config.ID)
		assert.NoError(t, err)
		assert.Equal(t, config.ID, retrieved.ID)
	})

	t.Run("get non-existent config", func(t *testing.T) {
		_, err := svc.GetConfig(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestLDAPService_ListConfigs(t *testing.T) {
	svc, _, _, _ := setupLDAPService()
	ctx := context.Background()

	t.Run("list configs", func(t *testing.T) {
		// Create multiple configs
		for i := 0; i < 3; i++ {
			req := &models.CreateLDAPConfigRequest{
				Server:           "ldap" + string(rune(i)) + ".example.com",
				Port:             389,
				BaseDN:           "dc=example,dc=com",
				BindDN:           "cn=admin,dc=example,dc=com",
				BindPassword:     "password",
				UserSearchFilter: "(objectClass=person)",
			}
			_, err := svc.CreateConfig(ctx, req)
			require.NoError(t, err)
		}

		configs, err := svc.ListConfigs(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(configs), 3)
	})
}

func TestLDAPService_UpdateConfig(t *testing.T) {
	svc, _, _, _ := setupLDAPService()
	ctx := context.Background()

	t.Run("update config", func(t *testing.T) {
		req := &models.CreateLDAPConfigRequest{
			Server:           "ldap.example.com",
			Port:             389,
			BaseDN:           "dc=example,dc=com",
			BindDN:           "cn=admin,dc=example,dc=com",
			BindPassword:     "password",
			UserSearchFilter: "(objectClass=person)",
		}
		config, err := svc.CreateConfig(ctx, req)
		require.NoError(t, err)

		updateReq := &models.UpdateLDAPConfigRequest{
			Server: stringPtr("ldap-updated.example.com"),
			Port:   intPtr(636),
		}

		updated, err := svc.UpdateConfig(ctx, config.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, "ldap-updated.example.com", updated.Server)
		assert.Equal(t, 636, updated.Port)
	})
}

func TestLDAPService_DeleteConfig(t *testing.T) {
	svc, _, _, _ := setupLDAPService()
	ctx := context.Background()

	t.Run("delete config", func(t *testing.T) {
		req := &models.CreateLDAPConfigRequest{
			Server:           "ldap.example.com",
			Port:             389,
			BaseDN:           "dc=example,dc=com",
			BindDN:           "cn=admin,dc=example,dc=com",
			BindPassword:     "password",
			UserSearchFilter: "(objectClass=person)",
		}
		config, err := svc.CreateConfig(ctx, req)
		require.NoError(t, err)

		err = svc.DeleteConfig(ctx, config.ID)
		assert.NoError(t, err)

		_, err = svc.GetConfig(ctx, config.ID)
		assert.Error(t, err)
	})
}

func intPtr(i int) *int {
	return &i
}
