package vault

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Vault represents a vault in the system
type Vault struct {
	ID                 int64   `json:"id"`
	VaultID            string  `json:"vault_id"`
	Name               string  `json:"name"`
	Path               string  `json:"path"`
	CreatedAt          int64   `json:"created_at"`
	LastSync           *int64  `json:"last_sync,omitempty"`
	LastSyncDeviceID   *string `json:"last_sync_device_id,omitempty"`
	LastSyncDeviceName *string `json:"last_sync_device_name,omitempty"`
}

const vaultSelectColumns = `id, vault_id, name, path, created_at, last_sync, last_sync_device_id, last_sync_device_name`

func scanVault(scanner interface {
	Scan(dest ...any) error
}) (*Vault, error) {
	var v Vault
	var lastSyncDeviceID, lastSyncDeviceName sql.NullString

	err := scanner.Scan(
		&v.ID,
		&v.VaultID,
		&v.Name,
		&v.Path,
		&v.CreatedAt,
		&v.LastSync,
		&lastSyncDeviceID,
		&lastSyncDeviceName,
	)
	if err != nil {
		return nil, err
	}

	if lastSyncDeviceID.Valid {
		v.LastSyncDeviceID = &lastSyncDeviceID.String
	}
	if lastSyncDeviceName.Valid {
		v.LastSyncDeviceName = &lastSyncDeviceName.String
	}

	return &v, nil
}

// Manager handles vault operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new vault manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// GetVault retrieves a vault by its vault_id
func (m *Manager) GetVault(vaultID string) (*Vault, error) {
	if err := ValidateVaultID(vaultID); err != nil {
		return nil, err
	}
	query := `SELECT ` + vaultSelectColumns + ` FROM vaults WHERE vault_id = ?`
	vault, err := scanVault(m.db.QueryRow(query, vaultID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vault not found: %s", vaultID)
		}
		return nil, fmt.Errorf("failed to get vault: %w", err)
	}

	return vault, nil
}

// RecordVaultSync updates last-sync metadata after a successful write from a device.
func (m *Manager) RecordVaultSync(vaultID, deviceID, deviceName string) error {
	now := time.Now().Unix()
	_, err := m.db.Exec(
		`UPDATE vaults SET last_sync = ?, last_sync_device_id = ?, last_sync_device_name = ? WHERE vault_id = ?`,
		now,
		deviceID,
		deviceName,
		vaultID,
	)
	if err != nil {
		return fmt.Errorf("failed to record vault sync: %w", err)
	}
	return nil
}

// GetOrCreateVault retrieves a vault by vault_id, creating it if it doesn't exist yet.
// This supports the "plugin-first" flow where Obsidian picks the vault_id and the server
// should accept it without requiring manual pre-registration in the dashboard.
func (m *Manager) GetOrCreateVault(vaultID string) (*Vault, error) {
	if err := ValidateVaultID(vaultID); err != nil {
		return nil, err
	}

	v, err := m.GetVault(vaultID)
	if err == nil {
		return v, nil
	}

	path, err := VaultDirectoryPath(vaultID)
	if err != nil {
		return nil, err
	}

	name := strings.Title(strings.ReplaceAll(vaultID, "-", " "))
	now := time.Now().Unix()

	_, insertErr := m.db.Exec(
		"INSERT INTO vaults (vault_id, name, path, created_at) VALUES (?, ?, ?, ?)",
		vaultID, name, path, now,
	)
	if insertErr != nil {
		return nil, fmt.Errorf("failed to create vault %s: %w", vaultID, insertErr)
	}

	created, getErr := m.GetVault(vaultID)
	if getErr != nil {
		return nil, getErr
	}

	return created, nil
}

// ValidatePath checks if a file path is valid and safe for the given vault
func (m *Manager) ValidatePath(vault *Vault, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("invalid path: empty")
	}
	if len(filePath) > maxFilePathLen {
		return fmt.Errorf("invalid path: too long")
	}
	if filepath.IsAbs(filePath) {
		return fmt.Errorf("invalid path: absolute paths not allowed")
	}
	if strings.Contains(filePath, "\x00") {
		return fmt.Errorf("invalid path: null byte not allowed")
	}

	normalized := filepath.ToSlash(filepath.Clean(filePath))
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return fmt.Errorf("invalid path: directory traversal not allowed")
	}
	if pathUsesObsidianConfig(normalized) {
		return fmt.Errorf("access to .obsidian directory is restricted")
	}

	vaultRoot := filepath.Clean(vault.Path)
	absolutePath := filepath.Clean(filepath.Join(vaultRoot, normalized))
	relative, err := filepath.Rel(vaultRoot, absolutePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid path: path escapes vault directory")
	}

	return nil
}

// GetVaultPath returns the absolute path for a file within a vault
func (m *Manager) GetVaultPath(vault *Vault, filePath string) (string, error) {
	if err := m.ValidatePath(vault, filePath); err != nil {
		return "", err
	}
	
	return filepath.Join(vault.Path, filePath), nil
}

// EnsureVaultDirectory verifies the vault directory exists and is accessible.
func (m *Manager) EnsureVaultDirectory(vault *Vault) error {
	info, err := os.Stat(vault.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(vault.Path, 0o755); mkErr != nil {
				return fmt.Errorf("failed to create vault directory: %w", mkErr)
			}
			return nil
		}
		return fmt.Errorf("failed to access vault directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("vault path is not a directory: %s", vault.Path)
	}
	return nil
}

