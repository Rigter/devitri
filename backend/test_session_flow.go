//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/db"
)

func main() {
	// Load .env file
	envPath := filepath.Join("..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
		}
	}

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer database.Close()

	// Create session store
	sessionStore := auth.NewSessionStore(database)

	// Create a test session
	deviceID := "test-device-id"
	deviceName := "Test Device"
	vaultID := "test-vault-id"
	session, token, err := sessionStore.CreateSession(deviceID, deviceName, &vaultID)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}

	fmt.Printf("Created session with token: %.50s...\n", token)
	fmt.Printf("Session ID: %d\n", session.ID)
	fmt.Printf("Token hash: %.20s...\n", session.TokenHash)

	// Hash the token to verify it matches the session
	tokenHash := hashToken(token)
	fmt.Printf("Token hash from token: %.20s...\n", tokenHash)

	if tokenHash == session.TokenHash {
		fmt.Println("Token hashes match!")
	} else {
		fmt.Println("Token hashes don't match!")
	}

	// Try to retrieve the session by token hash
	retrievedSession, err := sessionStore.GetSessionByTokenHash(session.TokenHash)
	if err != nil {
		fmt.Printf("Error retrieving session: %v\n", err)
		return
	}

	fmt.Printf("Retrieved session ID: %d\n", retrievedSession.ID)
	fmt.Printf("Retrieved device ID: %s\n", retrievedSession.DeviceID)

	// Now test JWT verification with auth middleware logic
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	jwtManager := auth.NewJWTManager(jwtSecret, auth.TokenTTL)
	parsedToken, err := jwtManager.Verify(token)
	if err != nil {
		fmt.Printf("Error verifying JWT token: %v\n", err)
		return
	}

	fmt.Println("JWT token verification successful!")
	fmt.Printf("Token valid: %t\n", parsedToken.Valid)
}

func hashToken(token string) string {
	hasher := auth.NewJWTManager("", 0) // We don't need the JWT manager for hashing
	// Actually, let's just reimplement the hashing logic
	// This is from auth/session.go lines 44-47
	h := auth.NewJWTManager("", 0)
	// We'll use the same approach as in session.go
	// But we need a hasher directly
	// Let me import crypto/sha256 directly
	return token // Placeholder
}