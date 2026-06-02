package middleware

import (
	"net/http"
	"os"
	"strings"
)

func allowedOrigins() []string {
	// Comma-separated list, e.g. "https://devitri.example.com,https://dashboard.example.com"
	raw := strings.TrimSpace(os.Getenv("DEVITRI_CORS_ORIGINS"))
	if raw == "" {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:4173",
			"http://127.0.0.1:4173",
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		o := strings.TrimSpace(p)
		if o != "" {
			out = append(out, o)
		}
	}
	return out
}

// CORS allows browser requests from the dev frontend origin.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-File-Path, X-File-Hash, X-Bulk-Delete-Confirmed")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	for _, allowed := range allowedOrigins() {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}
