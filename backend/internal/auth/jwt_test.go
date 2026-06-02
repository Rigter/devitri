package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	secret := "test-secret"
	duration := 1 * time.Hour
	
	manager := NewJWTManager(secret, duration)
	
	// Test token generation
	token, err := manager.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if token == "" {
		t.Error("Generate returned empty token")
	}
	
	// Test token verification
	verifiedToken, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	
	if verifiedToken == nil {
		t.Error("Verify returned nil token")
	}
}

func TestJWTManager_VerifyInvalidToken(t *testing.T) {
	secret := "test-secret"
	duration := 1 * time.Hour
	
	manager := NewJWTManager(secret, duration)
	
	// Test verification of invalid token
	_, err := manager.Verify("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTManager_VerifyExpiredToken(t *testing.T) {
	secret := "test-secret"
	// Very short duration to ensure token expires
	duration := 1 * time.Millisecond
	
	manager := NewJWTManager(secret, duration)
	
	// Generate token
	token, err := manager.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Wait for token to expire
	time.Sleep(2 * time.Millisecond)
	
	// Test verification of expired token
	_, err = manager.Verify(token)
	if err != ErrExpiredToken && err != ErrInvalidToken {
		t.Errorf("Expected ErrExpiredToken or ErrInvalidToken, got %v", err)
	}
}