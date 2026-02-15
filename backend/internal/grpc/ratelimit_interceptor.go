package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type RateLimiter interface {
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
}

func rateLimitInterceptor(limiter RateLimiter, maxPerMinute int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		key := extractAuthKey(ctx)
		if key == "" {
			return handler(ctx, req)
		}

		rateKey := fmt.Sprintf("grpc_ratelimit:%s", key)
		count, err := limiter.IncrementRateLimit(ctx, rateKey, time.Minute)
		if err != nil {
			return handler(ctx, req)
		}

		if count > int64(maxPerMinute) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %d requests per minute", maxPerMinute)
		}

		return handler(ctx, req)
	}
}

func extractAuthKey(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if keys := md.Get("x-api-key"); len(keys) > 0 {
		return keys[0]
	}
	if keys := md.Get("authorization"); len(keys) > 0 {
		return keys[0]
	}
	return ""
}
