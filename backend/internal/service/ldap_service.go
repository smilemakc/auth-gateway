package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/ldap"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// LDAPService handles LDAP configuration and operations
type LDAPService struct {
	ldapRepo      LDAPConfigRepository
	userRepo      UserStore
	groupRepo     LDAPGroupRepository
	logger        *logger.Logger
	encryptionKey string
}

// LDAPGroupRepository defines interface for group operations needed by LDAP sync
type LDAPGroupRepository interface {
	GetByName(ctx context.Context, name string) (*models.Group, error)
	Create(ctx context.Context, group *models.Group) error
	Update(ctx context.Context, group *models.Group) error
	AddUser(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error)
}

// LDAPConfigRepository defines interface for LDAP config operations
type LDAPConfigRepository interface {
	Create(ctx context.Context, config *models.LDAPConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error)
	GetActive(ctx context.Context) (*models.LDAPConfig, error)
	List(ctx context.Context) ([]*models.LDAPConfig, error)
	Update(ctx context.Context, config *models.LDAPConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error
	GetSyncLogs(ctx context.Context, configID uuid.UUID, limit, offset int) ([]*models.LDAPSyncLog, int, error)
	UpdateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error
}

// NewLDAPService creates a new LDAP service
func NewLDAPService(
	ldapRepo LDAPConfigRepository,
	userRepo UserStore,
	groupRepo LDAPGroupRepository,
	logger *logger.Logger,
	encryptionKey string,
) *LDAPService {
	return &LDAPService{
		ldapRepo:      ldapRepo,
		userRepo:      userRepo,
		groupRepo:     groupRepo,
		logger:        logger,
		encryptionKey: encryptionKey,
	}
}

// CreateConfig creates a new LDAP configuration
func (s *LDAPService) CreateConfig(ctx context.Context, req *models.CreateLDAPConfigRequest) (*models.LDAPConfig, error) {
	bindPassword, err := s.encryptBindPassword(req.BindPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bind password: %w", err)
	}

	config := &models.LDAPConfig{
		Server:               req.Server,
		Port:                 req.Port,
		UseTLS:               req.UseTLS,
		UseSSL:               req.UseSSL,
		Insecure:             req.Insecure,
		BindDN:               req.BindDN,
		BindPassword:         bindPassword,
		BaseDN:               req.BaseDN,
		UserSearchBase:       req.UserSearchBase,
		GroupSearchBase:      req.GroupSearchBase,
		UserSearchFilter:     req.UserSearchFilter,
		GroupSearchFilter:    req.GroupSearchFilter,
		UserIDAttribute:      req.UserIDAttribute,
		UserEmailAttribute:   req.UserEmailAttribute,
		UserNameAttribute:    req.UserNameAttribute,
		GroupIDAttribute:     req.GroupIDAttribute,
		GroupNameAttribute:   req.GroupNameAttribute,
		GroupMemberAttribute: req.GroupMemberAttribute,
		SyncEnabled:          req.SyncEnabled,
		IsActive:             true,
	}

	if req.UserSearchFilter == "" {
		config.UserSearchFilter = "(objectClass=person)"
	}
	if req.GroupSearchFilter == "" {
		config.GroupSearchFilter = "(objectClass=group)"
	}
	if req.UserIDAttribute == "" {
		config.UserIDAttribute = "uid"
	}
	if req.UserEmailAttribute == "" {
		config.UserEmailAttribute = "mail"
	}
	if req.UserNameAttribute == "" {
		config.UserNameAttribute = "cn"
	}
	if req.GroupIDAttribute == "" {
		config.GroupIDAttribute = "cn"
	}
	if req.GroupNameAttribute == "" {
		config.GroupNameAttribute = "cn"
	}
	if req.GroupMemberAttribute == "" {
		config.GroupMemberAttribute = "member"
	}

	if req.SyncInterval > 0 {
		config.SyncInterval = time.Duration(req.SyncInterval) * time.Second
	} else {
		config.SyncInterval = time.Hour
	}

	if err := s.ldapRepo.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetConfig retrieves an LDAP configuration by ID
func (s *LDAPService) GetConfig(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error) {
	return s.ldapRepo.GetByID(ctx, id)
}

// GetActiveConfig retrieves the active LDAP configuration
func (s *LDAPService) GetActiveConfig(ctx context.Context) (*models.LDAPConfig, error) {
	return s.ldapRepo.GetActive(ctx)
}

// ListConfigs retrieves all LDAP configurations
func (s *LDAPService) ListConfigs(ctx context.Context) ([]*models.LDAPConfig, error) {
	return s.ldapRepo.List(ctx)
}

// UpdateConfig updates an LDAP configuration
func (s *LDAPService) UpdateConfig(ctx context.Context, id uuid.UUID, req *models.UpdateLDAPConfigRequest) (*models.LDAPConfig, error) {
	config, err := s.ldapRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Server != nil {
		config.Server = *req.Server
	}
	if req.Port != nil {
		config.Port = *req.Port
	}
	if req.UseTLS != nil {
		config.UseTLS = *req.UseTLS
	}
	if req.UseSSL != nil {
		config.UseSSL = *req.UseSSL
	}
	if req.Insecure != nil {
		config.Insecure = *req.Insecure
	}
	if req.BindDN != nil {
		config.BindDN = *req.BindDN
	}
	if req.BindPassword != nil {
		encryptedPassword, encErr := s.encryptBindPassword(*req.BindPassword)
		if encErr != nil {
			return nil, fmt.Errorf("failed to encrypt bind password: %w", encErr)
		}
		config.BindPassword = encryptedPassword
	}
	if req.BaseDN != nil {
		config.BaseDN = *req.BaseDN
	}
	if req.UserSearchBase != nil {
		config.UserSearchBase = *req.UserSearchBase
	}
	if req.GroupSearchBase != nil {
		config.GroupSearchBase = *req.GroupSearchBase
	}
	if req.UserSearchFilter != nil {
		config.UserSearchFilter = *req.UserSearchFilter
	}
	if req.GroupSearchFilter != nil {
		config.GroupSearchFilter = *req.GroupSearchFilter
	}
	if req.SyncEnabled != nil {
		config.SyncEnabled = *req.SyncEnabled
	}
	if req.SyncInterval != nil {
		config.SyncInterval = time.Duration(*req.SyncInterval) * time.Second
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	if err := s.ldapRepo.Update(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// DeleteConfig deletes an LDAP configuration
func (s *LDAPService) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	return s.ldapRepo.Delete(ctx, id)
}

// TestConnection tests the LDAP connection
func (s *LDAPService) TestConnection(ctx context.Context, req *models.LDAPTestConnectionRequest) (*models.LDAPTestConnectionResponse, error) {
	// Create temporary config for testing
	testConfig := &models.LDAPConfig{
		Server:               req.Server,
		Port:                 req.Port,
		UseTLS:               req.UseTLS,
		UseSSL:               req.UseSSL,
		Insecure:             req.Insecure,
		BindDN:               req.BindDN,
		BindPassword:         req.BindPassword,
		BaseDN:               req.BaseDN,
		UserSearchFilter:     "(objectClass=person)",
		GroupSearchFilter:    "(objectClass=group)",
		UserIDAttribute:      "uid",
		UserEmailAttribute:   "mail",
		UserNameAttribute:    "cn",
		GroupIDAttribute:     "cn",
		GroupNameAttribute:   "cn",
		GroupMemberAttribute: "member",
	}

	connector := ldap.NewConnector(testConfig)
	if err := connector.Connect(); err != nil {
		return &models.LDAPTestConnectionResponse{
			Success: false,
			Message: "Connection failed",
			Error:   err.Error(),
		}, nil
	}
	defer connector.Close()

	// Try to search for users and groups
	users, err := connector.SearchUsers("")
	userCount := 0
	if err == nil {
		userCount = len(users)
	}

	groups, err2 := connector.SearchGroups("")
	groupCount := 0
	if err2 == nil {
		groupCount = len(groups)
	}

	if err != nil && err2 != nil {
		return &models.LDAPTestConnectionResponse{
			Success: false,
			Message: "Connection successful but search failed",
			Error:   fmt.Sprintf("User search: %v; Group search: %v", err, err2),
		}, nil
	}

	return &models.LDAPTestConnectionResponse{
		Success:    true,
		Message:    "Connection successful",
		UserCount:  userCount,
		GroupCount: groupCount,
	}, nil
}

// Sync performs LDAP synchronization
func (s *LDAPService) Sync(ctx context.Context, configID uuid.UUID, req *models.LDAPSyncRequest) (*models.LDAPSyncResponse, error) {
	config, err := s.ldapRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	decryptedPassword, err := s.decryptBindPassword(config.BindPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bind password: %w", err)
	}
	config.BindPassword = decryptedPassword

	connector := ldap.NewConnector(config)
	syncService := ldap.NewSyncService(
		connector,
		s.userRepo,
		s.groupRepo,
		s.ldapRepo,
		s.logger,
	)

	syncLog, err := syncService.Sync(ctx, config, req.SyncUsers, req.SyncGroups, req.DryRun)
	if err != nil {
		return &models.LDAPSyncResponse{
			Status:  "failed",
			Message: "Sync failed",
			Error:   err.Error(),
		}, nil
	}

	return &models.LDAPSyncResponse{
		Status:        syncLog.Status,
		SyncLogID:     syncLog.ID,
		UsersSynced:   syncLog.UsersSynced,
		UsersCreated:  syncLog.UsersCreated,
		UsersUpdated:  syncLog.UsersUpdated,
		UsersDeleted:  syncLog.UsersDeleted,
		GroupsSynced:  syncLog.GroupsSynced,
		GroupsCreated: syncLog.GroupsCreated,
		GroupsUpdated: syncLog.GroupsUpdated,
		Message:       "Sync completed successfully",
	}, nil
}

// GetSyncLogs retrieves sync logs for an LDAP configuration
func (s *LDAPService) GetSyncLogs(ctx context.Context, configID uuid.UUID, page, pageSize int) ([]*models.LDAPSyncLog, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	return s.ldapRepo.GetSyncLogs(ctx, configID, pageSize, offset)
}

func (s *LDAPService) encryptBindPassword(plaintext string) (string, error) {
	if s.encryptionKey == "" {
		return plaintext, nil
	}
	return utils.EncryptAESGCM(plaintext, s.encryptionKey)
}

func (s *LDAPService) decryptBindPassword(ciphertext string) (string, error) {
	if s.encryptionKey == "" {
		return ciphertext, nil
	}
	return utils.DecryptAESGCM(ciphertext, s.encryptionKey)
}

// Adapters to convert service interfaces to LDAP sync interfaces

type ldapUserRepoAdapter struct {
	repo UserStore
}

func (a *ldapUserRepoAdapter) GetByEmail(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
	return a.repo.GetByEmail(ctx, email, isActive, opts...)
}

func (a *ldapUserRepoAdapter) Create(ctx context.Context, user *models.User) error {
	return a.repo.Create(ctx, user)
}

func (a *ldapUserRepoAdapter) Update(ctx context.Context, user *models.User) error {
	return a.repo.Update(ctx, user)
}

func (a *ldapUserRepoAdapter) List(ctx context.Context, opts ...UserListOption) ([]*models.User, error) {
	return a.repo.List(ctx, opts...)
}

type ldapGroupRepoAdapter struct {
	repo LDAPGroupRepository
}

func (a *ldapGroupRepoAdapter) GetByName(ctx context.Context, name string) (*models.Group, error) {
	return a.repo.GetByName(ctx, name)
}

func (a *ldapGroupRepoAdapter) Create(ctx context.Context, group *models.Group) error {
	return a.repo.Create(ctx, group)
}

func (a *ldapGroupRepoAdapter) Update(ctx context.Context, group *models.Group) error {
	return a.repo.Update(ctx, group)
}

func (a *ldapGroupRepoAdapter) AddUser(ctx context.Context, groupID, userID uuid.UUID) error {
	return a.repo.AddUser(ctx, groupID, userID)
}

func (a *ldapGroupRepoAdapter) RemoveUser(ctx context.Context, groupID, userID uuid.UUID) error {
	return a.repo.RemoveUser(ctx, groupID, userID)
}

func (a *ldapGroupRepoAdapter) GetGroupMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
	return a.repo.GetGroupMembers(ctx, groupID, page, pageSize)
}

type ldapConfigRepoAdapter struct {
	repo LDAPConfigRepository
}

func (a *ldapConfigRepoAdapter) Update(ctx context.Context, config *models.LDAPConfig) error {
	return a.repo.Update(ctx, config)
}

func (a *ldapConfigRepoAdapter) CreateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	return a.repo.CreateSyncLog(ctx, log)
}

func (a *ldapConfigRepoAdapter) UpdateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	return a.repo.UpdateSyncLog(ctx, log)
}

func (a *ldapConfigRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error) {
	return a.repo.GetByID(ctx, id)
}
