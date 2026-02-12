package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/smilemakc/auth-gateway/pkg/keys"
)

func main() {
	fmt.Println("=== OIDC Key Manager Example ===\n")

	keyDir := os.Getenv("KEY_DIR")
	if keyDir == "" {
		keyDir = "./keys"
	}

	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		keyID = "key-20251215"
	}

	fmt.Printf("Using keys from: %s\n", keyDir)
	fmt.Printf("Key ID: %s\n\n", keyID)

	configs := []keys.KeyConfig{
		{
			ID:             keyID,
			Algorithm:      keys.RS256,
			PrivateKeyPath: fmt.Sprintf("%s/rsa_private_%s.pem", keyDir, keyID),
		},
	}

	manager, err := keys.NewManager(configs, keyID)
	if err != nil {
		log.Fatalf("Failed to create key manager: %v", err)
	}

	fmt.Println("✓ Key manager initialized successfully\n")

	currentKey, err := manager.GetCurrentKey()
	if err != nil {
		log.Fatalf("Failed to get current key: %v", err)
	}

	fmt.Printf("Current signing key:\n")
	fmt.Printf("  - Key ID: %s\n", currentKey.KID)
	fmt.Printf("  - Algorithm: %s\n\n", currentKey.Algorithm)

	data := []byte("This is a sample JWT payload")
	fmt.Printf("Data to sign: %s\n\n", string(data))

	signature, kid, err := manager.Sign(data)
	if err != nil {
		log.Fatalf("Failed to sign data: %v", err)
	}

	fmt.Printf("✓ Data signed successfully\n")
	fmt.Printf("  - Key ID used: %s\n", kid)
	fmt.Printf("  - Signature length: %d bytes\n\n", len(signature))

	err = manager.Verify(data, signature, kid)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}

	fmt.Println("✓ Signature verified successfully\n")

	jwks := manager.GetJWKS()
	fmt.Println("JWKS (JSON Web Key Set):")
	jwksJSON, err := json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JWKS: %v", err)
	}

	fmt.Println(string(jwksJSON))
	fmt.Println()

	fmt.Println("=== Example completed successfully ===")
	fmt.Println("\nUsage with multiple keys for rotation:")
	fmt.Println("  configs := []keys.KeyConfig{")
	fmt.Println("    {ID: \"key-new\", Algorithm: keys.RS256, PrivateKeyPath: \"./keys/rsa_private_key-new.pem\"},")
	fmt.Println("    {ID: \"key-old\", Algorithm: keys.RS256, PrivateKeyPath: \"./keys/rsa_private_key-old.pem\"},")
	fmt.Println("  }")
	fmt.Println("  manager, _ := keys.NewManager(configs, \"key-new\") // Signs with key-new, verifies with both")
}
