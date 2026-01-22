// Example: Using the gRPC client for server-to-server communication
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
)

func main() {
	// Create a gRPC client
	// The gRPC API is ideal for microservice-to-microservice communication
	grpcClient, err := authgateway.NewGRPCClient(authgateway.GRPCConfig{
		Address:     "localhost:50051",
		Insecure:    true, // Use true for development, false for production with TLS
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

	ctx := context.Background()

	// Example JWT token (replace with a real token)
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

	// Example 1: Validate a token
	fmt.Println("=== Validate Token ===")
	validateResp, err := grpcClient.ValidateToken(ctx, accessToken)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
	} else {
		if validateResp.Valid {
			fmt.Printf("Token is valid for user: %s (%s)\n", validateResp.Username, validateResp.Email)
			fmt.Printf("Roles: %v\n", validateResp.Roles)
			fmt.Printf("Expires at: %d\n", validateResp.ExpiresAt)
		} else {
			fmt.Println("Token is invalid")
		}
	}

	// Example 2: Get user by ID
	fmt.Println("\n=== Get User ===")
	userID := "550e8400-e29b-41d4-a716-446655440000" // Replace with real user ID
	user, err := grpcClient.GetUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
	} else {
		fmt.Printf("User: %s (%s)\n", user.FullName, user.Email)
		fmt.Printf("Active: %t, Email Verified: %t\n", user.IsActive, user.EmailVerified)
	}

	// Example 3: Check permission
	fmt.Println("\n=== Check Permission ===")
	hasPermission, err := grpcClient.HasPermission(ctx, userID, "users", "read")
	if err != nil {
		log.Printf("Permission check failed: %v", err)
	} else {
		fmt.Printf("User has 'users:read' permission: %t\n", hasPermission)
	}

	// Example 4: Introspect token
	fmt.Println("\n=== Introspect Token ===")
	introspectResp, err := grpcClient.IntrospectToken(ctx, accessToken)
	if err != nil {
		log.Printf("Token introspection failed: %v", err)
	} else {
		fmt.Printf("Token active: %t\n", introspectResp.Active)
		fmt.Printf("Issued at: %d, Expires at: %d\n", introspectResp.IssuedAt, introspectResp.ExpiresAt)
		fmt.Printf("Blacklisted: %t\n", introspectResp.Blacklisted)
	}

	// Example 5: Create user via gRPC
	fmt.Println("\n=== Create User via gRPC ===")
	createResp, err := grpcClient.CreateUser(ctx, &proto.CreateUserRequest{
		Email:       "newuser@example.com",
		Username:    "newuser",
		Password:    "securepassword123",
		FullName:    "New User",
		AccountType: "human",
	})
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		fmt.Printf("User created: %s (%s)\n", createResp.User.Username, createResp.User.Email)
		fmt.Printf("Access token: %s...\n", createResp.AccessToken[:50])
	}

	// Example 6: Login with email and password
	fmt.Println("\n=== Login via gRPC ===")
	loginResp, err := grpcClient.LoginWithEmail(ctx, "user@example.com", "password123")
	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		fmt.Printf("Logged in as: %s (%s)\n", loginResp.User.Username, loginResp.User.Email)
		fmt.Printf("Access token expires in: %d seconds\n", loginResp.ExpiresIn)
	}

	// Example 7: OTP-based passwordless login
	fmt.Println("\n=== OTP Login Flow ===")
	// Step 1: Request OTP
	otpLoginResp, err := grpcClient.LoginWithOTP(ctx, &proto.LoginWithOTPRequest{
		Email: "user@example.com",
	})
	if err != nil {
		log.Printf("OTP login request failed: %v", err)
	} else {
		fmt.Printf("OTP sent: %s (expires in %d seconds)\n", otpLoginResp.Message, otpLoginResp.ExpiresIn)

		// Step 2: Verify OTP (in real app, user would provide this)
		verifyResp, err := grpcClient.VerifyLoginOTP(ctx, &proto.VerifyLoginOTPRequest{
			Email: "user@example.com",
			Code:  "123456", // Replace with actual OTP
		})
		if err != nil {
			log.Printf("OTP verification failed: %v", err)
		} else {
			fmt.Printf("OTP login successful for: %s\n", verifyResp.User.Email)
		}
	}

	// Example 8: Send and verify generic OTP
	fmt.Println("\n=== Send/Verify OTP ===")
	sendOTPResp, err := grpcClient.SendOTP(ctx, &proto.SendOTPRequest{
		Email:   "user@example.com",
		OtpType: proto.OTPType_OTP_TYPE_VERIFICATION,
	})
	if err != nil {
		log.Printf("Failed to send OTP: %v", err)
	} else {
		fmt.Printf("OTP sent: %s (expires in %d seconds)\n", sendOTPResp.Message, sendOTPResp.ExpiresIn)
	}

	// Example 9: OAuth token introspection (RFC 7662)
	fmt.Println("\n=== OAuth Token Introspection ===")
	oauthToken := "oauth_access_token_here"
	oauthIntrospect, err := grpcClient.IntrospectOAuthToken(ctx, oauthToken, "access_token")
	if err != nil {
		log.Printf("OAuth introspection failed: %v", err)
	} else {
		fmt.Printf("Token active: %t\n", oauthIntrospect.Active)
		fmt.Printf("Client ID: %s, Scope: %s\n", oauthIntrospect.ClientId, oauthIntrospect.Scope)
		fmt.Printf("Subject: %s, Expires: %d\n", oauthIntrospect.Sub, oauthIntrospect.Exp)
	}

	// Example 10: Validate OAuth client credentials
	fmt.Println("\n=== Validate OAuth Client ===")
	clientValidation, err := grpcClient.ValidateOAuthClient(ctx, "my_client_id", "my_client_secret")
	if err != nil {
		log.Printf("Client validation failed: %v", err)
	} else {
		fmt.Printf("Client valid: %t\n", clientValidation.Valid)
		fmt.Printf("Client name: %s, Type: %s\n", clientValidation.ClientName, clientValidation.ClientType)
		fmt.Printf("Scopes: %v\n", clientValidation.Scopes)
	}

	// Example 11: Get OAuth client info
	fmt.Println("\n=== Get OAuth Client ===")
	oauthClient, err := grpcClient.GetOAuthClient(ctx, "my_client_id")
	if err != nil {
		log.Printf("Failed to get OAuth client: %v", err)
	} else {
		fmt.Printf("Client: %s (%s)\n", oauthClient.ClientName, oauthClient.ClientId)
		fmt.Printf("Type: %s, Active: %t\n", oauthClient.ClientType, oauthClient.IsActive)
		fmt.Printf("Grant types: %v\n", oauthClient.GrantTypes)
	}
}
