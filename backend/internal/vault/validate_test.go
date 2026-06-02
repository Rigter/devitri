package vault

import (
	"strings"
	"testing"
)

func TestValidateVaultID(t *testing.T) {
	valid := []string{"Devitri", "my-vault", "vault2", "a"}
	for _, id := range valid {
		if err := ValidateVaultID(id); err != nil {
			t.Errorf("expected valid vault_id %q, got %v", id, err)
		}
	}

	invalid := []string{"", "..", "../x", "a/b", "vault.id", ".hidden", strings.Repeat("a", 65)}
	for _, id := range invalid {
		if err := ValidateVaultID(id); err == nil {
			t.Errorf("expected invalid vault_id %q", id)
		}
	}
}

func TestVaultDirectoryPath(t *testing.T) {
	path, err := VaultDirectoryPath("Devitri")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	_, err = VaultDirectoryPath("../etc")
	if err == nil {
		t.Fatal("expected error for traversal vault_id")
	}
}

func TestPathUsesObsidianConfig(t *testing.T) {
	cases := []string{".obsidian/config", "notes/.obsidian/plugins/x", ".obsidian"}
	for _, p := range cases {
		if !pathUsesObsidianConfig(p) {
			t.Errorf("expected obsidian path blocked: %s", p)
		}
	}
	if pathUsesObsidianConfig("notes/readme.md") {
		t.Error("expected normal path allowed")
	}
}
