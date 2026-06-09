package api

import (
	"database/sql"

	"github.com/rigter/devitri/backend/internal/auth"
	"github.com/rigter/devitri/backend/internal/config"
	"github.com/rigter/devitri/backend/internal/vault"
	"github.com/rigter/devitri/backend/internal/api/handlers"
	"github.com/rigter/devitri/backend/internal/api/middleware"
	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the HTTP router.
// jwtSecret must come from config.ResolveJWTSecret() at process startup.
func SetupRouter(db *sql.DB, jwtSecret string) *mux.Router {
	// Create services
	sessionStore := auth.NewSessionStore(db, jwtSecret)
	vaultManager := vault.NewManager(db)
	
	authHandler := handlers.NewAuthHandler(sessionStore)
	syncHandler := handlers.NewSyncHandler(db)
	vaultHandler := handlers.NewVaultHandler(vaultManager)
	setupHandler := handlers.NewSetupHandler()
	settingsHandler := handlers.NewSettingsHandler()
	
	// Create middleware
	rateLimiter := middleware.NewRateLimiter(config.Current.LoginRateLimitPerMinute)
	authMiddleware := middleware.NewAuthMiddleware(sessionStore, jwtSecret)
	
	// Create router
	router := mux.NewRouter()
	
	// Apply security headers to all routes
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.FirstRun)
	
	// Health check endpoint (no auth required)
	router.HandleFunc("/health", authHandler.HealthCheck).Methods("GET")

	// Setup routes (Public)
	setupRouter := router.PathPrefix("/api/setup").Subrouter()
	setupRouter.Use(middleware.SetupGuard)
	setupRouter.HandleFunc("/check", setupHandler.CheckSetup).Methods("GET")
	setupMutations := setupRouter.NewRoute().Subrouter()
	setupMutations.Use(rateLimiter.RateLimit)
	setupMutations.HandleFunc("/generate", setupHandler.GenerateConfig).Methods("POST")
	setupMutations.HandleFunc("/generate-master-hash", setupHandler.GenerateMasterHash).Methods("POST")
	setupMutations.HandleFunc("/generate-jwt-secret", setupHandler.GenerateJWTSecret).Methods("POST")
	
	// Public auth routes (with rate limiting)
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	authRouter.Use(rateLimiter.RateLimit)
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	
	// Protected auth routes (require authentication)
	protectedAuthRouter := authRouter.NewRoute().Subrouter()
	protectedAuthRouter.Use(authMiddleware.Auth)
	protectedAuthRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	protectedAuthRouter.HandleFunc("/session", authHandler.Session).Methods("GET")
	protectedAuthRouter.HandleFunc("/sessions", authHandler.ListSessions).Methods("GET")
	protectedAuthRouter.HandleFunc("/sessions/{id}", authHandler.RevokeSession).Methods("DELETE")
	protectedAuthRouter.HandleFunc("/token", authHandler.GenerateToken).Methods("POST")
	
	// Vault routes (require authentication)
	vaultsRouter := router.PathPrefix("/api/vaults").Subrouter()
	vaultsRouter.Use(authMiddleware.Auth)
	vaultsRouter.HandleFunc("", vaultHandler.ListVaults).Methods("GET")
	vaultsRouter.HandleFunc("/{vault_id}", vaultHandler.GetVault).Methods("GET")
	
	// Protected sync routes (require authentication)
	syncRouter := router.PathPrefix("/api/vaults/{vault_id}/sync").Subrouter()
	syncRouter.Use(authMiddleware.Auth)
	syncRouter.HandleFunc("/manifest", syncHandler.GetManifest).Methods("GET")
	syncRouter.HandleFunc("/upload", syncHandler.UploadFile).Methods("POST")
	syncRouter.HandleFunc("/download", syncHandler.DownloadFile).Methods("GET")
	syncRouter.HandleFunc("/file", syncHandler.DeleteFile).Methods("DELETE")
	syncRouter.HandleFunc("/batch", syncHandler.BatchSync).Methods("POST")

	// Settings (require authentication)
	settingsRouter := router.PathPrefix("/api/settings").Subrouter()
	settingsRouter.Use(authMiddleware.Auth)
	settingsRouter.HandleFunc("", settingsHandler.GetSettings).Methods("GET")
	
	return router
}