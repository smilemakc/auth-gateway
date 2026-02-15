package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

type AdminAPIKeyService struct {
	apiKeyRepo APIKeyStore
	userRepo   UserStore
}

func (s *AdminAPIKeyService) ListAPIKeys(ctx context.Context, appID *uuid.UUID, page, pageSize int) (*models.AdminAPIKeyListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var apiKeys []*models.APIKey
	var err error

	if appID != nil {
		apiKeys, err = s.apiKeyRepo.ListByApp(ctx, *appID)
	} else {
		apiKeys, err = s.apiKeyRepo.ListAll(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	total := len(apiKeys)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return &models.AdminAPIKeyListResponse{
			APIKeys:  []*models.AdminAPIKeyResponse{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}
	if end > total {
		end = total
	}

	adminAPIKeys := make([]*models.AdminAPIKeyResponse, 0, end-start)
	for i := start; i < end; i++ {
		key := apiKeys[i]
		user, _ := s.userRepo.GetByID(ctx, key.UserID, nil)

		var scopes []string
		if err := json.Unmarshal(key.Scopes, &scopes); err != nil {
			scopes = []string{}
		}

		resp := &models.AdminAPIKeyResponse{
			ID:         key.ID,
			UserID:     key.UserID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scopes:     scopes,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			IsActive:   key.IsActive,
			CreatedAt:  key.CreatedAt,
		}
		if user != nil {
			resp.Username = user.Username
			resp.UserEmail = user.Email
			resp.OwnerName = user.FullName
		}
		adminAPIKeys = append(adminAPIKeys, resp)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.AdminAPIKeyListResponse{
		APIKeys:    adminAPIKeys,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminAPIKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}
