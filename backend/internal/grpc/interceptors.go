package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// Context keys for gRPC
const (
	GRPCApplicationIDKey = "grpc_application_id"
	GRPCTenantIDKey      = "grpc_tenant_id"
	GRPCAPIKeyKey        = "grpc_api_key"
	GRPCUserIDKey        = "grpc_user_id"
	GRPCUserEmailKey     = "grpc_user_email"
)

// contextExtractorInterceptor extracts additional context from gRPC metadata.
// Note: x-application-id is handled by apiKeyAuthInterceptor to avoid conflicts
// with application_id set during app secret authentication.
//
// Supported metadata keys:
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

// ResolveApplicationID returns application_id from request or falls back to context
// (set by app secret authentication). Returns empty string if neither is available.
func ResolveApplicationID(ctx context.Context, reqApplicationID string) string {
	if reqApplicationID != "" {
		return reqApplicationID
	}
	if appID := GetApplicationIDFromGRPCContext(ctx); appID != nil {
		return *appID
	}
	return ""
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

// extractCredential extracts API key or application secret from gRPC metadata.
// Checks x-api-key header first, then Authorization: Bearer header.
func extractCredential(md metadata.MD) string {
	if keys := md.Get("x-api-key"); len(keys) > 0 && keys[0] != "" {
		return keys[0]
	}
	if authHeaders := md.Get("authorization"); len(authHeaders) > 0 && authHeaders[0] != "" {
		authHeader := authHeaders[0]
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	return ""
}

// methodScopes maps gRPC method names to required API key scopes
var methodScopes = map[string]models.APIKeyScope{
	"/auth.AuthService/ValidateToken":                    models.ScopeValidateToken,
	"/auth.AuthService/IntrospectToken":                  models.ScopeIntrospectToken,
	"/auth.AuthService/GetUser":                          models.ScopeReadUsers,
	"/auth.AuthService/CheckPermission":                  models.ScopeReadUsers,
	"/auth.AuthService/GetApplicationAuthConfig":         models.ScopeReadUsers,
	"/auth.AuthService/GetUserApplicationProfile":  models.ScopeReadProfile,
	"/auth.AuthService/UpdateUserProfile":           models.ScopeReadUsers,
	"/auth.AuthService/CreateUserProfile":           models.ScopeAuthRegister,
	"/auth.AuthService/DeleteUserProfile":           models.ScopeReadUsers,
	"/auth.AuthService/BanUser":                     models.ScopeReadUsers,
	"/auth.AuthService/UnbanUser":                   models.ScopeReadUsers,
	"/auth.AuthService/ListApplicationUsers":        models.ScopeReadUsers,
	"/auth.AuthService/CreateUser":                       models.ScopeAuthRegister,
	"/auth.AuthService/Login":                            models.ScopeAuthLogin,
	"/auth.AuthService/SendOTP":                          models.ScopeAuthOTP,
	"/auth.AuthService/VerifyOTP":                        models.ScopeAuthOTP,
	"/auth.AuthService/LoginWithOTP":                     models.ScopeAuthOTP,
	"/auth.AuthService/VerifyLoginOTP":                   models.ScopeAuthOTP,
	"/auth.AuthService/RegisterWithOTP":                  models.ScopeAuthRegister,
	"/auth.AuthService/VerifyRegistrationOTP":            models.ScopeAuthRegister,
	"/auth.AuthService/InitPasswordlessRegistration":     models.ScopeAuthRegister,
	"/auth.AuthService/CompletePasswordlessRegistration": models.ScopeAuthRegister,
	"/auth.AuthService/SyncUsers":                        models.ScopeSyncUsers,
	"/auth.AuthService/SendEmail":                        models.ScopeEmailSend,
	"/auth.AuthService/IntrospectOAuthToken":             models.ScopeOAuthRead,
	"/auth.AuthService/ValidateOAuthClient":              models.ScopeOAuthRead,
	"/auth.AuthService/GetOAuthClient":                   models.ScopeOAuthRead,
	"/auth.AuthService/CreateTokenExchange":              models.ScopeExchangeManage,
	"/auth.AuthService/RedeemTokenExchange":              models.ScopeExchangeManage,
}

// NOTE: GetUserTelegramBots is excluded from methodScopes until fully implemented.
// The handler exists but returns codes.Unimplemented. Deny-by-default interceptor blocks it.

// apiKeyAuthInterceptor validates API key authentication for all gRPC requests
func apiKeyAuthInterceptor(apiKeyService *service.APIKeyService, appService *service.ApplicationService, log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Warn("gRPC auth failed: no metadata", map[string]interface{}{
				"method": info.FullMethod,
			})
			return nil, status.Error(codes.Unauthenticated, "missing API key: provide x-api-key metadata")
		}

		// Extract credential from metadata
		credential := extractCredential(md)
		if credential == "" {
			log.Warn("gRPC auth failed: credential not found", map[string]interface{}{
				"method": info.FullMethod,
			})
			return nil, status.Error(codes.Unauthenticated, "missing API key or application secret: provide x-api-key metadata")
		}

		// Verify method is allowed (deny-by-default)
		_, scopeRequired := methodScopes[info.FullMethod]
		if !scopeRequired {
			log.Warn("gRPC auth failed: method not configured", map[string]interface{}{
				"method": info.FullMethod,
			})
			return nil, status.Errorf(codes.PermissionDenied, "method %s is not configured for access", info.FullMethod)
		}

		// Authenticate based on credential type
		if strings.HasPrefix(credential, "app_") {
			// Application secret authentication
			app, err := appService.ValidateSecret(ctx, credential)
			if err != nil {
				log.Warn("gRPC auth failed: invalid application secret", map[string]interface{}{
					"method": info.FullMethod,
					"error":  err.Error(),
				})
				return nil, status.Error(codes.Unauthenticated, "invalid application secret")
			}

			// Scope enforcement: if application has AllowedGRPCScopes defined, restrict access
			if len(app.AllowedGRPCScopes) > 0 {
				requiredScope := string(methodScopes[info.FullMethod])
				if !containsScope(app.AllowedGRPCScopes, requiredScope) {
					log.Warn("gRPC auth failed: application scope restriction", map[string]interface{}{
						"method":         info.FullMethod,
						"required_scope": requiredScope,
						"app_name":       app.Name,
						"allowed_scopes": app.AllowedGRPCScopes,
					})
					return nil, status.Errorf(codes.PermissionDenied, "application not authorized for scope %q", requiredScope)
				}
			}

			ctx = context.WithValue(ctx, GRPCApplicationIDKey, app.ID.String())

			log.Debug("gRPC auth successful (application secret)", map[string]interface{}{
				"method":         info.FullMethod,
				"application_id": app.ID.String(),
				"app_name":       app.Name,
			})

			return handler(ctx, req)
		}

		// API key authentication
		apiKeyObj, user, err := apiKeyService.ValidateAPIKey(ctx, credential)
		if err != nil {
			log.Warn("gRPC auth failed: invalid API key", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Error(codes.Unauthenticated, "invalid API key")
		}

		requiredScope := methodScopes[info.FullMethod]
		if !apiKeyService.HasScope(apiKeyObj, requiredScope) {
			log.Warn("gRPC auth failed: insufficient scope", map[string]interface{}{
				"method":         info.FullMethod,
				"required_scope": string(requiredScope),
				"user_id":        user.ID.String(),
			})
			return nil, status.Errorf(codes.PermissionDenied, "insufficient scope: requires %s", string(requiredScope))
		}

		if appIDs := md.Get("x-application-id"); len(appIDs) > 0 && appIDs[0] != "" {
			appID, err := uuid.Parse(appIDs[0])
			if err != nil {
				log.Warn("gRPC auth failed: invalid application ID", map[string]interface{}{
					"method":         info.FullMethod,
					"application_id": appIDs[0],
					"error":          err.Error(),
				})
				return nil, status.Error(codes.InvalidArgument, "invalid application ID format")
			}

			_, err = appService.GetByID(ctx, appID)
			if err != nil {
				log.Warn("gRPC auth failed: application not found", map[string]interface{}{
					"method":         info.FullMethod,
					"application_id": appID.String(),
					"error":          err.Error(),
				})
				return nil, status.Error(codes.InvalidArgument, "application not found")
			}

			ctx = context.WithValue(ctx, GRPCApplicationIDKey, appID.String())
		}

		ctx = context.WithValue(ctx, GRPCAPIKeyKey, apiKeyObj)
		ctx = context.WithValue(ctx, GRPCUserIDKey, user.ID.String())
		ctx = context.WithValue(ctx, GRPCUserEmailKey, user.Email)

		log.Debug("gRPC auth successful (API key)", map[string]interface{}{
			"method":  info.FullMethod,
			"user_id": user.ID.String(),
			"api_key": apiKeyObj.KeyPrefix,
		})

		return handler(ctx, req)
	}
}

// streamAPIKeyAuthInterceptor validates API key or application secret authentication for streaming gRPC requests.
func streamAPIKeyAuthInterceptor(apiKeyService *service.APIKeyService, appService *service.ApplicationService, log *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Allow gRPC reflection without authentication
		if strings.HasPrefix(info.FullMethod, "/grpc.reflection.") {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing API key or application secret: provide x-api-key metadata")
		}

		credential := extractCredential(md)
		if credential == "" {
			return status.Error(codes.Unauthenticated, "missing API key or application secret: provide x-api-key metadata")
		}

		// Application secret — full access
		if strings.HasPrefix(credential, "app_") {
			_, err := appService.ValidateSecret(ss.Context(), credential)
			if err != nil {
				log.Warn("gRPC stream auth failed: invalid application secret", map[string]interface{}{
					"method": info.FullMethod,
					"error":  err.Error(),
				})
				return status.Error(codes.Unauthenticated, "invalid application secret")
			}
			return handler(srv, ss)
		}

		// API key
		_, _, err := apiKeyService.ValidateAPIKey(ss.Context(), credential)
		if err != nil {
			log.Warn("gRPC stream auth failed: invalid API key", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return status.Error(codes.Unauthenticated, "invalid API key")
		}

		return handler(srv, ss)
	}
}

// containsScope checks if a slice of scopes contains the given scope
func containsScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}
