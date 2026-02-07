package authgateway

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
)

const (
	contextKeyUserID         = "user_id"
	contextKeyEmail          = "email"
	contextKeyUsername       = "username"
	contextKeyRoles          = "roles"
	contextKeyAppRoles       = "app_roles"
	contextKeyApplicationID  = "application_id"
	contextKeyAuthValidation = "auth_validation"
)

type TokenExtractor func(r *http.Request) string

type MiddlewareConfig struct {
	TokenExtractors []TokenExtractor
	SkipPaths       []string
	OnError         func(*gin.Context, error)
	CacheEnabled    bool
	CacheTTL        time.Duration
}

// GinMiddleware returns a Gin handler that validates JWT tokens via gRPC
// and populates the request context with user information.
func (c *GRPCClient) GinMiddleware(opts ...MiddlewareOption) gin.HandlerFunc {
	cfg := defaultMiddlewareConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	var cache *validationCache
	if cfg.CacheEnabled {
		cache = newValidationCache(cfg.CacheTTL)
	}

	skipSet := buildSkipPathSet(cfg.SkipPaths)

	return func(ctx *gin.Context) {
		if _, shouldSkip := skipSet[ctx.Request.URL.Path]; shouldSkip {
			ctx.Next()
			return
		}

		token := extractToken(ctx.Request, cfg.TokenExtractors)
		if token == "" {
			handleMiddlewareError(ctx, cfg, &APIError{
				StatusCode: http.StatusUnauthorized,
				Code:       "MISSING_TOKEN",
				Message:    "Missing authorization token",
			})
			return
		}

		if cached := getFromCache(cache, token); cached != nil {
			setContextFromValidation(ctx, cached)
			ctx.Next()
			return
		}

		validation, err := c.ValidateToken(ctx.Request.Context(), token)
		if err != nil {
			handleMiddlewareError(ctx, cfg, err)
			return
		}

		if !validation.GetValid() {
			handleMiddlewareError(ctx, cfg, &APIError{
				StatusCode: http.StatusUnauthorized,
				Code:       "INVALID_TOKEN",
				Message:    "Invalid or expired token",
			})
			return
		}

		setToCache(cache, token, validation)
		setContextFromValidation(ctx, validation)
		ctx.Next()
	}
}

// RequireRole returns middleware that checks whether the authenticated user
// holds at least one of the specified global roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	required := toSet(roles)
	return func(ctx *gin.Context) {
		if !hasAnyFromContext(ctx, contextKeyRoles, required) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "Insufficient role",
			})
			return
		}
		ctx.Next()
	}
}

// RequireAppRole returns middleware that checks whether the authenticated user
// holds at least one of the specified application-scoped roles.
func RequireAppRole(roles ...string) gin.HandlerFunc {
	required := toSet(roles)
	return func(ctx *gin.Context) {
		if !hasAnyFromContext(ctx, contextKeyAppRoles, required) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "Insufficient application role",
			})
			return
		}
		ctx.Next()
	}
}

// RequirePermission returns middleware that checks a specific resource/action
// permission for the authenticated user via gRPC.
func (c *GRPCClient) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := GetUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "User not authenticated",
			})
			return
		}

		allowed, err := c.HasPermission(ctx.Request.Context(), userID, resource, action)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "PERMISSION_CHECK_FAILED",
				"message": "Failed to check permission",
			})
			return
		}

		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "Permission denied for " + resource + ":" + action,
			})
			return
		}

		ctx.Next()
	}
}

// --- Context accessors ---

func GetUserID(ctx *gin.Context) string {
	return stringFromContext(ctx, contextKeyUserID)
}

func GetEmail(ctx *gin.Context) string {
	return stringFromContext(ctx, contextKeyEmail)
}

func GetUsername(ctx *gin.Context) string {
	return stringFromContext(ctx, contextKeyUsername)
}

func GetRoles(ctx *gin.Context) []string {
	return stringSliceFromContext(ctx, contextKeyRoles)
}

func GetAppRoles(ctx *gin.Context) []string {
	return stringSliceFromContext(ctx, contextKeyAppRoles)
}

func GetApplicationID(ctx *gin.Context) string {
	return stringFromContext(ctx, contextKeyApplicationID)
}

