package kdbx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

const repoTestPassword = "correct horse battery staple test only"

func writeVault(t *testing.T, path string, doc *Document) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create() error = %v", err)
	}
	if err := Save(f, doc); err != nil {
		f.Close()
		t.Fatalf("Save() error = %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing written file: %v", err)
	}
}

func openVault(t *testing.T, path, password string) *Document {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	doc, err := Open(f, gokeepasslib.NewPasswordCredentials(password))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return doc
}

func TestSaveThenOpenRoundTrips(t *testing.T) {
	root := domain.Group{ID: domain.NewGroupID(), Name: "My Vault"}
	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		t.Fatalf("NewVaultFromRoot() error = %v", err)
	}

	entry := domain.NewEntry(root.ID, "GitHub")
	entry.Username = "octocat"
	entry.Password = domain.NewSecretFromString("s3cr3t-P@ssw0rd")
	if err := vault.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Credentials = gokeepasslib.NewPasswordCredentials(repoTestPassword)

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	writeVault(t, path, &Document{Vault: vault, Database: db})

	doc := openVault(t, path, repoTestPassword)

	if !doc.Database.Header.IsKdbx41() {
		t.Error("reopened database is not KDBX 4.1")
	}
	if doc.Vault.RootGroupID() != root.ID {
		t.Errorf("RootGroupID() = %x, want %x", doc.Vault.RootGroupID(), root.ID)
	}

	got, err := doc.Vault.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if got.Title != entry.Title || got.Username != entry.Username {
		t.Errorf("entry = %+v, want Title=%q Username=%q", got, entry.Title, entry.Username)
	}
	gotPassword, err := revealString(got.Password)
	if err != nil {
		t.Fatalf("revealing password: %v", err)
	}
	if gotPassword != "s3cr3t-P@ssw0rd" {
		t.Errorf("Password = %q, want %q", gotPassword, "s3cr3t-P@ssw0rd")
	}
}

func TestOpenRejectsWrongPassword(t *testing.T) {
	root := domain.Group{ID: domain.NewGroupID(), Name: "My Vault"}
	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		t.Fatalf("NewVaultFromRoot() error = %v", err)
	}

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Credentials = gokeepasslib.NewPasswordCredentials(repoTestPassword)

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	writeVault(t, path, &Document{Vault: vault, Database: db})

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	if _, err := Open(f, gokeepasslib.NewPasswordCredentials("wrong password")); err == nil {
		t.Error("Open() with wrong password succeeded, want error")
	}
}

// TestSaveAfterEditReusesReopenedDatabase exercises the realistic edit
// cycle: open a file, add an entry to the decoded Vault, save it back out
// using the same *gokeepasslib.Database Open returned, then reopen and
// confirm both the original and the new entry survived. This is the case
// that would break if Save's Lock/Encode sequence relied on doc.Database
// still being in some pristine post-Decode state.
func TestSaveAfterEditReusesReopenedDatabase(t *testing.T) {
	root := domain.Group{ID: domain.NewGroupID(), Name: "My Vault"}
	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		t.Fatalf("NewVaultFromRoot() error = %v", err)
	}
	first := domain.NewEntry(root.ID, "GitHub")
	first.Password = domain.NewSecretFromString("first-secret")
	if err := vault.AddEntry(first); err != nil {
		t.Fatalf("AddEntry(first) error = %v", err)
	}

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Credentials = gokeepasslib.NewPasswordCredentials(repoTestPassword)

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	writeVault(t, path, &Document{Vault: vault, Database: db})

	doc := openVault(t, path, repoTestPassword)

	second := domain.NewEntry(root.ID, "GitLab")
	second.Password = domain.NewSecretFromString("second-secret")
	if err := doc.Vault.AddEntry(second); err != nil {
		t.Fatalf("AddEntry(second) error = %v", err)
	}

	writeVault(t, path, doc)

	reopened := openVault(t, path, repoTestPassword)

	gotFirst, err := reopened.Vault.Entry(first.ID)
	if err != nil {
		t.Fatalf("Entry(first) error = %v", err)
	}
	if gotFirst.Title != "GitHub" {
		t.Errorf("first.Title = %q, want %q", gotFirst.Title, "GitHub")
	}

	gotSecond, err := reopened.Vault.Entry(second.ID)
	if err != nil {
		t.Fatalf("Entry(second) error = %v", err)
	}
	if gotSecond.Title != "GitLab" {
		t.Errorf("second.Title = %q, want %q", gotSecond.Title, "GitLab")
	}
}
