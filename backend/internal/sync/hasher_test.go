package sync

import (
	"strings"
	"testing"
)

func TestCalculateSHA256FromReader(t *testing.T) {
	// Test with a known string
	input := "This is a test file for hashing."
	expectedHash := "d8344f8ca8e5f17d7a2e2e48ffc35edb2bfecec976038096172bbcd169c1d48d" // SHA256 of input
	
	hash, err := CalculateSHA256FromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to calculate hash: %v", err)
	}
	
	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}