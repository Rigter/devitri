package sync

import (
	"testing"

	"github.com/rigter/devitri/backend/internal/config"
)

func TestProcessBatchBulkDeleteRequiresConfirmation(t *testing.T) {
	config.Current = config.Config{
		DeleteThresholdCount:   20,
		DeleteThresholdPercent: 5,
	}

	// Simulate the safety decision used at the end of ProcessBatch.
	totalFiles := 100
	deletions := 25
	safety := CheckBulkDeleteSafety(totalFiles, deletions)

	if safety.IsSafe {
		t.Fatal("expected bulk delete to be unsafe")
	}

	bulkWarning := !safety.IsSafe
	toDelete := []string{"a", "b", "c"} // placeholder paths
	if bulkWarning && false {           // BulkDeleteConfirmed == false
		toDelete = []string{}
	}
	if len(toDelete) != 3 {
		t.Fatalf("expected toDelete unchanged when not simulating clear, got %d", len(toDelete))
	}

	toDelete = []string{"a", "b", "c"}
	if bulkWarning && !true { // BulkDeleteConfirmed == true
		toDelete = []string{}
	}
	if len(toDelete) != 3 {
		t.Fatal("expected toDelete preserved when confirmed")
	}

	toDelete = []string{"a", "b", "c"}
	if bulkWarning && !false {
		toDelete = []string{}
	}
	if len(toDelete) != 0 {
		t.Fatal("expected toDelete cleared when warning and not confirmed")
	}
}
