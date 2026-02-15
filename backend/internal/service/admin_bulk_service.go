package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

type AdminBulkService struct {
	userRepo   UserStore
	appRepo    ApplicationStore
	bcryptCost int
}

func (s *AdminBulkService) SyncUsers(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.userRepo.GetUsersUpdatedAfter(ctx, updatedAfter, appID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to sync users: %w", err)
	}

	syncUsers := make([]models.SyncUserResponse, len(users))
	for i, user := range users {
		syncUser := models.SyncUserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Username:      user.Username,
			FullName:      user.FullName,
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
			UpdatedAt:     user.UpdatedAt,
		}

		if len(user.ApplicationProfiles) > 0 {
			profile := user.ApplicationProfiles[0]
			appProfile := &models.SyncUserAppProfile{
				AppRoles: profile.AppRoles,
				IsActive: profile.IsActive,
				IsBanned: profile.IsBanned,
			}
			if profile.DisplayName != nil {
				appProfile.DisplayName = *profile.DisplayName
			}
			if profile.AvatarURL != nil {
				appProfile.AvatarURL = *profile.AvatarURL
			}
			syncUser.AppProfile = appProfile
		}

		syncUsers[i] = syncUser
	}

	hasMore := offset+limit < total

	return &models.SyncUsersResponse{
		Users:         syncUsers,
		Total:         total,
		HasMore:       hasMore,
		SyncTimestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *AdminBulkService) ImportUsers(ctx context.Context, req *models.BulkImportUsersRequest, appID *uuid.UUID) (*models.ImportUsersResponse, error) {
	var imported, skipped, updated, errCount int
	var details []models.ImportDetail

	for _, entry := range req.Users {
		detail := models.ImportDetail{
			Email:    entry.Email,
			Username: entry.Username,
		}
		if entry.Phone != nil {
			detail.Phone = *entry.Phone
		}

		email := utils.NormalizeEmail(entry.Email)
		var phone string
		if entry.Phone != nil && *entry.Phone != "" {
			phone = utils.NormalizePhone(*entry.Phone)
		}
		username := entry.Username

		if email == "" && phone == "" && username == "" {
			detail.Status = "error"
			detail.Reason = "at least one of email, phone, or username is required"
			errCount++
			details = append(details, detail)
			continue
		}

		if email != "" && !utils.IsValidEmail(email) {
			detail.Status = "error"
			detail.Reason = "invalid email format"
			errCount++
			details = append(details, detail)
			continue
		}

		if phone != "" && !utils.IsValidPhone(phone) {
			detail.Status = "error"
			detail.Reason = "invalid phone format"
			errCount++
			details = append(details, detail)
			continue
		}

		var existingUser *models.User
		if email != "" {
			existingUser, _ = s.userRepo.GetByEmail(ctx, email, nil)
		}
		if existingUser == nil && phone != "" {
			existingUser, _ = s.userRepo.GetByPhone(ctx, phone, nil)
		}
		if existingUser == nil && username != "" {
			existingUser, _ = s.userRepo.GetByUsername(ctx, username, nil)
		}

		if existingUser != nil {
			switch req.OnConflict {
			case "skip":
				detail.Status = "skipped"
				detail.Reason = "user already exists"
				detail.UserID = existingUser.ID.String()
				skipped++
			case "update":
				if entry.FullName != "" {
					existingUser.FullName = entry.FullName
				}
				if entry.IsActive != nil {
					existingUser.IsActive = *entry.IsActive
				}
				if entry.PasswordHashImport != "" {
					existingUser.PasswordHash = entry.PasswordHashImport
				}
				if phone != "" && existingUser.Phone == nil {
					existingUser.Phone = &phone
				}
				if email != "" && existingUser.Email == "" {
					existingUser.Email = email
				}
				if err := s.userRepo.Update(ctx, existingUser); err != nil {
					detail.Status = "error"
					detail.Reason = fmt.Sprintf("failed to update: %s", err.Error())
					errCount++
				} else {
					detail.Status = "updated"
					detail.UserID = existingUser.ID.String()
					updated++
				}
			case "error":
				detail.Status = "error"
				detail.Reason = "user already exists"
				errCount++
			}
			details = append(details, detail)
			continue
		}

		userID := uuid.New()
		if entry.ID != nil {
			userID = *entry.ID
		}

		if username == "" && email != "" {
			username = strings.Split(email, "@")[0]
		}
		if username == "" && phone != "" {
			username = strings.TrimPrefix(phone, "+")
		}
		username = utils.NormalizeUsername(username)

		if entry.Username == "" && username != "" {
			if _, err := s.userRepo.GetByUsername(ctx, username, nil); err == nil {
				suffix := userID.String()[:6]
				username = username + "-" + suffix
			}
		}

		passwordHash := entry.PasswordHashImport
		if passwordHash == "" {
			randomBytes := make([]byte, 16)
			if _, err := rand.Read(randomBytes); err != nil {
				detail.Status = "error"
				detail.Reason = fmt.Sprintf("failed to generate random password: %s", err.Error())
				errCount++
				details = append(details, detail)
				continue
			}
			randomPassword := base64.URLEncoding.EncodeToString(randomBytes)
			hash, hashErr := utils.HashPassword(randomPassword, s.bcryptCost)
			if hashErr != nil {
				detail.Status = "error"
				detail.Reason = fmt.Sprintf("failed to hash password: %s", hashErr.Error())
				errCount++
				details = append(details, detail)
				continue
			}
			passwordHash = hash
		}

		isActive := true
		if entry.IsActive != nil {
			isActive = *entry.IsActive
		}

		emailVerified := false
		if entry.EmailVerified != nil {
			emailVerified = *entry.EmailVerified
		} else if entry.SkipEmailVerification {
			emailVerified = true
		}

		phoneVerified := false
		if entry.PhoneVerified != nil {
			phoneVerified = *entry.PhoneVerified
		}

		user := &models.User{
			ID:            userID,
			Email:         email,
			Username:      username,
			PasswordHash:  passwordHash,
			FullName:      entry.FullName,
			IsActive:      isActive,
			EmailVerified: emailVerified,
			PhoneVerified: phoneVerified,
			AccountType:   string(models.AccountTypeHuman),
		}

		if phone != "" {
			user.Phone = &phone
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			detail.Status = "error"
			detail.Reason = fmt.Sprintf("failed to create: %s", err.Error())
			errCount++
			details = append(details, detail)
			continue
		}

		if appID != nil {
			profile := &models.UserApplicationProfile{
				UserID:        userID,
				ApplicationID: *appID,
				IsActive:      true,
				AppRoles:      entry.AppRoles,
			}
			_ = s.appRepo.CreateUserProfile(ctx, profile)
		}

		detail.Status = "imported"
		detail.UserID = userID.String()
		imported++
		details = append(details, detail)
	}

	return &models.ImportUsersResponse{
		Imported: imported,
		Skipped:  skipped,
		Updated:  updated,
		Errors:   errCount,
		Details:  details,
	}, nil
}
