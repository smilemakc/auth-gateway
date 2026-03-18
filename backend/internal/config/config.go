package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	GRPC      GRPCConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	OAuth     OAuthConfig
	SMTP      SMTPConfig
	SMS       SMSConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Security  SecurityConfig
	Metrics   MetricsConfig
	GeoIP     GeoIPConfig
	OIDC      OIDCConfig
	LDAP      LDAPConfig
	SAML      SAMLConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port           string
	Env            string
	LogLevel       string
	ExternalURL    string   // Base URL for Swagger docs (e.g., https://api.example.com)
	TrustedProxies []string // Trusted proxy IPs for X-Forwarded-For
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	Port                 string
	TLSEnabled           bool
	TLSCert              string // Path to TLS certificate file
	TLSKey               string // Path to TLS private key file
	ReflectionEnabled    bool   // Enable gRPC reflection (disable in production)
	MaxRequestsPerMinute int    // Rate limit: max requests per minute per API key
}

// Validate validates gRPC configuration
func (c *GRPCConfig) Validate() error {
	if c.TLSEnabled {
		if c.TLSCert == "" {
			return fmt.Errorf("GRPC_TLS_CERT_FILE is required when GRPC_TLS_ENABLED is true")
		}
		if c.TLSKey == "" {
			return fmt.Errorf("GRPC_TLS_KEY_FILE is required when GRPC_TLS_ENABLED is true")
		}
	}
	return nil
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	DBName         string
	SSLMode        string
	MaxOpenConns   int
	MaxIdleConns   int
	EnableQueryLog bool // Enable query logging (should be false in production)
}

// RedisConfig contains Redis-related configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig contains JWT-related configuration
type JWTConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessExpires  time.Duration
	RefreshExpires time.Duration
}

// Validate validates JWT configuration
// Returns error if secrets are too short or missing
func (c *JWTConfig) Validate() error {
	const minSecretLength = 32 // Minimum 32 characters for HS256

	if c.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if len(c.AccessSecret) < minSecretLength {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least %d characters long (current: %d). This is a security requirement for HS256 algorithm", minSecretLength, len(c.AccessSecret))
	}

	if c.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if len(c.RefreshSecret) < minSecretLength {
		return fmt.Errorf("JWT_REFRESH_SECRET must be at least %d characters long (current: %d). This is a security requirement for HS256 algorithm", minSecretLength, len(c.RefreshSecret))
	}

	// Warn if secrets are the same (security risk)
	if c.AccessSecret == c.RefreshSecret {
		log.Printf("WARNING: JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are the same. This is a security risk. Use different secrets.")
	}

	return nil
}

// OAuthConfig contains OAuth provider configurations
type OAuthConfig struct {
	Google           OAuthProvider
	Yandex           OAuthProvider
	GitHub           OAuthProvider
	Instagram        OAuthProvider
	OneC             CustomOAuthProvider
	FrontendURL      string
	TelegramBotToken string
}

// OAuthProvider represents a single OAuth provider configuration
type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

// CustomOAuthProvider represents a custom OAuth provider with configurable URLs
type CustomOAuthProvider struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	CallbackURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       string
}

// SMTPConfig contains SMTP email configuration
type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// SMSConfig contains SMS provider configuration
type SMSConfig struct {
	Provider string // "twilio", "aws_sns", "vonage", "mock"
	Enabled  bool

	// Twilio configuration
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// AWS SNS configuration
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSSenderID        string

	// Rate limiting for SMS
	SMSMaxPerHour   int
	SMSMaxPerDay    int
	SMSMaxPerNumber int // Max SMS per phone number per hour
}

// CORSConfig contains CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// Validate checks for insecure CORS configurations
func (c *CORSConfig) Validate() error {
	if c.AllowCredentials {
		for _, origin := range c.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("CORS: wildcard origin '*' with AllowCredentials is insecure (RFC 6454)")
			}
		}
	}
	return nil
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	SignupMax     int
	SignupWindow  time.Duration
	SigninMax     int
	SigninWindow  time.Duration
	RefreshMax    int           // Max refresh token requests per window
	RefreshWindow time.Duration // Time window for refresh token rate limiting
	APIMax        int
	APIWindow     time.Duration
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	BcryptCost                    int
	TokenBlacklistCleanupInterval time.Duration
	PasswordPolicy                PasswordPolicyConfig
	JITProvisioning               bool   // Enable Just-In-Time user provisioning for OAuth/OIDC logins
	EncryptionKey                 string
	StrictTokenBinding            bool   // Reject refresh if IP/UserAgent changed
	CSRFEnabled                   bool   // Enable Double Submit Cookie CSRF protection
	OTPHMACSecret                 string // HMAC secret for OTP code hashing (prevents brute-force on 6-digit codes)
	MaxActiveSessions             int    // Maximum active sessions per user (0 = unlimited)
}

