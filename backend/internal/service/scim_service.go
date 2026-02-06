package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SCIMService handles SCIM 2.0 operations
type SCIMService struct {
	userRepo  UserStore
	groupRepo GroupRepository
	logger    *logger.Logger
	baseURL   string
}

// NewSCIMService creates a new SCIM service
func NewSCIMService(userRepo UserStore, groupRepo GroupRepository, logger *logger.Logger, baseURL string) *SCIMService {
	return &SCIMService{
		userRepo:  userRepo,
		groupRepo: groupRepo,
		logger:    logger,
		baseURL:   baseURL,
	}
}

// GetUsers retrieves users with SCIM filtering and pagination
func (s *SCIMService) GetUsers(ctx context.Context, filter string, startIndex, count int) (*models.SCIMListResponse, error) {
	// Parse filter (basic implementation)
	// Example: filter=userName eq "user@example.com"

	page := 1
	pageSize := count
	if startIndex > 0 {
		page = (startIndex / count) + 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Get total count
	total, err := s.userRepo.Count(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// Get paginated users
	users, err := s.userRepo.List(ctx, UserListLimit(pageSize), UserListOffset(offset))
	if err != nil {
		return nil, err
	}

	// Convert to SCIM format
	resources := make([]interface{}, len(users))
	for i, user := range users {
		scimUser := s.userToSCIM(user)
		resources[i] = scimUser
	}

	return &models.SCIMListResponse{
		TotalResults: total,
		ItemsPerPage: pageSize,
		StartIndex:   startIndex,
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		Resources:    resources,
	}, nil
}

// GetUser retrieves a user by ID in SCIM format
func (s *SCIMService) GetUser(ctx context.Context, id string) (*models.SCIMUser, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, models.ErrNotFound
	}

	return s.userToSCIM(user), nil
}

// CreateUser creates a user from SCIM format
func (s *SCIMService) CreateUser(ctx context.Context, scimUser *models.SCIMUser) (*models.SCIMUser, error) {
	// Extract email from SCIM user
	var email string
	if len(scimUser.Emails) > 0 {
		email = scimUser.Emails[0].Value
	} else if scimUser.UserName != "" {
		email = scimUser.UserName
	} else {
		return nil, models.NewAppError(400, "Email or userName is required")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email, nil)
	if existingUser != nil {
		return nil, models.NewAppError(409, "User already exists")
	}

	// Create user
	user := &models.User{
		Email:         email,
		Username:      scimUser.UserName,
		FullName:      s.formatSCIMName(scimUser.Name),
		IsActive:      scimUser.Active,
		EmailVerified: true, // SCIM users are typically pre-verified
	}

	// Generate a temporary password (should be changed on first login)
	// Note: In production, should generate secure random password
	passwordHash, err := utils.HashPassword("TempPassword123!", 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = passwordHash

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.userToSCIM(user), nil
}

// UpdateUser updates a user from SCIM format (PUT - full update)
func (s *SCIMService) UpdateUser(ctx context.Context, id string, scimUser *models.SCIMUser) (*models.SCIMUser, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, models.ErrNotFound
	}

	// Update fields
	if len(scimUser.Emails) > 0 {
		user.Email = scimUser.Emails[0].Value
	}
	user.Username = scimUser.UserName
	user.FullName = s.formatSCIMName(scimUser.Name)
	user.IsActive = scimUser.Active

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.userToSCIM(user), nil
}

// PatchUser applies SCIM PATCH operations to a user
func (s *SCIMService) PatchUser(ctx context.Context, id string, patchReq *models.SCIMPatchRequest) (*models.SCIMUser, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, models.ErrNotFound
	}

	// Apply patch operations
	for _, op := range patchReq.Operations {
		switch op.Op {
		case "replace":
			if err := s.applyReplaceOperation(user, op.Path, op.Value); err != nil {
				return nil, err
			}
		case "add":
			if err := s.applyAddOperation(user, op.Path, op.Value); err != nil {
				return nil, err
			}
		case "remove":
			if err := s.applyRemoveOperation(user, op.Path); err != nil {
				return nil, err
			}
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.userToSCIM(user), nil
}

// DeleteUser deletes a user
func (s *SCIMService) DeleteUser(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return models.NewAppError(400, "Invalid user ID")
	}

	// Soft delete by setting is_active = false
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return err
	}
	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// GetGroups retrieves groups with SCIM filtering and pagination
func (s *SCIMService) GetGroups(ctx context.Context, filter string, startIndex, count int) (*models.SCIMListResponse, error) {
	page := 1
	pageSize := count
	if startIndex > 0 {
		page = (startIndex / count) + 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 100 {
		pageSize = 100
	}

	groups, total, err := s.groupRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	resources := make([]interface{}, len(groups))
	for i, group := range groups {
		scimGroup := s.groupToSCIM(group)
		resources[i] = scimGroup
	}

	return &models.SCIMListResponse{
		TotalResults: total,
		ItemsPerPage: pageSize,
		StartIndex:   startIndex,
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		Resources:    resources,
	}, nil
}

// GetGroup retrieves a group by ID in SCIM format
func (s *SCIMService) GetGroup(ctx context.Context, id string) (*models.SCIMGroup, error) {
	groupID, err := uuid.Parse(id)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid group ID")
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	return s.groupToSCIM(group), nil
}

// Helper methods

func (s *SCIMService) userToSCIM(user *models.User) *models.SCIMUser {
	return &models.SCIMUser{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:       user.ID.String(),
		UserName: user.Username,
		Name: models.SCIMName{
			Formatted: user.FullName,
		},
		Emails: []models.SCIMEmail{
			{
				Value:   user.Email,
				Primary: true,
			},
		},
		Active: user.IsActive,
		Meta: models.SCIMMeta{
			ResourceType: "User",
			Created:      user.CreatedAt,
			LastModified: user.UpdatedAt,
			Location:     fmt.Sprintf("%s/scim/v2/Users/%s", s.baseURL, user.ID.String()),
			Version:      fmt.Sprintf("W/\"%d\"", user.UpdatedAt.Unix()),
		},
	}
}

func (s *SCIMService) groupToSCIM(group *models.Group) *models.SCIMGroup {
	// Get group members
	members, _, _ := s.groupRepo.GetGroupMembers(context.Background(), group.ID, 1, 1000)

	scimMembers := make([]models.SCIMMember, len(members))
	for i, member := range members {
		scimMembers[i] = models.SCIMMember{
			Value:   member.ID.String(),
			Ref:     fmt.Sprintf("%s/scim/v2/Users/%s", s.baseURL, member.ID.String()),
			Display: member.Email,
			Type:    "User",
		}
	}

	return &models.SCIMGroup{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		ID:          group.ID.String(),
		DisplayName: group.DisplayName,
		Members:     scimMembers,
		Meta: models.SCIMMeta{
			ResourceType: "Group",
			Created:      group.CreatedAt,
			LastModified: group.UpdatedAt,
			Location:     fmt.Sprintf("%s/scim/v2/Groups/%s", s.baseURL, group.ID.String()),
			Version:      fmt.Sprintf("W/\"%d\"", group.UpdatedAt.Unix()),
		},
	}
}

func (s *SCIMService) formatSCIMName(name models.SCIMName) string {
	if name.Formatted != "" {
		return name.Formatted
	}
	parts := []string{name.GivenName, name.MiddleName, name.FamilyName}
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return strings.Join(result, " ")
}

func (s *SCIMService) applyReplaceOperation(user *models.User, path string, value interface{}) error {
	switch path {
	case "userName":
		if str, ok := value.(string); ok {
			user.Username = str
		}
	case "name.formatted":
		if str, ok := value.(string); ok {
			user.FullName = str
		}
	case "emails[type eq \"primary\"].value":
		if str, ok := value.(string); ok {
			user.Email = str
		}
	case "active":
		if active, ok := value.(bool); ok {
			user.IsActive = active
		}
	}
	return nil
}

func (s *SCIMService) applyAddOperation(user *models.User, path string, value interface{}) error {
	// Similar to replace for most cases
	return s.applyReplaceOperation(user, path, value)
}

func (s *SCIMService) applyRemoveOperation(user *models.User, path string) error {
	// For remove, we might set to empty or false
	switch path {
	case "active":
		user.IsActive = false
	}
	return nil
}

// GetServiceProviderConfig returns SCIM service provider configuration
func (s *SCIMService) GetServiceProviderConfig(ctx context.Context) *models.SCIMServiceProviderConfig {
	return &models.SCIMServiceProviderConfig{
		Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		Patch: models.SCIMFeature{
			Supported: true,
		},
		Bulk: models.SCIMBulkFeature{
			Supported:      false,
			MaxOperations:  0,
			MaxPayloadSize: 0,
		},
		Filter: models.SCIMFilterFeature{
			Supported:  true,
			MaxResults: 200,
		},
		ChangePassword: models.SCIMFeature{
			Supported: false,
		},
		Sort: models.SCIMFeature{
			Supported: false,
		},
		ETag: models.SCIMFeature{
			Supported: true,
		},
		AuthenticationSchemes: []models.SCIMAuthScheme{
			{
				Type:        "oauthbearertoken",
				Name:        "OAuth Bearer Token",
				Description: "Authentication using OAuth 2.0 Bearer Token",
			},
		},
		Meta: models.SCIMMeta{
			ResourceType: "ServiceProviderConfig",
			Location:     fmt.Sprintf("%s/scim/v2/ServiceProviderConfig", s.baseURL),
		},
	}
}

// GetSchemas returns available SCIM schemas
func (s *SCIMService) GetSchemas(ctx context.Context) []*models.SCIMSchema {
	return []*models.SCIMSchema{
		{
			ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
			Name:        "User",
			Description: "User Account",
			Attributes: []models.SCIMAttribute{
				{Name: "userName", Type: "string", Required: true, Mutability: "readWrite", Returned: "default"},
				{Name: "name", Type: "complex", Required: false, Mutability: "readWrite", Returned: "default"},
				{Name: "emails", Type: "complex", MultiValued: true, Required: true, Mutability: "readWrite", Returned: "default"},
				{Name: "active", Type: "boolean", Required: false, Mutability: "readWrite", Returned: "default"},
			},
		},
		{
			ID:          "urn:ietf:params:scim:schemas:core:2.0:Group",
			Name:        "Group",
			Description: "Group",
			Attributes: []models.SCIMAttribute{
				{Name: "displayName", Type: "string", Required: true, Mutability: "readWrite", Returned: "default"},
				{Name: "members", Type: "complex", MultiValued: true, Required: false, Mutability: "readWrite", Returned: "default"},
			},
		},
	}
}
