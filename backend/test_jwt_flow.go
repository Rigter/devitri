//go:build ignore

package main

import (
	"fmt"
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
			fmt.Printf("Error loading .env file: %v\n", err)
		}
	}

	// Check JWT secret from environment
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	fmt.Printf("JWT Secret from environment: %.20s...\n", jwtSecret)
	
	// Test token generation with the loaded secret
	jwtManager := auth.NewJWTManager(jwtSecret, auth.TokenTTL)
	token, err := jwtManager.Generate()
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return
	}
	fmt.Printf("Generated token: %.50s...\n", token)
	
	// Test token verification with the same secret
	_, err = jwtManager.Verify(token)
	if err != nil {
		fmt.Printf("Error verifying token: %v\n", err)
		return
	}
	fmt.Println("Token verification successful with auth.JWTManager!")
	
	// Test with router's getJWTSecret function indirectly
	// We can't directly call it, but we can test the same logic
	routerSecret := getRouterJWTSecret()
	routerJWTManager := auth.NewJWTManager(routerSecret, auth.TokenTTL)
	_, err = routerJWTManager.Verify(token)
	if err != nil {
		fmt.Printf("Error verifying token with router secret: %v\n", err)
		return
	}
	fmt.Println("Token verification successful with router JWT secret!")
	
	// Test with session.go's getJWTSecret function
	sessionSecret := getSessionJWTSecret()
	if jwtSecret == sessionSecret {
		fmt.Println("Session JWT secrets match!")
	} else {
		fmt.Printf("Warning: Session JWT secrets don't match!\n")
		fmt.Printf("  Environment: %.20s...\n", jwtSecret)
		fmt.Printf("  Session:     %.20s...\n", sessionSecret)
	}
}

// Simulate router's getJWTSecret function
func getRouterJWTSecret() string {
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	if jwtSecret == "" {
		return "default-secret-key-for-development"
	}
	return jwtSecret
}

// Simulate session.go's getJWTSecret function
func getSessionJWTSecret() string {
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	if jwtSecret == "" {
		return "default-secret-key-for-development"
	}
	return jwtSecret
}