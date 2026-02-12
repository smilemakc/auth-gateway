package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestIPFilterService_CreateIPFilter(t *testing.T) {
	mockStore := &mockIPFilterStore{}
	svc := NewIPFilterService(mockStore)

	ctx := context.Background()
	adminID := uuid.New()

	req := &models.CreateIPFilterRequest{
		IPCIDR:     "192.168.1.0/24",
		FilterType: "whitelist",
		Reason:     "Office",
	}

	t.Run("Success", func(t *testing.T) {
		mockStore.CreateIPFilterFunc = func(ctx context.Context, filter *models.IPFilter) error {
			assert.Equal(t, "192.168.1.0/24", filter.IPCIDR)
			assert.Equal(t, "whitelist", filter.FilterType)
			assert.True(t, filter.IsActive)
			return nil
		}

		filter, err := svc.CreateIPFilter(ctx, req, adminID)
		assert.NoError(t, err)
		assert.NotNil(t, filter)
	})

	t.Run("InvalidCIDR", func(t *testing.T) {
		reqInvalid := &models.CreateIPFilterRequest{
			IPCIDR: "invalid",
		}

		filter, err := svc.CreateIPFilter(ctx, reqInvalid, adminID)
		assert.Error(t, err)
		assert.Nil(t, filter)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockStore.CreateIPFilterFunc = func(ctx context.Context, filter *models.IPFilter) error {
			return errors.New("db error")
		}

		filter, err := svc.CreateIPFilter(ctx, req, adminID)
		assert.Error(t, err)
		assert.Nil(t, filter)
	})
}

func TestIPFilterService_CheckIPAllowed(t *testing.T) {
	mockStore := &mockIPFilterStore{}
	svc := NewIPFilterService(mockStore)
	ctx := context.Background()

	t.Run("NoFilters_Allowed", func(t *testing.T) {
		mockStore.GetActiveIPFiltersFunc = func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{}, nil
		}

		resp, err := svc.CheckIPAllowed(ctx, "1.2.3.4")
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("Whitelist_Allowed", func(t *testing.T) {
		mockStore.GetActiveIPFiltersFunc = func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{IPCIDR: "1.2.3.4", FilterType: "whitelist"},
			}, nil
		}

		resp, err := svc.CheckIPAllowed(ctx, "1.2.3.4")
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("Whitelist_Denied", func(t *testing.T) {
		mockStore.GetActiveIPFiltersFunc = func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{IPCIDR: "1.2.3.4", FilterType: "whitelist"},
			}, nil
		}

		resp, err := svc.CheckIPAllowed(ctx, "5.6.7.8")
		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Equal(t, "whitelist", resp.FilterType)
	})

	t.Run("Blacklist_Denied", func(t *testing.T) {
		mockStore.GetActiveIPFiltersFunc = func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{IPCIDR: "1.2.3.4", FilterType: "blacklist", Reason: "bad ip"},
			}, nil
		}

		resp, err := svc.CheckIPAllowed(ctx, "1.2.3.4")
		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Equal(t, "blacklist", resp.FilterType)
	})

	t.Run("Blacklist_Allowed", func(t *testing.T) {
		mockStore.GetActiveIPFiltersFunc = func(ctx context.Context) ([]models.IPFilter, error) {
			return []models.IPFilter{
				{IPCIDR: "1.2.3.4", FilterType: "blacklist"},
			}, nil
		}

		resp, err := svc.CheckIPAllowed(ctx, "5.6.7.8")
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})
}

func TestIPFilterService_ListIPFilters(t *testing.T) {
	mockStore := &mockIPFilterStore{}
	svc := NewIPFilterService(mockStore)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockStore.ListIPFiltersFunc = func(ctx context.Context, page, perPage int, filterType string) ([]models.IPFilterWithCreator, int, error) {
			return []models.IPFilterWithCreator{}, 0, nil
		}

		resp, err := svc.ListIPFilters(ctx, 1, 10, "")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Total)
	})
}

func TestIPFilterService_UpdateAndDelete(t *testing.T) {
	mockStore := &mockIPFilterStore{}
	svc := NewIPFilterService(mockStore)
	ctx := context.Background()
	id := uuid.New()

	t.Run("Update", func(t *testing.T) {
		mockStore.UpdateIPFilterFunc = func(ctx context.Context, id uuid.UUID, reason string, isActive bool) error {
			return nil
		}

		err := svc.UpdateIPFilter(ctx, id, &models.UpdateIPFilterRequest{})
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		mockStore.DeleteIPFilterFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		err := svc.DeleteIPFilter(ctx, id)
		assert.NoError(t, err)
	})
}
