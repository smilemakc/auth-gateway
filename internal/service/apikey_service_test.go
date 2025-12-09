package service

import (
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	service := &APIKeyService{}

	// Test key generation
	key1, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if key1 == "" {
		t.Error("Generated key should not be empty")
	}

	if len(key1) < 20 {
		t.Error("Generated key is too short")
	}

	if key1[:4] != "agw_" {
		t.Errorf("Key should start with 'agw_', got: %s", key1[:4])
	}

	// Test uniqueness
	key2, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate second API key: %v", err)
	}

	if key1 == key2 {
		t.Error("Generated keys should be unique")
	}
}
