package authgateway

import (
	"context"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// Health performs a full health check on the Auth Gateway.
func (c *Client) Health(ctx context.Context) (*models.HealthStatus, error) {
	var resp models.HealthStatus
	if err := c.get(ctx, "/auth/health", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Ready checks if the Auth Gateway is ready to accept requests.
func (c *Client) Ready(ctx context.Context) (*models.HealthStatus, error) {
	var resp models.HealthStatus
	if err := c.get(ctx, "/auth/ready", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Live checks if the Auth Gateway is alive.
func (c *Client) Live(ctx context.Context) (*models.HealthStatus, error) {
	var resp models.HealthStatus
	if err := c.get(ctx, "/auth/live", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MaintenanceStatus checks the maintenance mode status.
func (c *Client) MaintenanceStatus(ctx context.Context) (*models.MaintenanceStatus, error) {
	var resp models.MaintenanceStatus
	if err := c.get(ctx, "/system/maintenance", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
