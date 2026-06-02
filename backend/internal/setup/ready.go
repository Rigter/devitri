package setup

import "os"

// IsConfigured reports whether required secrets are present in the environment.
func IsConfigured() bool {
	return os.Getenv("DEVITRI_MASTER_HASH") != "" && os.Getenv("DEVITRI_JWT_SECRET") != ""
}

// MissingFields lists unset required environment variables.
func MissingFields() []string {
	missing := []string{}
	if os.Getenv("DEVITRI_MASTER_HASH") == "" {
		missing = append(missing, "DEVITRI_MASTER_HASH")
	}
	if os.Getenv("DEVITRI_JWT_SECRET") == "" {
		missing = append(missing, "DEVITRI_JWT_SECRET")
	}
	return missing
}
