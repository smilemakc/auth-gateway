package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

// UserService provides user operations
type UserService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetProfile retrieves a user's profile
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByIDWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user.PublicUser(), nil
}

// UpdateProfile updates a user's profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest, ip, userAgent string) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.ProfilePictureURL != "" {
		user.ProfilePictureURL = req.ProfilePictureURL
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logAudit(&userID, models.ActionUpdateProfile, models.StatusFailed, ip, userAgent, map[string]interface{}{
			"reason": "update_failed",
		})
		return nil, err
	}

	// Reload user with roles before returning
	user, err = s.userRepo.GetByIDWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Log successful update
	s.logAudit(&userID, models.ActionUpdateProfile, models.StatusSuccess, ip, userAgent, nil)

	return user.PublicUser(), nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user.PublicUser(), nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user.PublicUser(), nil
}

// List retrieves a list of users with pagination
func (s *UserService) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return public user data
	publicUsers := make([]*models.User, len(users))
	for i, user := range users {
		publicUsers[i] = user.PublicUser()
	}

	return publicUsers, nil
}

// Count returns the total number of users
func (s *UserService) Count(ctx context.Context) (int, error) {
	return s.userRepo.Count(ctx)
}

// logAudit logs an audit entry
func (s *UserService) logAudit(userID *uuid.UUID, action models.AuditAction, status models.AuditStatus, ip, userAgent string, details map[string]interface{}) {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	auditLog := models.CreateAuditLog(userID, action, status, ip, userAgent, detailsJSON)

	// Log asynchronously
	go func() {
		_ = s.auditRepo.Create(context.Background(), auditLog)
	}()
}
