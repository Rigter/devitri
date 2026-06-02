package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rigter/devitri/backend/internal/config"
)

// SettingsHandler serves read-only platform configuration.
type SettingsHandler struct{}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

type settingsSecurityResponse struct {
	SessionTTLHours         int  `json:"session_ttl_hours"`
	DeviceTokenTTLDays      int  `json:"device_token_ttl_days"`
	LoginRateLimitPerMinute int  `json:"login_rate_limit_per_minute"`
	BcryptCost              int  `json:"bcrypt_cost"`
	MasterHashConfigured    bool `json:"master_hash_configured"`
	JWTSecretConfigured     bool `json:"jwt_secret_configured"`
}

type settingsSyncResponse struct {
	DeleteThresholdCount   int     `json:"delete_threshold_count"`
	DeleteThresholdPercent float64 `json:"delete_threshold_percent"`
}

type settingsOperationalResponse struct {
	TZ   string `json:"tz"`
	PUID int    `json:"puid"`
	PGID int    `json:"pgid"`
}

type settingsResponse struct {
	Security    settingsSecurityResponse    `json:"security"`
	Sync        settingsSyncResponse        `json:"sync"`
	Operational settingsOperationalResponse `json:"operational"`
}

// GetSettings handles GET /api/settings.
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	cfg := config.Current

	response := settingsResponse{
		Security: settingsSecurityResponse{
			SessionTTLHours:         cfg.SessionTTLHours,
			DeviceTokenTTLDays:      cfg.DeviceTokenTTLDays,
			LoginRateLimitPerMinute: cfg.LoginRateLimitPerMinute,
			BcryptCost:              cfg.BcryptCost,
			MasterHashConfigured:    cfg.MasterHashConfigured,
			JWTSecretConfigured:     cfg.JWTSecretConfigured,
		},
		Sync: settingsSyncResponse{
			DeleteThresholdCount:   cfg.DeleteThresholdCount,
			DeleteThresholdPercent: cfg.DeleteThresholdPercent,
		},
		Operational: settingsOperationalResponse{
			TZ:   cfg.TZ,
			PUID: cfg.PUID,
			PGID: cfg.PGID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
