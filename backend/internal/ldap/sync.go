package ldap

import (
	"context"
	"fmt"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/queryopt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SyncService handles LDAP synchronization
type SyncService struct {
	connector      *Connector
	userRepo       UserRepository
	groupRepo      GroupRepository
	ldapConfigRepo LDAPConfigRepository
	logger         *logger.Logger
}

// UserRepository defines interface for user operations
type UserRepository interface {
	GetByEmail(ctx context.Context, email string, isActive *bool, opts ...queryopt.UserGetOption) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	List(ctx context.Context, opts ...queryopt.UserListOption) ([]*models.User, error)
}

// GroupRepository defines interface for group operations
type GroupRepository interface {
	GetByName(ctx context.Context, name string) (*models.Group, error)
	Create(ctx context.Context, group *models.Group) error
	Update(ctx context.Context, group *models.Group) error
	AddUser(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error)
}

// LDAPConfigRepository defines interface for LDAP config operations
type LDAPConfigRepository interface {
	Update(ctx context.Context, config *models.LDAPConfig) error
	CreateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error
	UpdateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error)
}

// NewSyncService creates a new LDAP sync service
func NewSyncService(
	connector *Connector,
	userRepo UserRepository,
	groupRepo GroupRepository,
	ldapConfigRepo LDAPConfigRepository,
	logger *logger.Logger,
) *SyncService {
	return &SyncService{
		connector:      connector,
		userRepo:       userRepo,
		groupRepo:      groupRepo,
		ldapConfigRepo: ldapConfigRepo,
		logger:         logger,
	}
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	UsersSynced   int
	UsersCreated  int
	UsersUpdated  int
	UsersDeleted  int
	GroupsSynced  int
	GroupsCreated int
	GroupsUpdated int
	Errors        []error
}

// Sync performs a full synchronization of users and groups from LDAP
func (s *SyncService) Sync(ctx context.Context, config *models.LDAPConfig, syncUsers, syncGroups, dryRun bool) (*models.LDAPSyncLog, error) {
	startTime := time.Now()

	// Create sync log
	syncLog := &models.LDAPSyncLog{
		ID:           uuid.New(),
		LDAPConfigID: config.ID,
		Status:       "running",
		StartedAt:    startTime,
	}

	if err := s.ldapConfigRepo.CreateSyncLog(ctx, syncLog); err != nil {
		return nil, fmt.Errorf("failed to create sync log: %w", err)
	}

	result := &SyncResult{}

	// Sync users
	if syncUsers {
		userResult, err := s.syncUsers(ctx, config, dryRun)
		if err != nil {
			syncLog.Status = "failed"
			syncLog.ErrorMessage = err.Error()
			result.Errors = append(result.Errors, err)
		} else {
			result.UsersSynced = userResult.UsersSynced
			result.UsersCreated = userResult.UsersCreated
			result.UsersUpdated = userResult.UsersUpdated
			result.UsersDeleted = userResult.UsersDeleted
		}
	}

	// Sync groups
	if syncGroups {
		groupResult, err := s.syncGroups(ctx, config, dryRun)
		if err != nil {
			if syncLog.Status != "failed" {
				syncLog.Status = "partial"
			}
			if syncLog.ErrorMessage != "" {
				syncLog.ErrorMessage += "; " + err.Error()
			} else {
				syncLog.ErrorMessage = err.Error()
			}
			result.Errors = append(result.Errors, err)
		} else {
			result.GroupsSynced = groupResult.GroupsSynced
			result.GroupsCreated = groupResult.GroupsCreated
			result.GroupsUpdated = groupResult.GroupsUpdated
		}
	}

	// Update sync log
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)

	if syncLog.Status == "running" {
		if len(result.Errors) == 0 {
			syncLog.Status = "success"
		} else {
			syncLog.Status = "partial"
		}
	}

	syncLog.CompletedAt = &completedAt
	syncLog.Duration = duration.Milliseconds()
	syncLog.UsersSynced = result.UsersSynced
	syncLog.UsersCreated = result.UsersCreated
	syncLog.UsersUpdated = result.UsersUpdated
	syncLog.UsersDeleted = result.UsersDeleted
	syncLog.GroupsSynced = result.GroupsSynced
	syncLog.GroupsCreated = result.GroupsCreated
	syncLog.GroupsUpdated = result.GroupsUpdated

	if err := s.ldapConfigRepo.UpdateSyncLog(ctx, syncLog); err != nil {
		s.logger.Error("Failed to update sync log", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Update config with last sync time
	config.LastSyncAt = &completedAt
	if config.SyncInterval > 0 {
		nextSync := completedAt.Add(config.SyncInterval)
		config.NextSyncAt = &nextSync
	}
	if err := s.ldapConfigRepo.Update(ctx, config); err != nil {
		s.logger.Warn("Failed to update LDAP config sync times", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return syncLog, nil
}

// syncUsers synchronizes users from LDAP
func (s *SyncService) syncUsers(ctx context.Context, config *models.LDAPConfig, dryRun bool) (*SyncResult, error) {
	connector := NewConnector(config)
	if err := connector.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer connector.Close()

	// Search all users
	ldapUsers, err := connector.SearchUsers("")
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP users: %w", err)
	}

	result := &SyncResult{
		UsersSynced: len(ldapUsers),
	}

	// Process each user
	for _, ldapUser := range ldapUsers {
		if ldapUser.Email == "" {
			s.logger.Warn("LDAP user missing email, skipping", map[string]interface{}{
				"dn": ldapUser.DN,
				"id": ldapUser.ID,
			})
			continue
		}

		// Check if user exists
		existingUser, err := s.userRepo.GetByEmail(ctx, ldapUser.Email, nil)
		if err != nil && err != models.ErrUserNotFound {
			result.Errors = append(result.Errors, fmt.Errorf("failed to check user %s: %w", ldapUser.Email, err))
			continue
		}

		if existingUser == nil {
			// Create new user
			if !dryRun {
				newUser := &models.User{
					Email:         ldapUser.Email,
					Username:      ldapUser.ID,
					FullName:      ldapUser.FullName,
					IsActive:      true,
					EmailVerified: true, // LDAP users are pre-verified
				}
				// Generate a temporary password (should be changed on first login)
				// In production, should use secure random password
				newUser.PasswordHash = "LDAP_USER" // Placeholder, user must use LDAP auth

				if err := s.userRepo.Create(ctx, newUser); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create user %s: %w", ldapUser.Email, err))
					continue
				}
			}
			result.UsersCreated++
		} else {
			// Update existing user
			if !dryRun {
				existingUser.FullName = ldapUser.FullName
				existingUser.Username = ldapUser.ID
				if err := s.userRepo.Update(ctx, existingUser); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to update user %s: %w", ldapUser.Email, err))
					continue
				}
			}
			result.UsersUpdated++
		}
	}

	return result, nil
}

