package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// BulkService handles bulk operations for users
type BulkService struct {
	userRepo   UserStore
	rbacRepo   RBACStore
	logger     *logger.Logger
	bcryptCost int
}

// NewBulkService creates a new bulk service
func NewBulkService(
	userRepo UserStore,
	rbacRepo RBACStore,
	logger *logger.Logger,
	bcryptCost int,
) *BulkService {
	return &BulkService{
		userRepo:   userRepo,
		rbacRepo:   rbacRepo,
		logger:     logger,
		bcryptCost: bcryptCost,
	}
}

// BulkCreateUsers creates multiple users
func (s *BulkService) BulkCreateUsers(ctx context.Context, req *models.BulkCreateUsersRequest) (*models.BulkOperationResult, error) {
	result := &models.BulkOperationResult{
		Total:   len(req.Users),
		Errors:  []models.BulkOperationError{},
		Results: []models.BulkOperationItemResult{},
	}

	for i, userReq := range req.Users {
		// Validate user data
		if userReq.Email == "" || userReq.Username == "" || userReq.Password == "" {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				Email:   userReq.Email,
				Message: "Email, username, and password are required",
			})
			continue
		}

		// Check if user already exists
		existingUser, _ := s.userRepo.GetByEmail(ctx, userReq.Email, nil)
		if existingUser != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				Email:   userReq.Email,
				Message: "User already exists",
			})
			continue
		}

		// Hash password
		passwordHash, err := utils.HashPassword(userReq.Password, s.bcryptCost)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				Email:   userReq.Email,
				Message: fmt.Sprintf("Failed to hash password: %v", err),
			})
			continue
		}

		// Create user
		user := &models.User{
			Email:         userReq.Email,
			Username:      utils.SanitizeUsername(userReq.Username),
			FullName:      utils.SanitizeHTML(userReq.FullName),
			PasswordHash:  passwordHash,
			IsActive:      userReq.IsActive,
			EmailVerified: userReq.EmailVerified,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				Email:   userReq.Email,
				Message: fmt.Sprintf("Failed to create user: %v", err),
			})
			continue
		}

		result.Success++
		result.Results = append(result.Results, models.BulkOperationItemResult{
			Index:   i,
			ID:      user.ID,
			Email:   user.Email,
			Success: true,
			Message: "User created successfully",
		})
	}

	return result, nil
}

// BulkUpdateUsers updates multiple users
func (s *BulkService) BulkUpdateUsers(ctx context.Context, req *models.BulkUpdateUsersRequest) (*models.BulkOperationResult, error) {
	result := &models.BulkOperationResult{
		Total:   len(req.Users),
		Errors:  []models.BulkOperationError{},
		Results: []models.BulkOperationItemResult{},
	}

	for i, userReq := range req.Users {
		// Get existing user
		user, err := s.userRepo.GetByID(ctx, userReq.ID, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userReq.ID.String(),
				Message: fmt.Sprintf("User not found: %v", err),
			})
			continue
		}

		// Update fields
		if userReq.Email != nil {
			user.Email = *userReq.Email
		}
		if userReq.Username != nil {
			user.Username = utils.SanitizeUsername(*userReq.Username)
		}
		if userReq.FullName != nil {
			user.FullName = utils.SanitizeHTML(*userReq.FullName)
		}
		if userReq.IsActive != nil {
			user.IsActive = *userReq.IsActive
		}

		if err := s.userRepo.Update(ctx, user); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userReq.ID.String(),
				Email:   user.Email,
				Message: fmt.Sprintf("Failed to update user: %v", err),
			})
			continue
		}

		result.Success++
		result.Results = append(result.Results, models.BulkOperationItemResult{
			Index:   i,
			ID:      user.ID,
			Email:   user.Email,
			Success: true,
			Message: "User updated successfully",
		})
	}

	return result, nil
}

// BulkDeleteUsers deletes multiple users (soft delete)
func (s *BulkService) BulkDeleteUsers(ctx context.Context, req *models.BulkDeleteUsersRequest) (*models.BulkOperationResult, error) {
	result := &models.BulkOperationResult{
		Total:   len(req.UserIDs),
		Errors:  []models.BulkOperationError{},
		Results: []models.BulkOperationItemResult{},
	}

	for i, userID := range req.UserIDs {
		// Get user to get email for error reporting
		user, _ := s.userRepo.GetByID(ctx, userID, nil)
		email := ""
		if user != nil {
			email = user.Email
		}

		// Delete user (soft delete) - get user and set is_active = false
		user, err := s.userRepo.GetByID(ctx, userID, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userID.String(),
				Email:   email,
				Message: fmt.Sprintf("Failed to get user: %v", err),
			})
			continue
		}
		user.IsActive = false
		if err := s.userRepo.Update(ctx, user); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userID.String(),
				Email:   email,
				Message: fmt.Sprintf("Failed to delete user: %v", err),
			})
			continue
		}

		result.Success++
		result.Results = append(result.Results, models.BulkOperationItemResult{
			Index:   i,
			ID:      userID,
			Email:   email,
			Success: true,
			Message: "User deleted successfully",
		})
	}

	return result, nil
}

// BulkAssignRoles assigns roles to multiple users
func (s *BulkService) BulkAssignRoles(ctx context.Context, req *models.BulkAssignRolesRequest, assignedBy uuid.UUID) (*models.BulkOperationResult, error) {
	result := &models.BulkOperationResult{
		Total:   len(req.UserIDs),
		Errors:  []models.BulkOperationError{},
		Results: []models.BulkOperationItemResult{},
	}

	for i, userID := range req.UserIDs {
		// Get user to get email for error reporting
		user, err := s.userRepo.GetByID(ctx, userID, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userID.String(),
				Message: fmt.Sprintf("User not found: %v", err),
			})
			continue
		}

		// Assign each role
		allSuccess := true
		for _, roleID := range req.RoleIDs {
			if err := s.rbacRepo.AssignRoleToUser(ctx, userID, roleID, assignedBy); err != nil {
				allSuccess = false
				s.logger.Warn("Failed to assign role to user", map[string]interface{}{
					"user_id": userID,
					"role_id": roleID,
					"error":   err.Error(),
				})
			}
		}

		if !allSuccess {
			result.Failed++
			result.Errors = append(result.Errors, models.BulkOperationError{
				Index:   i,
				ID:      userID.String(),
				Email:   user.Email,
				Message: "Failed to assign some roles",
			})
			continue
		}

		result.Success++
		result.Results = append(result.Results, models.BulkOperationItemResult{
			Index:   i,
			ID:      userID,
			Email:   user.Email,
			Success: true,
			Message: "Roles assigned successfully",
		})
	}

	return result, nil
}
