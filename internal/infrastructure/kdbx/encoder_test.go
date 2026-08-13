package kdbx

import (
	"testing"

	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// buildSampleVault returns a vault shaped root -> Personal -> Banking,
// with one entry ("GitHub") directly inside Personal, so tests can assert
// on both nesting depth and where entries land.
func buildSampleVault(t *testing.T) (*domain.Vault, domain.Group, domain.Group, domain.Group, domain.Entry) {
	t.Helper()

	root := domain.Group{ID: domain.NewGroupID(), Name: "My Vault", Notes: "root notes"}
	v, err := domain.NewVaultFromRoot(root)
	if err != nil {
		t.Fatalf("NewVaultFromRoot() error = %v", err)
	}

	child := domain.NewGroup(root.ID, "Personal")
	if err := v.AddGroup(child); err != nil {
		t.Fatalf("AddGroup(child) error = %v", err)
	}

	grandchild := domain.NewGroup(child.ID, "Banking")
	if err := v.AddGroup(grandchild); err != nil {
		t.Fatalf("AddGroup(grandchild) error = %v", err)
	}

	entry := domain.NewEntry(child.ID, "GitHub")
	entry.Username = "octocat"
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	return v, root, child, grandchild, entry
}

func TestRootGroupFromVaultPreservesRootIdentity(t *testing.T) {
	v, root, _, _, _ := buildSampleVault(t)

	gg, err := RootGroupFromVault(v)
	if err != nil {
		t.Fatalf("RootGroupFromVault() error = %v", err)
	}

	if [16]byte(gg.UUID) != [16]byte(root.ID) {
		t.Errorf("UUID = %x, want %x", gg.UUID, root.ID)
	}
	if gg.Name != root.Name {
		t.Errorf("Name = %q, want %q", gg.Name, root.Name)
	}
	if gg.Notes != root.Notes {
		t.Errorf("Notes = %q, want %q", gg.Notes, root.Notes)
	}
}

func TestRootGroupFromVaultNestsChildGroupsAndEntries(t *testing.T) {
	v, _, child, grandchild, entry := buildSampleVault(t)

	gg, err := RootGroupFromVault(v)
	if err != nil {
		t.Fatalf("RootGroupFromVault() error = %v", err)
	}

	if got := len(gg.Groups); got != 1 {
		t.Fatalf("root has %d child groups, want 1", got)
	}
	gotChild := gg.Groups[0]
	if gotChild.Name != child.Name {
		t.Errorf("child.Name = %q, want %q", gotChild.Name, child.Name)
	}

	if got := len(gotChild.Groups); got != 1 {
		t.Fatalf("child has %d nested groups, want 1", got)
	}
	if gotChild.Groups[0].Name != grandchild.Name {
		t.Errorf("grandchild.Name = %q, want %q", gotChild.Groups[0].Name, grandchild.Name)
	}

	if got := len(gotChild.Entries); got != 1 {
		t.Fatalf("child has %d entries, want 1", got)
	}
	if gotChild.Entries[0].GetTitle() != entry.Title {
		t.Errorf("entry.Title = %q, want %q", gotChild.Entries[0].GetTitle(), entry.Title)
	}
	if gotChild.Entries[0].GetPassword() != "hunter2" {
		t.Errorf("entry.Password = %q, want %q", gotChild.Entries[0].GetPassword(), "hunter2")
	}
}

// TestEncodeThenDecodeRoundTrips is the in-memory analogue of
// roundtrip_test.go's file-level check, but exercising this package's own
// mapper/encoder/decoder instead of gokeepasslib's raw types directly: a
// vault mapped out to a Database's tree and back in must describe the
// same groups, nesting and entries it started with.
func TestEncodeThenDecodeRoundTrips(t *testing.T) {
	v, root, child, grandchild, entry := buildSampleVault(t)

	gg, err := RootGroupFromVault(v)
	if err != nil {
		t.Fatalf("RootGroupFromVault() error = %v", err)
	}

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Content.Root.Groups = []gokeepasslib.Group{gg}

	got, err := VaultFromDatabase(db)
	if err != nil {
		t.Fatalf("VaultFromDatabase() error = %v", err)
	}

	if got.RootGroupID() != root.ID {
		t.Errorf("RootGroupID() = %x, want %x", got.RootGroupID(), root.ID)
	}

	gotChild, err := got.Group(child.ID)
	if err != nil {
		t.Fatalf("Group(child) error = %v", err)
	}
	if gotChild.ParentID == nil || *gotChild.ParentID != root.ID {
		t.Errorf("child.ParentID = %v, want %x", gotChild.ParentID, root.ID)
	}

	gotGrandchild, err := got.Group(grandchild.ID)
	if err != nil {
		t.Fatalf("Group(grandchild) error = %v", err)
	}
	if gotGrandchild.ParentID == nil || *gotGrandchild.ParentID != child.ID {
		t.Errorf("grandchild.ParentID = %v, want %x", gotGrandchild.ParentID, child.ID)
	}

	gotEntry, err := got.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if gotEntry.ParentGroup != child.ID {
		t.Errorf("entry.ParentGroup = %x, want %x", gotEntry.ParentGroup, child.ID)
	}
	if gotEntry.Title != entry.Title || gotEntry.Username != entry.Username {
		t.Errorf("entry = %+v, want Title=%q Username=%q", gotEntry, entry.Title, entry.Username)
	}
}
