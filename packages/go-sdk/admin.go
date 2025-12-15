package authgateway

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// AdminService handles administrative operations.
// All methods require admin privileges.
type AdminService struct {
	client *Client
}

// --- Statistics ---

// GetStats retrieves system statistics.
func (s *AdminService) GetStats(ctx context.Context) (*models.SystemStats, error) {
	var resp models.SystemStats
	if err := s.client.get(ctx, "/api/admin/stats", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- User Management ---

// ListUsers retrieves all users with pagination and filtering.
func (s *AdminService) ListUsers(ctx context.Context, params *models.ListUsersParams) (*models.PaginatedList[models.User], error) {
	query := ""
	if params != nil {
		query = buildQueryString(params)
	}

	var resp models.PaginatedList[models.User]
	if err := s.client.get(ctx, "/api/admin/users"+query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateUser creates a new user account.
func (s *AdminService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	var resp models.User
	if err := s.client.post(ctx, "/api/admin/users", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUser retrieves a user by ID.
func (s *AdminService) GetUser(ctx context.Context, id string) (*models.User, error) {
	var resp models.User
	if err := s.client.get(ctx, fmt.Sprintf("/api/admin/users/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUser updates a user.
func (s *AdminService) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error) {
	var resp models.User
	if err := s.client.put(ctx, fmt.Sprintf("/api/admin/users/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteUser deletes a user.
func (s *AdminService) DeleteUser(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/users/%s", id), nil)
}

// AssignRole assigns a role to a user.
func (s *AdminService) AssignRole(ctx context.Context, userID, roleID string) (*models.MessageResponse, error) {
	req := &models.AssignRoleRequest{RoleID: roleID}

	var resp models.MessageResponse
	if err := s.client.post(ctx, fmt.Sprintf("/api/admin/users/%s/roles", userID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveRole removes a role from a user.
func (s *AdminService) RemoveRole(ctx context.Context, userID, roleID string) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.delete(ctx, fmt.Sprintf("/api/admin/users/%s/roles/%s", userID, roleID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- RBAC Management ---

// ListPermissions retrieves all permissions.
func (s *AdminService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	var resp []models.Permission
	if err := s.client.get(ctx, "/api/admin/rbac/permissions", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreatePermission creates a new permission.
func (s *AdminService) CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*models.Permission, error) {
	var resp models.Permission
	if err := s.client.post(ctx, "/api/admin/rbac/permissions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRoles retrieves all roles.
func (s *AdminService) ListRoles(ctx context.Context) ([]models.Role, error) {
	var resp []models.Role
	if err := s.client.get(ctx, "/api/admin/rbac/roles", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateRole creates a new role.
func (s *AdminService) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	var resp models.Role
	if err := s.client.post(ctx, "/api/admin/rbac/roles", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRole retrieves a role by ID.
func (s *AdminService) GetRole(ctx context.Context, id string) (*models.Role, error) {
	var resp models.Role
	if err := s.client.get(ctx, fmt.Sprintf("/api/admin/rbac/roles/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRole updates a role.
func (s *AdminService) UpdateRole(ctx context.Context, id string, req *models.UpdateRoleRequest) (*models.Role, error) {
	var resp models.Role
	if err := s.client.put(ctx, fmt.Sprintf("/api/admin/rbac/roles/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRole deletes a role.
func (s *AdminService) DeleteRole(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/rbac/roles/%s", id), nil)
}

// GetPermissionMatrix retrieves the permission matrix for UI.
func (s *AdminService) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrixResponse, error) {
	var resp models.PermissionMatrixResponse
	if err := s.client.get(ctx, "/api/admin/rbac/permission-matrix", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- API Keys Management ---

// ListAllAPIKeys retrieves all API keys across all users.
func (s *AdminService) ListAllAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	var resp []models.APIKey
	if err := s.client.get(ctx, "/api/admin/api-keys", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// RevokeUserAPIKey revokes a user's API key.
func (s *AdminService) RevokeUserAPIKey(ctx context.Context, id string) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.post(ctx, fmt.Sprintf("/api/admin/api-keys/%s/revoke", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Audit Logs ---

// ListAuditLogs retrieves audit logs with filtering.
func (s *AdminService) ListAuditLogs(ctx context.Context, params *models.ListAuditLogsParams) (*models.PaginatedList[models.AuditLog], error) {
	query := ""
	if params != nil {
		query = buildQueryString(params)
	}

	var resp models.PaginatedList[models.AuditLog]
	if err := s.client.get(ctx, "/api/admin/audit-logs"+query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Session Management ---

// ListAllSessions retrieves all sessions across all users.
func (s *AdminService) ListAllSessions(ctx context.Context, params *models.ListSessionsParams) (*models.PaginatedList[models.Session], error) {
	query := ""
	if params != nil {
		query = buildQueryString(params)
	}

	var resp models.PaginatedList[models.Session]
	if err := s.client.get(ctx, "/api/admin/sessions"+query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSessionStats retrieves session statistics.
func (s *AdminService) GetSessionStats(ctx context.Context) (*models.SessionStats, error) {
	var resp models.SessionStats
	if err := s.client.get(ctx, "/api/admin/sessions/stats", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- IP Filters ---

// ListIPFilters retrieves all IP filters.
func (s *AdminService) ListIPFilters(ctx context.Context) ([]models.IPFilter, error) {
	var resp []models.IPFilter
	if err := s.client.get(ctx, "/api/admin/ip-filters", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateIPFilter creates a new IP filter.
func (s *AdminService) CreateIPFilter(ctx context.Context, req *models.CreateIPFilterRequest) (*models.IPFilter, error) {
	var resp models.IPFilter
	if err := s.client.post(ctx, "/api/admin/ip-filters", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteIPFilter deletes an IP filter.
func (s *AdminService) DeleteIPFilter(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/ip-filters/%s", id), nil)
}

// --- System Configuration ---

// UpdateBranding updates branding settings.
func (s *AdminService) UpdateBranding(ctx context.Context, req *models.UpdateBrandingRequest) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.put(ctx, "/api/admin/branding", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetMaintenanceMode enables or disables maintenance mode.
func (s *AdminService) SetMaintenanceMode(ctx context.Context, req *models.MaintenanceModeRequest) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.put(ctx, "/api/admin/system/maintenance", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSystemHealth retrieves detailed system health information.
func (s *AdminService) GetSystemHealth(ctx context.Context) (*models.HealthStatus, error) {
	var resp models.HealthStatus
	if err := s.client.get(ctx, "/api/admin/system/health", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Analytics ---

// GetGeoDistribution retrieves user geographic distribution.
func (s *AdminService) GetGeoDistribution(ctx context.Context) ([]models.GeoDistribution, error) {
	var resp []models.GeoDistribution
	if err := s.client.get(ctx, "/api/admin/analytics/geo-distribution", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// --- OAuth Client Management ---

// CreateOAuthClient creates a new OAuth client.
func (s *AdminService) CreateOAuthClient(ctx context.Context, req *models.CreateOAuthClientRequest) (*models.CreateOAuthClientResponse, error) {
	var resp models.CreateOAuthClientResponse
	if err := s.client.post(ctx, "/api/admin/oauth/clients", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListOAuthClients lists OAuth clients with pagination.
func (s *AdminService) ListOAuthClients(ctx context.Context, page, perPage int, ownerID *string) (*models.ListOAuthClientsResponse, error) {
	params := map[string]string{
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}
	if ownerID != nil {
		params["owner_id"] = *ownerID
	}

	path := "/api/admin/oauth/clients"
	if len(params) > 0 {
		query := buildQueryStringFromMap(params)
		path += query
	}

	var resp models.ListOAuthClientsResponse
	if err := s.client.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOAuthClient retrieves an OAuth client by ID.
func (s *AdminService) GetOAuthClient(ctx context.Context, id string) (*models.OAuthClient, error) {
	var resp models.OAuthClient
	if err := s.client.get(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOAuthClient updates an OAuth client.
func (s *AdminService) UpdateOAuthClient(ctx context.Context, id string, req *models.UpdateOAuthClientRequest) (*models.OAuthClient, error) {
	var resp models.OAuthClient
	if err := s.client.put(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteOAuthClient deletes an OAuth client.
func (s *AdminService) DeleteOAuthClient(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s", id), nil)
}

// RotateOAuthClientSecret rotates an OAuth client's secret.
func (s *AdminService) RotateOAuthClientSecret(ctx context.Context, id string) (*models.RotateSecretResponse, error) {
	var resp models.RotateSecretResponse
	if err := s.client.post(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s/rotate-secret", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- OAuth Scope Management ---

// ListOAuthScopes lists all OAuth scopes.
func (s *AdminService) ListOAuthScopes(ctx context.Context) (*models.ListScopesResponse, error) {
	var resp models.ListScopesResponse
	if err := s.client.get(ctx, "/api/admin/oauth/scopes", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateOAuthScope creates a custom OAuth scope.
func (s *AdminService) CreateOAuthScope(ctx context.Context, req *models.CreateScopeRequest) (*models.OAuthScope, error) {
	var resp models.OAuthScope
	if err := s.client.post(ctx, "/api/admin/oauth/scopes", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteOAuthScope deletes a non-system OAuth scope.
func (s *AdminService) DeleteOAuthScope(ctx context.Context, id string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/oauth/scopes/%s", id), nil)
}

// --- User Consent Management ---

// ListOAuthClientConsents lists all user consents for an OAuth client.
func (s *AdminService) ListOAuthClientConsents(ctx context.Context, clientID string) (*models.ListConsentsResponse, error) {
	var resp models.ListConsentsResponse
	if err := s.client.get(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s/consents", clientID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeOAuthUserConsent revokes a user's consent for an OAuth client.
func (s *AdminService) RevokeOAuthUserConsent(ctx context.Context, clientID, userID string) error {
	return s.client.delete(ctx, fmt.Sprintf("/api/admin/oauth/clients/%s/consents/%s", clientID, userID), nil)
}
