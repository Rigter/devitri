package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rigter/devitri/backend/internal/api/middleware"
	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/config"
	"github.com/rigter/devitri/backend/internal/sync"
	"github.com/rigter/devitri/backend/internal/vault"
)

// SyncHandler handles sync-related HTTP requests
type SyncHandler struct {
	db     *sql.DB
	sync   *sync.SyncService
	vaults *vault.Manager
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(db *sql.DB) *SyncHandler {
	vaultManager := vault.NewManager(db)
	syncService := sync.NewSyncService(db, vaultManager)
	
	return &SyncHandler{
		db:     db,
		sync:   syncService,
		vaults: vaultManager,
	}
}

func (h *SyncHandler) recordVaultSync(r *http.Request, vaultID string) {
	session, ok := r.Context().Value(middleware.SessionContextKey).(*auth.Session)
	if !ok || session == nil {
		return
	}

	deviceName := session.DeviceName
	if deviceName == "" {
		deviceName = session.DeviceID
	}

	if err := h.vaults.RecordVaultSync(vaultID, session.DeviceID, deviceName); err != nil {
		fmt.Printf("Warning: failed to record vault sync for %s: %v\n", vaultID, err)
	}
}

// GetManifest returns the server's current file manifest
func (h *SyncHandler) GetManifest(w http.ResponseWriter, r *http.Request) {
	// Get vault ID from URL parameters
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	// Validate and get vault (auto-create on first sync)
	vaultObj, err := h.vaults.GetOrCreateVault(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("vault not found: %s", err), http.StatusNotFound)
		return
	}

	// Ensure vault directory exists
	if err := h.vaults.EnsureVaultDirectory(vaultObj); err != nil {
		http.Error(w, fmt.Sprintf("failed to access vault directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Generate manifest
	manifest, err := h.sync.GetManifest(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate manifest: %s", err), http.StatusInternalServerError)
		return
	}

	// Set content type and write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
}

// UploadFile handles file uploads
func (h *SyncHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Get vault ID from URL parameters
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	// Validate and get vault (auto-create on first sync)
	vaultObj, err := h.vaults.GetOrCreateVault(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("vault not found: %s", err), http.StatusNotFound)
		return
	}

	// Ensure vault directory exists
	if err := h.vaults.EnsureVaultDirectory(vaultObj); err != nil {
		http.Error(w, fmt.Sprintf("failed to access vault directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Get file path and hash from headers
	filePath := r.Header.Get("X-File-Path")
	fileHash := r.Header.Get("X-File-Hash")

	if filePath == "" {
		http.Error(w, "X-File-Path header is required", http.StatusBadRequest)
		return
	}

	if fileHash == "" {
		http.Error(w, "X-File-Hash header is required", http.StatusBadRequest)
		return
	}

	// Validate file path
	if err := h.vaults.ValidatePath(vaultObj, filePath); err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	// Read file data (bounded)
	maxBytes := config.Current.MaxUploadBytes
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read file data: %s", err), http.StatusBadRequest)
		return
	}

	// Calculate hash of received data to verify integrity
	calculatedHash, err := sync.CalculateSHA256FromReader(strings.NewReader(string(data)))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to calculate hash: %s", err), http.StatusInternalServerError)
		return
	}

	// Verify hash matches
	if calculatedHash != fileHash {
		http.Error(w, "file hash mismatch", http.StatusBadRequest)
		return
	}

	// Write file to vault
	if err := h.sync.WriteFile(vaultObj, filePath, data); err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %s", err), http.StatusInternalServerError)
		return
	}

	// Update database record
	if err := h.sync.UpdateFile(vaultID, filePath, fileHash, int64(len(data))); err != nil {
		http.Error(w, fmt.Sprintf("failed to update file record: %s", err), http.StatusInternalServerError)
		return
	}

	h.recordVaultSync(r, vaultID)

	// Return success response
	response := map[string]interface{}{
		"status": "success",
		"path":   filePath,
		"hash":   fileHash,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
}

// DownloadFile handles file downloads
func (h *SyncHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	// Get vault ID from URL parameters
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	// Get file path from query parameters
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "path query parameter is required", http.StatusBadRequest)
		return
	}

	// Validate and get vault (auto-create on first sync)
	vaultObj, err := h.vaults.GetOrCreateVault(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("vault not found: %s", err), http.StatusNotFound)
		return
	}

	// Ensure vault directory exists
	if err := h.vaults.EnsureVaultDirectory(vaultObj); err != nil {
		http.Error(w, fmt.Sprintf("failed to access vault directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Validate file path
	if err := h.vaults.ValidatePath(vaultObj, filePath); err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	// Check if file exists in database and is not deleted
	var deletedAt *int64
	err = h.db.QueryRow(`SELECT deleted_at FROM files WHERE vault_id = ? AND path = ? ORDER BY id DESC LIMIT 1`, 
		vaultID, filePath).Scan(&deletedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to query file: %s", err), http.StatusInternalServerError)
		return
	}
	
	if deletedAt != nil {
		http.Error(w, "file has been deleted", http.StatusNotFound)
		return
	}

	// Read file data
	data, err := h.sync.ReadFile(vaultObj, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found on disk", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to read file: %s", err), http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))
	
	// Write file data
	if _, err := w.Write(data); err != nil {
		// Log error but don't send HTTP error as headers are already sent
		fmt.Printf("failed to write file data: %s\n", err)
		return
	}
}

// DeleteFile handles file deletion
func (h *SyncHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	// Get vault ID from URL parameters
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	// Validate and get vault (auto-create on first sync)
	vaultObj, err := h.vaults.GetOrCreateVault(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("vault not found: %s", err), http.StatusNotFound)
		return
	}

	// Ensure vault directory exists
	if err := h.vaults.EnsureVaultDirectory(vaultObj); err != nil {
		http.Error(w, fmt.Sprintf("failed to access vault directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Parse JSON request body
	var request struct {
		Path string `json:"path"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request: %s", err), http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}

	// Validate file path
	if err := h.vaults.ValidatePath(vaultObj, request.Path); err != nil {
		http.Error(w, fmt.Sprintf("invalid file path: %s", err), http.StatusBadRequest)
		return
	}

	confirmed := strings.EqualFold(r.Header.Get("X-Bulk-Delete-Confirmed"), "true")
	if err := h.sync.AssertDeleteAllowed(vaultID, confirmed); err != nil {
		if errors.Is(err, sync.ErrBulkDeleteBlocked) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":               "bulk_delete_blocked",
				"message":             err.Error(),
				"bulk_delete_warning": true,
			})
			return
		}
		http.Error(w, fmt.Sprintf("failed to check delete safety: %s", err), http.StatusInternalServerError)
		return
	}

	// Delete file from database (mark as deleted)
	if err := h.sync.DeleteFile(vaultID, request.Path); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete file record: %s", err), http.StatusInternalServerError)
		return
	}

	h.recordVaultSync(r, vaultID)

	// Delete file from filesystem
	if err := h.sync.DeletePhysicalFile(vaultObj, request.Path); err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("failed to delete file from disk: %s", err), http.StatusInternalServerError)
			return
		}
		// File doesn't exist on disk, which is fine
	}

	// Return success response
	response := map[string]interface{}{
		"status": "success",
		"path":   request.Path,
		"deleted_at": time.Now().Unix(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
}

// BatchSync handles delta-sync negotiations
func (h *SyncHandler) BatchSync(w http.ResponseWriter, r *http.Request) {
	// Get vault ID from URL parameters
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	// Validate and get vault (auto-create on first sync)
	vaultObj, err := h.vaults.GetOrCreateVault(vaultID)
	if err != nil {
		http.Error(w, fmt.Sprintf("vault not found: %s", err), http.StatusNotFound)
		return
	}

	// Ensure vault directory exists
	if err := h.vaults.EnsureVaultDirectory(vaultObj); err != nil {
		http.Error(w, fmt.Sprintf("failed to access vault directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Parse JSON request body
	var request sync.SyncBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request: %s", err), http.StatusBadRequest)
		return
	}

	// Process batch sync
	response, err := h.sync.ProcessBatch(vaultID, &request)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid path") ||
			strings.Contains(err.Error(), "access to .obsidian") ||
			strings.Contains(err.Error(), "directory traversal") {
			status = http.StatusBadRequest
		}
		http.Error(w, fmt.Sprintf("failed to process batch sync: %s", err), status)
		return
	}

	h.recordVaultSync(r, vaultID)

	// Set content type and write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
}