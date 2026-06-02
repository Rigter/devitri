package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database
func InitDB() (*sql.DB, error) {
	// Create data directory if it doesn't exist
	dataDir := "/data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		// For development, use a local directory
		dataDir = "./data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Database file path
	dbPath := filepath.Join(dataDir, "devitri.db")

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations executes all database migrations
func runMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name    TEXT    NOT NULL,
		applied_at INTEGER NOT NULL
	);
	`
	_, err := db.Exec(createMigrationsTable)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Find migration files
	migrationsDir := "internal/db/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try alternative path (for development)
		migrationsDir = "backend/internal/db/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found: %s", migrationsDir)
		}
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}

	sort.Strings(migrationFiles)

	// Apply migrations in order
	for _, filename := range migrationFiles {
		// Extract version number from filename (e.g., "001_init.sql" -> 1)
		var version int
		_, err := fmt.Sscanf(filename, "%d_", &version)
		if err != nil {
			return fmt.Errorf("invalid migration filename: %s", filename)
		}

		// Check if migration has already been applied
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}

		if exists {
			continue
		}

		fmt.Printf("Applying migration: %s\n", filename)

		// Read migration file
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			version, filename, time.Now().Unix())
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}
	}

	return nil
}