package service

import (
	"context"
	"testing"
)

func TestPasswordChecker_Disabled(t *testing.T) {
	pc := NewPasswordChecker(false)
	compromised, count := pc.IsCompromised(context.Background(), "password123")
	if compromised {
		t.Error("disabled checker should never return compromised")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestPasswordChecker_HashFormat(t *testing.T) {
	pc := NewPasswordChecker(true)
	_, _ = pc.IsCompromised(context.Background(), "")
}
