package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rigter/devitri/backend/internal/config"
)

func TestGetSettings(t *testing.T) {
	config.Current = config.Config{
		SessionTTLHours:         24,
		DeviceTokenTTLDays:      365,
		LoginRateLimitPerMinute: 5,
		DeleteThresholdCount:    20,
		DeleteThresholdPercent:  5,
		BcryptCost:              14,
		TZ:                      "America/Mexico_City",
		PUID:                    1000,
		PGID:                    1000,
		MasterHashConfigured:    true,
		JWTSecretConfigured:     true,
	}

	handler := NewSettingsHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()

	handler.GetSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body settingsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Security.SessionTTLHours != 24 {
		t.Errorf("session_ttl_hours = %d, want 24", body.Security.SessionTTLHours)
	}
	if body.Sync.DeleteThresholdCount != 20 {
		t.Errorf("delete_threshold_count = %d, want 20", body.Sync.DeleteThresholdCount)
	}
	if !body.Security.MasterHashConfigured || !body.Security.JWTSecretConfigured {
		t.Error("expected secrets to be marked configured")
	}
}
