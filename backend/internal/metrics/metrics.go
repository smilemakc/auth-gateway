package metrics

import (
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_gateway_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// Authentication metrics
	loginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_login_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // success, failure
	)

	signupTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_gateway_signup_total",
			Help: "Total number of user registrations",
		},
	)

	tokenValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_token_validations_total",
			Help: "Total number of token validations",
		},
		[]string{"type", "status"}, // type: access, refresh; status: valid, invalid
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"}, // operation: select, insert, update, delete; status: success, error
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_gateway_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	dbConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_gateway_database_connections",
			Help: "Current database connection pool statistics",
		},
		[]string{"state"}, // open, in_use, idle, max_open
	)

	// Redis metrics
	redisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"}, // operation: get, set, del; status: success, error
	)

	redisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_gateway_redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"operation"},
	)

	// Session metrics
	activeSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_gateway_active_sessions",
			Help: "Current number of active sessions",
		},
	)

	// Rate limiting metrics
	rateLimitTriggers = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_rate_limit_triggers_total",
			Help: "Total number of rate limit triggers",
		},
		[]string{"endpoint", "ip"},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "severity"}, // type: auth, db, redis, validation; severity: warning, error, critical
	)

	// LDAP sync metrics
	ldapSyncTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_ldap_sync_total",
			Help: "Total number of LDAP synchronizations",
		},
		[]string{"status"}, // success, failed
	)

	ldapSyncDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_gateway_ldap_sync_duration_seconds",
			Help:    "LDAP synchronization duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
	)

	ldapSyncUsers = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_gateway_ldap_sync_users_total",
			Help: "Total number of users synced from LDAP",
		},
		[]string{"action"}, // created, updated, deleted
	)
)

// MetricsCollector collects and exposes Prometheus metrics
type MetricsCollector struct {
	registry *prometheus.Registry
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		registry: prometheus.NewRegistry(),
	}
}

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := "2xx"
	if statusCode >= 300 && statusCode < 400 {
		status = "3xx"
	} else if statusCode >= 400 && statusCode < 500 {
		status = "4xx"
	} else if statusCode >= 500 {
		status = "5xx"
	}

	httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordLogin records a login attempt
func RecordLogin(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	loginTotal.WithLabelValues(status).Inc()
}

// RecordSignup records a user registration
func RecordSignup() {
	signupTotal.Inc()
}

// RecordTokenValidation records a token validation
func RecordTokenValidation(tokenType string, valid bool) {
	status := "invalid"
	if valid {
		status = "valid"
	}
	tokenValidations.WithLabelValues(tokenType, status).Inc()
}

// RecordDBQuery records a database query
func RecordDBQuery(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	dbQueriesTotal.WithLabelValues(operation, status).Inc()
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateDBConnections updates database connection pool metrics
func UpdateDBConnections(stats sql.DBStats) {
	dbConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
	dbConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
	dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))
	dbConnections.WithLabelValues("max_open").Set(float64(stats.MaxOpenConnections))
}

// RecordRedisOperation records a Redis operation
func RecordRedisOperation(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	redisOperationsTotal.WithLabelValues(operation, status).Inc()
	redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateActiveSessions updates the active sessions count
func UpdateActiveSessions(count int) {
	activeSessions.Set(float64(count))
}

// RecordRateLimitTrigger records a rate limit trigger
func RecordRateLimitTrigger(endpoint, ip string) {
	rateLimitTriggers.WithLabelValues(endpoint, ip).Inc()
}

// RecordError records an error
func RecordError(errorType, severity string) {
	errorsTotal.WithLabelValues(errorType, severity).Inc()
}

// RecordLDAPSync records an LDAP synchronization
func RecordLDAPSync(success bool, duration time.Duration, usersCreated, usersUpdated, usersDeleted int) {
	status := "failed"
	if success {
		status = "success"
	}
	ldapSyncTotal.WithLabelValues(status).Inc()
	ldapSyncDuration.Observe(duration.Seconds())
	ldapSyncUsers.WithLabelValues("created").Add(float64(usersCreated))
	ldapSyncUsers.WithLabelValues("updated").Add(float64(usersUpdated))
	ldapSyncUsers.WithLabelValues("deleted").Add(float64(usersDeleted))
}
