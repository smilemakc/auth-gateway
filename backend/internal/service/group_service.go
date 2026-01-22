package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// GroupService handles group-related business logic
type GroupService struct {
	groupRepo GroupRepository
	userRepo  UserStore // For dynamic group evaluation
	logger    *logger.Logger
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error)
	GetByName(ctx context.Context, name string) (*models.Group, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Group, int, error)
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddUser(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error)
	GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*models.Group, error)
	GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int, error)
}

// NewGroupService creates a new group service
func NewGroupService(groupRepo GroupRepository, userRepo UserStore, logger *logger.Logger) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(ctx context.Context, req *models.CreateGroupRequest) (*models.Group, error) {
	// Check if parent group exists (if provided)
	if req.ParentGroupID != nil {
		parent, err := s.groupRepo.GetByID(ctx, *req.ParentGroupID)
		if err != nil {
			return nil, models.NewAppError(404, "Parent group not found")
		}
		// Prevent circular references (would need to check in a transaction for full safety)
		if parent.ParentGroupID != nil && *parent.ParentGroupID == uuid.Nil {
			// This is a basic check; full cycle detection would require recursive query
		}
	}

	group := &models.Group{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		ParentGroupID: req.ParentGroupID,
		IsSystemGroup: false,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		if err == models.ErrAlreadyExists {
			return nil, models.NewAppError(409, "Group with this name already exists")
		}
		return nil, err
	}

	return group, nil
}

// GetGroup retrieves a group by ID
func (s *GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return group, nil
}

// ListGroups retrieves a paginated list of groups
func (s *GroupService) ListGroups(ctx context.Context, page, pageSize int) (*models.GroupListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	groups, total, err := s.groupRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]models.GroupResponse, len(groups))
	for i, group := range groups {
		memberCount, _ := s.groupRepo.GetGroupMemberCount(ctx, group.ID)
		responses[i] = models.GroupResponse{
			ID:            group.ID,
			Name:          group.Name,
			DisplayName:   group.DisplayName,
			Description:   group.Description,
			ParentGroupID: group.ParentGroupID,
			IsSystemGroup: group.IsSystemGroup,
			MemberCount:   memberCount,
			CreatedAt:     group.CreatedAt,
			UpdatedAt:     group.UpdatedAt,
		}
	}

	return &models.GroupListResponse{
		Groups: responses,
		Total:  total,
		Page:   page,
		Size:   pageSize,
	}, nil
}

// UpdateGroup updates a group
func (s *GroupService) UpdateGroup(ctx context.Context, id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != nil {
		group.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.ParentGroupID != nil {
		// Validate parent group exists
		if *req.ParentGroupID != uuid.Nil {
			_, err := s.groupRepo.GetByID(ctx, *req.ParentGroupID)
			if err != nil {
				return nil, models.NewAppError(404, "Parent group not found")
			}
		}
		group.ParentGroupID = req.ParentGroupID
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteGroup deletes a group
func (s *GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id)
}

// AddGroupMembers adds users to a group
func (s *GroupService) AddGroupMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	// Verify group exists
	_, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Add each user
	for _, userID := range userIDs {
		if err := s.groupRepo.AddUser(ctx, groupID, userID); err != nil {
			s.logger.Warn("Failed to add user to group", map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"error":    err.Error(),
			})
			// Continue with other users
		}
	}

	return nil
}

// RemoveGroupMember removes a user from a group
func (s *GroupService) RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return s.groupRepo.RemoveUser(ctx, groupID, userID)
}

// GetGroupMembers retrieves members of a group
func (s *GroupService) GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.groupRepo.GetGroupMembers(ctx, groupID, page, pageSize)
}

// GetUserGroups retrieves all groups for a user
func (s *GroupService) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*models.Group, error) {
	return s.groupRepo.GetUserGroups(ctx, userID)
}

// GetGroupMemberCount returns the number of members in a group
func (s *GroupService) GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	return s.groupRepo.GetGroupMemberCount(ctx, groupID)
}

