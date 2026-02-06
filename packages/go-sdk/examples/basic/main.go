// Example: Basic authentication flow with the Auth Gateway Go SDK
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

func main() {
	// Create a new client
	client := authgateway.NewClient(authgateway.Config{
		BaseURL:     "http://localhost:8811",
		Timeout:     30 * time.Second,
		AutoRefresh: true, // Automatically refresh tokens when they expire
	})

	ctx := context.Background()

	// Example 1: Sign up a new user
	fmt.Println("=== Sign Up ===")
	authResp, err := client.Auth.SignUp(ctx, &models.SignUpRequest{
		Email:    "user@example.com",
		Username: "testuser",
		Password: "securepassword123",
		FullName: "Test User",
	})
	if err != nil {
		log.Printf("Sign up failed: %v", err)
	} else {
		fmt.Printf("User signed up: %s\n", authResp.User.Email)
	}

	// Example 2: Sign in
	fmt.Println("\n=== Sign In ===")
	authResp, err = client.Auth.SignInWithEmail(ctx, "user@example.com", "securepassword123")
	if err != nil {
		// Check if 2FA is required
		if tfaErr, ok := err.(*authgateway.TwoFactorRequiredError); ok {
			fmt.Printf("2FA required. Token: %s\n", tfaErr.TwoFactorToken)
			// Handle 2FA verification:
			// authResp, err = client.Auth.Verify2FA(ctx, tfaErr.TwoFactorToken, "123456")
		} else {
			log.Printf("Sign in failed: %v", err)
		}
	} else {
		fmt.Printf("Signed in successfully! Token expires in: %d seconds\n", authResp.ExpiresIn)
	}

	// Example 3: Get profile
	fmt.Println("\n=== Get Profile ===")
	user, err := client.Profile.Get(ctx)
	if err != nil {
		log.Printf("Failed to get profile: %v", err)
	} else {
		fmt.Printf("Profile: %s (%s)\n", user.FullName, user.Email)
	}

	// Example 4: Update profile
	fmt.Println("\n=== Update Profile ===")
	user, err = client.Profile.Update(ctx, &models.UpdateProfileRequest{
		FullName: "Updated Name",
	})
	if err != nil {
		log.Printf("Failed to update profile: %v", err)
	} else {
		fmt.Printf("Updated profile: %s\n", user.FullName)
	}

	// Example 5: List sessions
	fmt.Println("\n=== List Sessions ===")
	sessions, err := client.Sessions.List(ctx)
	if err != nil {
		log.Printf("Failed to list sessions: %v", err)
	} else {
		for _, session := range sessions {
			current := ""
			if session.IsCurrent {
				current = " (current)"
			}
			fmt.Printf("- %s: %s %s%s\n", session.ID[:8], session.DeviceType, session.Browser, current)
		}
	}

	// Example 6: Create API key
	fmt.Println("\n=== Create API Key ===")
	apiKeyResp, err := client.APIKeys.Create(ctx, &models.CreateAPIKeyRequest{
		Name:        "My API Key",
		Description: "For testing",
		Scopes:      []string{"profile:read", "users:read"},
	})
	if err != nil {
		log.Printf("Failed to create API key: %v", err)
	} else {
		fmt.Printf("API Key created! Key: %s (save this, it won't be shown again)\n", apiKeyResp.PlainKey)
	}

	// Example 7: Logout
	fmt.Println("\n=== Logout ===")
	err = client.Auth.Logout(ctx)
	if err != nil {
		log.Printf("Logout failed: %v", err)
	} else {
		fmt.Println("Logged out successfully")
	}
}
