//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rigter/devitri/backend/internal/auth"
)

func main() {
	// Load .env file
	envPath := filepath.Join("..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	masterHash := os.Getenv("DEVITRI_MASTER_HASH")
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	
	fmt.Printf("DEVITRI_MASTER_HASH: '%s'\n", masterHash)
	fmt.Printf("DEVITRI_JWT_SECRET: '%s'\n", jwtSecret)
	
	if masterHash == "" {
		fmt.Println("DEVITRI_MASTER_HASH is empty!")
		return
	}
	
	// Test password
	password := "D3v1tr1$$"
	
	// Compare password with hash
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", masterHash)
	fmt.Printf("Match: %t\n", auth.ComparePassword(password, masterHash))
}