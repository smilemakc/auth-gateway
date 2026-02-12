package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// Context keys for gRPC
const (
	GRPCApplicationIDKey = "grpc_application_id"
	GRPCTenantIDKey      = "grpc_tenant_id"
)

// contextExtractorInterceptor extracts application context from gRPC metadata
// and adds it to the context for use in handlers.
//
// Supported metadata keys:
// - x-application-id: Application UUID
// - x-tenant-id: Tenant UUID (for future multi-tenancy support)
func contextExtractorInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		if appIDs := md.Get("x-application-id"); len(appIDs) > 0 && appIDs[0] != "" {
			ctx = context.WithValue(ctx, GRPCApplicationIDKey, appIDs[0])
			if log != nil {
				log.Debug("gRPC application context extracted", map[string]interface{}{
					"application_id": appIDs[0],
					"method":         info.FullMethod,
				})
			}
		}

		if tenantIDs := md.Get("x-tenant-id"); len(tenantIDs) > 0 && tenantIDs[0] != "" {
			ctx = context.WithValue(ctx, GRPCTenantIDKey, tenantIDs[0])
		}

		return handler(ctx, req)
	}
}

// loggingInterceptor logs all gRPC requests
func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Log request
		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		fields := map[string]interface{}{
			"method":      info.FullMethod,
			"duration_ms": duration.Milliseconds(),
			"status":      statusCode.String(),
		}

		if err != nil {
			fields["error"] = err.Error()
			log.Warn("gRPC request failed", fields)
		} else {
			log.Info("gRPC request", fields)
		}

		return resp, err
	}
}

// recoveryInterceptor recovers from panics in gRPC handlers
func recoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("gRPC panic recovered", map[string]interface{}{
					"method": info.FullMethod,
					"panic":  fmt.Sprintf("%v", r),
				})
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// GetApplicationIDFromGRPCContext extracts application_id from gRPC context
func GetApplicationIDFromGRPCContext(ctx context.Context) *string {
	val := ctx.Value(GRPCApplicationIDKey)
	if val == nil {
		return nil
	}

	appID, ok := val.(string)
	if !ok || appID == "" {
		return nil
	}

	return &appID
}

// GetApplicationUUIDFromGRPCContext extracts and parses application_id as UUID
func GetApplicationUUIDFromGRPCContext(ctx context.Context) *uuid.UUID {
	appIDStr := GetApplicationIDFromGRPCContext(ctx)
	if appIDStr == nil {
		return nil
	}

	appID, err := uuid.Parse(*appIDStr)
	if err != nil {
		return nil
	}

	return &appID
}

// GetTenantIDFromGRPCContext extracts tenant_id from gRPC context
func GetTenantIDFromGRPCContext(ctx context.Context) *string {
	val := ctx.Value(GRPCTenantIDKey)
	if val == nil {
		return nil
	}

	tenantID, ok := val.(string)
	if !ok || tenantID == "" {
		return nil
	}

	return &tenantID
}
