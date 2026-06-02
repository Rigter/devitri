package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rigter/devitri/backend/internal/vault"
	"github.com/gorilla/mux"
)

// VaultHandler handles vault-related requests
type VaultHandler struct {
	manager *vault.Manager
}

// NewVaultHandler creates a new VaultHandler
func NewVaultHandler(manager *vault.Manager) *VaultHandler {
	return &VaultHandler{manager: manager}
}

// ListVaults returns all registered vaults
func (h *VaultHandler) ListVaults(w http.ResponseWriter, r *http.Request) {
	vaults, err := h.manager.ListVaultsWithStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"vaults": vaults})
}

// GetVault returns a specific vault by ID
func (h *VaultHandler) GetVault(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vaultID := vars["vault_id"]

	v, err := h.manager.GetVault(vaultID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	item, err := h.manager.VaultToListItem(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
