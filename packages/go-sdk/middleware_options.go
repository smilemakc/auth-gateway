package authgateway

import (
	"time"

	"github.com/gin-gonic/gin"
)

type MiddlewareOption func(*MiddlewareConfig)

func WithTokenExtractors(extractors ...TokenExtractor) MiddlewareOption {
	return func(cfg *MiddlewareConfig) {
		cfg.TokenExtractors = extractors
	}
}

func WithSkipPaths(paths ...string) MiddlewareOption {
	return func(cfg *MiddlewareConfig) {
		cfg.SkipPaths = paths
	}
}

func WithErrorHandler(handler func(*gin.Context, error)) MiddlewareOption {
	return func(cfg *MiddlewareConfig) {
		cfg.OnError = handler
	}
}

func WithCache(ttl time.Duration) MiddlewareOption {
	return func(cfg *MiddlewareConfig) {
		cfg.CacheEnabled = true
		cfg.CacheTTL = ttl
	}
}

func WithoutCache() MiddlewareOption {
	return func(cfg *MiddlewareConfig) {
		cfg.CacheEnabled = false
	}
}
