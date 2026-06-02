package sync

import (
	"testing"
)

func TestBulkDeleteCheckResult(t *testing.T) {
	result := &BulkDeleteCheckResult{
		IsSafe:     true,
		Reason:     "Test reason",
		FileCount:  5,
		Threshold:  20,
		Percentage: 2.5,
	}
	
	if !result.IsSafe {
		t.Error("Expected IsSafe to be true")
	}
	
	if result.Reason != "Test reason" {
		t.Errorf("Expected reason 'Test reason', got '%s'", result.Reason)
	}
	
	if result.FileCount != 5 {
		t.Errorf("Expected FileCount 5, got %d", result.FileCount)
	}
	
	if result.Threshold != 20 {
		t.Errorf("Expected Threshold 20, got %d", result.Threshold)
	}
	
	if result.Percentage != 2.5 {
		t.Errorf("Expected Percentage 2.5, got %f", result.Percentage)
	}
}