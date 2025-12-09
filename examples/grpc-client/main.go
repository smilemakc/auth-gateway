package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client messages (should match server proto)
type ValidateTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type ValidateTokenResponse struct {
	Valid        bool   `json:"valid"`
	UserId       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	ErrorMessage string `json:"error_message,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to Auth Gateway gRPC server")

	// Example 1: Validate Token
	fmt.Println("\n📝 Example 1: Validate Token")
	validateTokenExample(conn)

	// Example 2: Get User
	fmt.Println("\n📝 Example 2: Get User")
	getUserExample(conn)

	// Example 3: Check Permission
	fmt.Println("\n📝 Example 3: Check Permission")
	checkPermissionExample(conn)

	// Example 4: Introspect Token
	fmt.Println("\n📝 Example 4: Introspect Token")
	introspectTokenExample(conn)
}

func validateTokenExample(conn *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a placeholder token - replace with real token from signin
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.token"

	// Call ValidateToken method
	// In real implementation, use generated client from proto
	fmt.Printf("Validating token: %s...\n", token[:50]+"...")
	fmt.Println("⚠️  This is an example - use real token from /auth/signin")

	// Example response structure
	fmt.Println("\nExpected Response:")
	fmt.Println("  valid: true/false")
	fmt.Println("  user_id: uuid")
	fmt.Println("  email: user@example.com")
	fmt.Println("  username: johndoe")
	fmt.Println("  role: user")
	fmt.Println("  expires_at: unix timestamp")

	_ = ctx // avoid unused warning
}

func getUserExample(conn *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Example user ID
	userID := "123e4567-e89b-12d3-a456-426614174000"

	fmt.Printf("Getting user: %s\n", userID)
	fmt.Println("⚠️  This is an example - replace with real user ID")

	// Example response structure
	fmt.Println("\nExpected Response:")
	fmt.Println("  user:")
	fmt.Println("    id: uuid")
	fmt.Println("    email: user@example.com")
	fmt.Println("    username: johndoe")
	fmt.Println("    full_name: John Doe")
	fmt.Println("    role: user")
	fmt.Println("    email_verified: true/false")
	fmt.Println("    is_active: true/false")

	_ = ctx
}

func checkPermissionExample(conn *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Checking permission:")
	fmt.Println("  user_id: 123e4567-e89b-12d3-a456-426614174000")
	fmt.Println("  resource: orders")
	fmt.Println("  action: read")

	// Example response structure
	fmt.Println("\nExpected Response:")
	fmt.Println("  allowed: true/false")
	fmt.Println("  role: user/moderator/admin")

	_ = ctx
}

func introspectTokenExample(conn *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.token"

	fmt.Printf("Introspecting token: %s...\n", token[:50]+"...")
	fmt.Println("⚠️  This is an example - use real token from /auth/signin")

	// Example response structure
	fmt.Println("\nExpected Response:")
	fmt.Println("  active: true/false")
	fmt.Println("  user_id: uuid")
	fmt.Println("  email: user@example.com")
	fmt.Println("  username: johndoe")
	fmt.Println("  role: user")
	fmt.Println("  issued_at: unix timestamp")
	fmt.Println("  expires_at: unix timestamp")
	fmt.Println("  not_before: unix timestamp")
	fmt.Println("  blacklisted: true/false")

	_ = ctx
}
