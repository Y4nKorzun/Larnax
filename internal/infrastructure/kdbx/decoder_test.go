package kdbx

import (
	"errors"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestVaultFromDatabaseRejectsEmptyDatabase(t *testing.T) {
	if _, err := VaultFromDatabase(nil); !errors.Is(err, ErrEmptyDatabase) {
		t.Errorf("VaultFromDatabase(nil) error = %v, want %v", err, ErrEmptyDatabase)
	}

	if _, err := VaultFromDatabase(&gokeepasslib.Database{}); !errors.Is(err, ErrEmptyDatabase) {
		t.Errorf("VaultFromDatabase(empty) error = %v, want %v", err, ErrEmptyDatabase)
	}
}

func TestVaultFromDatabasePreservesRootIdentity(t *testing.T) {
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	root := &db.Content.Root.Groups[0]
	root.Entries = nil
	root.Name = "My Vault"
	root.Notes = "root notes"

	v, err := VaultFromDatabase(db)
	if err != nil {
		t.Fatalf("VaultFromDatabase() error = %v", err)
	}

	if v.RootGroupID() != domain.GroupID(root.UUID) {
		t.Errorf("RootGroupID() = %x, want %x", v.RootGroupID(), root.UUID)
	}
	got, err := v.Group(v.RootGroupID())
	if err != nil {
		t.Fatalf("Group(root) error = %v", err)
	}
	if got.Name != "My Vault" {
		t.Errorf("root.Name = %q, want %q", got.Name, "My Vault")
	}
	if got.Notes != "root notes" {
		t.Errorf("root.Notes = %q, want %q", got.Notes, "root notes")
	}
}

func TestVaultFromDatabaseWalksNestedGroupsAndEntries(t *testing.T) {
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	root := &db.Content.Root.Groups[0]
	root.Entries = nil

	child := gokeepasslib.NewGroup()
	child.Name = "Personal"

	grandchild := gokeepasslib.NewGroup()
	grandchild.Name = "Banking"

	entry := gokeepasslib.NewEntry()
	entry.Values = append(entry.Values,
		valueData(fieldTitle, "GitHub", false),
		valueData(fieldUserName, "octocat", false),
		valueData(fieldPassword, "hunter2", true),
	)

	child.Entries = append(child.Entries, entry)
	child.Groups = append(child.Groups, grandchild)
	root.Groups = append(root.Groups, child)

	v, err := VaultFromDatabase(db)
	if err != nil {
		t.Fatalf("VaultFromDatabase() error = %v", err)
	}

	childID := domain.GroupID(child.UUID)
	gotChild, err := v.Group(childID)
	if err != nil {
		t.Fatalf("Group(child) error = %v", err)
	}
	if gotChild.ParentID == nil || *gotChild.ParentID != v.RootGroupID() {
		t.Errorf("child.ParentID = %v, want %x", gotChild.ParentID, v.RootGroupID())
	}

	grandchildID := domain.GroupID(grandchild.UUID)
	gotGrandchild, err := v.Group(grandchildID)
	if err != nil {
		t.Fatalf("Group(grandchild) error = %v", err)
	}
	if gotGrandchild.ParentID == nil || *gotGrandchild.ParentID != childID {
		t.Errorf("grandchild.ParentID = %v, want %x", gotGrandchild.ParentID, childID)
	}

	entryID := domain.EntryID(entry.UUID)
	gotEntry, err := v.Entry(entryID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if gotEntry.ParentGroup != childID {
		t.Errorf("entry.ParentGroup = %x, want %x", gotEntry.ParentGroup, childID)
	}
	if gotEntry.Title != "GitHub" {
		t.Errorf("entry.Title = %q, want %q", gotEntry.Title, "GitHub")
	}
	if gotEntry.Username != "octocat" {
		t.Errorf("entry.Username = %q, want %q", gotEntry.Username, "octocat")
	}
}
