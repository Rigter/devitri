package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "test-password"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}
	
	if hash == password {
		t.Error("HashPassword returned plain text password")
	}
}

func TestComparePassword(t *testing.T) {
	password := "test-password"
	wrongPassword := "wrong-password"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	// Test correct password
	if !ComparePassword(password, hash) {
		t.Error("ComparePassword failed to validate correct password")
	}
	
	// Test wrong password
	if ComparePassword(wrongPassword, hash) {
		t.Error("ComparePassword incorrectly validated wrong password")
	}
}