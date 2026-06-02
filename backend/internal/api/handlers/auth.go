package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/config"
	"github.com/rigter/devitri/backend/internal/api/middleware"
	"github.com/gorilla/mux"
)

// LoginRequest represents the request body for login
type LoginRequest struct {
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

// LoginResponse represents the response body for login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	DeviceID  string `json:"device_id"`
}

// SessionResponse represents the response body for session info
type SessionResponse struct {
	ID         int64   `json:"id,omitempty"`
	DeviceID   string  `json:"device_id"`
	DeviceName string  `json:"device_name"`
	VaultID    *string `json:"vault_id"`
	CreatedAt  int64   `json:"created_at"`
	ExpiresAt  int64   `json:"expires_at"`
	LastSeen   int64   `json:"last_seen"`
}

// TokenRequest represents the request body for generating a new token
type TokenRequest struct {
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	sessionStore *auth.SessionStore
	masterHash   string
	jwtSecret    string
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(sessionStore *auth.SessionStore) *AuthHandler {
	return &AuthHandler{
		sessionStore: sessionStore,
		masterHash:   os.Getenv("DEVITRI_MASTER_HASH"),
		jwtSecret:    os.Getenv("DEVITRI_JWT_SECRET"),
	}
}

// Login handles the POST /api/auth/login endpoint
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Check if we're in first-run mode
	if h.masterHash == "" || h.jwtSecret == "" {
		http.Error(w, `{"error": "first_run_mode"}`, http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid_json"}`, http.StatusBadRequest)
		return
	}

	// Validate password
	if !auth.ComparePassword(req.Password, h.masterHash) {
		http.Error(w, `{"error": "invalid_credentials"}`, http.StatusUnauthorized)
		return
	}

	// Create a new session
	session, token, err := h.sessionStore.CreateSessionWithTTL(req.DeviceID, req.DeviceName, nil, config.Current.SessionTTL())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed_to_create_session", "message": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
		DeviceID:  session.DeviceID,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout handles the POST /api/auth/logout endpoint
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session, ok := r.Context().Value(middleware.SessionContextKey).(*auth.Session)
	if !ok {
		http.Error(w, `{"error": "session_not_found"}`, http.StatusUnauthorized)
		return
	}

	// Delete session from database
	err := h.sessionStore.DeleteSession(session.TokenHash)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed_to_logout", "message": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged_out"})
}

// Session handles the GET /api/auth/session endpoint
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session, ok := r.Context().Value(middleware.SessionContextKey).(*auth.Session)
	if !ok {
		http.Error(w, `{"error": "session_not_found"}`, http.StatusUnauthorized)
		return
	}

	// Prepare response
	response := SessionResponse{
		ID:         session.ID,
		DeviceID:   session.DeviceID,
		DeviceName: session.DeviceName,
		VaultID:    session.VaultID,
		CreatedAt:  session.CreatedAt,
		ExpiresAt:  session.ExpiresAt,
		LastSeen:   session.LastSeen,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GenerateToken handles the POST /api/auth/token endpoint
func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid_json"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.DeviceName == "" {
		http.Error(w, `{"error": "missing_fields", "message": "device_id and device_name are required"}`, http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, `{"error": "password_required"}`, http.StatusBadRequest)
		return
	}
	if !auth.ComparePassword(req.Password, h.masterHash) {
		http.Error(w, `{"error": "invalid_credentials"}`, http.StatusUnauthorized)
		return
	}

	// Create a new long-lived session for Obsidian / Connect
	session, token, err := h.sessionStore.CreateSessionWithTTL(req.DeviceID, req.DeviceName, nil, config.Current.DeviceTokenTTL())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed_to_create_token", "message": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
		DeviceID:  session.DeviceID,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListSessions handles the GET /api/auth/sessions endpoint
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.sessionStore.ListSessions()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed_to_list_sessions", "message": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	var response []SessionResponse
	for _, s := range sessions {
		response = append(response, SessionResponse{
			ID:         s.ID,
			DeviceID:   s.DeviceID,
			DeviceName: s.DeviceName,
			VaultID:    s.VaultID,
			CreatedAt:  s.CreatedAt,
			ExpiresAt:  s.ExpiresAt,
			LastSeen:   s.LastSeen,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if response == nil {
		json.NewEncoder(w).Encode([]SessionResponse{})
	} else {
		json.NewEncoder(w).Encode(response)
	}
}

// RevokeSession handles the DELETE /api/auth/sessions/{id} endpoint
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	current, ok := r.Context().Value(middleware.SessionContextKey).(*auth.Session)
	if !ok {
		http.Error(w, `{"error": "session_not_found"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		http.Error(w, `{"error": "invalid_id"}`, http.StatusBadRequest)
		return
	}
	if id != current.ID {
		var body struct {
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Password == "" || !auth.ComparePassword(body.Password, h.masterHash) {
			http.Error(w, `{"error": "password_required", "message": "master password required to revoke other sessions"}`, http.StatusForbidden)
			return
		}
	}

	err := h.sessionStore.DeleteSessionByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed_to_revoke_session", "message": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "session_revoked"})
}

// HealthCheck handles the GET /health endpoint
func (h *AuthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check if we're in first-run mode
	firstRun := h.masterHash == "" || h.jwtSecret == ""
	
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"firstRun":  firstRun,
	}
	
	if firstRun {
		response["message"] = "Server is in first-run mode. Please complete setup."
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}