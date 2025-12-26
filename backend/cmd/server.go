// @title Auth Gateway API
// @version 1.0
// @description Centralized authentication and authorization system for microservices ecosystem.
// @description
// @description ## Overview
// @description Auth Gateway provides a complete authentication and authorization solution including:
// @description - User registration and authentication (email/password, OAuth, passwordless)
// @description - Two-factor authentication (TOTP)
// @description - API key management for service-to-service communication
// @description - Role-based access control (RBAC)
// @description - Session management
// @description - Webhooks for event notifications
// @description
// @description ## Authentication
// @description Most endpoints require authentication via JWT Bearer token or API key.
// @description
// @description ### JWT Authentication
// @description Include the access token in the Authorization header: `Authorization: Bearer {token}`
// @description
// @description ### API Key Authentication
// @description Include the API key in the X-API-Key header: `X-API-Key: agw_{key}`

// @contact.name API Support
// @contact.email maksbalashov@gmail.com

// @host localhost:8811
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token for user authentication. Format: "Bearer {access_token}"

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key for service-to-service authentication. Format: "agw_{key}"

package cmd

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
	_ "github.com/smilemakc/auth-gateway/docs"
	"github.com/smilemakc/auth-gateway/internal/config"
	grpcserver "github.com/smilemakc/auth-gateway/internal/grpc"
	"github.com/smilemakc/auth-gateway/internal/handler"
	"github.com/smilemakc/auth-gateway/internal/jobs"
	"github.com/smilemakc/auth-gateway/internal/metrics"
	"github.com/smilemakc/auth-gateway/internal/middleware"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/sms"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/keys"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type infra struct {
	cfg            *config.Config
	log            *logger.Logger
	keyManager     *keys.Manager
	db             *repository.Database
	redis          *service.RedisService
	jwtService     *jwt.Service
	oidcJWTService *jwt.OIDCService
	smsProvider    sms.SMSProvider
}

type repoSet struct {
	User          *repository.UserRepository
	Token         *repository.TokenRepository
	OAuth         *repository.OAuthRepository
	Audit         *repository.AuditRepository
	APIKey        *repository.APIKeyRepository
	OTP           *repository.OTPRepository
	BackupCode    *repository.BackupCodeRepository
	SMSLog        *repository.SMSLogRepository
	RBAC          *repository.RBACRepository
	Session       *repository.SessionRepository
	IPFilter      *repository.IPFilterRepository
	Webhook       *repository.WebhookRepository
	Template      *repository.TemplateRepository
	Branding      *repository.BrandingRepository
	System        *repository.SystemRepository
	Geo           *repository.GeoRepository
	OAuthProvider *repository.OAuthProviderRepository
	Group         *repository.GroupRepository
	LDAP          *repository.LDAPRepository
	SAML          *repository.SAMLRepository
}

type serviceSet struct {
	Geo             *service.GeoService
	Audit           *service.AuditService
	Blacklist       *service.BlacklistService
	Session         *service.SessionService
	Auth            *service.AuthService
	User            *service.UserService
	APIKey          *service.APIKeyService
	Email           *service.EmailService
	OTP             *service.OTPService
	OAuth           *service.OAuthService
	TwoFA           *service.TwoFactorService
	Admin           *service.AdminService
	RBAC            *service.RBACService
	IPFilter        *service.IPFilterService
	Webhook         *service.WebhookService
	Template        *service.TemplateService
	OAuthProvider   *service.OAuthProviderService
	MinimalOAuthSvc *service.OAuthProviderService
	Group           *service.GroupService
	LDAP            *service.LDAPService
	Bulk            *service.BulkService
	SCIM            *service.SCIMService
	SAML            *service.SAMLService
}

type handlerSet struct {
	Auth          *handler.AuthHandler
	Health        *handler.HealthHandler
	APIKey        *handler.APIKeyHandler
	OTP           *handler.OTPHandler
	OAuth         *handler.OAuthHandler
	TwoFA         *handler.TwoFactorHandler
	Admin         *handler.AdminHandler
	AdvancedAdmin *handler.AdvancedAdminHandler
	Webhook       *handler.WebhookHandler
	Template      *handler.TemplateHandler
	OAuthProvider *handler.OAuthProviderHandler
	OAuthAdmin    *handler.OAuthAdminHandler
	Login         *handler.LoginHandler
	Group         *handler.GroupHandler
	SCIM          *handler.SCIMHandler
	LDAP          *handler.LDAPHandler
	Bulk          *handler.BulkHandler
	SAML          *handler.SAMLHandler
}

