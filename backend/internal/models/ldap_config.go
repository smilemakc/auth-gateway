package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// LDAPConfig represents LDAP/Active Directory configuration
type LDAPConfig struct {
	bun.BaseModel `bun:"table:ldap_configs"`

	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`

	// Connection settings
	Server   string `json:"server" bun:"server,notnull" example:"ldap.example.com"`
	Port     int    `json:"port" bun:"port,notnull,default:389" example:"389"`
	UseTLS   bool   `json:"use_tls" bun:"use_tls,notnull,default:false" example:"false"`
	UseSSL   bool   `json:"use_ssl" bun:"use_ssl,notnull,default:false" example:"false"`
	Insecure bool   `json:"insecure" bun:"insecure,notnull,default:false" example:"false"` // Skip certificate verification

	// Authentication
	BindDN       string `json:"bind_dn" bun:"bind_dn,notnull" example:"cn=admin,dc=example,dc=com"`
	BindPassword string `json:"bind_password" bun:"bind_password,notnull"` // Encrypted in DB

	// Search configuration
	BaseDN            string `json:"base_dn" bun:"base_dn,notnull" example:"dc=example,dc=com"`
	UserSearchBase    string `json:"user_search_base" bun:"user_search_base" example:"ou=users,dc=example,dc=com"`
	GroupSearchBase   string `json:"group_search_base" bun:"group_search_base" example:"ou=groups,dc=example,dc=com"`
	UserSearchFilter  string `json:"user_search_filter" bun:"user_search_filter,notnull,default:(objectClass=person)" example:"(objectClass=person)"`
	GroupSearchFilter string `json:"group_search_filter" bun:"group_search_filter,notnull,default:(objectClass=group)" example:"(objectClass=group)"`

	// Attribute mappings
	UserIDAttribute    string `json:"user_id_attribute" bun:"user_id_attribute,notnull,default:uid" example:"uid"`
	UserEmailAttribute string `json:"user_email_attribute" bun:"user_email_attribute,notnull,default:mail" example:"mail"`
	UserNameAttribute  string `json:"user_name_attribute" bun:"user_name_attribute,notnull,default:cn" example:"cn"`
	UserDNAttribute    string `json:"user_dn_attribute" bun:"user_dn_attribute,notnull,default:dn" example:"dn"`

	GroupIDAttribute     string `json:"group_id_attribute" bun:"group_id_attribute,notnull,default:cn" example:"cn"`
	GroupNameAttribute   string `json:"group_name_attribute" bun:"group_name_attribute,notnull,default:cn" example:"cn"`
	GroupMemberAttribute string `json:"group_member_attribute" bun:"group_member_attribute,notnull,default:member" example:"member"`

	// Sync settings
	SyncEnabled  bool          `json:"sync_enabled" bun:"sync_enabled,notnull,default:false" example:"false"`
	SyncInterval time.Duration `json:"sync_interval" bun:"sync_interval,notnull,default:3600000000000" example:"3600000000000"` // 1 hour in nanoseconds
	LastSyncAt   *time.Time    `json:"last_sync_at,omitempty" bun:"last_sync_at"`
	NextSyncAt   *time.Time    `json:"next_sync_at,omitempty" bun:"next_sync_at"`

	// Status
	IsActive       bool       `json:"is_active" bun:"is_active,notnull,default:true" example:"true"`
	LastTestAt     *time.Time `json:"last_test_at,omitempty" bun:"last_test_at"`
	LastTestResult string     `json:"last_test_result,omitempty" bun:"last_test_result" example:"success"`

	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// LDAPSyncLog represents a log entry for LDAP synchronization
type LDAPSyncLog struct {
	bun.BaseModel `bun:"table:ldap_sync_logs"`

	ID uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`

	LDAPConfigID uuid.UUID `json:"ldap_config_id" bun:"ldap_config_id,type:uuid,notnull"`

	// Sync results
	Status        string `json:"status" bun:"status,notnull" example:"success"` // success, failed, partial
	UsersSynced   int    `json:"users_synced" bun:"users_synced,notnull,default:0" example:"10"`
	UsersCreated  int    `json:"users_created" bun:"users_created,notnull,default:0" example:"5"`
	UsersUpdated  int    `json:"users_updated" bun:"users_updated,notnull,default:0" example:"3"`
	UsersDeleted  int    `json:"users_deleted" bun:"users_deleted,notnull,default:0" example:"2"`
	GroupsSynced  int    `json:"groups_synced" bun:"groups_synced,notnull,default:0" example:"5"`
	GroupsCreated int    `json:"groups_created" bun:"groups_created,notnull,default:0" example:"2"`
	GroupsUpdated int    `json:"groups_updated" bun:"groups_updated,notnull,default:0" example:"1"`

	// Error information
	ErrorMessage string `json:"error_message,omitempty" bun:"error_message" example:"Connection timeout"`

	StartedAt   time.Time  `json:"started_at" bun:"started_at,nullzero,notnull,default:current_timestamp"`
	CompletedAt *time.Time `json:"completed_at,omitempty" bun:"completed_at"`
	Duration    int64      `json:"duration_ms" bun:"duration_ms" example:"5000"` // Duration in milliseconds
}

