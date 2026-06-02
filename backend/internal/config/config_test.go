package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DEVITRI_SESSION_TTL_HOURS")
	os.Unsetenv("DEVITRI_DEVICE_TOKEN_TTL_DAYS")
	os.Unsetenv("DEVITRI_LOGIN_RATE_LIMIT")
	os.Unsetenv("DEVITRI_DELETE_THRESHOLD_COUNT")
	os.Unsetenv("DEVITRI_DELETE_THRESHOLD_PERCENT")

	cfg := Load()

	if cfg.SessionTTLHours != 24 {
		t.Errorf("SessionTTLHours = %d, want 24", cfg.SessionTTLHours)
	}
	if cfg.DeviceTokenTTLDays != 365 {
		t.Errorf("DeviceTokenTTLDays = %d, want 365", cfg.DeviceTokenTTLDays)
	}
	if cfg.LoginRateLimitPerMinute != 5 {
		t.Errorf("LoginRateLimitPerMinute = %d, want 5", cfg.LoginRateLimitPerMinute)
	}
	if cfg.DeleteThresholdCount != 20 {
		t.Errorf("DeleteThresholdCount = %d, want 20", cfg.DeleteThresholdCount)
	}
	if cfg.DeleteThresholdPercent != 5 {
		t.Errorf("DeleteThresholdPercent = %f, want 5", cfg.DeleteThresholdPercent)
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("DEVITRI_SESSION_TTL_HOURS", "48")
	t.Setenv("DEVITRI_DEVICE_TOKEN_TTL_DAYS", "30")
	t.Setenv("DEVITRI_LOGIN_RATE_LIMIT", "10")
	t.Setenv("DEVITRI_DELETE_THRESHOLD_COUNT", "15")
	t.Setenv("DEVITRI_DELETE_THRESHOLD_PERCENT", "3")
	t.Setenv("DEVITRI_MASTER_HASH", "hash")
	t.Setenv("DEVITRI_JWT_SECRET", "secret")

	cfg := Load()

	if cfg.SessionTTLHours != 48 {
		t.Errorf("SessionTTLHours = %d, want 48", cfg.SessionTTLHours)
	}
	if !cfg.MasterHashConfigured || !cfg.JWTSecretConfigured {
		t.Error("expected secrets configured")
	}
}
