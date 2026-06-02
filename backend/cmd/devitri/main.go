package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rigter/devitri/backend/internal/api"
	"github.com/rigter/devitri/backend/internal/api/middleware"
	"github.com/rigter/devitri/backend/internal/config"
	"github.com/rigter/devitri/backend/internal/db"
	"github.com/rigter/devitri/backend/internal/setup"
	"github.com/rigter/devitri/backend/internal/vault"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	envPath := filepath.Join("..", "..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	} else {
		// Try to load from current directory
		if _, err := os.Stat(".env"); err == nil {
			if err := godotenv.Load(); err != nil {
				log.Printf("Warning: Error loading .env file: %v", err)
			}
		}
	}

	cfg := config.Load()
	if err := config.ValidateStartupSecrets(setup.IsConfigured()); err != nil {
		log.Fatalf("Security configuration error: %v", err)
	}

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initial vault scan
	vaultManager := vault.NewManager(database)
	if err := vaultManager.ScanVaults(); err != nil {
		log.Printf("Warning: Initial vault scan failed: %v", err)
	}

	jwtSecret := ""
	if setup.IsConfigured() || config.AllowInsecureJWT() {
		resolved, err := config.ResolveJWTSecret()
		if err != nil {
			log.Fatalf("JWT configuration error: %v", err)
		}
		jwtSecret = resolved
	} else {
		// First-run: generate an ephemeral secret so the router can initialize.
		// Auth is blocked by FirstRun middleware until DEVITRI_* are configured.
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			log.Fatalf("JWT configuration error: failed to generate ephemeral secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(secretBytes)
	}
	if setup.IsConfigured() && cfg.JWTSecretConfigured {
		log.Println("Devitri: JWT secret loaded from environment")
	} else if config.AllowInsecureJWT() {
		log.Println("Warning: DEVITRI_ALLOW_INSECURE_JWT=true — using development JWT secret")
	}

	// Setup router (CORS wraps the router so OPTIONS preflight works before route matching)
	router := api.SetupRouter(database, jwtSecret)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.CORS(router),
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Starting Devitri backend server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}