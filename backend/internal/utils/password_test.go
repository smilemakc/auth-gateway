package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	err = CheckPassword(hash, password)
	if err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}

	// Test incorrect password
	err = CheckPassword(hash, "wrongPassword")
	if err == nil {
		t.Error("CheckPassword should fail for incorrect password")
	}
}

func TestIsPasswordValid(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"Valid password", "testPassword123", true},
		{"Valid minimum length", "abcdefgh", true},
		{"Too short", "abcdefg", false},
		{"Empty", "", false},
		{"Only digits no lowercase", "12345678", false},
		{"Mixed valid", "pass1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPasswordValid(tt.password)
			if got != tt.want {
				t.Errorf("IsPasswordValid(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}
