package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rigter/devitri/backend/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

// Session represents a user session in the database
type Session struct {
	ID         int64  `db:"id"`
	TokenHash  string `db:"token_hash"`
	DeviceID   string `db:"device_id"`
	DeviceName string `db:"device_name"`
	VaultID    *string `db:"vault_id"`
	CreatedAt  int64  `db:"created_at"`
	ExpiresAt  int64  `db:"expires_at"`
	LastSeen   int64  `db:"last_seen"`
}

// SessionStore handles session persistence
type SessionStore struct {
	db        *sql.DB
	jwtSecret string
}

// NewSessionStore creates a new SessionStore
func NewSessionStore(db *sql.DB, jwtSecret string) *SessionStore {
	return &SessionStore{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// CreateSession creates a new session in the database with default TTL
func (s *SessionStore) CreateSession(deviceID, deviceName string, vaultID *string) (*Session, string, error) {
	return s.CreateSessionWithTTL(deviceID, deviceName, vaultID, config.Current.SessionTTL())
}

// CreateSessionWithTTL creates a new session in the database with a custom TTL
func (s *SessionStore) CreateSessionWithTTL(deviceID, deviceName string, vaultID *string, ttl time.Duration) (*Session, string, error) {
	// Generate a new JWT token
	jwtManager := NewJWTManager(s.jwtSecret, ttl)
	token, err := jwtManager.Generate()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token for storage
	hasher := sha256.New()
	hasher.Write([]byte(token))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// Create session struct
	now := time.Now().Unix()
	session := &Session{
		TokenHash:  tokenHash,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		VaultID:    vaultID,
		CreatedAt:  now,
		ExpiresAt:  now + int64(ttl.Seconds()),
		LastSeen:   now,
	}

	// Insert into database
	query := `
		INSERT INTO sessions (token_hash, device_id, device_name, vault_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query, session.TokenHash, session.DeviceID, session.DeviceName, session.VaultID, 
		session.CreatedAt, session.ExpiresAt, session.LastSeen)
	if err != nil {
		return nil, "", fmt.Errorf("failed to insert session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get last insert ID: %w", err)
	}
	session.ID = id

	return session, token, nil
}

// ListSessions retrieves all sessions from the database
func (s *SessionStore) ListSessions() ([]*Session, error) {
	query := `
		SELECT id, token_hash, device_id, device_name, vault_id, created_at, expires_at, last_seen
		FROM sessions
		ORDER BY last_seen DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(&session.ID, &session.TokenHash, &session.DeviceID, &session.DeviceName, 
			&session.VaultID, &session.CreatedAt, &session.ExpiresAt, &session.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetSessionByTokenHash retrieves a session by its token hash
func (s *SessionStore) GetSessionByTokenHash(tokenHash string) (*Session, error) {
	query := `
		SELECT id, token_hash, device_id, device_name, vault_id, created_at, expires_at, last_seen
		FROM sessions
		WHERE token_hash = ?
	`
	row := s.db.QueryRow(query, tokenHash)

	session := &Session{}
	err := row.Scan(&session.ID, &session.TokenHash, &session.DeviceID, &session.DeviceName, 
		&session.VaultID, &session.CreatedAt, &session.ExpiresAt, &session.LastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to scan session: %w", err)
	}

	// Check if session has expired
	if time.Now().Unix() > session.ExpiresAt {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// DeleteSessionByID deletes a session by its ID
func (s *SessionStore) DeleteSessionByID(id int64) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// DeleteSession deletes a session by its token hash
func (s *SessionStore) DeleteSession(tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = ?`
	_, err := s.db.Exec(query, tokenHash)
	return err
}

// UpdateLastSeen updates the last_seen timestamp for a session
func (s *SessionStore) UpdateLastSeen(tokenHash string) error {
	query := `UPDATE sessions SET last_seen = ? WHERE token_hash = ?`
	_, err := s.db.Exec(query, time.Now().Unix(), tokenHash)
	return err
}