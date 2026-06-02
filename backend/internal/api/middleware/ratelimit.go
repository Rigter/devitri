package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements rate limiting per IP address
type RateLimiter struct {
	limiterMap map[string]*rate.Limiter
	limit      rate.Limit
	burst      int
	mutex      sync.RWMutex
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 5
	}
	return &RateLimiter{
		limiterMap: make(map[string]*rate.Limiter),
		limit:      rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:      requestsPerMinute,
	}
}

// getLimiter returns the rate limiter for a specific IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiterMap[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.limiterMap[ip] = limiter
	}

	return limiter
}

// getIP extracts the IP address from an HTTP request
func getIP(r *http.Request) (string, error) {
	// Only trust proxy headers when explicitly enabled.
	trustProxy := os.Getenv("DEVITRI_TRUST_PROXY_HEADERS") == "true"
	if trustProxy {
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			// Use the first IP in the list.
			if idx := strings.IndexByte(xff, ','); idx >= 0 {
				xff = xff[:idx]
			}
			ip := strings.TrimSpace(xff)
			if ip != "" {
				return ip, nil
			}
		}
		if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
			return xrip, nil
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse remote address: %w", err)
	}
	return host, nil
}

// RateLimit applies rate limiting to HTTP requests per IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP address
		ip, err := getIP(r)
		if err != nil {
			http.Error(w, "failed to get IP address", http.StatusInternalServerError)
			return
		}

		// Get the rate limiter for this IP
		limiter := rl.getLimiter(ip)

		// Check if request is allowed
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error": "rate_limit_exceeded"}`)
			return
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}