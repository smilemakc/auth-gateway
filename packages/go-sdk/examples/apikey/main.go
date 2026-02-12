// Example: Using API key authentication with the Auth Gateway Go SDK
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

func main() {
	// Create a client with API key authentication
	// API keys are useful for server-to-server communication
	client := authgateway.NewClient(authgateway.Config{
		BaseURL: "http://localhost:8811",
		APIKey:  "agw_your_api_key_here", // Replace with your actual API key
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// With API key authentication, you can access endpoints based on the key's scopes

	// Example: Get profile (requires profile:read scope)
	fmt.Println("=== Get Profile with API Key ===")
	user, err := client.Profile.Get(ctx)
	if err != nil {
		if apiErr, ok := err.(*authgateway.APIError); ok {
			log.Printf("API Error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
		} else {
			log.Printf("Error: %v\n", err)
		}
		return
	}
	fmt.Printf("Profile: %s (%s)\n", user.FullName, user.Email)

	// Example: List API keys (requires api_keys:read scope)
	fmt.Println("\n=== List API Keys ===")
	keys, err := client.APIKeys.List(ctx)
	if err != nil {
		log.Printf("Failed to list API keys: %v", err)
		return
	}
	for _, key := range keys {
		active := "active"
		if !key.IsActive {
			active = "revoked"
		}
		fmt.Printf("- %s: %s (%s)\n", key.KeyPrefix, key.Name, active)
	}

	// Example: Health check (no authentication required)
	fmt.Println("\n=== Health Check ===")
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	fmt.Printf("Status: %s, Database: %s, Redis: %s\n", health.Status, health.Database, health.Redis)
}