type middlewareSet struct {
	Auth        *middleware.AuthMiddleware
	APIKey      *middleware.APIKeyMiddleware
	RateLimit   *middleware.RateLimitMiddleware
	IPFilter    *middleware.IPFilterMiddleware
	Maintenance *middleware.MaintenanceMiddleware
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runServer() {
	deps, cleanup, err := buildInfra()
	if err != nil {
		fmt.Printf("Failed to initialize infra: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	repos := buildRepositories(deps)
	services := buildServices(deps, repos)
	handlers := buildHandlers(deps, repos, services)
	middlewares := buildMiddlewares(deps, repos, services)
	router := buildRouter(deps, services, handlers, middlewares)

	startTokenCleanup(repos.Token, deps.cfg.Security.TokenBlacklistCleanupInterval, deps.log)

	// Start LDAP sync job if LDAP service is available
	var ldapSyncJob *jobs.LDAPSyncJob
	if services.LDAP != nil {
		ldapSyncJob = jobs.NewLDAPSyncJob(services.LDAP, deps.log)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			ldapSyncJob.Start(ctx)
			cancel()
		}()
		deps.log.Info("LDAP sync job started")
	}

	if deps.cfg.OIDC.Enabled {
		deps.log.Info("OIDC Provider enabled", map[string]interface{}{
			"issuer":    deps.cfg.OIDC.Issuer,
			"algorithm": deps.cfg.OIDC.SigningAlgorithm,
		})
	}

	srv := &http.Server{
		Addr:    ":" + deps.cfg.Server.Port,
		Handler: router,
	}

	grpcSrv, err := grpcserver.NewServer(
		deps.cfg.Server.GRPCPort,
		deps.jwtService,
		repos.User,
		repos.Token,
		repos.RBAC,
		services.APIKey,
		services.Auth,
		services.OAuthProvider,
		deps.redis,
		deps.log,
	)
	if err != nil {
		deps.log.Fatal("Failed to create gRPC server", map[string]interface{}{
			"error": err.Error(),
		})
	}

	go func() {
		deps.log.Info("HTTP server starting", map[string]interface{}{
			"port": deps.cfg.Server.Port,
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			deps.log.Fatal("Failed to start HTTP server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	go func() {
		if err := grpcSrv.Start(); err != nil {
			deps.log.Fatal("Failed to start gRPC server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	deps.log.Info("Shutting down servers...")

	// Stop LDAP sync job
	if ldapSyncJob != nil {
		ldapSyncJob.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		deps.log.Error("HTTP server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	grpcSrv.Stop()

	deps.log.Info("Servers exited successfully")
}

func buildInfra() (*infra, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	log := logger.New("auth-gateway", logger.LogLevel(cfg.Server.LogLevel), true)
	logger.SetDefault(log)
	log.Info("Starting Auth Gateway", map[string]interface{}{
		"env":  cfg.Server.Env,
		"port": cfg.Server.Port,
	})

	var keyManager *keys.Manager
	if cfg.OIDC.Enabled {
		keyConfigs := buildKeyConfigs(&cfg.OIDC)
		if len(keyConfigs) > 0 {
			keyManager, err = keys.NewManager(keyConfigs, cfg.OIDC.SigningKeyID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to init key manager: %w", err)
			}
			log.Info("OIDC key manager initialized", map[string]interface{}{
				"keys": len(keyConfigs),
			})
		}
	}

	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info("Database connected successfully")

	redis, err := service.NewRedisService(&cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	log.Info("Redis connected successfully")

	jwtService := jwt.NewService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpires,
		cfg.JWT.RefreshExpires,
	)

	var oidcJWTService *jwt.OIDCService
	if keyManager != nil {
		oidcJWTService = jwt.NewOIDCService(keyManager, cfg.OIDC.Issuer)
	}

	deps := &infra{
		cfg:            cfg,
		log:            log,
		keyManager:     keyManager,
		db:             db,
		redis:          redis,
		jwtService:     jwtService,
		oidcJWTService: oidcJWTService,
		smsProvider:    initSMSProvider(cfg, log),
	}

	cleanup := func() {
		_ = redis.Close()
		_ = db.Close()
	}

	return deps, cleanup, nil
}

func buildRepositories(deps *infra) *repoSet {
	return &repoSet{
		User:          repository.NewUserRepository(deps.db),
		Token:         repository.NewTokenRepository(deps.db),
		OAuth:         repository.NewOAuthRepository(deps.db),
		Audit:         repository.NewAuditRepository(deps.db),
		APIKey:        repository.NewAPIKeyRepository(deps.db),
		OTP:           repository.NewOTPRepository(deps.db),
		BackupCode:    repository.NewBackupCodeRepository(deps.db),
		SMSLog:        repository.NewSMSLogRepository(deps.db),
		RBAC:          repository.NewRBACRepository(deps.db),
		Session:       repository.NewSessionRepository(deps.db),
		IPFilter:      repository.NewIPFilterRepository(deps.db),
		Webhook:       repository.NewWebhookRepository(deps.db),
		Template:      repository.NewTemplateRepository(deps.db),
		Branding:      repository.NewBrandingRepository(deps.db),
		System:        repository.NewSystemRepository(deps.db),
		Geo:           repository.NewGeoRepository(deps.db),
		OAuthProvider: repository.NewOAuthProviderRepository(deps.db),
		Group:         repository.NewGroupRepository(deps.db),
		LDAP:          repository.NewLDAPRepository(deps.db),
		SAML:          repository.NewSAMLRepository(deps.db),
	}
}

func buildServices(deps *infra, repos *repoSet) *serviceSet {
	var geoService *service.GeoService
	if deps.cfg.GeoIP.Enabled {
		geoService = service.NewGeoService(deps.cfg.GeoIP.APIKey)
		deps.log.Info("GeoIP service enabled")
	}

	auditService := service.NewAuditService(repos.Audit, geoService)
	blacklistService := service.NewBlacklistService(deps.redis, repos.Token, deps.jwtService, deps.log, auditService)

	// Synchronize blacklist from database to Redis on startup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := blacklistService.SyncFromDatabase(ctx); err != nil {
		deps.log.Warn("Failed to sync blacklist from database", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail startup, but log the warning
	} else {
		deps.log.Info("Blacklist synchronized from database to Redis")
	}

	sessionService := service.NewSessionService(repos.Session, blacklistService, deps.log)
	userService := service.NewUserService(repos.User, auditService)
	apiKeyService := service.NewAPIKeyService(repos.APIKey, repos.User, auditService)
	emailService := service.NewEmailService(&deps.cfg.SMTP)
	otpService := service.NewOTPService(
		repos.OTP,
		repos.User,
		auditService,
		service.OTPServiceOptions{
			EmailSender: emailService,
			SMSProvider: deps.smsProvider,
			SMSLogRepo:  repos.SMSLog,
			Cache:       deps.redis,
			Config:      deps.cfg,
		},
	)
	twoFAService := service.NewTwoFactorService(repos.User, repos.BackupCode, "Auth Gateway")
	// Convert password policy config to utils PasswordPolicy
	passwordPolicy := utils.PasswordPolicy{
		MinLength:        deps.cfg.Security.PasswordPolicy.MinLength,
		RequireUppercase: deps.cfg.Security.PasswordPolicy.RequireUppercase,
		RequireLowercase: deps.cfg.Security.PasswordPolicy.RequireLowercase,
		RequireNumbers:   deps.cfg.Security.PasswordPolicy.RequireNumbers,
		RequireSpecial:   deps.cfg.Security.PasswordPolicy.RequireSpecial,
		MaxLength:        deps.cfg.Security.PasswordPolicy.MaxLength,
	}
	authService := service.NewAuthService(repos.User, repos.Token, repos.RBAC, auditService, deps.jwtService, blacklistService, deps.redis, sessionService, twoFAService, deps.cfg.Security.BcryptCost, passwordPolicy, deps.db)
	oauthService := service.NewOAuthService(repos.User, repos.OAuth, repos.Token, repos.Audit, repos.RBAC, deps.jwtService, sessionService, &http.Client{Timeout: 10 * time.Second}, deps.cfg.Security.JITProvisioning)
	adminService := service.NewAdminService(repos.User, repos.APIKey, repos.Audit, repos.OAuth, repos.RBAC, deps.cfg.Security.BcryptCost, deps.db)
	rbacService := service.NewRBACService(repos.RBAC, auditService)
	ipFilterService := service.NewIPFilterService(repos.IPFilter)
	webhookService := service.NewWebhookService(repos.Webhook, auditService)
	templateService := service.NewTemplateService(repos.Template, auditService)

	var oauthProviderService *service.OAuthProviderService
	if deps.cfg.OIDC.Enabled && deps.oidcJWTService != nil {
		baseURL := deps.cfg.OIDC.Issuer
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", deps.cfg.Server.Port)
		}
		oauthProviderService = service.NewOAuthProviderService(
			repos.OAuthProvider,
			repos.User,
			repos.Audit,
			sessionService,
			deps.oidcJWTService,
			deps.keyManager,
			deps.log,
			deps.cfg.OIDC.Issuer,
			baseURL,
		)
	}

	var minimalOAuth *service.OAuthProviderService
	if oauthProviderService != nil {
		minimalOAuth = oauthProviderService
	} else {
		minimalOAuth = service.NewOAuthProviderServiceMinimal(repos.OAuthProvider, repos.Audit, deps.log)
	}

	groupService := service.NewGroupService(repos.Group, repos.User, deps.log)

	// SCIM Service
	baseURL := deps.cfg.OIDC.Issuer
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", deps.cfg.Server.Port)
	}
	scimService := service.NewSCIMService(repos.User, repos.Group, deps.log, baseURL)

	// SAML Service
	samlService := service.NewSAMLService(
		repos.SAML,
		repos.User,
		repos.RBAC,
		deps.log,
		baseURL,
		baseURL,
	)

	// LDAP Service
	ldapService := service.NewLDAPService(
		repos.LDAP,
		repos.User,
		repos.Group,
		deps.log,
	)

	// Bulk Service
	bulkService := service.NewBulkService(
		repos.User,
		repos.RBAC,
		deps.log,
		deps.cfg.Security.BcryptCost,
	)

	return &serviceSet{
		Geo:             geoService,
		Audit:           auditService,
		Blacklist:       blacklistService,
		Session:         sessionService,
		Auth:            authService,
		User:            userService,
		APIKey:          apiKeyService,
		Email:           emailService,
		OTP:             otpService,
		OAuth:           oauthService,
		TwoFA:           twoFAService,
		Admin:           adminService,
		RBAC:            rbacService,
		IPFilter:        ipFilterService,
		Webhook:         webhookService,
		Template:        templateService,
		OAuthProvider:   oauthProviderService,
		MinimalOAuthSvc: minimalOAuth,
		Group:           groupService,
		LDAP:            ldapService,
		Bulk:            bulkService,
		SCIM:            scimService,
		SAML:            samlService,
	}
}

func buildHandlers(deps *infra, repos *repoSet, services *serviceSet) *handlerSet {
	authHandler := handler.NewAuthHandler(services.Auth, services.User, services.OTP, deps.log)
	healthHandler := handler.NewHealthHandler(deps.db, deps.redis)
	apiKeyHandler := handler.NewAPIKeyHandler(services.APIKey, deps.log)
	otpHandler := handler.NewOTPHandler(services.OTP, services.Auth, deps.log)
	oauthHandler := handler.NewOAuthHandler(services.OAuth, deps.log)
	twoFAHandler := handler.NewTwoFactorHandler(services.TwoFA, services.User, deps.log)
	adminHandler := handler.NewAdminHandler(services.Admin, deps.log)
	advancedAdminHandler := handler.NewAdvancedAdminHandler(services.RBAC, services.Session, services.IPFilter, repos.Branding, repos.System, repos.Geo, deps.log)
	webhookHandler := handler.NewWebhookHandler(services.Webhook, deps.log)
	templateHandler := handler.NewTemplateHandler(services.Template, deps.log)

	var oauthProviderHandler *handler.OAuthProviderHandler
	if services.OAuthProvider != nil {
		oauthProviderHandler = handler.NewOAuthProviderHandler(services.OAuthProvider, deps.log)
	}
	oauthAdminHandler := handler.NewOAuthAdminHandler(services.MinimalOAuthSvc, deps.log)

	secureCookie := deps.cfg.Server.Env == "production"
	loginHandler := handler.NewLoginHandler(services.Auth, services.OTP, deps.jwtService, deps.log, secureCookie)
	groupHandler := handler.NewGroupHandler(services.Group, deps.log)
	ldapHandler := handler.NewLDAPHandler(services.LDAP, deps.log)
	bulkHandler := handler.NewBulkHandler(services.Bulk, deps.log)
	scimHandler := handler.NewSCIMHandler(services.SCIM, deps.log)
	samlHandler := handler.NewSAMLHandler(services.SAML, deps.log)

	return &handlerSet{
		Auth:          authHandler,
		Health:        healthHandler,
		APIKey:        apiKeyHandler,
		OTP:           otpHandler,
		OAuth:         oauthHandler,
		TwoFA:         twoFAHandler,
		Admin:         adminHandler,
		AdvancedAdmin: advancedAdminHandler,
		Webhook:       webhookHandler,
		Template:      templateHandler,
		OAuthProvider: oauthProviderHandler,
		OAuthAdmin:    oauthAdminHandler,
		Login:         loginHandler,
		Group:         groupHandler,
		SCIM:          scimHandler,
		LDAP:          ldapHandler,
		Bulk:          bulkHandler,
		SAML:          samlHandler,
	}
}

func buildMiddlewares(deps *infra, repos *repoSet, services *serviceSet) *middlewareSet {
	authMiddleware := middleware.NewAuthMiddleware(deps.jwtService, services.Blacklist)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(services.APIKey, repos.RBAC)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(deps.redis, &deps.cfg.RateLimit)
	ipFilterMiddleware := middleware.NewIPFilterMiddleware(services.IPFilter)
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(repos.System)

	return &middlewareSet{
		Auth:        authMiddleware,
		APIKey:      apiKeyMiddleware,
		RateLimit:   rateLimitMiddleware,
		IPFilter:    ipFilterMiddleware,
		Maintenance: maintenanceMiddleware,
	}
}

func buildRouter(deps *infra, services *serviceSet, handlers *handlerSet, middlewares *middlewareSet) *gin.Engine {
	if deps.cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	if deps.cfg.OIDC.Enabled {
		tmpl, err := template.ParseGlob("internal/templates/*.html")
		if err != nil {
			deps.log.Warn("Failed to load OAuth templates", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			router.SetHTMLTemplate(tmpl)
			deps.log.Info("OAuth templates loaded successfully")
		}
	}

	router.Use(middleware.Recovery(deps.log))
	router.Use(middleware.Logger(deps.log))
	router.Use(middleware.SetupCORS(&deps.cfg.CORS))
	router.Use(middlewares.Maintenance.CheckMaintenance())
	router.Use(middlewares.IPFilter.CheckIPFilter())

	// Metrics middleware (if enabled)
	if deps.cfg.Metrics.Enabled {
		router.Use(middleware.MetricsMiddleware())
		router.Use(middleware.MetricsErrorMiddleware())
	}

	router.GET("/api/swagger.json", func(c *gin.Context) {
		c.File("docs/swagger.json")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/api/swagger.json"),
	))

	router.GET("/health", handlers.Health.Health)
	router.GET("/ready", handlers.Health.Readiness)
	router.GET("/live", handlers.Health.Liveness)
	router.GET("/system/maintenance", handlers.AdvancedAdmin.GetMaintenanceMode)

	// Metrics endpoint (if enabled)
	if deps.cfg.Metrics.Enabled {
		metricsHandler := handler.NewMetricsHandler()
		router.GET("/metrics", metricsHandler.PrometheusMetrics)
		deps.log.Info("Prometheus metrics enabled", map[string]interface{}{
			"endpoint": "/metrics",
		})
	}

	// SAML 2.0 Endpoints
	if handlers.SAML != nil {
		router.GET("/saml/metadata", handlers.SAML.GetMetadata)
		router.POST("/saml/sso", middlewares.Auth.Authenticate(), handlers.SAML.SSO)
		router.POST("/saml/slo", handlers.SAML.SLO)
	}

	// SCIM 2.0 API endpoints
	if handlers.SCIM != nil {
		scimGroup := router.Group("/scim/v2")
		scimGroup.Use(middlewares.Auth.Authenticate()) // Require authentication
		{
			// Users endpoints
			scimGroup.GET("/Users", handlers.SCIM.GetUsers)
			scimGroup.POST("/Users", handlers.SCIM.CreateUser)
			scimGroup.GET("/Users/:id", handlers.SCIM.GetUser)
			scimGroup.PUT("/Users/:id", handlers.SCIM.UpdateUser)
			scimGroup.PATCH("/Users/:id", handlers.SCIM.PatchUser)
			scimGroup.DELETE("/Users/:id", handlers.SCIM.DeleteUser)

			// Groups endpoints
			scimGroup.GET("/Groups", handlers.SCIM.GetGroups)
			scimGroup.GET("/Groups/:id", handlers.SCIM.GetGroup)

			// Service Provider Config
			scimGroup.GET("/ServiceProviderConfig", handlers.SCIM.GetServiceProviderConfig)

			// Schemas
			scimGroup.GET("/Schemas", handlers.SCIM.GetSchemas)
		}
	}

	loginGroup := router.Group("")
	loginGroup.Use(handlers.Login.SessionMiddleware())
	{
		loginGroup.GET("/login", handlers.Login.LoginPage)
		loginGroup.POST("/login", handlers.Login.LoginSubmit)
		loginGroup.GET("/login/otp", handlers.Login.OTPVerifyPage)
		loginGroup.POST("/login/otp", handlers.Login.OTPVerifySubmit)
		loginGroup.GET("/logout", handlers.Login.Logout)
	}

	if handlers.OAuthProvider != nil {
		router.GET("/.well-known/openid-configuration", handlers.OAuthProvider.Discovery)
		router.GET("/.well-known/jwks.json", handlers.OAuthProvider.JWKS)

		oauth := router.Group("/oauth")
		{
			oauth.GET("/authorize", handlers.OAuthProvider.Authorize)
			oauth.POST("/token", handlers.OAuthProvider.Token)
			oauth.POST("/introspect", handlers.OAuthProvider.Introspect)
			oauth.POST("/revoke", handlers.OAuthProvider.Revoke)
			oauth.GET("/userinfo", handlers.OAuthProvider.UserInfo)
			oauth.POST("/device/code", handlers.OAuthProvider.DeviceCode)
			oauth.POST("/device/token", handlers.OAuthProvider.DeviceToken)
			oauth.GET("/device", handlers.OAuthProvider.DeviceVerification)
			oauth.POST("/device/approve", handlers.Login.SessionMiddleware(), handlers.OAuthProvider.DeviceApprove)
			oauth.GET("/consent", handlers.Login.SessionMiddleware(), handlers.OAuthProvider.ConsentPage)
			oauth.POST("/consent", handlers.Login.SessionMiddleware(), handlers.OAuthProvider.ConsentSubmit)
		}
	}

	apiGroup := router.Group("/api")
	{
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/signup", middlewares.RateLimit.LimitSignup(), handlers.Auth.SignUp)
			authGroup.POST("/signin", middlewares.RateLimit.LimitSignin(), handlers.Auth.SignIn)
			authGroup.POST("/refresh", middlewares.RateLimit.LimitRefreshToken(), handlers.Auth.RefreshToken)
			authGroup.POST("/verify/resend", handlers.OTP.ResendVerification)
			authGroup.POST("/verify/email", handlers.OTP.VerifyEmailOTP)
			authGroup.POST("/password/reset/request", handlers.Auth.RequestPasswordReset)
			authGroup.POST("/password/reset/complete", handlers.Auth.ResetPassword)
			authGroup.POST("/2fa/login/verify", handlers.Auth.Verify2FA)
		}

		otpGroup := apiGroup.Group("/otp")
		{
			otpGroup.POST("/send", handlers.OTP.SendOTP)
			otpGroup.POST("/verify", handlers.OTP.VerifyOTP)
		}

		passwordlessGroup := apiGroup.Group("/auth/passwordless")
		{
			passwordlessGroup.POST("/request", handlers.OTP.RequestPasswordlessLogin)
			passwordlessGroup.POST("/verify", handlers.OTP.VerifyPasswordlessLogin)
		}

		signupPhoneGroup := apiGroup.Group("/auth/signup/phone")
		{
			signupPhoneGroup.POST("", middlewares.RateLimit.LimitSignup(), handlers.Auth.InitPasswordlessRegistration)
			signupPhoneGroup.POST("/verify", handlers.Auth.CompletePasswordlessRegistration)
		}

		oauthGroup := apiGroup.Group("/auth")
		{
			oauthGroup.GET("/providers", handlers.OAuth.GetProviders)
			oauthGroup.GET("/:provider", handlers.OAuth.Login)
			oauthGroup.GET("/:provider/callback", handlers.OAuth.Callback)
			oauthGroup.POST("/telegram/callback", handlers.OAuth.TelegramCallback)
		}

		protectedAuth := apiGroup.Group("/auth")
		protectedAuth.Use(middlewares.Auth.Authenticate())
		{
			protectedAuth.POST("/logout", handlers.Auth.Logout)
			protectedAuth.GET("/profile", handlers.Auth.GetProfile)
			protectedAuth.PUT("/profile", handlers.Auth.UpdateProfile)
			protectedAuth.POST("/change-password", handlers.Auth.ChangePassword)
			protectedAuth.POST("/2fa/setup", handlers.TwoFA.Setup)
			protectedAuth.POST("/2fa/verify", handlers.TwoFA.Verify)
			protectedAuth.POST("/2fa/disable", handlers.TwoFA.Disable)
			protectedAuth.GET("/2fa/status", handlers.TwoFA.GetStatus)
			protectedAuth.POST("/2fa/backup-codes/regenerate", handlers.TwoFA.RegenerateBackupCodes)
		}

		apiKeysGroup := apiGroup.Group("/api-keys")
		apiKeysGroup.Use(middlewares.Auth.Authenticate())
		{
			apiKeysGroup.POST("", handlers.APIKey.Create)
			apiKeysGroup.GET("", handlers.APIKey.List)
			apiKeysGroup.GET("/:id", handlers.APIKey.Get)
			apiKeysGroup.PUT("/:id", handlers.APIKey.Update)
			apiKeysGroup.POST("/:id/revoke", handlers.APIKey.Revoke)
			apiKeysGroup.DELETE("/:id", handlers.APIKey.Delete)
		}

		adminGroup := apiGroup.Group("/admin")
		adminGroup.Use(middlewares.Auth.Authenticate())
		adminGroup.Use(middleware.RequireAdmin())
		{
			adminGroup.GET("/stats", handlers.Admin.GetStats)
			adminGroup.GET("/users", handlers.Admin.ListUsers)
			adminGroup.POST("/users", handlers.Admin.CreateUser)
			adminGroup.GET("/users/:id", handlers.Admin.GetUser)
			adminGroup.PUT("/users/:id", handlers.Admin.UpdateUser)
			adminGroup.DELETE("/users/:id", handlers.Admin.DeleteUser)
			adminGroup.POST("/users/:id/roles", handlers.Admin.AssignRole)
			adminGroup.DELETE("/users/:id/roles/:roleId", handlers.Admin.RemoveRole)
			adminGroup.GET("/api-keys", handlers.Admin.ListAPIKeys)
			adminGroup.POST("/api-keys/:id/revoke", handlers.Admin.RevokeAPIKey)
			adminGroup.GET("/audit-logs", handlers.Admin.ListAuditLogs)

			rbacGroup := adminGroup.Group("/rbac")
			{
				rbacGroup.GET("/permissions", handlers.AdvancedAdmin.ListPermissions)
				rbacGroup.POST("/permissions", handlers.AdvancedAdmin.CreatePermission)
				rbacGroup.DELETE("/permissions/:id", handlers.AdvancedAdmin.DeletePermission)
				rbacGroup.GET("/roles", handlers.AdvancedAdmin.ListRoles)
				rbacGroup.POST("/roles", handlers.AdvancedAdmin.CreateRole)
				rbacGroup.GET("/roles/:id", handlers.AdvancedAdmin.GetRole)
				rbacGroup.PUT("/roles/:id", handlers.AdvancedAdmin.UpdateRole)
				rbacGroup.DELETE("/roles/:id", handlers.AdvancedAdmin.DeleteRole)
				rbacGroup.GET("/permission-matrix", handlers.AdvancedAdmin.GetPermissionMatrix)
			}

			adminGroup.GET("/sessions", handlers.AdvancedAdmin.ListAllSessions)
			adminGroup.GET("/sessions/stats", handlers.AdvancedAdmin.GetSessionStats)
			adminGroup.DELETE("/sessions/:id", handlers.AdvancedAdmin.AdminRevokeSession)

			ipFilterGroup := adminGroup.Group("/ip-filters")
			{
				ipFilterGroup.GET("", handlers.AdvancedAdmin.ListIPFilters)
				ipFilterGroup.POST("", handlers.AdvancedAdmin.CreateIPFilter)
				ipFilterGroup.DELETE("/:id", handlers.AdvancedAdmin.DeleteIPFilter)
			}

			adminGroup.PUT("/branding", handlers.AdvancedAdmin.UpdateBranding)

			systemGroup := adminGroup.Group("/system")
			{
				systemGroup.PUT("/maintenance", handlers.AdvancedAdmin.SetMaintenanceMode)
				systemGroup.GET("/health", handlers.AdvancedAdmin.GetSystemHealth)
			}

			analyticsGroup := adminGroup.Group("/analytics")
			{
				analyticsGroup.GET("/geo-distribution", handlers.AdvancedAdmin.GetGeoDistribution)
			}

			webhooksGroup := adminGroup.Group("/webhooks")
			{
				webhooksGroup.GET("", handlers.Webhook.ListWebhooks)
				webhooksGroup.POST("", handlers.Webhook.CreateWebhook)
				webhooksGroup.GET("/events", handlers.Webhook.GetAvailableEvents)
				webhooksGroup.GET("/:id", handlers.Webhook.GetWebhook)
				webhooksGroup.PUT("/:id", handlers.Webhook.UpdateWebhook)
				webhooksGroup.DELETE("/:id", handlers.Webhook.DeleteWebhook)
				webhooksGroup.POST("/:id/test", handlers.Webhook.TestWebhook)
				webhooksGroup.GET("/:id/deliveries", handlers.Webhook.ListWebhookDeliveries)
			}

			templatesGroup := adminGroup.Group("/templates")
			{
				templatesGroup.GET("", handlers.Template.ListEmailTemplates)
				templatesGroup.POST("", handlers.Template.CreateEmailTemplate)
				templatesGroup.GET("/types", handlers.Template.GetAvailableTemplateTypes)
				templatesGroup.POST("/preview", handlers.Template.PreviewEmailTemplate)
				templatesGroup.GET("/variables/:type", handlers.Template.GetDefaultVariables)
				templatesGroup.GET("/:id", handlers.Template.GetEmailTemplate)
				templatesGroup.PUT("/:id", handlers.Template.UpdateEmailTemplate)
				templatesGroup.DELETE("/:id", handlers.Template.DeleteEmailTemplate)
			}

			adminOAuth := adminGroup.Group("/oauth")
			{
				adminOAuth.POST("/clients", handlers.OAuthAdmin.CreateClient)
				adminOAuth.GET("/clients", handlers.OAuthAdmin.ListClients)
				adminOAuth.GET("/clients/:id", handlers.OAuthAdmin.GetClient)
				adminOAuth.PUT("/clients/:id", handlers.OAuthAdmin.UpdateClient)
				adminOAuth.DELETE("/clients/:id", handlers.OAuthAdmin.DeleteClient)
				adminOAuth.POST("/clients/:id/rotate-secret", handlers.OAuthAdmin.RotateSecret)

				adminOAuth.GET("/scopes", handlers.OAuthAdmin.ListScopes)
				adminOAuth.POST("/scopes", handlers.OAuthAdmin.CreateScope)
				adminOAuth.DELETE("/scopes/:id", handlers.OAuthAdmin.DeleteScope)

				adminOAuth.GET("/clients/:id/consents", handlers.OAuthAdmin.ListClientConsents)
				adminOAuth.DELETE("/clients/:id/consents/:user_id", handlers.OAuthAdmin.RevokeUserConsent)
			}

			groupsGroup := adminGroup.Group("/groups")
			{
				groupsGroup.GET("", handlers.Group.ListGroups)
				groupsGroup.POST("", handlers.Group.CreateGroup)
				groupsGroup.GET("/:id", handlers.Group.GetGroup)
				groupsGroup.PUT("/:id", handlers.Group.UpdateGroup)
				groupsGroup.DELETE("/:id", handlers.Group.DeleteGroup)
				groupsGroup.GET("/:id/members", handlers.Group.GetGroupMembers)
				groupsGroup.POST("/:id/members", handlers.Group.AddGroupMembers)
				groupsGroup.DELETE("/:id/members/:user_id", handlers.Group.RemoveGroupMember)
			}

			ldapGroup := adminGroup.Group("/ldap")
			{
				ldapGroup.GET("/config", handlers.LDAP.GetActiveConfig)
				ldapGroup.GET("/configs", handlers.LDAP.ListConfigs)
				ldapGroup.POST("/config", handlers.LDAP.CreateConfig)
				ldapGroup.GET("/config/:id", handlers.LDAP.GetConfig)
				ldapGroup.PUT("/config/:id", handlers.LDAP.UpdateConfig)
				ldapGroup.DELETE("/config/:id", handlers.LDAP.DeleteConfig)
				ldapGroup.POST("/test-connection", handlers.LDAP.TestConnection)
				ldapGroup.POST("/config/:id/sync", handlers.LDAP.Sync)
				ldapGroup.GET("/config/:id/sync-logs", handlers.LDAP.GetSyncLogs)
			}

			bulkGroup := adminGroup.Group("/users")
			{
				bulkGroup.POST("/bulk-create", handlers.Bulk.BulkCreateUsers)
				bulkGroup.PUT("/bulk-update", handlers.Bulk.BulkUpdateUsers)
				bulkGroup.POST("/bulk-delete", handlers.Bulk.BulkDeleteUsers)
				bulkGroup.POST("/bulk-assign-roles", handlers.Bulk.BulkAssignRoles)
			}

			samlGroup := adminGroup.Group("/saml")
			{
				samlGroup.GET("/sp", handlers.SAML.ListSPs)
				samlGroup.POST("/sp", handlers.SAML.CreateSP)
				samlGroup.GET("/sp/:id", handlers.SAML.GetSP)
				samlGroup.PUT("/sp/:id", handlers.SAML.UpdateSP)
				samlGroup.DELETE("/sp/:id", handlers.SAML.DeleteSP)
			}
		}

		sessionsGroup := apiGroup.Group("/sessions")
		sessionsGroup.Use(middlewares.Auth.Authenticate())
		{
			sessionsGroup.GET("", handlers.AdvancedAdmin.ListUserSessions)
			sessionsGroup.DELETE("/:id", handlers.AdvancedAdmin.RevokeSession)
			sessionsGroup.POST("/revoke-all", handlers.AdvancedAdmin.RevokeAllSessions)
		}

		protectedAPI := apiGroup.Group("/v1")
		protectedAPI.Use(func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && !strings.HasPrefix(authHeader, "Bearer agw_") {
				middlewares.Auth.Authenticate()(c)
				return
			}
			middlewares.APIKey.Authenticate()(c)
		})
		{
			protectedAPI.GET("/profile", handlers.Auth.GetProfile)
		}
	}

	return router
}

// startTokenCleanup starts a background routine to clean up expired tokens
func startTokenCleanup(tokenRepo *repository.TokenRepository, interval time.Duration, log *logger.Logger) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Debug("Running token cleanup")

			// Use context with timeout for cleanup operation
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := tokenRepo.CleanupExpiredTokens(ctx); err != nil {
				log.Error("Token cleanup failed", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				log.Debug("Token cleanup completed successfully")
			}
			cancel()
		}
	}()
}

// startMetricsCollection starts a background routine to collect and update metrics
func startMetricsCollection(db *repository.Database, redis *service.RedisService, log *logger.Logger) {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
		defer ticker.Stop()

		for range ticker.C {
			// Update database connection pool metrics
			stats := db.Stats()
			metrics.UpdateDBConnections(stats)

			// TODO: Update Redis metrics if Redis service exposes stats
			// TODO: Update active sessions count
			_ = redis
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

// initSMSProvider initializes an SMS provider based on configuration; returns nil if disabled or misconfigured.
func initSMSProvider(cfg *config.Config, log *logger.Logger) sms.SMSProvider {
	if !cfg.SMS.Enabled {
		return nil
	}

	providerCfg := sms.ProviderConfig{
		Provider:        sms.ProviderType(cfg.SMS.Provider),
		EnableMockInDev: cfg.Server.Env != "production",
	}

	switch sms.ProviderType(cfg.SMS.Provider) {
	case sms.ProviderTwilio:
		providerCfg.TwilioConfig = &sms.TwilioConfig{
			AccountSID: cfg.SMS.TwilioAccountSID,
			AuthToken:  cfg.SMS.TwilioAuthToken,
			FromNumber: cfg.SMS.TwilioFromNumber,
		}
	case sms.ProviderAWSSNS:
		providerCfg.AWSSNSConfig = &sms.AWSSNSConfig{
			Region:          cfg.SMS.AWSRegion,
			AccessKeyID:     cfg.SMS.AWSAccessKeyID,
			SecretAccessKey: cfg.SMS.AWSSecretAccessKey,
			FromNumber:      cfg.SMS.AWSSenderID,
		}
	default:
		log.Warn("Unsupported SMS provider configured; SMS disabled", map[string]interface{}{
			"provider": cfg.SMS.Provider,
		})
		return nil
	}

	// Use context with timeout for SMS provider initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	provider, err := sms.NewProvider(ctx, providerCfg)
	if err != nil {
		log.Warn("SMS provider initialization failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil
	}

	log.Info("SMS provider initialized", map[string]interface{}{
		"provider": cfg.SMS.Provider,
	})
	return provider
}
