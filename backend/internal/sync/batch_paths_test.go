package sync

import (
	"path/filepath"
	"testing"

	"github.com/rigter/devitri/backend/internal/vault"
)

func TestValidateBatchPaths(t *testing.T) {
	root := t.TempDir()
	vaultsRoot := filepath.Join(root, "vaults")
	vaultDir := filepath.Join(vaultsRoot, "test-vault")
	if err := vault.ValidateVaultID("test-vault"); err != nil {
		t.Fatal(err)
	}

	v := &vault.Vault{
		VaultID: "test-vault",
		Path:    vaultDir,
	}

	svc := &SyncService{vaults: vault.NewManager(nil)}

	if err := svc.validateBatchPaths(v, []FileMetadata{
		{Path: "notes/ok.md", Hash: "a", Size: 1, ModifiedAt: 1},
	}); err != nil {
		t.Fatalf("expected valid path: %v", err)
	}

	if err := svc.validateBatchPaths(v, []FileMetadata{
		{Path: "../escape.md", Hash: "a", Size: 1, ModifiedAt: 1},
	}); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}

	if err := svc.validateBatchPaths(v, []FileMetadata{
		{Path: ".obsidian/config", Hash: "a", Size: 1, ModifiedAt: 1},
	}); err == nil {
		t.Fatal("expected .obsidian path to be rejected")
	}

	if err := svc.validateBatchPaths(v, []FileMetadata{
		{Path: "a.md", Hash: "a", Size: 1, ModifiedAt: 1},
		{Path: "a.md", Hash: "b", Size: 1, ModifiedAt: 1},
	}); err == nil {
		t.Fatal("expected duplicate path to be rejected")
	}
}
