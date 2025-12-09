package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	// oauthRepo := repository.NewOAuthRepository(db) // TODO: будет использоваться для OAuth
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, auditRepo, jwtService, redis, cfg.Security.BcryptCost)
	userService := service.NewUserService(userRepo, auditRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService, log)
	healthHandler := handler.NewHealthHandler(db, redis)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, redis, tokenRepo)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redis, &cfg.RateLimit)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery(log))
	router.Use(middleware.Logger(log))
	router.Use(middleware.SetupCORS(&cfg.CORS))

	// Health check endpoints (no auth required)
	router.GET("/auth/health", healthHandler.Health)
	router.GET("/auth/ready", healthHandler.Readiness)
	router.GET("/auth/live", healthHandler.Liveness)

	// Public auth endpoints
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/signup", rateLimitMiddleware.LimitSignup(), authHandler.SignUp)
		authGroup.POST("/signin", rateLimitMiddleware.LimitSignin(), authHandler.SignIn)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected auth endpoints (require authentication)
	protectedAuth := router.Group("/auth")
	protectedAuth.Use(authMiddleware.Authenticate())
	{
		protectedAuth.POST("/logout", authHandler.Logout)
		protectedAuth.GET("/profile", authHandler.GetProfile)
		protectedAuth.PUT("/profile", authHandler.UpdateProfile)
		protectedAuth.POST("/change-password", authHandler.ChangePassword)
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

			if err := tokenRepo.CleanupExpiredTokens(); err != nil {
				log.Error("Token cleanup failed", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				log.Debug("Token cleanup completed successfully")
			}
		}
	}()
}