func HasRole(ctx *gin.Context, role string) bool {
	return containsString(GetRoles(ctx), role)
}

func HasAppRole(ctx *gin.Context, role string) bool {
	return containsString(GetAppRoles(ctx), role)
}

// --- Token extractors ---

func BearerTokenExtractor() TokenExtractor {
	return func(r *http.Request) string {
		header := r.Header.Get("Authorization")
		if len(header) > 7 && strings.EqualFold(header[:7], "bearer ") {
			return header[7:]
		}
		return ""
	}
}

func CookieTokenExtractor(name string) TokenExtractor {
	return func(r *http.Request) string {
		cookie, err := r.Cookie(name)
		if err != nil {
			return ""
		}
		return cookie.Value
	}
}

func QueryTokenExtractor(param string) TokenExtractor {
	return func(r *http.Request) string {
		return r.URL.Query().Get(param)
	}
}

func HeaderTokenExtractor(header string) TokenExtractor {
	return func(r *http.Request) string {
		return r.Header.Get(header)
	}
}

// --- Internal helpers ---

func defaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		TokenExtractors: []TokenExtractor{BearerTokenExtractor()},
		CacheTTL:        30 * time.Second,
	}
}

func buildSkipPathSet(paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return set
}

func extractToken(r *http.Request, extractors []TokenExtractor) string {
	for _, extractor := range extractors {
		if token := extractor(r); token != "" {
			return token
		}
	}
	return ""
}

func handleMiddlewareError(ctx *gin.Context, cfg MiddlewareConfig, err error) {
	if cfg.OnError != nil {
		cfg.OnError(ctx, err)
		return
	}

	statusCode := http.StatusUnauthorized
	code := "UNAUTHORIZED"
	message := err.Error()

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 0 {
			statusCode = apiErr.StatusCode
		}
		code = apiErr.Code
		message = apiErr.Message
	}

	ctx.AbortWithStatusJSON(statusCode, gin.H{
		"error":   code,
		"message": message,
	})
}

func setContextFromValidation(ctx *gin.Context, resp *proto.ValidateTokenResponse) {
	ctx.Set(contextKeyUserID, resp.GetUserId())
	ctx.Set(contextKeyEmail, resp.GetEmail())
	ctx.Set(contextKeyUsername, resp.GetUsername())
	ctx.Set(contextKeyRoles, resp.GetRoles())
	ctx.Set(contextKeyAppRoles, resp.GetAppRoles())
	ctx.Set(contextKeyApplicationID, resp.GetApplicationId())
	ctx.Set(contextKeyAuthValidation, resp)
}

func stringFromContext(ctx *gin.Context, key string) string {
	val, exists := ctx.Get(key)
	if !exists {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

func stringSliceFromContext(ctx *gin.Context, key string) []string {
	val, exists := ctx.Get(key)
	if !exists {
		return nil
	}
	slice, ok := val.([]string)
	if !ok {
		return nil
	}
	return slice
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return set
}

func hasAnyFromContext(ctx *gin.Context, key string, required map[string]struct{}) bool {
	values := stringSliceFromContext(ctx, key)
	for _, v := range values {
		if _, found := required[v]; found {
			return true
		}
	}
	return false
}

func getFromCache(cache *validationCache, token string) *proto.ValidateTokenResponse {
	if cache == nil {
		return nil
	}
	return cache.get(token)
}

func setToCache(cache *validationCache, token string, resp *proto.ValidateTokenResponse) {
	if cache == nil {
		return
	}
	cache.set(token, resp)
}

// --- Validation cache ---

type cachedValidation struct {
	resp      *proto.ValidateTokenResponse
	expiresAt time.Time
}

type validationCache struct {
	mu    sync.RWMutex
	items map[string]*cachedValidation
	ttl   time.Duration
}

func newValidationCache(ttl time.Duration) *validationCache {
	return &validationCache{
		items: make(map[string]*cachedValidation),
		ttl:   ttl,
	}
}

func (vc *validationCache) get(token string) *proto.ValidateTokenResponse {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	entry, exists := vc.items[token]
	if !exists {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.resp
}

func (vc *validationCache) set(token string, resp *proto.ValidateTokenResponse) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.items[token] = &cachedValidation{
		resp:      resp,
		expiresAt: time.Now().Add(vc.ttl),
	}
}