// CreateLDAPConfigRequest represents a request to create LDAP configuration
type CreateLDAPConfigRequest struct {
	Server   string `json:"server" binding:"required" example:"ldap.example.com"`
	Port     int    `json:"port" binding:"required" example:"389"`
	UseTLS   bool   `json:"use_tls" example:"false"`
	UseSSL   bool   `json:"use_ssl" example:"false"`
	Insecure bool   `json:"insecure" example:"false"`

	BindDN       string `json:"bind_dn" binding:"required" example:"cn=admin,dc=example,dc=com"`
	BindPassword string `json:"bind_password" binding:"required"`

	BaseDN            string `json:"base_dn" binding:"required" example:"dc=example,dc=com"`
	UserSearchBase    string `json:"user_search_base,omitempty" example:"ou=users,dc=example,dc=com"`
	GroupSearchBase   string `json:"group_search_base,omitempty" example:"ou=groups,dc=example,dc=com"`
	UserSearchFilter  string `json:"user_search_filter,omitempty" example:"(objectClass=person)"`
	GroupSearchFilter string `json:"group_search_filter,omitempty" example:"(objectClass=group)"`

	UserIDAttribute    string `json:"user_id_attribute,omitempty" example:"uid"`
	UserEmailAttribute string `json:"user_email_attribute,omitempty" example:"mail"`
	UserNameAttribute  string `json:"user_name_attribute,omitempty" example:"cn"`

	GroupIDAttribute     string `json:"group_id_attribute,omitempty" example:"cn"`
	GroupNameAttribute   string `json:"group_name_attribute,omitempty" example:"cn"`
	GroupMemberAttribute string `json:"group_member_attribute,omitempty" example:"member"`

	SyncEnabled  bool `json:"sync_enabled" example:"false"`
	SyncInterval int  `json:"sync_interval" example:"3600"` // in seconds
}

// UpdateLDAPConfigRequest represents a request to update LDAP configuration
type UpdateLDAPConfigRequest struct {
	Server   *string `json:"server,omitempty"`
	Port     *int    `json:"port,omitempty"`
	UseTLS   *bool   `json:"use_tls,omitempty"`
	UseSSL   *bool   `json:"use_ssl,omitempty"`
	Insecure *bool   `json:"insecure,omitempty"`

	BindDN       *string `json:"bind_dn,omitempty"`
	BindPassword *string `json:"bind_password,omitempty"`

	BaseDN            *string `json:"base_dn,omitempty"`
	UserSearchBase    *string `json:"user_search_base,omitempty"`
	GroupSearchBase   *string `json:"group_search_base,omitempty"`
	UserSearchFilter  *string `json:"user_search_filter,omitempty"`
	GroupSearchFilter *string `json:"group_search_filter,omitempty"`

	SyncEnabled  *bool `json:"sync_enabled,omitempty"`
	SyncInterval *int  `json:"sync_interval,omitempty"` // in seconds
	IsActive     *bool `json:"is_active,omitempty"`
}

// LDAPTestConnectionRequest represents a request to test LDAP connection
type LDAPTestConnectionRequest struct {
	Server   string `json:"server" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	UseTLS   bool   `json:"use_tls"`
	UseSSL   bool   `json:"use_ssl"`
	Insecure bool   `json:"insecure"`

	BindDN       string `json:"bind_dn" binding:"required"`
	BindPassword string `json:"bind_password" binding:"required"`

	BaseDN string `json:"base_dn" binding:"required"`
}

// LDAPTestConnectionResponse represents a response from LDAP connection test
type LDAPTestConnectionResponse struct {
	Success    bool   `json:"success" example:"true"`
	Message    string `json:"message" example:"Connection successful"`
	Error      string `json:"error,omitempty" example:""`
	UserCount  int    `json:"user_count,omitempty" example:"100"`
	GroupCount int    `json:"group_count,omitempty" example:"20"`
}

// LDAPSyncRequest represents a request to trigger manual LDAP sync
type LDAPSyncRequest struct {
	SyncUsers  bool `json:"sync_users" example:"true"`
	SyncGroups bool `json:"sync_groups" example:"true"`
	DryRun     bool `json:"dry_run" example:"false"` // If true, only simulate without making changes
}

// LDAPSyncResponse represents a response from LDAP sync operation
type LDAPSyncResponse struct {
	Status        string    `json:"status" example:"success"`
	SyncLogID     uuid.UUID `json:"sync_log_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UsersSynced   int       `json:"users_synced" example:"10"`
	UsersCreated  int       `json:"users_created" example:"5"`
	UsersUpdated  int       `json:"users_updated" example:"3"`
	UsersDeleted  int       `json:"users_deleted" example:"2"`
	GroupsSynced  int       `json:"groups_synced" example:"5"`
	GroupsCreated int       `json:"groups_created" example:"2"`
	GroupsUpdated int       `json:"groups_updated" example:"1"`
	Message       string    `json:"message" example:"Sync completed successfully"`
	Error         string    `json:"error,omitempty"`
}

// LDAPSyncLogListResponse represents paginated LDAP sync log list
type LDAPSyncLogListResponse struct {
	// List of sync logs
	Logs []*LDAPSyncLog `json:"logs"`
	// Total number of logs
	Total int `json:"total" example:"50"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PageSize int `json:"page_size" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"3"`
}

// LDAPConfigListResponse represents LDAP configurations list
type LDAPConfigListResponse struct {
	// List of LDAP configurations
	Configs []*LDAPConfig `json:"configs"`
	// Total number of configurations
	Total int `json:"total" example:"2"`
}