// ScanVaults scans the /vaults directory and registers any subdirectories as vaults in the database
func (m *Manager) ScanVaults() error {
	vaultsRoot := "/vaults"
	
	// If /vaults doesn't exist, create it (or use current directory for dev if it exists)
	if _, err := os.Stat(vaultsRoot); os.IsNotExist(err) {
		// For development fallback to local vaults folder if it exists
		if _, err := os.Stat("./vaults"); err == nil {
			vaultsRoot = "./vaults"
		} else {
			return fmt.Errorf("vaults directory not found: %s", vaultsRoot)
		}
	}

	entries, err := os.ReadDir(vaultsRoot)
	if err != nil {
		return fmt.Errorf("failed to read vaults directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			vaultID := entry.Name()
			name := strings.Title(strings.ReplaceAll(vaultID, "-", " "))
			path := filepath.Join(vaultsRoot, vaultID)
			
			// Check if already exists
			var exists bool
			err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM vaults WHERE vault_id = ?)", vaultID).Scan(&exists)
			if err != nil {
				return fmt.Errorf("failed to check if vault exists: %w", err)
			}
			
			if !exists {
				// Insert new vault
				_, err = m.db.Exec(
					"INSERT INTO vaults (vault_id, name, path, created_at) VALUES (?, ?, ?, ?)",
					vaultID, name, path, 1622332800, // Static timestamp for now or use time.Now().Unix()
				)
				if err != nil {
					return fmt.Errorf("failed to register vault %s: %w", vaultID, err)
				}
				fmt.Printf("Registered new vault: %s (%s)\n", name, vaultID)
			}
		}
	}
	
	return nil
}

// ListVaults returns all registered vaults
func (m *Manager) ListVaults() ([]*Vault, error) {
	query := `SELECT ` + vaultSelectColumns + ` FROM vaults ORDER BY name ASC`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list vaults: %w", err)
	}
	defer rows.Close()

	vaults := []*Vault{}
	for rows.Next() {
		vault, err := scanVault(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vault: %w", err)
		}
		vaults = append(vaults, vault)
	}
	
	return vaults, nil
}

// VaultStats holds aggregate file counts for dashboard display.
type VaultStats struct {
	Files     int   `json:"files"`
	Folders   int   `json:"folders"`
	SizeBytes int64 `json:"size_bytes"`
}

// VaultListItem is the API shape returned by GET /api/vaults and GET /api/vaults/:id.
type VaultListItem struct {
	VaultID            string     `json:"vault_id"`
	Name               string     `json:"name"`
	Path               string     `json:"path"`
	CreatedAt          int64      `json:"created_at"`
	LastSync           *int64     `json:"last_sync,omitempty"`
	LastSyncDeviceID   *string    `json:"last_sync_device_id,omitempty"`
	LastSyncDeviceName *string    `json:"last_sync_device_name,omitempty"`
	TotalFiles         int        `json:"total_files"`
	TotalSizeBytes     int64      `json:"total_size_bytes"`
	Stats              VaultStats `json:"stats"`
}

// GetVaultStats returns file/folder/size aggregates for a vault.
func (m *Manager) GetVaultStats(vaultID string) (VaultStats, error) {
	var stats VaultStats
	err := m.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(size_bytes), 0) FROM files WHERE vault_id = ? AND deleted_at IS NULL`,
		vaultID,
	).Scan(&stats.Files, &stats.SizeBytes)
	if err != nil {
		return stats, fmt.Errorf("failed to get vault stats: %w", err)
	}

	rows, err := m.db.Query(
		`SELECT path FROM files WHERE vault_id = ? AND deleted_at IS NULL`,
		vaultID,
	)
	if err != nil {
		return stats, fmt.Errorf("failed to query file paths: %w", err)
	}
	defer rows.Close()

	folders := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return stats, fmt.Errorf("failed to scan path: %w", err)
		}
		dir := filepath.Dir(path)
		for dir != "." && dir != "" {
			folders[dir] = struct{}{}
			dir = filepath.Dir(dir)
		}
	}
	stats.Folders = len(folders)

	return stats, nil
}

// VaultToListItem builds the API response for a vault including stats.
func (m *Manager) VaultToListItem(v *Vault) (VaultListItem, error) {
	stats, err := m.GetVaultStats(v.VaultID)
	if err != nil {
		return VaultListItem{}, err
	}
	return VaultListItem{
		VaultID:            v.VaultID,
		Name:               v.Name,
		Path:               v.Path,
		CreatedAt:          v.CreatedAt,
		LastSync:           v.LastSync,
		LastSyncDeviceID:   v.LastSyncDeviceID,
		LastSyncDeviceName: v.LastSyncDeviceName,
		TotalFiles:         stats.Files,
		TotalSizeBytes:     stats.SizeBytes,
		Stats:              stats,
	}, nil
}

// ListVaultsWithStats returns all vaults enriched with file statistics.
func (m *Manager) ListVaultsWithStats() ([]VaultListItem, error) {
	vaults, err := m.ListVaults()
	if err != nil {
		return nil, err
	}

	items := make([]VaultListItem, 0, len(vaults))
	for _, v := range vaults {
		item, err := m.VaultToListItem(v)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
