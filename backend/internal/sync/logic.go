package sync

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rigter/devitri/backend/internal/vault"
)

// FileRecord represents a file record from the database
type FileRecord struct {
	ID         int64  `json:"id"`
	VaultID    string `json:"vault_id"`
	Path       string `json:"path"`
	HashSHA256 string `json:"hash_sha256"`
	SizeBytes  int64  `json:"size_bytes"`
	ModifiedAt int64  `json:"modified_at"`
	DeletedAt  *int64 `json:"deleted_at,omitempty"`
}

// SyncService handles sync operations
type SyncService struct {
	db     *sql.DB
	vaults *vault.Manager
}

// NewSyncService creates a new sync service
func NewSyncService(db *sql.DB, vaultManager *vault.Manager) *SyncService {
	return &SyncService{
		db:     db,
		vaults: vaultManager,
	}
}

// GetFiles retrieves all files for a vault from the database
func (s *SyncService) GetFiles(vaultID string) ([]FileRecord, error) {
	query := `SELECT id, vault_id, path, hash_sha256, size_bytes, modified_at, deleted_at 
	          FROM files WHERE vault_id = ?`
	rows, err := s.db.Query(query, vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []FileRecord
	for rows.Next() {
		var file FileRecord
		err := rows.Scan(&file.ID, &file.VaultID, &file.Path, &file.HashSHA256, &file.SizeBytes, &file.ModifiedAt, &file.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return files, nil
}

// GetManifest generates a manifest for a vault
func (s *SyncService) GetManifest(vaultID string) (*ManifestResponse, error) {
	files, err := s.GetFiles(vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	// Convert to FileMetadata (use non-nil slice so JSON encodes [] not null)
	manifestFiles := make([]FileMetadata, 0)
	for _, file := range files {
		// Only include non-deleted files in the manifest
		if file.DeletedAt == nil {
			manifestFiles = append(manifestFiles, FileMetadata{
				Path:       file.Path,
				Hash:       file.HashSHA256,
				Size:       file.SizeBytes,
				ModifiedAt: file.ModifiedAt,
			})
		}
	}

	return &ManifestResponse{
		VaultID:     vaultID,
		GeneratedAt: time.Now().Unix(),
		Files:       manifestFiles,
	}, nil
}

// CountActiveFiles returns non-deleted file records for a vault.
func (s *SyncService) CountActiveFiles(vaultID string) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM files WHERE vault_id = ? AND deleted_at IS NULL`,
		vaultID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active files: %w", err)
	}
	return count, nil
}

// CountRecentDeletions returns tombstones created at or after sinceUnix.
func (s *SyncService) CountRecentDeletions(vaultID string, sinceUnix int64) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM files WHERE vault_id = ? AND deleted_at IS NOT NULL AND deleted_at >= ?`,
		vaultID,
		sinceUnix,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent deletions: %w", err)
	}
	return count, nil
}

// AssertDeleteAllowed enforces bulk-delete thresholds for single-file DELETE API calls.
func (s *SyncService) AssertDeleteAllowed(vaultID string, confirmed bool) error {
	if confirmed {
		return nil
	}

	active, err := s.CountActiveFiles(vaultID)
	if err != nil {
		return err
	}

	since := time.Now().Add(-RecentDeletionWindow).Unix()
	recent, err := s.CountRecentDeletions(vaultID, since)
	if err != nil {
		return err
	}

	proposed := recent + 1
	safety := CheckBulkDeleteSafety(active, proposed)
	if !safety.IsSafe {
		return fmt.Errorf("%w: %s", ErrBulkDeleteBlocked, safety.Reason)
	}
	return nil
}

func (s *SyncService) validateBatchPaths(vaultObj *vault.Vault, files []FileMetadata) error {
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		if strings.TrimSpace(file.Path) == "" {
			return fmt.Errorf("invalid path: empty")
		}
		if _, dup := seen[file.Path]; dup {
			return fmt.Errorf("invalid path: duplicate %q", file.Path)
		}
		seen[file.Path] = struct{}{}
		if err := s.vaults.ValidatePath(vaultObj, file.Path); err != nil {
			return err
		}
	}
	return nil
}

// ProcessBatch processes a batch sync request
func (s *SyncService) ProcessBatch(vaultID string, request *SyncBatchRequest) (*SyncBatchResponse, error) {
	vaultObj, err := s.vaults.GetVault(vaultID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}
	if err := s.validateBatchPaths(vaultObj, request.Files); err != nil {
		return nil, err
	}

	// Get server's current file state
	serverFiles, err := s.GetFiles(vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server files: %w", err)
	}

	response := &SyncBatchResponse{
		ToUpload:          []string{},
		ToDownload:        []string{},
		Conflicts:         []string{},
		ToDelete:          []string{},
		BulkDeleteWarning: false,
	}

	// Create map for O(1) lookup of server files
	serverMap := make(map[string]FileRecord)
	for _, file := range serverFiles {
		// Only consider non-deleted files
		if file.DeletedAt == nil {
			serverMap[file.Path] = file
		}
	}

	// Create map for client files
	clientMap := make(map[string]FileMetadata)
	for _, file := range request.Files {
		clientMap[file.Path] = file
	}

	// Iterate through client files
	for _, clientFile := range request.Files {
		serverFile, existsOnServer := serverMap[clientFile.Path]

		if !existsOnServer {
			// File exists on client but not on server
			// Check if it was deleted on server by looking for tombstones
			var deletedAt *int64
			err := s.db.QueryRow(`SELECT deleted_at FROM files WHERE vault_id = ? AND path = ? ORDER BY id DESC LIMIT 1`, 
				vaultID, clientFile.Path).Scan(&deletedAt)
			
			if err != nil && err != sql.ErrNoRows {
				// Database error
				return nil, fmt.Errorf("failed to check file deletion status: %w", err)
			}
			
			if deletedAt != nil {
				// File was deleted on server, client should delete it too
				response.ToDelete = append(response.ToDelete, clientFile.Path)
			} else {
				// File is new on client, upload it
				response.ToUpload = append(response.ToUpload, clientFile.Path)
			}
		} else {
			// File exists in both
			if clientFile.Hash == serverFile.HashSHA256 {
				// Synchronized, no action needed
				continue
			}
			
			// Hashes differ - potential conflict
			// For MVP, we'll mark as conflict and let client handle it
			response.Conflicts = append(response.Conflicts, clientFile.Path)
		}
	}

	// Iterate through server files to find missing on client
	for _, serverFile := range serverFiles {
		// Only consider non-deleted files
		if serverFile.DeletedAt == nil {
			if _, existsOnClient := clientMap[serverFile.Path]; !existsOnClient {
				// File exists on server but not on client, download it
				response.ToDownload = append(response.ToDownload, serverFile.Path)
			}
		}
	}

	// Bulk delete protection
	totalFiles := len(serverMap)
	deletions := len(response.ToDelete)
	safety := CheckBulkDeleteSafety(totalFiles, deletions)
	response.BulkDeleteWarning = !safety.IsSafe
	if !safety.IsSafe && !request.BulkDeleteConfirmed {
		response.ToDelete = []string{}
	}

	return response, nil
}

// UpdateFile updates a file record in the database
func (s *SyncService) UpdateFile(vaultID, filePath, hash string, size int64) error {
	// Get current timestamp
	now := time.Now().Unix()

	// Check if file already exists
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM files WHERE vault_id = ? AND path = ?)`, 
		vaultID, filePath).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if exists {
		// Update existing file
		query := `UPDATE files SET hash_sha256 = ?, size_bytes = ?, modified_at = ? WHERE vault_id = ? AND path = ?`
		_, err = s.db.Exec(query, hash, size, now, vaultID, filePath)
	} else {
		// Insert new file
		query := `INSERT INTO files (vault_id, path, hash_sha256, size_bytes, modified_at) VALUES (?, ?, ?, ?, ?)`
		_, err = s.db.Exec(query, vaultID, filePath, hash, size, now)
	}

	if err != nil {
		return fmt.Errorf("failed to update file record: %w", err)
	}

	return nil
}

// DeleteFile marks a file as deleted in the database
func (s *SyncService) DeleteFile(vaultID, filePath string) error {
	// Get current timestamp
	now := time.Now().Unix()

	// Check if file already exists
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM files WHERE vault_id = ? AND path = ?)`, 
		vaultID, filePath).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if exists {
		// Update existing file to mark as deleted
		query := `UPDATE files SET deleted_at = ? WHERE vault_id = ? AND path = ?`
		_, err = s.db.Exec(query, now, vaultID, filePath)
	} else {
		// Insert new file record marked as deleted (tombstone)
		query := `INSERT INTO files (vault_id, path, hash_sha256, size_bytes, modified_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = s.db.Exec(query, vaultID, filePath, "", 0, now, now)
	}

	if err != nil {
		return fmt.Errorf("failed to mark file as deleted: %w", err)
	}

	return nil
}

// GetFileInfo retrieves file information from the filesystem
func (s *SyncService) GetFileInfo(vaultObj *vault.Vault, filePath string) (os.FileInfo, error) {
	fullPath, err := s.vaults.GetVaultPath(vaultObj, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault path: %w", err)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return info, nil
}

// WriteFile writes data to a file in the vault
func (s *SyncService) WriteFile(vaultObj *vault.Vault, filePath string, data []byte) error {
	fullPath, err := s.vaults.GetVaultPath(vaultObj, filePath)
	if err != nil {
		return fmt.Errorf("failed to get vault path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads data from a file in the vault
func (s *SyncService) ReadFile(vaultObj *vault.Vault, filePath string) ([]byte, error) {
	fullPath, err := s.vaults.GetVaultPath(vaultObj, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault path: %w", err)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// DeletePhysicalFile deletes a file from the filesystem
func (s *SyncService) DeletePhysicalFile(vaultObj *vault.Vault, filePath string) error {
	fullPath, err := s.vaults.GetVaultPath(vaultObj, filePath)
	if err != nil {
		return fmt.Errorf("failed to get vault path: %w", err)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}