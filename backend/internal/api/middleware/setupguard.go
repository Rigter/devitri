package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/rigter/devitri/backend/internal/setup"
)

// SetupGuard blocks setup mutation endpoints after the server is configured.
func SetupGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/setup/check" || !setup.IsConfigured() {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "setup_already_configured",
		})
	})
}