// Validate checks security configuration for common misconfigurations
func (c *SecurityConfig) Validate(env string) error {
	if env == "production" && c.OTPHMACSecret == "change-me-in-production-otp-hmac-secret-32-chars-minimum" {
		return fmt.Errorf("OTP_HMAC_SECRET must be changed from default value in production")
	}
	if len(c.OTPHMACSecret) < 32 {
		return fmt.Errorf("OTP_HMAC_SECRET must be at least 32 characters long (current: %d)", len(c.OTPHMACSecret))
	}
	return nil
}

// PasswordPolicyConfig contains password policy configuration
type PasswordPolicyConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSpecial   bool
	MaxLength        int  // 0 means no maximum
	CommonPasswords  bool // Check against common passwords list
	CheckCompromised bool // Check passwords against HaveIBeenPwned API
}

type MetricsConfig struct {
	Enabled bool
	Port    string
}

type GeoIPConfig struct {
	APIKey  string
	Enabled bool
}

// LDAPConfig contains LDAP/Active Directory configuration
type LDAPConfig struct {
	Enabled bool
	// Default LDAP server settings (can be overridden per-config in database)
	DefaultServer       string
	DefaultPort         string
	DefaultBaseDN       string
	DefaultBindDN       string
	DefaultBindPassword string
	SyncInterval        time.Duration // Default sync interval
	AutoSyncEnabled     bool          // Enable automatic periodic sync
}

// SAMLConfig contains SAML 2.0 IdP configuration
type SAMLConfig struct {
	Enabled     bool
	Issuer      string // SAML Entity ID
	SSOURL      string // SSO endpoint URL
	SLOURL      string // Single Logout endpoint URL
	Certificate string // Path to certificate file for signing
	PrivateKey  string // Path to private key file for signing
	MetadataURL string // URL for SAML metadata
}

