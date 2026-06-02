//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	var vaultID *string = nil
	
	fmt.Println("Creating session...")
	session, token, err := sessionStore.CreateSession(deviceID, deviceName, vaultID)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}

	fmt.Printf("Created session successfully!\n")
	fmt.Printf("Token: %.50s...\n", token)
	fmt.Printf("Session ID: %d\n", session.ID)
	fmt.Printf("Token Hash: %.20s...\n", session.TokenHash)

	// Hash the token manually to verify it matches
	hasher := sha256.New()
	hasher.Write([]byte(token))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("Manual hash: %.20s...\n", tokenHash)

	if tokenHash == session.TokenHash {
		fmt.Println("Token hashes match!")
	} else {
		fmt.Println("Token hashes don't match!")
	}

	// Try to retrieve the session
	fmt.Println("\nRetrieving session...")
	retrievedSession, err := sessionStore.GetSessionByTokenHash(session.TokenHash)
	if err != nil {
		fmt.Printf("Error retrieving session: %v\n", err)
		return
	}

	fmt.Printf("Retrieved session successfully!\n")
	fmt.Printf("Retrieved Session ID: %d\n", retrievedSession.ID)
	fmt.Printf("Retrieved Device ID: %s\n", retrievedSession.DeviceID)

	// Test JWT verification
	fmt.Println("\nTesting JWT verification...")
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	jwtManager := auth.NewJWTManager(jwtSecret, auth.TokenTTL)
	
	verifiedToken, err := jwtManager.Verify(token)
	if err != nil {
		fmt.Printf("Error verifying JWT: %v\n", err)
		return
	}

	fmt.Println("JWT verification successful!")
	fmt.Printf("Token valid: %t\n", verifiedToken.Valid)
}