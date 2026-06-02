//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load("test.env"); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	masterHash := os.Getenv("DEVITRI_MASTER_HASH")
	jwtSecret := os.Getenv("DEVITRI_JWT_SECRET")
	
	fmt.Printf("DEVITRI_MASTER_HASH: '%s'\n", masterHash)
	fmt.Printf("DEVITRI_JWT_SECRET: '%s'\n", jwtSecret)
}