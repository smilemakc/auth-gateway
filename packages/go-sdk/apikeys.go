package authgateway

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// APIKeysService handles API key operations.
type APIKeysService struct {
	client *Client
}

// Create creates a new API key.
// The plain key is only returned once at creation - store it securely.
func (s *APIKeysService) Create(ctx context.Context, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	var resp models.CreateAPIKeyResponse
	if err := s.client.post(ctx, "/api/api-keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// List retrieves all API keys for the current user.
func (s *APIKeysService) List(ctx context.Context) ([]models.APIKey, error) {
	var resp []models.APIKey
	if err := s.client.get(ctx, "/api/api-keys", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Get retrieves a specific API key by ID.
func (s *APIKeysService) Get(ctx context.Context, id string) (*models.APIKey, error) {
	var resp models.APIKey
	if err := s.client.get(ctx, fmt.Sprintf("/api/api-keys/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update updates an API key.
func (s *APIKeysService) Update(ctx context.Context, id string, req *models.UpdateAPIKeyRequest) (*models.APIKey, error) {
	var resp models.APIKey
	if err := s.client.put(ctx, fmt.Sprintf("/api/api-keys/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Revoke revokes an API key (deactivates it).
func (s *APIKeysService) Revoke(ctx context.Context, id string) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.post(ctx, fmt.Sprintf("/api/api-keys/%s/revoke", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes an API key.
func (s *APIKeysService) Delete(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/api-keys/%s", id), nil)
}