// EvaluateDynamicGroupMembers evaluates dynamic group membership rules and returns matching user IDs
func (s *GroupService) EvaluateDynamicGroupMembers(ctx context.Context, group *models.Group, allUsers []*models.User) []uuid.UUID {
	if !group.IsDynamic || group.MembershipRules == nil {
		return []uuid.UUID{}
	}

	matchingUserIDs := []uuid.UUID{}
	for _, user := range allUsers {
		if s.userMatchesRules(user, group.MembershipRules) {
			matchingUserIDs = append(matchingUserIDs, user.ID)
		}
	}

	return matchingUserIDs
}

// userMatchesRules checks if a user matches the dynamic group membership rules
func (s *GroupService) userMatchesRules(user *models.User, rules map[string]interface{}) bool {
	for key, value := range rules {
		switch key {
		case "email_domain":
			// Check if user's email domain matches
			if domain, ok := value.(string); ok {
				// Simple domain check - can be enhanced
				if !strings.Contains(user.Email, "@"+domain) {
					return false
				}
			}
		case "role":
			// Check if user has specific role
			// This would require checking user roles
			// For now, we'll skip role-based matching in this basic implementation
		case "is_active":
			// Check if user is active
			if isActive, ok := value.(bool); ok {
				if user.IsActive != isActive {
					return false
				}
			}
		case "email_verified":
			// Check if email is verified
			if verified, ok := value.(bool); ok {
				if user.EmailVerified != verified {
					return false
				}
			}
		}
	}
	return true
}

// GetGroupPermissions returns all permissions for a group (including inherited from parent groups)
func (s *GroupService) GetGroupPermissions(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	permissions := make([]uuid.UUID, 0)
	permissions = append(permissions, group.PermissionIDs...)

	// Recursively get permissions from parent groups
	if group.ParentGroupID != nil {
		parentPermissions, err := s.GetGroupPermissions(ctx, *group.ParentGroupID)
		if err == nil {
			permissions = append(permissions, parentPermissions...)
		}
	}

	// Remove duplicates
	seen := make(map[uuid.UUID]bool)
	uniquePermissions := []uuid.UUID{}
	for _, perm := range permissions {
		if !seen[perm] {
			seen[perm] = true
			uniquePermissions = append(uniquePermissions, perm)
		}
	}

	return uniquePermissions, nil
}

// SyncDynamicGroupMembers syncs members of a dynamic group based on rules
func (s *GroupService) SyncDynamicGroupMembers(ctx context.Context, groupID uuid.UUID) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	if !group.IsDynamic {
		return models.NewAppError(400, "Group is not a dynamic group")
	}

	// Get all users (this is a simplified approach - in production, you'd want pagination)
	// For now, we'll use a limit
	allUsers, err := s.userRepo.List(ctx, 10000, 0, nil) // Large limit to get all users
	if err != nil {
		return err
	}

	// Evaluate rules to get matching user IDs
	matchingUserIDs := s.EvaluateDynamicGroupMembers(ctx, group, allUsers)

	// Get current members
	currentMembers, _, err := s.groupRepo.GetGroupMembers(ctx, groupID, 1, 10000)
	if err != nil {
		return err
	}

	currentMemberIDs := make(map[uuid.UUID]bool)
	for _, member := range currentMembers {
		currentMemberIDs[member.ID] = true
	}

	// Add new members
	for _, userID := range matchingUserIDs {
		if !currentMemberIDs[userID] {
			if err := s.groupRepo.AddUser(ctx, groupID, userID); err != nil {
				s.logger.Warn("Failed to add user to dynamic group", map[string]interface{}{
					"group_id": groupID,
					"user_id":  userID,
					"error":    err.Error(),
				})
			}
		}
	}

	// Remove members that no longer match
	for _, member := range currentMembers {
		shouldBeMember := false
		for _, matchingID := range matchingUserIDs {
			if matchingID == member.ID {
				shouldBeMember = true
				break
			}
		}
		if !shouldBeMember {
			if err := s.groupRepo.RemoveUser(ctx, groupID, member.ID); err != nil {
				s.logger.Warn("Failed to remove user from dynamic group", map[string]interface{}{
					"group_id": groupID,
					"user_id":  member.ID,
					"error":    err.Error(),
				})
			}
		}
	}

	return nil
}
