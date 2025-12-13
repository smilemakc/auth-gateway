package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemSetting represents a system-wide configuration setting
type SystemSetting struct {
	Key         string     `json:"key" bun:"key,pk"`
	Value       string     `json:"value" bun:"value" binding:"required"`
	Description string     `json:"description,omitempty" bun:"description"`
	SettingType string     `json:"setting_type" bun:"setting_type"` // "string", "boolean", "integer", "json"
	IsPublic    bool       `json:"is_public" bun:"is_public"`       // Can be exposed to public API
	UpdatedAt   time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty" bun:"updated_by,type:uuid"`
}

// UpdateSystemSettingRequest is the request to update a system setting
type UpdateSystemSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

// SystemSettingsListResponse contains all system settings
type SystemSettingsListResponse struct {
	Settings []SystemSetting `json:"settings"`
}

// MaintenanceModeRequest is used to toggle maintenance mode
type MaintenanceModeRequest struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

// MaintenanceModeResponse returns the current maintenance mode status
type MaintenanceModeResponse struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

// System setting keys
const (
	SettingMaintenanceMode          = "maintenance_mode"
	SettingMaintenanceMessage       = "maintenance_message"
	SettingAllowNewRegistrations    = "allow_new_registrations"
	SettingRequireEmailVerification = "require_email_verification"
	SettingMaxSessionsPerUser       = "max_sessions_per_user"
	SettingSessionTimeoutHours      = "session_timeout_hours"
)

// HealthMetric represents a system health metric
type HealthMetric struct {
	ID          uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	MetricName  string    `json:"metric_name" bun:"metric_name"`
	MetricValue float64   `json:"metric_value" bun:"metric_value"`
	MetricUnit  string    `json:"metric_unit,omitempty" bun:"metric_unit"` // "bytes", "percentage", "count", "milliseconds"
	Metadata    string    `json:"metadata,omitempty" bun:"metadata"`       // JSON metadata
	RecordedAt  time.Time `json:"recorded_at" bun:"recorded_at,nullzero,notnull,default:current_timestamp"`
}

// SystemHealthResponse contains current system health metrics
type SystemHealthResponse struct {
	Status              string                 `json:"status"` // "healthy", "degraded", "down"
	DatabaseStatus      string                 `json:"database_status"`
	RedisStatus         string                 `json:"redis_status"`
	DatabaseConnections DatabaseConnectionInfo `json:"database_connections"`
	RedisMemory         RedisMemoryInfo        `json:"redis_memory"`
	Uptime              int64                  `json:"uptime_seconds"`
	Metrics             []HealthMetric         `json:"metrics,omitempty"`
}

// DatabaseConnectionInfo contains database connection pool stats
type DatabaseConnectionInfo struct {
	MaxOpen      int `json:"max_open"`
	Open         int `json:"open"`
	InUse        int `json:"in_use"`
	Idle         int `json:"idle"`
	WaitCount    int `json:"wait_count"`
	WaitDuration int `json:"wait_duration_ms"`
}

// RedisMemoryInfo contains Redis memory statistics
type RedisMemoryInfo struct {
	UsedMemory      int64   `json:"used_memory_bytes"`
	UsedMemoryHuman string  `json:"used_memory_human"`
	MaxMemory       int64   `json:"max_memory_bytes"`
	MaxMemoryHuman  string  `json:"max_memory_human"`
	MemoryUsage     float64 `json:"memory_usage_percentage"`
}

// MetricType defines available metric types
const (
	MetricTypeDatabaseConnections = "database.connections"
	MetricTypeDatabaseLatency     = "database.latency_ms"
	MetricTypeRedisMemory         = "redis.memory_bytes"
	MetricTypeRedisLatency        = "redis.latency_ms"
	MetricTypeAPIRequests         = "api.requests_count"
	MetricTypeAPIErrors           = "api.errors_count"
	MetricTypeActiveUsers         = "users.active_count"
	MetricTypeActiveSessions      = "sessions.active_count"
)