// syncGroups synchronizes groups from LDAP
func (s *SyncService) syncGroups(ctx context.Context, config *models.LDAPConfig, dryRun bool) (*SyncResult, error) {
	connector := NewConnector(config)
	if err := connector.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer connector.Close()

	// Search all groups
	ldapGroups, err := connector.SearchGroups("")
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP groups: %w", err)
	}

	result := &SyncResult{
		GroupsSynced: len(ldapGroups),
	}

	// Process each group
	for _, ldapGroup := range ldapGroups {
		if ldapGroup.Name == "" {
			s.logger.Warn("LDAP group missing name, skipping", map[string]interface{}{
				"dn": ldapGroup.DN,
				"id": ldapGroup.ID,
			})
			continue
		}

		// Check if group exists
		existingGroup, err := s.groupRepo.GetByName(ctx, ldapGroup.Name)
		if err != nil && err != models.ErrNotFound {
			result.Errors = append(result.Errors, fmt.Errorf("failed to check group %s: %w", ldapGroup.Name, err))
			continue
		}

		var groupID uuid.UUID
		if existingGroup == nil {
			// Create new group
			if !dryRun {
				newGroup := &models.Group{
					Name:        ldapGroup.Name,
					DisplayName: ldapGroup.Name,
					Description: fmt.Sprintf("Synced from LDAP: %s", ldapGroup.DN),
				}

				if err := s.groupRepo.Create(ctx, newGroup); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create group %s: %w", ldapGroup.Name, err))
					continue
				}
				groupID = newGroup.ID
			} else {
				// In dry run, use a placeholder UUID
				groupID = uuid.New()
			}
			result.GroupsCreated++
		} else {
			// Update existing group
			if !dryRun {
				existingGroup.DisplayName = ldapGroup.Name
				if err := s.groupRepo.Update(ctx, existingGroup); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to update group %s: %w", ldapGroup.Name, err))
					continue
				}
			}
			groupID = existingGroup.ID
			result.GroupsUpdated++
		}

		// Sync group members
		if !dryRun {
			if err := s.syncGroupMembers(ctx, connector, config, groupID, ldapGroup.Members); err != nil {
				s.logger.Warn("Failed to sync group members", map[string]interface{}{
					"group": ldapGroup.Name,
					"error": err.Error(),
				})
				result.Errors = append(result.Errors, fmt.Errorf("failed to sync members for group %s: %w", ldapGroup.Name, err))
			}
		}
	}

	return result, nil
}

