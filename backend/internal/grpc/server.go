package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	pb "github.com/smilemakc/auth-gateway/proto"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
}

// NewServer creates a new gRPC server
func NewServer(
	grpcConfig *config.GRPCConfig,
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	rbacRepo *repository.RBACRepository,
	apiKeyService *service.APIKeyService,
	authService *service.AuthService,
	oauthProviderService *service.OAuthProviderService,
	otpService *service.OTPService,
	emailProfileService *service.EmailProfileService,
	adminService *service.AdminService,
	appService *service.ApplicationService,
	redis *service.RedisService,
	tokenExchangeService *service.TokenExchangeService,
	log *logger.Logger,
) (*Server, error) {
	// Create listener
	lis, err := net.Listen("tcp", ":"+grpcConfig.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", grpcConfig.Port, err)
	}

	// Build server options
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			apiKeyAuthInterceptor(apiKeyService, appService, log),
			contextExtractorInterceptor(log),
			loggingInterceptor(log),
			recoveryInterceptor(log),
		),
	}

	// Add TLS credentials if enabled
	if grpcConfig.TLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(grpcConfig.TLSCert, grpcConfig.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
		log.Info("gRPC TLS enabled", map[string]interface{}{
			"cert": grpcConfig.TLSCert,
		})
	}

	// Create gRPC server with options
	grpcServer := grpc.NewServer(serverOpts...)

	// Register auth service handler
	handler := NewAuthHandlerV2(jwtService, userRepo, tokenRepo, rbacRepo, apiKeyService, authService, oauthProviderService, otpService, emailProfileService, adminService, appService, redis, tokenExchangeService, log)
	pb.RegisterAuthServiceServer(grpcServer, handler)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
		logger:     log,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	s.logger.Info("Starting gRPC server", map[string]interface{}{
		"address": s.listener.Addr().String(),
	})

	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
