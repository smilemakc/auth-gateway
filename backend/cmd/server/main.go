package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/config"
	grpcserver "github.com/smilemakc/auth-gateway/internal/grpc"
	"github.com/smilemakc/auth-gateway/internal/handler"
	"github.com/smilemakc/auth-gateway/internal/middleware"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New("auth-gateway", logger.LogLevel(cfg.Server.LogLevel), true)
	logger.SetDefault(log)

	log.Info("Starting Auth Gateway", map[string]interface{}{
		"env":  cfg.Server.Env,
		"port": cfg.Server.Port,
	})

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer db.Close()
	log.Info("Database connected successfully")

	// Initialize Redis
	redis, err := service.NewRedisService(&cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer redis.Close()
	log.Info("Redis connected successfully")

	// Initialize JWT service
	jwtService := jwt.NewService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpires,
		cfg.JWT.RefreshExpires,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	oauthRepo := repository.NewOAuthRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	backupCodeRepo := repository.NewBackupCodeRepository(db)

	// Advanced feature repositories
	rbacRepo := repository.NewRBACRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ipFilterRepo := repository.NewIPFilterRepository(db)
	// webhookRepo := repository.NewWebhookRepository(db)
	// templateRepo := repository.NewTemplateRepository(db)
	brandingRepo := repository.NewBrandingRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	geoRepo := repository.NewGeoRepository(db)

	// Initialize geo and audit services
	var geoService *service.GeoService
	if cfg.GeoIP.Enabled {
		geoService = service.NewGeoService(cfg.GeoIP.APIKey)
		log.Info("GeoIP service enabled")
	}
	auditService := service.NewAuditService(auditRepo, geoService)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, rbacRepo, auditService, jwtService, redis, cfg.Security.BcryptCost)
	userService := service.NewUserService(userRepo, auditService)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo, auditService)
	emailService := service.NewEmailService(&cfg.SMTP)
	otpService := service.NewOTPService(otpRepo, userRepo, emailService, auditService)
	oauthService := service.NewOAuthService(userRepo, oauthRepo, tokenRepo, auditRepo, rbacRepo, jwtService, &http.Client{Timeout: 10 * time.Second})
	twoFAService := service.NewTwoFactorService(userRepo, backupCodeRepo, "Auth Gateway")
	adminService := service.NewAdminService(userRepo, apiKeyRepo, auditRepo, oauthRepo, rbacRepo, cfg.Security.BcryptCost)

	// Advanced feature services
	rbacService := service.NewRBACService(rbacRepo, auditService)
	sessionService := service.NewSessionService(sessionRepo)
	ipFilterService := service.NewIPFilterService(ipFilterRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService, otpService, log)
	healthHandler := handler.NewHealthHandler(db, redis)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, log)
	otpHandler := handler.NewOTPHandler(otpService, authService, jwtService, log)
	oauthHandler := handler.NewOAuthHandler(oauthService, log)
	twoFAHandler := handler.NewTwoFactorHandler(twoFAService, userService, log)
	adminHandler := handler.NewAdminHandler(adminService, log)
	advancedAdminHandler := handler.NewAdvancedAdminHandler(rbacService, sessionService, ipFilterService, brandingRepo, systemRepo, geoRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, redis, tokenRepo)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(apiKeyService, rbacRepo)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redis, &cfg.RateLimit)

	// Advanced feature middleware
	// rbacMiddleware := middleware.NewRBACMiddleware(rbacService)
	ipFilterMiddleware := middleware.NewIPFilterMiddleware(ipFilterService)
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(systemRepo)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery(log))
	router.Use(middleware.Logger(log))
	router.Use(middleware.SetupCORS(&cfg.CORS))
	router.Use(maintenanceMiddleware.CheckMaintenance())
	router.Use(ipFilterMiddleware.CheckIPFilter())

	// Health check endpoints (no auth required)
	router.GET("/auth/health", healthHandler.Health)
	router.GET("/auth/ready", healthHandler.Readiness)
	router.GET("/auth/live", healthHandler.Liveness)

	// Public system status endpoint
	router.GET("/system/maintenance", advancedAdminHandler.GetMaintenanceMode)

	// API Group
	apiGroup := router.Group("/api")
	{
		// Public auth endpoints
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/signup", rateLimitMiddleware.LimitSignup(), authHandler.SignUp)
			authGroup.POST("/signin", rateLimitMiddleware.LimitSignin(), authHandler.SignIn)
			authGroup.POST("/refresh", authHandler.RefreshToken)

			// Email verification
			authGroup.POST("/verify/resend", otpHandler.SendOTP)
			authGroup.POST("/verify/email", authHandler.VerifyEmail)

			// Password reset
			authGroup.POST("/password/reset/request", authHandler.RequestPasswordReset)
			authGroup.POST("/password/reset/complete", authHandler.ResetPassword)

			// 2FA login verification (public - no auth required)
			authGroup.POST("/2fa/login/verify", authHandler.Verify2FA)
		}

		// OTP endpoints
		otpGroup := apiGroup.Group("/otp")
		{
			otpGroup.POST("/send", otpHandler.SendOTP)
			otpGroup.POST("/verify", otpHandler.VerifyOTP)
		}

		// Passwordless login endpoints
		passwordlessGroup := apiGroup.Group("/auth/passwordless")
		{
			passwordlessGroup.POST("/request", otpHandler.RequestPasswordlessLogin)
			passwordlessGroup.POST("/verify", otpHandler.VerifyPasswordlessLogin)
		}

		// OAuth endpoints
		oauthGroup := apiGroup.Group("/auth")
		{
			oauthGroup.GET("/providers", oauthHandler.GetProviders)
			oauthGroup.GET("/:provider", oauthHandler.Login)
			oauthGroup.GET("/:provider/callback", oauthHandler.Callback)
			oauthGroup.POST("/telegram/callback", oauthHandler.TelegramCallback)
		}

		// Protected auth endpoints (require authentication)
		protectedAuth := apiGroup.Group("/auth")
		protectedAuth.Use(authMiddleware.Authenticate())
		{
			protectedAuth.POST("/logout", authHandler.Logout)
			protectedAuth.GET("/profile", authHandler.GetProfile)
			protectedAuth.PUT("/profile", authHandler.UpdateProfile)
			protectedAuth.POST("/change-password", authHandler.ChangePassword)

			// 2FA management (protected)
			protectedAuth.POST("/2fa/setup", twoFAHandler.Setup)
			protectedAuth.POST("/2fa/verify", twoFAHandler.Verify)
			protectedAuth.POST("/2fa/disable", twoFAHandler.Disable)
			protectedAuth.GET("/2fa/status", twoFAHandler.GetStatus)
			protectedAuth.POST("/2fa/backup-codes/regenerate", twoFAHandler.RegenerateBackupCodes)
		}

		// API Keys endpoints (require JWT authentication)
		apiKeysGroup := apiGroup.Group("/api-keys")
		apiKeysGroup.Use(authMiddleware.Authenticate())
		{
			apiKeysGroup.POST("", apiKeyHandler.Create)
			apiKeysGroup.GET("", apiKeyHandler.List)
			apiKeysGroup.GET("/:id", apiKeyHandler.Get)
			apiKeysGroup.PUT("/:id", apiKeyHandler.Update)
			apiKeysGroup.POST("/:id/revoke", apiKeyHandler.Revoke)
			apiKeysGroup.DELETE("/:id", apiKeyHandler.Delete)
		}

		// Admin endpoints (require admin role)
		adminGroup := apiGroup.Group("/admin")
		adminGroup.Use(authMiddleware.Authenticate())
		adminGroup.Use(middleware.RequireAdmin())
		{
			// Statistics
			adminGroup.GET("/stats", adminHandler.GetStats)

			// User management
			adminGroup.GET("/users", adminHandler.ListUsers)
			adminGroup.POST("/users", adminHandler.CreateUser) // Added CreateUser if it exists in handler, checking imports later
			adminGroup.GET("/users/:id", adminHandler.GetUser)
			adminGroup.PUT("/users/:id", adminHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

			// User role management
			adminGroup.POST("/users/:id/roles", adminHandler.AssignRole)
			adminGroup.DELETE("/users/:id/roles/:roleId", adminHandler.RemoveRole)

			// API key management
			adminGroup.GET("/api-keys", adminHandler.ListAPIKeys)
			adminGroup.POST("/api-keys/:id/revoke", adminHandler.RevokeAPIKey)

			// Audit logs
			adminGroup.GET("/audit-logs", adminHandler.ListAuditLogs)

			// RBAC Management
			rbacGroup := adminGroup.Group("/rbac")
			{
				rbacGroup.GET("/permissions", advancedAdminHandler.ListPermissions)
				rbacGroup.POST("/permissions", advancedAdminHandler.CreatePermission)

				rbacGroup.GET("/roles", advancedAdminHandler.ListRoles)
				rbacGroup.POST("/roles", advancedAdminHandler.CreateRole)
				rbacGroup.GET("/roles/:id", advancedAdminHandler.GetRole)
				rbacGroup.PUT("/roles/:id", advancedAdminHandler.UpdateRole)
				rbacGroup.DELETE("/roles/:id", advancedAdminHandler.DeleteRole)

				rbacGroup.GET("/permission-matrix", advancedAdminHandler.GetPermissionMatrix)
			}

			// Session Management (Admin view)
			adminGroup.GET("/sessions", advancedAdminHandler.ListAllSessions)
			adminGroup.GET("/sessions/stats", advancedAdminHandler.GetSessionStats)

			// IP Filter Management
			ipFilterGroup := adminGroup.Group("/ip-filters")
			{
				ipFilterGroup.GET("", advancedAdminHandler.ListIPFilters)
				ipFilterGroup.POST("", advancedAdminHandler.CreateIPFilter)
				ipFilterGroup.DELETE("/:id", advancedAdminHandler.DeleteIPFilter)
			}

			// Branding Management
			adminGroup.PUT("/branding", advancedAdminHandler.UpdateBranding)

			// System Settings
			systemGroup := adminGroup.Group("/system")
			{
				systemGroup.PUT("/maintenance", advancedAdminHandler.SetMaintenanceMode)
				systemGroup.GET("/health", advancedAdminHandler.GetSystemHealth)
			}

			// Analytics
			analyticsGroup := adminGroup.Group("/analytics")
			{
				analyticsGroup.GET("/geo-distribution", advancedAdminHandler.GetGeoDistribution)
			}
		}

		// Session Management (User endpoints)
		sessionsGroup := apiGroup.Group("/sessions")
		sessionsGroup.Use(authMiddleware.Authenticate())
		{
			sessionsGroup.GET("", advancedAdminHandler.ListUserSessions)
			sessionsGroup.DELETE("/:id", advancedAdminHandler.RevokeSession)
			sessionsGroup.POST("/revoke-all", advancedAdminHandler.RevokeAllSessions)
		}

		// Example: Protected endpoint that accepts both JWT and API key
		protectedAPI := apiGroup.Group("/v1")
		// This middleware will try JWT first, then API key
		protectedAPI.Use(func(c *gin.Context) {
			// Try JWT authentication first
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && !strings.HasPrefix(authHeader, "Bearer agw_") {
				authMiddleware.Authenticate()(c)
				return
			}

			// Try API key authentication
			apiKeyMiddleware.Authenticate()(c)
		})
		{
			// Example endpoint that can be accessed with either JWT or API key
			protectedAPI.GET("/profile", authHandler.GetProfile)
		}
	}

	// Start token cleanup routine
	startTokenCleanup(tokenRepo, cfg.Security.TokenBlacklistCleanupInterval, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Create gRPC server
	grpcSrv, err := grpcserver.NewServer(
		cfg.Server.GRPCPort,
		jwtService,
		userRepo,
		tokenRepo,
		rbacRepo,
		apiKeyService,
		redis,
		log,
	)
	if err != nil {
		log.Fatal("Failed to create gRPC server", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Info("HTTP server starting", map[string]interface{}{
			"port": cfg.Server.Port,
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcSrv.Start(); err != nil {
			log.Fatal("Failed to start gRPC server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("HTTP server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Shutdown gRPC server
	grpcSrv.Stop()

	log.Info("Servers exited successfully")
}

// startTokenCleanup starts a background routine to clean up expired tokens
func startTokenCleanup(tokenRepo *repository.TokenRepository, interval time.Duration, log *logger.Logger) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Debug("Running token cleanup")

			if err := tokenRepo.CleanupExpiredTokens(context.Background()); err != nil {
				log.Error("Token cleanup failed", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				log.Debug("Token cleanup completed successfully")
			}
		}
	}()
}
