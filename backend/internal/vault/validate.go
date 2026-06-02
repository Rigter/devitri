package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// VaultIDPattern matches safe vault identifiers (plugin / Obsidian vault names).
var VaultIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

const maxFilePathLen = 1024

// ValidateVaultID rejects path traversal and unsafe vault identifiers.
func ValidateVaultID(vaultID string) error {
	if vaultID == "" {
		return fmt.Errorf("invalid vault_id: empty")
	}
	if len(vaultID) > 64 {
		return fmt.Errorf("invalid vault_id: too long")
	}
	if vaultID == "." || vaultID == ".." {
		return fmt.Errorf("invalid vault_id: reserved name")
	}
	if strings.Contains(vaultID, "/") || strings.Contains(vaultID, "\\") {
		return fmt.Errorf("invalid vault_id: path separators not allowed")
	}
	if !VaultIDPattern.MatchString(vaultID) {
		return fmt.Errorf("invalid vault_id: must match [a-zA-Z0-9][a-zA-Z0-9_-]{0,63}")
	}
	return nil
}

// ResolveVaultsRoot returns the directory where vault data is stored.
func ResolveVaultsRoot() string {
	if _, err := os.Stat("/vaults"); err == nil {
		return "/vaults"
	}
	if _, err := os.Stat("./vaults"); err == nil {
		return "./vaults"
	}
	return "/vaults"
}

// VaultDirectoryPath returns the absolute path for a vault after validation.
func VaultDirectoryPath(vaultID string) (string, error) {
	if err := ValidateVaultID(vaultID); err != nil {
		return "", err
	}
	root := filepath.Clean(ResolveVaultsRoot())
	joined := filepath.Join(root, vaultID)
	cleaned := filepath.Clean(joined)
	rel, err := filepath.Rel(root, cleaned)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("invalid vault_id: path escapes vaults root")
	}
	return cleaned, nil
}

// pathUsesObsidianConfig reports whether a normalized relative path touches .obsidian.
func pathUsesObsidianConfig(filePath string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(filePath))
	if cleaned == ".obsidian" || strings.HasPrefix(cleaned, ".obsidian/") {
		return true
	}
	for _, segment := range strings.Split(cleaned, "/") {
		if segment == ".obsidian" {
			return true
		}
	}
	return false
}
