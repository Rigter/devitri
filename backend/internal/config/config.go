package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultSessionTTLHours       = 24
	defaultDeviceTokenTTLDays    = 365
	defaultLoginRateLimitPerMin  = 5
	defaultDeleteThresholdCount  = 20
	defaultDeleteThresholdPercent = 5.0
	defaultBcryptCost            = 14
	defaultTZ                    = "America/Mexico_City"
	defaultPUID                  = 1000
	defaultPGID                  = 1000
	defaultMaxUploadBytes        = 52_428_800 // 50 MiB
)

// Config holds platform settings loaded from environment variables.
type Config struct {
	SessionTTLHours          int
	DeviceTokenTTLDays       int
	LoginRateLimitPerMinute  int
	DeleteThresholdCount     int
	DeleteThresholdPercent   float64
	BcryptCost               int
	TZ                       string
	PUID                     int
	PGID                     int
	MaxUploadBytes           int64
	MasterHashConfigured     bool
	JWTSecretConfigured      bool
}

// Current is populated by Load at startup.
var Current Config

// Load reads configuration from the environment with defaults.
func Load() Config {
	cfg := Config{
		SessionTTLHours:         envInt("DEVITRI_SESSION_TTL_HOURS", defaultSessionTTLHours),
		DeviceTokenTTLDays:      envInt("DEVITRI_DEVICE_TOKEN_TTL_DAYS", defaultDeviceTokenTTLDays),
		LoginRateLimitPerMinute: envInt("DEVITRI_LOGIN_RATE_LIMIT", defaultLoginRateLimitPerMin),
		DeleteThresholdCount:    envInt("DEVITRI_DELETE_THRESHOLD_COUNT", defaultDeleteThresholdCount),
		DeleteThresholdPercent:  envFloat("DEVITRI_DELETE_THRESHOLD_PERCENT", defaultDeleteThresholdPercent),
		BcryptCost:              defaultBcryptCost,
		TZ:                      envString("TZ", envString("DEVITRI_TZ", defaultTZ)),
		PUID:                    envInt("PUID", envInt("DEVITRI_PUID", defaultPUID)),
		PGID:                    envInt("PGID", envInt("DEVITRI_PGID", defaultPGID)),
		MaxUploadBytes:          envInt64("DEVITRI_MAX_UPLOAD_BYTES", defaultMaxUploadBytes),
		MasterHashConfigured:    os.Getenv("DEVITRI_MASTER_HASH") != "",
		JWTSecretConfigured:     os.Getenv("DEVITRI_JWT_SECRET") != "",
	}
	Current = cfg
	return cfg
}

func (c Config) SessionTTL() time.Duration {
	return time.Duration(c.SessionTTLHours) * time.Hour
}

func (c Config) DeviceTokenTTL() time.Duration {
	return time.Duration(c.DeviceTokenTTLDays) * 24 * time.Hour
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
