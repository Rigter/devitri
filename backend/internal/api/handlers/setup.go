package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/setup"
)

// SetupHandler handles first-run setup operations
type SetupHandler struct{}

// NewSetupHandler creates a new SetupHandler
func NewSetupHandler() *SetupHandler {
	return &SetupHandler{}
}

// CheckSetup checks if the server is properly configured
func (h *SetupHandler) CheckSetup(w http.ResponseWriter, r *http.Request) {
	missing := setup.MissingFields()

	response := map[string]interface{}{
		"ready":          setup.IsConfigured(),
		"missing_fields": missing,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateConfig creates bcrypt hash and JWT secret from the user's password.
func (h *SetupHandler) GenerateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, `{"error":"password_too_short"}`, http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusInternalServerError)
		return
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		http.Error(w, "Failed to generate JWT secret", http.StatusInternalServerError)
		return
	}
	jwtSecret := hex.EncodeToString(secretBytes)

	response := map[string]string{
		"hash":        hash,
		"jwt_secret":  jwtSecret,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateMasterHash hashes a user-provided password (legacy endpoint).
func (h *SetupHandler) GenerateMasterHash(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		http.Error(w, `{"error":"password_required"}`, http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"hash": hash})
}

// GenerateJWTSecret generates a random JWT secret
func (h *SetupHandler) GenerateJWTSecret(w http.ResponseWriter, r *http.Request) {
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		http.Error(w, "Failed to generate JWT secret", http.StatusInternalServerError)
		return
	}
	secret := hex.EncodeToString(secretBytes)

	response := map[string]interface{}{
		"secret": secret,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