// OIDCConfig contains OIDC provider configuration
type OIDCConfig struct {
	// Issuer URL (e.g., "https://auth.example.com")
	// Used in ID tokens and discovery document
	Issuer string

	// Signing key configuration
	SigningKeyPath   string
	SigningKeyID     string
	SigningAlgorithm string // RS256 or ES256

	// Optional: Additional keys for rotation
	// Format: "kid1:/path/to/key1.pem,kid2:/path/to/key2.pem"
	AdditionalKeys string

	// Token TTL defaults (can be overridden per client)
	AccessTokenTTL  int // seconds, default 900 (15 min)
	RefreshTokenTTL int // seconds, default 604800 (7 days)
	IDTokenTTL      int // seconds, default 3600 (1 hour)
	AuthCodeTTL     int // seconds, default 600 (10 min)
	DeviceCodeTTL   int // seconds, default 1800 (30 min)

	// Device flow settings
	DeviceCodeInterval int // seconds, default 5

	// Enable/disable OIDC provider
	Enabled bool
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found will use environment variables instead.")
	}
	cfg := &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8181"),
			Env:            getEnv("ENV", "development"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			ExternalURL:    getEnv("EXTERNAL_URL", ""), // e.g., https://api.example.com
			TrustedProxies: getEnvAsSlice("TRUSTED_PROXIES", []string{}),
		},
		GRPC: GRPCConfig{
			Port:                 getEnv("GRPC_PORT", "50051"),
			TLSEnabled:           getEnvAsBool("GRPC_TLS_ENABLED", false),
			TLSCert:              getEnv("GRPC_TLS_CERT_FILE", ""),
			TLSKey:               getEnv("GRPC_TLS_KEY_FILE", ""),
			ReflectionEnabled:    getEnvAsBool("GRPC_REFLECTION_ENABLED", false),
			MaxRequestsPerMinute: getEnvAsInt("GRPC_MAX_REQUESTS_PER_MINUTE", 100),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", "postgres"),
			DBName:         getEnv("DB_NAME", "auth_gateway"),
			SSLMode:        getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:   getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			EnableQueryLog: getEnvAsBool("DB_ENABLE_QUERY_LOG", false),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:   getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:  getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpires:  getEnvAsDuration("JWT_ACCESS_EXPIRES", "15m"),
			RefreshExpires: getEnvAsDuration("JWT_REFRESH_EXPIRES", "168h"),
		},
		OAuth: OAuthConfig{
			Google: OAuthProvider{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("GOOGLE_CALLBACK_URL", ""),
			},
			Yandex: OAuthProvider{
				ClientID:     getEnv("YANDEX_CLIENT_ID", ""),
				ClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("YANDEX_CALLBACK_URL", ""),
			},
			GitHub: OAuthProvider{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("GITHUB_CALLBACK_URL", ""),
			},
			Instagram: OAuthProvider{
				ClientID:     getEnv("INSTAGRAM_CLIENT_ID", ""),
				ClientSecret: getEnv("INSTAGRAM_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("INSTAGRAM_CALLBACK_URL", ""),
			},
			OneC: CustomOAuthProvider{
				Enabled:      getEnvAsBool("OAUTH_ONEC_ENABLED", false),
				ClientID:     getEnv("OAUTH_ONEC_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_ONEC_CLIENT_SECRET", ""),
				CallbackURL:  getEnv("OAUTH_ONEC_REDIRECT_URI", ""),
				AuthURL:      getEnv("OAUTH_ONEC_AUTH_URL", ""),
				TokenURL:     getEnv("OAUTH_ONEC_TOKEN_URL", ""),
				UserInfoURL:  getEnv("OAUTH_ONEC_USERINFO_URL", ""),
				Scopes:       getEnv("OAUTH_ONEC_SCOPES", "openid profile email"),
			},
			FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3001"),
			TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		},
		SMTP: SMTPConfig{
			Host:      getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:      getEnvAsInt("SMTP_PORT", 587),
			Username:  getEnv("SMTP_USERNAME", ""),
			Password:  getEnv("SMTP_PASSWORD", ""),
			FromEmail: getEnv("SMTP_FROM_EMAIL", "noreply@authgateway.com"),
			FromName:  getEnv("SMTP_FROM_NAME", "Auth Gateway"),
		},
		SMS: SMSConfig{
			Provider:           getEnv("SMS_PROVIDER", "mock"),
			Enabled:            getEnvAsBool("SMS_ENABLED", false),
			TwilioAccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
			TwilioAuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
			TwilioFromNumber:   getEnv("TWILIO_FROM_NUMBER", ""),
			AWSRegion:          getEnv("AWS_SNS_REGION", "us-east-1"),
			AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			AWSSenderID:        getEnv("AWS_SNS_SENDER_ID", ""),
			SMSMaxPerHour:      getEnvAsInt("SMS_MAX_PER_HOUR", 10),
			SMSMaxPerDay:       getEnvAsInt("SMS_MAX_PER_DAY", 50),
			SMSMaxPerNumber:    getEnvAsInt("SMS_MAX_PER_NUMBER", 5),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3001"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With", "X-Application-ID", "X-API-Key"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		RateLimit: RateLimitConfig{
			SignupMax:     getEnvAsInt("RATE_LIMIT_SIGNUP_MAX", 5),
			SignupWindow:  getEnvAsDuration("RATE_LIMIT_SIGNUP_WINDOW", "1h"),
			SigninMax:     getEnvAsInt("RATE_LIMIT_SIGNIN_MAX", 10),
			SigninWindow:  getEnvAsDuration("RATE_LIMIT_SIGNIN_WINDOW", "15m"),
			RefreshMax:    getEnvAsInt("RATE_LIMIT_REFRESH_MAX", 30),
			RefreshWindow: getEnvAsDuration("RATE_LIMIT_REFRESH_WINDOW", "5m"),
			APIMax:        getEnvAsInt("RATE_LIMIT_API_MAX", 100),
			APIWindow:     getEnvAsDuration("RATE_LIMIT_API_WINDOW", "1m"),
		},
		Security: SecurityConfig{
			BcryptCost:                    getEnvAsInt("BCRYPT_COST", 12),
			TokenBlacklistCleanupInterval: getEnvAsDuration("TOKEN_BLACKLIST_CLEANUP_INTERVAL", "1h"),
			JITProvisioning:               getEnvAsBool("JIT_PROVISIONING_ENABLED", true), // Enabled by default
			EncryptionKey:                 getEnv("ENCRYPTION_KEY", ""),
			StrictTokenBinding:            getEnvAsBool("STRICT_TOKEN_BINDING", false),
			CSRFEnabled:                   getEnvAsBool("CSRF_ENABLED", false),
			OTPHMACSecret:                 getEnv("OTP_HMAC_SECRET", "change-me-in-production-otp-hmac-secret-32-chars-minimum"),
			MaxActiveSessions:             getEnvAsInt("MAX_ACTIVE_SESSIONS", 0),
			PasswordPolicy: PasswordPolicyConfig{
				MinLength:        getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
				RequireUppercase: getEnvAsBool("PASSWORD_REQUIRE_UPPERCASE", false),
				RequireLowercase: getEnvAsBool("PASSWORD_REQUIRE_LOWERCASE", true),
				RequireNumbers:   getEnvAsBool("PASSWORD_REQUIRE_NUMBERS", false),
				RequireSpecial:   getEnvAsBool("PASSWORD_REQUIRE_SPECIAL", false),
				MaxLength:        getEnvAsInt("PASSWORD_MAX_LENGTH", 0),
				CommonPasswords:  getEnvAsBool("PASSWORD_CHECK_COMMON", false),
				CheckCompromised: getEnvAsBool("PASSWORD_CHECK_COMPROMISED", false),
			},
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
		},
		GeoIP: GeoIPConfig{
			APIKey:  getEnv("GEOIP_API_KEY", ""),
			Enabled: getEnvAsBool("GEOIP_ENABLED", true),
		},
		OIDC: OIDCConfig{
			Issuer:             getEnv("OIDC_ISSUER", ""),
			SigningKeyPath:     getEnv("OIDC_SIGNING_KEY_PATH", ""),
			SigningKeyID:       getEnv("OIDC_SIGNING_KEY_ID", ""),
			SigningAlgorithm:   getEnv("OIDC_SIGNING_ALGORITHM", "RS256"),
			AdditionalKeys:     getEnv("OIDC_ADDITIONAL_KEYS", ""),
			AccessTokenTTL:     getEnvAsInt("OIDC_ACCESS_TOKEN_TTL", 900),
			RefreshTokenTTL:    getEnvAsInt("OIDC_REFRESH_TOKEN_TTL", 604800),
			IDTokenTTL:         getEnvAsInt("OIDC_ID_TOKEN_TTL", 3600),
			AuthCodeTTL:        getEnvAsInt("OIDC_AUTH_CODE_TTL", 600),
			DeviceCodeTTL:      getEnvAsInt("OIDC_DEVICE_CODE_TTL", 1800),
			DeviceCodeInterval: getEnvAsInt("OIDC_DEVICE_CODE_INTERVAL", 5),
			Enabled:            getEnvAsBool("OIDC_ENABLED", false),
		},
	}

	setOIDCDefaults(cfg)

	// Validate JWT configuration (checks for presence and minimum length)
	if err := cfg.JWT.Validate(); err != nil {
		return nil, fmt.Errorf("JWT configuration validation failed: %w", err)
	}

	// Validate gRPC configuration
	if err := cfg.GRPC.Validate(); err != nil {
		return nil, fmt.Errorf("gRPC configuration validation failed: %w", err)
	}

	// Validate CORS configuration
	if err := cfg.CORS.Validate(); err != nil {
		return nil, fmt.Errorf("CORS configuration validation failed: %w", err)
	}

	// Validate security configuration
	if err := cfg.Security.Validate(cfg.Server.Env); err != nil {
		return nil, fmt.Errorf("security configuration validation failed: %w", err)
	}

	return cfg, nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// setOIDCDefaults sets default values for OIDC configuration
func setOIDCDefaults(cfg *Config) {
	if cfg.OIDC.AccessTokenTTL == 0 {
		cfg.OIDC.AccessTokenTTL = 900 // 15 minutes
	}
	if cfg.OIDC.RefreshTokenTTL == 0 {
		cfg.OIDC.RefreshTokenTTL = 604800 // 7 days
	}
	if cfg.OIDC.IDTokenTTL == 0 {
		cfg.OIDC.IDTokenTTL = 3600 // 1 hour
	}
	if cfg.OIDC.AuthCodeTTL == 0 {
		cfg.OIDC.AuthCodeTTL = 600 // 10 minutes
	}
	if cfg.OIDC.DeviceCodeTTL == 0 {
		cfg.OIDC.DeviceCodeTTL = 1800 // 30 minutes
	}
	if cfg.OIDC.DeviceCodeInterval == 0 {
		cfg.OIDC.DeviceCodeInterval = 5 // 5 seconds
	}
	if cfg.OIDC.SigningAlgorithm == "" {
		cfg.OIDC.SigningAlgorithm = "RS256"
	}
}

// GetKeyConfigs parses key configuration and returns slice of key configs
// Returns format: []struct{KID, KeyPath string}
func (c *OIDCConfig) GetKeyConfigs() []struct {
	KID     string
	KeyPath string
} {
	var configs []struct {
		KID     string
		KeyPath string
	}

	if c.SigningKeyPath != "" {
		configs = append(configs, struct {
			KID     string
			KeyPath string
		}{
			KID:     c.SigningKeyID,
			KeyPath: c.SigningKeyPath,
		})
	}

	if c.AdditionalKeys != "" {
		entries := splitAndTrim(c.AdditionalKeys, ",")
		for _, entry := range entries {
			parts := splitAndTrim(entry, ":")
			if len(parts) == 2 {
				configs = append(configs, struct {
					KID     string
					KeyPath string
				}{
					KID:     parts[0],
					KeyPath: parts[1],
				})
			}
		}
	}

	return configs
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return splitAndTrim(value, ",")
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		if trimmed := trimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	return splitByRune(s, rune(sep[0]))
}

func splitByRune(s string, sep rune) []string {
	var result []string
	var current string
	for _, char := range s {
		if char == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && isSpace(s[start]) {
		start++
	}

	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
