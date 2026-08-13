package application

import (
	"errors"
	"testing"
)

func TestCreateVaultRejectsEmptyPassphrase(t *testing.T) {
	if _, err := CreateVault("My Vault", ""); !errors.Is(err, ErrEmptyMasterPassphrase) {
		t.Errorf("CreateVault() error = %v, want %v", err, ErrEmptyMasterPassphrase)
	}
}

func TestCreateVaultNamesRootGroup(t *testing.T) {
	doc, err := CreateVault("My Vault", "correct horse battery staple")
	if err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}

	root, err := doc.Vault.Group(doc.Vault.RootGroupID())
	if err != nil {
		t.Fatalf("Group(RootGroupID()) error = %v", err)
	}
	if root.Name != "My Vault" {
		t.Errorf("root.Name = %q, want %q", root.Name, "My Vault")
	}
}

func TestCreateVaultProducesEmptyVault(t *testing.T) {
	doc, err := CreateVault("My Vault", "correct horse battery staple")
	if err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}

	if got := len(doc.Vault.ChildGroups(doc.Vault.RootGroupID())); got != 0 {
		t.Errorf("ChildGroups(root) = %d, want 0", got)
	}
	if got := len(doc.Vault.EntriesIn(doc.Vault.RootGroupID())); got != 0 {
		t.Errorf("EntriesIn(root) = %d, want 0", got)
	}
}
