package authgateway

import (
	"context"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// ProfileService handles user profile operations.
type ProfileService struct {
	client *Client
}

// Get retrieves the current user's profile.
func (s *ProfileService) Get(ctx context.Context) (*models.User, error) {
	var resp models.User
	if err := s.client.get(ctx, "/api/auth/profile", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update updates the current user's profile.
func (s *ProfileService) Update(ctx context.Context, req *models.UpdateProfileRequest) (*models.User, error) {
	var resp models.User
	if err := s.client.put(ctx, "/api/auth/profile", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChangePassword changes the current user's password.
func (s *ProfileService) ChangePassword(ctx context.Context, oldPassword, newPassword string) (*models.MessageResponse, error) {
	req := &models.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/change-password", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
