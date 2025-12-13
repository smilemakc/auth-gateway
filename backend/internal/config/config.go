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
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port     string
	GRPCPort string
	Env      string
	LogLevel string
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
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

// OAuthConfig contains OAuth provider configurations
type OAuthConfig struct {
	Google      OAuthProvider
	Yandex      OAuthProvider
	GitHub      OAuthProvider
	Instagram   OAuthProvider
	FrontendURL string
}

// OAuthProvider represents a single OAuth provider configuration
type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
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

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	SignupMax    int
	SignupWindow time.Duration
	SigninMax    int
	SigninWindow time.Duration
	APIMax       int
	APIWindow    time.Duration
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	BcryptCost                    int
	TokenBlacklistCleanupInterval time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Port    string
}

type GeoIPConfig struct {
	APIKey  string
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
			Port:     getEnv("PORT", "3000"),
			GRPCPort: getEnv("GRPC_PORT", "50051"),
			Env:      getEnv("ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			DBName:       getEnv("DB_NAME", "auth_gateway"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
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
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3001"),
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
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		RateLimit: RateLimitConfig{
			SignupMax:    getEnvAsInt("RATE_LIMIT_SIGNUP_MAX", 5),
			SignupWindow: getEnvAsDuration("RATE_LIMIT_SIGNUP_WINDOW", "1h"),
			SigninMax:    getEnvAsInt("RATE_LIMIT_SIGNIN_MAX", 10),
			SigninWindow: getEnvAsDuration("RATE_LIMIT_SIGNIN_WINDOW", "15m"),
			APIMax:       getEnvAsInt("RATE_LIMIT_API_MAX", 100),
			APIWindow:    getEnvAsDuration("RATE_LIMIT_API_WINDOW", "1m"),
		},
		Security: SecurityConfig{
			BcryptCost:                    getEnvAsInt("BCRYPT_COST", 10),
			TokenBlacklistCleanupInterval: getEnvAsDuration("TOKEN_BLACKLIST_CLEANUP_INTERVAL", "1h"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
		},
		GeoIP: GeoIPConfig{
			APIKey:  getEnv("GEOIP_API_KEY", ""),
			Enabled: getEnvAsBool("GEOIP_ENABLED", true),
		},
	}

	// Validate critical configuration
	if cfg.JWT.AccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
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
