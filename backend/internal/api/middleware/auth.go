package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/config"
)

// ContextKey is a custom type for context keys
type ContextKey string

// SessionContextKey is the key used to store session data in request context
const SessionContextKey ContextKey = "session"

// AuthMiddleware handles JWT authentication and session validation
type AuthMiddleware struct {
	sessionStore *auth.SessionStore
	jwtManager   *auth.JWTManager
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(sessionStore *auth.SessionStore, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore: sessionStore,
		jwtManager:   auth.NewJWTManager(jwtSecret, config.Current.SessionTTL()),
	}
}

// Auth checks for a valid JWT token and active session
func (am *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing_authorization_header"}`, http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "invalid_authorization_header"}`, http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify the JWT token
		_, err := am.jwtManager.Verify(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				http.Error(w, `{"error": "session_expired"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error": "invalid_token"}`, http.StatusUnauthorized)
			return
		}

		// Hash the token to look up in the database
		hasher := sha256.New()
		hasher.Write([]byte(token))
		tokenHash := hex.EncodeToString(hasher.Sum(nil))

		// Get the session from the database
		session, err := am.sessionStore.GetSessionByTokenHash(tokenHash)
		if err != nil {
			http.Error(w, `{"error": "invalid_session"}`, http.StatusUnauthorized)
			return
		}

		// Update last seen timestamp
		err = am.sessionStore.UpdateLastSeen(tokenHash)
		if err != nil {
			// Log the error but don't fail the request
			fmt.Printf("Warning: failed to update last seen for session %d: %v\n", session.ID, err)
		}

		// Add session to request context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}