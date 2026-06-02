package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rigter/devitri/backend/internal/setup"
)

// FirstRun blocks API routes until required secrets are configured.
func FirstRun(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if setup.IsConfigured() {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		if strings.HasPrefix(path, "/api/setup") {
			next.ServeHTTP(w, r)
			return
		}
		if path == "/api/auth/login" && r.Method == http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		if path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":          "first_run_mode",
			"missing_fields": setup.MissingFields(),
		})
	})
}
