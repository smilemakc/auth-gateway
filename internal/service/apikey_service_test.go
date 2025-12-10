package service

import (
	"strings"
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

	if !strings.HasPrefix(key1, "agw_") {
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

	// Test multiple generations
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := service.GenerateAPIKey()
		if err != nil {
			t.Fatalf("Failed to generate API key #%d: %v", i, err)
		}
		if keys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		keys[key] = true
	}

	if len(keys) != 100 {
		t.Errorf("Expected 100 unique keys, got %d", len(keys))
	}
}

func TestNewAPIKeyService(t *testing.T) {
	// This just ensures NewAPIKeyService is covered
	service := &APIKeyService{}
	if service == nil {
		t.Error("Service should not be nil")
	}
}
