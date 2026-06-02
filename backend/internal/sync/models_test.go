package sync

import (
	"encoding/json"
	"testing"
)

func TestFileMetadataJSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling of FileMetadata
	original := FileMetadata{
		Path:       "test/file.md",
		Hash:       "abc123",
		Size:       1024,
		ModifiedAt: 1234567890,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal FileMetadata: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled FileMetadata
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal FileMetadata: %v", err)
	}

	// Check that all fields match
	if original.Path != unmarshaled.Path {
		t.Errorf("Path mismatch: expected %s, got %s", original.Path, unmarshaled.Path)
	}
	
	if original.Hash != unmarshaled.Hash {
		t.Errorf("Hash mismatch: expected %s, got %s", original.Hash, unmarshaled.Hash)
	}
	
	if original.Size != unmarshaled.Size {
		t.Errorf("Size mismatch: expected %d, got %d", original.Size, unmarshaled.Size)
	}
	
	if original.ModifiedAt != unmarshaled.ModifiedAt {
		t.Errorf("ModifiedAt mismatch: expected %d, got %d", original.ModifiedAt, unmarshaled.ModifiedAt)
	}
}

func TestSyncBatchResponseJSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling of SyncBatchResponse
	original := SyncBatchResponse{
		ToUpload:          []string{"file1.md", "file2.md"},
		ToDownload:        []string{"file3.md"},
		Conflicts:         []string{"conflict.md"},
		ToDelete:          []string{"delete.md"},
		BulkDeleteWarning: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal SyncBatchResponse: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled SyncBatchResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SyncBatchResponse: %v", err)
	}

	// Check that all fields match
	if len(original.ToUpload) != len(unmarshaled.ToUpload) {
		t.Errorf("ToUpload length mismatch: expected %d, got %d", len(original.ToUpload), len(unmarshaled.ToUpload))
	}
	
	if len(original.ToDownload) != len(unmarshaled.ToDownload) {
		t.Errorf("ToDownload length mismatch: expected %d, got %d", len(original.ToDownload), len(unmarshaled.ToDownload))
	}
	
	if len(original.Conflicts) != len(unmarshaled.Conflicts) {
		t.Errorf("Conflicts length mismatch: expected %d, got %d", len(original.Conflicts), len(unmarshaled.Conflicts))
	}
	
	if len(original.ToDelete) != len(unmarshaled.ToDelete) {
		t.Errorf("ToDelete length mismatch: expected %d, got %d", len(original.ToDelete), len(unmarshaled.ToDelete))
	}
	
	if original.BulkDeleteWarning != unmarshaled.BulkDeleteWarning {
		t.Errorf("BulkDeleteWarning mismatch: expected %v, got %v", original.BulkDeleteWarning, unmarshaled.BulkDeleteWarning)
	}
}