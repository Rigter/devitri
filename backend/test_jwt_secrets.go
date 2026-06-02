//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	envPath := filepath.Join("..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
		}
	}

	// Check JWT secret from environment
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	fmt.Printf("JWT Secret from environment: %s\n", jwtSecret)

	// Compare with the default value
	defaultSecret := "default-secret-key-for-development"
	if jwtSecret == defaultSecret {
		fmt.Println("Warning: Using default JWT secret!")
	} else if jwtSecret == "" {
		fmt.Println("Error: JWT secret is empty!")
	} else {
		fmt.Println("JWT secret loaded correctly.")
	}
}