// syncGroupMembers synchronizes members of a group from LDAP
func (s *SyncService) syncGroupMembers(ctx context.Context, connector *Connector, config *models.LDAPConfig, groupID uuid.UUID, memberDNs []string) error {
	// Get current group members
	currentMembers, _, err := s.groupRepo.GetGroupMembers(ctx, groupID, 1, 10000)
	if err != nil {
		return fmt.Errorf("failed to get current group members: %w", err)
	}

	// Create a map of current member IDs for quick lookup
	currentMemberMap := make(map[uuid.UUID]bool)
	for _, member := range currentMembers {
		currentMemberMap[member.ID] = true
	}

	// Track which users we've added from LDAP
	ldapMemberMap := make(map[uuid.UUID]bool)

	// Process each member DN from LDAP
	for _, memberDN := range memberDNs {
		if memberDN == "" {
			continue
		}

		// Search for user by DN in LDAP
		// First try direct DN search
		searchRequest := ldap.NewSearchRequest(
			memberDN,
			ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=*)",
			[]string{"dn", config.UserEmailAttribute, config.UserIDAttribute},
			nil,
		)

		sr, err := connector.conn.Search(searchRequest)
		if err != nil {
			// If direct DN search fails, try to find user by searching in user search base
			// Extract RDN (Relative Distinguished Name) from DN
			// For example, from "cn=john,ou=users,dc=example,dc=com" extract "cn=john"
			rdn := memberDN
			for i := 0; i < len(memberDN); i++ {
				if memberDN[i] == ',' {
					rdn = memberDN[:i]
					break
				}
			}

			// Search by RDN in the user search base
			searchBase := config.UserSearchBase
			if searchBase == "" {
				searchBase = config.BaseDN
			}

			searchFilter := fmt.Sprintf("(%s)", rdn)
			searchRequest = ldap.NewSearchRequest(
				searchBase,
				ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
				searchFilter,
				[]string{"dn", config.UserEmailAttribute, config.UserIDAttribute},
				nil,
			)

			sr, err = connector.conn.Search(searchRequest)
			if err != nil {
				s.logger.Warn("Failed to find LDAP member", map[string]interface{}{
					"memberDN": memberDN,
					"error":    err.Error(),
				})
				continue
			}
		}

		if len(sr.Entries) == 0 {
			s.logger.Warn("LDAP member not found", map[string]interface{}{
				"memberDN": memberDN,
			})
			continue
		}

		entry := sr.Entries[0]
		userEmail := entry.GetAttributeValue(config.UserEmailAttribute)
		if userEmail == "" {
			s.logger.Warn("LDAP member missing email", map[string]interface{}{
				"memberDN": memberDN,
			})
			continue
		}

		// Find user in our system by email
		user, err := s.userRepo.GetByEmail(ctx, userEmail, nil)
		if err != nil {
			if err == models.ErrUserNotFound {
				s.logger.Warn("User not found in system for LDAP member", map[string]interface{}{
					"memberDN": memberDN,
					"email":    userEmail,
				})
				continue
			}
			return fmt.Errorf("failed to find user %s: %w", userEmail, err)
		}

		// Add user to group if not already a member
		if !currentMemberMap[user.ID] {
			if err := s.groupRepo.AddUser(ctx, groupID, user.ID); err != nil {
				s.logger.Warn("Failed to add user to group", map[string]interface{}{
					"groupID": groupID,
					"userID":  user.ID,
					"error":   err.Error(),
				})
				continue
			}
		}

		ldapMemberMap[user.ID] = true
	}

	// Remove users from group who are no longer in LDAP
	for _, member := range currentMembers {
		if !ldapMemberMap[member.ID] {
			if err := s.groupRepo.RemoveUser(ctx, groupID, member.ID); err != nil {
				s.logger.Warn("Failed to remove user from group", map[string]interface{}{
					"groupID": groupID,
					"userID":  member.ID,
					"error":   err.Error(),
				})
				continue
			}
		}
	}

	return nil
}
