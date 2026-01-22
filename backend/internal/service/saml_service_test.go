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

// mockSAMLSPRepository is a simple mock for SAML SP operations
type mockSAMLSPRepository struct {
	sps         map[uuid.UUID]*models.SAMLServiceProvider
	spsByEntity map[string]*models.SAMLServiceProvider
}

func newMockSAMLSPRepository() *mockSAMLSPRepository {
	return &mockSAMLSPRepository{
		sps:         make(map[uuid.UUID]*models.SAMLServiceProvider),
		spsByEntity: make(map[string]*models.SAMLServiceProvider),
	}
}

func (m *mockSAMLSPRepository) Create(ctx context.Context, sp *models.SAMLServiceProvider) error {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	m.sps[sp.ID] = sp
	m.spsByEntity[sp.EntityID] = sp
	return nil
}

func (m *mockSAMLSPRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SAMLServiceProvider, error) {
	sp, ok := m.sps[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return sp, nil
}

func (m *mockSAMLSPRepository) GetByEntityID(ctx context.Context, entityID string) (*models.SAMLServiceProvider, error) {
	sp, ok := m.spsByEntity[entityID]
	if !ok {
		return nil, models.ErrNotFound
	}
	return sp, nil
}

func (m *mockSAMLSPRepository) List(ctx context.Context, page, pageSize int) ([]*models.SAMLServiceProvider, int, error) {
	sps := make([]*models.SAMLServiceProvider, 0, len(m.sps))
	for _, sp := range m.sps {
		sps = append(sps, sp)
	}
	total := len(sps)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*models.SAMLServiceProvider{}, total, nil
	}
	if end > total {
		end = total
	}
	return sps[start:end], total, nil
}

func (m *mockSAMLSPRepository) Update(ctx context.Context, sp *models.SAMLServiceProvider) error {
	if _, ok := m.sps[sp.ID]; !ok {
		return models.ErrNotFound
	}
	m.sps[sp.ID] = sp
	m.spsByEntity[sp.EntityID] = sp
	return nil
}

func (m *mockSAMLSPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sp, ok := m.sps[id]
	if ok {
		delete(m.spsByEntity, sp.EntityID)
	}
	delete(m.sps, id)
	return nil
}

func setupSAMLService() (*SAMLService, *mockSAMLSPRepository, *mockUserStore, *mockRBACStore) {
	mSP := newMockSAMLSPRepository()
	mUser := &mockUserStore{}
	mRBAC := &mockRBACStore{}
	log := logger.New("test", logger.InfoLevel, false)
	svc := NewSAMLService(mSP, mUser, mRBAC, log, "https://auth.example.com", "https://auth.example.com")
	return svc, mSP, mUser, mRBAC
}

func TestSAMLService_CreateSP(t *testing.T) {
	svc, _, _, _ := setupSAMLService()
	ctx := context.Background()

	t.Run("create SP successfully", func(t *testing.T) {
		req := &models.CreateSAMLSPRequest{
			Name:     "Test SP",
			EntityID: "https://sp.example.com/saml",
			ACSURL:   "https://sp.example.com/saml/acs",
		}

		sp, err := svc.CreateSP(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, sp)
		assert.Equal(t, req.Name, sp.Name)
		assert.Equal(t, req.EntityID, sp.EntityID)
		assert.Equal(t, req.ACSURL, sp.ACSURL)
	})

	t.Run("fail on duplicate entity ID", func(t *testing.T) {
		req := &models.CreateSAMLSPRequest{
			Name:     "First SP",
			EntityID: "https://sp.example.com/saml",
			ACSURL:   "https://sp.example.com/saml/acs",
		}
		_, err := svc.CreateSP(ctx, req)
		require.NoError(t, err)

		req2 := &models.CreateSAMLSPRequest{
			Name:     "Second SP",
			EntityID: "https://sp.example.com/saml", // Same entity ID
			ACSURL:   "https://sp.example.com/saml/acs",
		}
		sp, err := svc.CreateSP(ctx, req2)
		assert.Error(t, err)
		assert.Nil(t, sp)
	})
}

func TestSAMLService_GetSP(t *testing.T) {
	svc, _, _, _ := setupSAMLService()
	ctx := context.Background()

	t.Run("get existing SP", func(t *testing.T) {
		req := &models.CreateSAMLSPRequest{
			Name:     "Test SP",
			EntityID: "https://sp.example.com/saml",
			ACSURL:   "https://sp.example.com/saml/acs",
		}
		sp, err := svc.CreateSP(ctx, req)
		require.NoError(t, err)

		retrieved, err := svc.GetSP(ctx, sp.ID)
		assert.NoError(t, err)
		assert.Equal(t, sp.ID, retrieved.ID)
	})

	t.Run("get non-existent SP", func(t *testing.T) {
		_, err := svc.GetSP(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestSAMLService_ListSPs(t *testing.T) {
	svc, _, _, _ := setupSAMLService()
	ctx := context.Background()

	t.Run("list SPs", func(t *testing.T) {
		// Create multiple SPs
		for i := 0; i < 3; i++ {
			req := &models.CreateSAMLSPRequest{
				Name:     "SP " + string(rune(i)),
				EntityID: "https://sp" + string(rune(i)) + ".example.com/saml",
				ACSURL:   "https://sp" + string(rune(i)) + ".example.com/saml/acs",
			}
			_, err := svc.CreateSP(ctx, req)
			require.NoError(t, err)
		}

		sps, total, err := svc.ListSPs(ctx, 1, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 3)
		assert.GreaterOrEqual(t, len(sps), 3)
	})
}

func TestSAMLService_UpdateSP(t *testing.T) {
	svc, _, _, _ := setupSAMLService()
	ctx := context.Background()

	t.Run("update SP", func(t *testing.T) {
		req := &models.CreateSAMLSPRequest{
			Name:     "Original SP",
			EntityID: "https://sp.example.com/saml",
			ACSURL:   "https://sp.example.com/saml/acs",
		}
		sp, err := svc.CreateSP(ctx, req)
		require.NoError(t, err)

		updateReq := &models.UpdateSAMLSPRequest{
			Name:   stringPtr("Updated SP"),
			ACSURL: stringPtr("https://sp.example.com/saml/acs/updated"),
		}

		updated, err := svc.UpdateSP(ctx, sp.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, "Updated SP", updated.Name)
		assert.Equal(t, "https://sp.example.com/saml/acs/updated", updated.ACSURL)
	})
}

func TestSAMLService_DeleteSP(t *testing.T) {
	svc, _, _, _ := setupSAMLService()
	ctx := context.Background()

	t.Run("delete SP", func(t *testing.T) {
		req := &models.CreateSAMLSPRequest{
			Name:     "Delete SP",
			EntityID: "https://sp.example.com/saml",
			ACSURL:   "https://sp.example.com/saml/acs",
		}
		sp, err := svc.CreateSP(ctx, req)
		require.NoError(t, err)

		err = svc.DeleteSP(ctx, sp.ID)
		assert.NoError(t, err)

		_, err = svc.GetSP(ctx, sp.ID)
		assert.Error(t, err)
	})
}
