// @title Auth Gateway API
// @version 1.0
// @description Centralized authentication and authorization system for microservices

// @contact.name API Support
// @contact.email maksbalashov@gmail.com

// @host localhost:8811
// @BasePath /
// @schemes http https

// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// SecurityDefinitions:
//
//	BearerAuth:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: JWT Bearer token. Format: "Bearer {token}"
//	ApiKeyAuth:
//	  type: apiKey
//	  name: X-API-Key
//	  in: header
//	  description: API Key for service-to-service authentication. Format: "agw_{key}"
//

package main

import (
	"context"
	"fmt"
	"html/template"
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
	"github.com/smilemakc/auth-gateway/pkg/keys"
	"github.com/smilemakc/auth-gateway/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/smilemakc/auth-gateway/docs"
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

	// Initialize signing key manager for OIDC
	var keyManager *keys.Manager
	if cfg.OIDC.Enabled {
		keyConfigs := buildKeyConfigs(&cfg.OIDC)
		if len(keyConfigs) > 0 {
			keyManager, err = keys.NewManager(keyConfigs, cfg.OIDC.SigningKeyID)
			if err != nil {
				log.Error("Failed to initialize key manager", map[string]interface{}{
					"error": err.Error(),
				})
				os.Exit(1)
			}
			log.Info("OIDC key manager initialized", map[string]interface{}{
				"keys": len(keyConfigs),
			})
		}
	}

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

	// Initialize OIDC JWT service
	var oidcJWTService *jwt.OIDCService
	if keyManager != nil {
		oidcJWTService = jwt.NewOIDCService(keyManager, cfg.OIDC.Issuer)
	}

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
	webhookRepo := repository.NewWebhookRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	brandingRepo := repository.NewBrandingRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	geoRepo := repository.NewGeoRepository(db)
	oauthProviderRepo := repository.NewOAuthProviderRepository(db)

	// Initialize geo and audit services
	var geoService *service.GeoService
	if cfg.GeoIP.Enabled {
		geoService = service.NewGeoService(cfg.GeoIP.APIKey)
		log.Info("GeoIP service enabled")
	}
	auditService := service.NewAuditService(auditRepo, geoService)

	// Initialize session creation service (universal session management)
	sessionCreationService := service.NewSessionCreationService(sessionRepo, log)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, rbacRepo, auditService, jwtService, redis, sessionCreationService, cfg.Security.BcryptCost)
	userService := service.NewUserService(userRepo, auditService)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo, auditService)
	emailService := service.NewEmailService(&cfg.SMTP)
	otpService := service.NewOTPService(otpRepo, userRepo, emailService, auditService)
	oauthService := service.NewOAuthService(userRepo, oauthRepo, tokenRepo, auditRepo, rbacRepo, jwtService, sessionCreationService, &http.Client{Timeout: 10 * time.Second})
	twoFAService := service.NewTwoFactorService(userRepo, backupCodeRepo, "Auth Gateway")
	adminService := service.NewAdminService(userRepo, apiKeyRepo, auditRepo, oauthRepo, rbacRepo, cfg.Security.BcryptCost)

	// Advanced feature services
	rbacService := service.NewRBACService(rbacRepo, auditService)
	sessionService := service.NewSessionService(sessionRepo)
	ipFilterService := service.NewIPFilterService(ipFilterRepo)
	webhookService := service.NewWebhookService(webhookRepo, auditService)
	templateService := service.NewTemplateService(templateRepo, auditService)

	// Initialize OAuth provider service
	var oauthProviderService *service.OAuthProviderService
	if cfg.OIDC.Enabled && oidcJWTService != nil {
		baseURL := cfg.OIDC.Issuer
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
		}
		oauthProviderService = service.NewOAuthProviderService(
			oauthProviderRepo,
			userRepo,
			auditRepo,
			sessionCreationService,
			oidcJWTService,
			keyManager,
			log,
			cfg.OIDC.Issuer,
			baseURL,
		)
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService, otpService, log)
	healthHandler := handler.NewHealthHandler(db, redis)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, log)
	otpHandler := handler.NewOTPHandler(otpService, authService, log)
	oauthHandler := handler.NewOAuthHandler(oauthService, log)
	twoFAHandler := handler.NewTwoFactorHandler(twoFAService, userService, log)
	adminHandler := handler.NewAdminHandler(adminService, log)
	advancedAdminHandler := handler.NewAdvancedAdminHandler(rbacService, sessionService, ipFilterService, brandingRepo, systemRepo, geoRepo)
	webhookHandler := handler.NewWebhookHandler(webhookService, log)
	templateHandler := handler.NewTemplateHandler(templateService, log)

	// Initialize OAuth provider handlers
	var oauthProviderHandler *handler.OAuthProviderHandler
	var oauthAdminHandler *handler.OAuthAdminHandler
	if oauthProviderService != nil {
		oauthProviderHandler = handler.NewOAuthProviderHandler(oauthProviderService, log)
		oauthAdminHandler = handler.NewOAuthAdminHandler(oauthProviderService, log)
	} else {
		// Create minimal OAuth admin service for client management even when OIDC is disabled
		minimalOAuthProviderService := service.NewOAuthProviderServiceMinimal(oauthProviderRepo, auditRepo, log)
		oauthAdminHandler = handler.NewOAuthAdminHandler(minimalOAuthProviderService, log)
	}

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

	// Load OAuth templates
	if cfg.OIDC.Enabled {
		tmpl, err := template.ParseGlob("internal/templates/*.html")
		if err != nil {
			log.Warn("Failed to load OAuth templates", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			router.SetHTMLTemplate(tmpl)
			log.Info("OAuth templates loaded successfully")
		}
	}

	// Global middleware
	router.Use(middleware.Recovery(log))
	router.Use(middleware.Logger(log))
	router.Use(middleware.SetupCORS(&cfg.CORS))
	router.Use(maintenanceMiddleware.CheckMaintenance())
	router.Use(ipFilterMiddleware.CheckIPFilter())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)
	router.GET("/live", healthHandler.Liveness)

	// Public system status endpoint
	router.GET("/system/maintenance", advancedAdminHandler.GetMaintenanceMode)

	// OIDC Discovery endpoints (public, no auth required)
	if oauthProviderHandler != nil {
		router.GET("/.well-known/openid-configuration", oauthProviderHandler.Discovery)
		router.GET("/.well-known/jwks.json", oauthProviderHandler.JWKS)

		// OAuth Provider endpoints
		oauth := router.Group("/oauth")
		{
			// Authorization endpoint (user-facing)
			oauth.GET("/authorize", oauthProviderHandler.Authorize)

			// Token endpoint (public, client auth in body/header)
			oauth.POST("/token", oauthProviderHandler.Token)

			// Token introspection (client auth required)
			oauth.POST("/introspect", oauthProviderHandler.Introspect)

			// Token revocation (client auth required)
			oauth.POST("/revoke", oauthProviderHandler.Revoke)

			// UserInfo endpoint (bearer token required)
			oauth.GET("/userinfo", oauthProviderHandler.UserInfo)

			// Device flow endpoints
			oauth.POST("/device/code", oauthProviderHandler.DeviceCode)
			oauth.POST("/device/token", oauthProviderHandler.DeviceToken)
			oauth.GET("/device", oauthProviderHandler.DeviceVerification)

			// Device approval requires authentication
			oauth.POST("/device/approve", authMiddleware.Authenticate(), oauthProviderHandler.DeviceApprove)

			// Consent endpoints require authentication
			oauth.GET("/consent", authMiddleware.Authenticate(), oauthProviderHandler.ConsentPage)
			oauth.POST("/consent", authMiddleware.Authenticate(), oauthProviderHandler.ConsentSubmit)
		}
	}

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

		// Passwordless registration endpoints (two-step signup without password)
		signupPhoneGroup := apiGroup.Group("/auth/signup/phone")
		{
			signupPhoneGroup.POST("", rateLimitMiddleware.LimitSignup(), authHandler.InitPasswordlessRegistration)
			signupPhoneGroup.POST("/verify", authHandler.CompletePasswordlessRegistration)
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
				rbacGroup.DELETE("/permissions/:id", advancedAdminHandler.DeletePermission)

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

			// Webhooks Management
			webhooksGroup := adminGroup.Group("/webhooks")
			{
				webhooksGroup.GET("", webhookHandler.ListWebhooks)
				webhooksGroup.POST("", webhookHandler.CreateWebhook)
				webhooksGroup.GET("/events", webhookHandler.GetAvailableEvents)
				webhooksGroup.GET("/:id", webhookHandler.GetWebhook)
				webhooksGroup.PUT("/:id", webhookHandler.UpdateWebhook)
				webhooksGroup.DELETE("/:id", webhookHandler.DeleteWebhook)
				webhooksGroup.POST("/:id/test", webhookHandler.TestWebhook)
				webhooksGroup.GET("/:id/deliveries", webhookHandler.ListWebhookDeliveries)
			}

			// Email Templates Management
			templatesGroup := adminGroup.Group("/templates")
			{
				templatesGroup.GET("", templateHandler.ListEmailTemplates)
				templatesGroup.POST("", templateHandler.CreateEmailTemplate)
				templatesGroup.GET("/types", templateHandler.GetAvailableTemplateTypes)
				templatesGroup.POST("/preview", templateHandler.PreviewEmailTemplate)
				templatesGroup.GET("/variables/:type", templateHandler.GetDefaultVariables)
				templatesGroup.GET("/:id", templateHandler.GetEmailTemplate)
				templatesGroup.PUT("/:id", templateHandler.UpdateEmailTemplate)
				templatesGroup.DELETE("/:id", templateHandler.DeleteEmailTemplate)
			}

			// OAuth Provider Management (always available for client management)
			adminOAuth := adminGroup.Group("/oauth")
			{
				// Client management
				adminOAuth.POST("/clients", oauthAdminHandler.CreateClient)
				adminOAuth.GET("/clients", oauthAdminHandler.ListClients)
				adminOAuth.GET("/clients/:id", oauthAdminHandler.GetClient)
				adminOAuth.PUT("/clients/:id", oauthAdminHandler.UpdateClient)
				adminOAuth.DELETE("/clients/:id", oauthAdminHandler.DeleteClient)
				adminOAuth.POST("/clients/:id/rotate-secret", oauthAdminHandler.RotateSecret)

				// Scope management
				adminOAuth.GET("/scopes", oauthAdminHandler.ListScopes)
				adminOAuth.POST("/scopes", oauthAdminHandler.CreateScope)
				adminOAuth.DELETE("/scopes/:id", oauthAdminHandler.DeleteScope)

				// Consent management
				adminOAuth.GET("/clients/:id/consents", oauthAdminHandler.ListClientConsents)
				adminOAuth.DELETE("/clients/:id/consents/:user_id", oauthAdminHandler.RevokeUserConsent)
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

	// Log OIDC provider status
	if cfg.OIDC.Enabled {
		log.Info("OIDC Provider enabled", map[string]interface{}{
			"issuer":    cfg.OIDC.Issuer,
			"algorithm": cfg.OIDC.SigningAlgorithm,
		})
	}

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
		authService,
		oauthProviderService,
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

// buildKeyConfigs converts OIDC config to key manager format
func buildKeyConfigs(oidcCfg *config.OIDCConfig) []keys.KeyConfig {
	var keyConfigs []keys.KeyConfig

	configs := oidcCfg.GetKeyConfigs()
	algorithm := keys.Algorithm(oidcCfg.SigningAlgorithm)

	for _, cfg := range configs {
		keyConfigs = append(keyConfigs, keys.KeyConfig{
			ID:             cfg.KID,
			Algorithm:      algorithm,
			PrivateKeyPath: cfg.KeyPath,
		})
	}

	return keyConfigs
}
