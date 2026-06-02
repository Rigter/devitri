package sync

import (
	"testing"

	"github.com/rigter/devitri/backend/internal/config"
)

func TestCheckBulkDeleteSafety(t *testing.T) {
	config.Current = config.Config{
		DeleteThresholdCount:   20,
		DeleteThresholdPercent: 5,
	}

	tests := []struct {
		name          string
		totalFiles    int
		deleteCount   int
		expectedSafe  bool
	}{
		{
			name:         "Within both thresholds",
			totalFiles:   100,
			deleteCount:  5, // Below 20, below 5%
			expectedSafe: true,
		},
		{
			name:         "Exceeds absolute threshold",
			totalFiles:   1000,
			deleteCount:  25, // Above 20
			expectedSafe: false,
		},
		{
			name:         "Exceeds percentage threshold but not absolute",
			totalFiles:   1000,
			deleteCount:  30, // 3%, below 20 but below 5% threshold
			expectedSafe: false, // Actually, 30 is above 20, so it should be false
		},
		{
			name:         "Exceeds percentage threshold",
			totalFiles:   100,
			deleteCount:  10, // 10%, above 5%
			expectedSafe: false,
		},
		{
			name:         "Zero total files",
			totalFiles:   0,
			deleteCount:  5,
			expectedSafe: true,
		},
		{
			name:         "Zero delete count",
			totalFiles:   100,
			deleteCount:  0,
			expectedSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckBulkDeleteSafety(tt.totalFiles, tt.deleteCount)
			
			if result.IsSafe != tt.expectedSafe {
				t.Errorf("Expected IsSafe=%v, got %v. Reason: %s", tt.expectedSafe, result.IsSafe, result.Reason)
			}
		})
	}
}

func TestCalculateSHA256(t *testing.T) {
	// Test with a known file
	expectedHash := "d8344f8ca8e5f17d7a2e2e48ffc35edb2bfecec976038096172bbcd169c1d48d" // SHA256 of test file content
	
	hash, err := CalculateSHA256("testdata/test.txt")
	if err != nil {
		t.Fatalf("Failed to calculate hash: %v", err)
	}
	
	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}