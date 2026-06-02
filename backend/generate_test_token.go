//go:build ignore

package main

import (
	"crypto/sha256"
	"database/sql"
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

	fmt.Printf("TOKEN=%s\n", token)
}