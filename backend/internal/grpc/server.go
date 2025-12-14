package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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
	port string,
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	rbacRepo *repository.RBACRepository,
	apiKeyService *service.APIKeyService,
	authService *service.AuthService,
	redis *service.RedisService,
	log *logger.Logger,
) (*Server, error) {
	// Create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}
	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
			recoveryInterceptor(log),
		),
	)

	// Register auth service handler
	handler := NewAuthHandlerV2(jwtService, userRepo, tokenRepo, rbacRepo, apiKeyService, authService, redis, log)
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
