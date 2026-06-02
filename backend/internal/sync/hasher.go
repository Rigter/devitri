package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateSHA256 calculates the SHA-256 hash of a file
func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// CalculateSHA256FromReader calculates the SHA-256 hash from an io.Reader
func CalculateSHA256FromReader(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("failed to hash from reader: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}