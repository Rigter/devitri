//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Get the master hash from environment
	masterHash := os.Getenv("DEVITRI_MASTER_HASH")
	if masterHash == "" {
		fmt.Println("DEVITRI_MASTER_HASH not set")
		return
	}
	
	// Test password
	password := "D3v1tr1$$"
	
	// Compare password with hash
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", masterHash)
	
	err := bcrypt.CompareHashAndPassword([]byte(masterHash), []byte(password))
	if err != nil {
		fmt.Printf("Match: false (error: %v)\n", err)
	} else {
		fmt.Printf("Match: true\n")
	}
}
