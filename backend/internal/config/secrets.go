package config

import (
	"fmt"
	"os"
)

const devJWTPlaceholder = "default-secret-key-for-development"

// AllowInsecureJWT is set when DEVITRI_ALLOW_INSECURE_JWT=true (local development only).
func AllowInsecureJWT() bool {
	return os.Getenv("DEVITRI_ALLOW_INSECURE_JWT") == "true"
}

// ResolveJWTSecret returns the JWT signing secret or an error if unset in production mode.
func ResolveJWTSecret() (string, error) {
	secret := os.Getenv("DEVITRI_JWT_SECRET")
	if secret != "" {
		return secret, nil
	}
	if AllowInsecureJWT() {
		return devJWTPlaceholder, nil
	}
	return "", fmt.Errorf("DEVITRI_JWT_SECRET is required (set DEVITRI_ALLOW_INSECURE_JWT=true only for local development)")
}

// ValidateStartupSecrets ensures the process can run securely once first-run is complete.
func ValidateStartupSecrets(configured bool) error {
	if !configured {
		return nil
	}
	_, err := ResolveJWTSecret()
	return err
}
