//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file
	envPath := filepath.Join("..", ".env")
	fmt.Printf("Looking for .env file at: %s\n", envPath)
	
	if _, err := os.Stat(envPath); err == nil {
		fmt.Println(".env file found, loading...")
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	} else {
		fmt.Println(".env file not found at expected location")
		// Try current directory
		if _, err := os.Stat(".env"); err == nil {
			fmt.Println(".env file found in current directory, loading...")
			if err := godotenv.Load(); err != nil {
				log.Printf("Error loading .env file: %v", err)
			}
		} else {
			fmt.Println(".env file not found in current directory either")
		}
	}

	masterHash := os.Getenv("DEVITRI_MASTER_HASH")
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	
	fmt.Printf("DEVITRI_MASTER_HASH: '%s'\n", masterHash)
	fmt.Printf("DEVITRI_JWT_SECRET: '%s'\n", jwtSecret)
	
	if masterHash == "" {
		fmt.Println("DEVITRI_MASTER_HASH is empty!")
	}
	
	if jwtSecret == "" {
		fmt.Println("DEVITRI_JWT_SECRET is empty!")
	}
}