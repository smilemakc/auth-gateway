package authgateway

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// SessionsService handles session management operations.
type SessionsService struct {
	client *Client
}

// List retrieves all active sessions for the current user.
func (s *SessionsService) List(ctx context.Context) ([]models.Session, error) {
	var resp []models.Session
	if err := s.client.get(ctx, "/api/sessions", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Revoke revokes a specific session.
func (s *SessionsService) Revoke(ctx context.Context, id string) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.delete(ctx, fmt.Sprintf("/api/sessions/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeAll revokes all sessions except the current one.
func (s *SessionsService) RevokeAll(ctx context.Context) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/sessions/revoke-all", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
