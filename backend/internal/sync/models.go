package sync

// FileMetadata represents metadata for a file
type FileMetadata struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
}

// SyncBatchRequest represents a batch sync request from the client
type SyncBatchRequest struct {
	DeviceID            string         `json:"device_id"`
	Files               []FileMetadata `json:"files"`
	BulkDeleteConfirmed bool           `json:"bulk_delete_confirmed,omitempty"`
}

// SyncBatchResponse represents the server's response to a batch sync request
type SyncBatchResponse struct {
	ToUpload          []string `json:"to_upload"`
	ToDownload        []string `json:"to_download"`
	Conflicts         []string `json:"conflicts"`
	ToDelete          []string `json:"to_delete"`
	BulkDeleteWarning bool     `json:"bulk_delete_warning"`
}

// ManifestResponse represents the server's file manifest
type ManifestResponse struct {
	VaultID     string         `json:"vault_id"`
	GeneratedAt int64          `json:"generated_at"`
	Files       []FileMetadata `json:"files"`
}