package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockRateLimiter struct {
	count int64
	err   error
}

func (m *mockRateLimiter) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.count, nil
}

func TestRateLimitInterceptor_AllowsUnderLimit(t *testing.T) {
	limiter := &mockRateLimiter{count: 1}
	interceptor := rateLimitInterceptor(limiter, 100)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", "test-key"))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Test"}, handler)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if resp != "success" {
		t.Errorf("expected success, got %v", resp)
	}
}

func TestRateLimitInterceptor_BlocksOverLimit(t *testing.T) {
	limiter := &mockRateLimiter{count: 101}
	interceptor := rateLimitInterceptor(limiter, 100)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", "test-key"))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Test"}, handler)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %v", resp)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("expected ResourceExhausted code, got %v", st.Code())
	}
}

func TestRateLimitInterceptor_NoAuthKey(t *testing.T) {
	limiter := &mockRateLimiter{count: 101}
	interceptor := rateLimitInterceptor(limiter, 100)

	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Test"}, handler)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if resp != "success" {
		t.Errorf("expected success, got %v", resp)
	}
}

func TestRateLimitInterceptor_RedisError(t *testing.T) {
	limiter := &mockRateLimiter{err: errors.New("redis connection failed")}
	interceptor := rateLimitInterceptor(limiter, 100)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", "test-key"))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Test"}, handler)

	if err != nil {
		t.Errorf("expected no error (fail open), got %v", err)
	}
	if resp != "success" {
		t.Errorf("expected success, got %v", resp)
	}
}
