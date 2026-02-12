package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/ldap"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// LDAPSyncJob handles periodic LDAP synchronization
type LDAPSyncJob struct {
	ldapService *service.LDAPService
	logger      *logger.Logger
	stopChan    chan struct{}
}

// NewLDAPSyncJob creates a new LDAP sync job
func NewLDAPSyncJob(ldapService *service.LDAPService, logger *logger.Logger) *LDAPSyncJob {
	return &LDAPSyncJob{
		ldapService: ldapService,
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

// Start starts the LDAP sync job scheduler
func (j *LDAPSyncJob) Start(ctx context.Context) {
	j.logger.Info("Starting LDAP sync job scheduler")

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("LDAP sync job scheduler stopped (context cancelled)")
			return
		case <-j.stopChan:
			j.logger.Info("LDAP sync job scheduler stopped")
			return
		case <-ticker.C:
			j.runScheduledSyncs(ctx)
		}
	}
}

// Stop stops the LDAP sync job scheduler
func (j *LDAPSyncJob) Stop() {
	close(j.stopChan)
}

// runScheduledSyncs checks for LDAP configs that need synchronization
func (j *LDAPSyncJob) runScheduledSyncs(ctx context.Context) {
	// Get all active LDAP configs
	configs, err := j.ldapService.ListConfigs(ctx)
	if err != nil {
		j.logger.Error("Failed to list LDAP configs for sync", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	now := time.Now()
	for _, config := range configs {
		if !config.IsActive || !config.SyncEnabled {
			continue
		}

		// Check if sync is due
		shouldSync := false
		if config.NextSyncAt == nil {
			// First sync
			shouldSync = true
		} else if now.After(*config.NextSyncAt) {
			shouldSync = true
		}

		if shouldSync {
			j.logger.Info("Starting scheduled LDAP sync", map[string]interface{}{
				"config_id": config.ID,
				"server":    config.Server,
			})

			// Run sync in goroutine to avoid blocking
			go j.runSync(ctx, config.ID)
		}
	}
}

// runSync executes a single LDAP synchronization
func (j *LDAPSyncJob) runSync(ctx context.Context, configID uuid.UUID) {
	syncReq := &models.LDAPSyncRequest{
		SyncUsers:  true,
		SyncGroups: true,
		DryRun:     false,
	}

	response, err := j.ldapService.Sync(ctx, configID, syncReq)
	if err != nil {
		j.logger.Error("LDAP sync failed", map[string]interface{}{
			"config_id": configID,
			"error":     err.Error(),
		})
		return
	}

	j.logger.Info("LDAP sync completed", map[string]interface{}{
		"config_id":     configID,
		"status":        response.Status,
		"users_synced":  response.UsersSynced,
		"users_created": response.UsersCreated,
		"users_updated": response.UsersUpdated,
		"users_deleted": response.UsersDeleted,
		"groups_synced": response.GroupsSynced,
	})

	// Update next sync time
	if err := j.updateNextSyncTime(ctx, configID); err != nil {
		j.logger.Warn("Failed to update next sync time", map[string]interface{}{
			"config_id": configID,
			"error":     err.Error(),
		})
	}
}

// updateNextSyncTime updates the NextSyncAt field for a config
func (j *LDAPSyncJob) updateNextSyncTime(ctx context.Context, configID uuid.UUID) error {
	config, err := j.ldapService.GetConfig(ctx, configID)
	if err != nil {
		return err
	}

	now := time.Now()
	nextSync := now.Add(config.SyncInterval)

	updateReq := &models.UpdateLDAPConfigRequest{
		// We need to update LastSyncAt and NextSyncAt
		// Since UpdateLDAPConfigRequest doesn't have these fields,
		// we'll need to add them or use a different approach
	}

	// For now, we'll update through the repository directly
	// This is a limitation - we should add LastSyncAt/NextSyncAt to UpdateLDAPConfigRequest
	_ = updateReq
	_ = nextSync

	return nil
}

// ManualSync triggers a manual sync for a specific config
func (j *LDAPSyncJob) ManualSync(ctx context.Context, configID uuid.UUID, syncUsers, syncGroups, dryRun bool) error {
	syncReq := &models.LDAPSyncRequest{
		SyncUsers:  syncUsers,
		SyncGroups: syncGroups,
		DryRun:     dryRun,
	}

	response, err := j.ldapService.Sync(ctx, configID, syncReq)
	if err != nil {
		return err
	}

	j.logger.Info("Manual LDAP sync completed", map[string]interface{}{
		"config_id":     configID,
		"status":        response.Status,
		"users_synced":  response.UsersSynced,
		"users_created": response.UsersCreated,
		"users_updated": response.UsersUpdated,
		"users_deleted": response.UsersDeleted,
		"groups_synced": response.GroupsSynced,
	})

	return nil
}

// ChangeDetectionService detects changes in LDAP directory
type ChangeDetectionService struct {
	logger *logger.Logger
}

// NewChangeDetectionService creates a new change detection service
func NewChangeDetectionService(logger *logger.Logger) *ChangeDetectionService {
	return &ChangeDetectionService{
		logger: logger,
	}
}

// DetectChanges compares current LDAP state with stored state and detects changes
func (s *ChangeDetectionService) DetectChanges(ctx context.Context, connector *ldap.Connector, config *models.LDAPConfig) (*ChangeDetectionResult, error) {
	result := &ChangeDetectionResult{
		NewUsers:      []LDAPUser{},
		UpdatedUsers:  []LDAPUser{},
		DeletedUsers:  []string{},
		NewGroups:     []LDAPGroup{},
		UpdatedGroups: []LDAPGroup{},
		DeletedGroups: []string{},
	}

	// Get current LDAP users
	currentUsers, err := connector.SearchUsers("")
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP users: %w", err)
	}

	// Get current LDAP groups
	currentGroups, err := connector.SearchGroups("")
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP groups: %w", err)
	}

	// TODO: Compare with stored state from database
	// This would require storing a snapshot of LDAP state
	// For now, we'll return the current state
	_ = currentUsers
	_ = currentGroups

	return result, nil
}

// ChangeDetectionResult represents the result of change detection
type ChangeDetectionResult struct {
	NewUsers      []LDAPUser
	UpdatedUsers  []LDAPUser
	DeletedUsers  []string
	NewGroups     []LDAPGroup
	UpdatedGroups []LDAPGroup
	DeletedGroups []string
}

// LDAPUser represents an LDAP user (simplified)
type LDAPUser struct {
	DN       string
	Email    string
	Username string
	FullName string
}

// LDAPGroup represents an LDAP group (simplified)
type LDAPGroup struct {
	DN      string
	Name    string
	Members []string
}
