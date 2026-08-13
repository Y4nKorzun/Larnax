package application

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/filesystem"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

const saveVaultTestPassword = "correct horse battery staple test only"

func newTestDoc(t *testing.T) *kdbx.Document {
	t.Helper()
	doc, err := CreateVault("My Vault", saveVaultTestPassword)
	if err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}
	return doc
}

func TestSaveVaultWritesFileOpenVaultCanRead(t *testing.T) {
	doc := newTestDoc(t)
	entry := domain.NewEntry(doc.Vault.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := doc.Vault.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if _, err := SaveVault(path, doc, filesystem.FileRevision{}, 0); err != nil {
		t.Fatalf("SaveVault() error = %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	opened, err := OpenVault(f, saveVaultTestPassword)
	if err != nil {
		t.Fatalf("OpenVault() error = %v", err)
	}
	got, err := opened.Document.Vault.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub")
	}
}

func TestSaveVaultReturnsRevisionMatchingDisk(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	rev, err := SaveVault(path, doc, filesystem.FileRevision{}, 0)
	if err != nil {
		t.Fatalf("SaveVault() error = %v", err)
	}

	onDisk, err := filesystem.ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}
	if rev.Hash != onDisk.Hash {
		t.Error("returned revision hash does not match the file just written")
	}
}

func TestSaveVaultDetectsExternalModification(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	rev, err := SaveVault(path, doc, filesystem.FileRevision{}, 0)
	if err != nil {
		t.Fatalf("first SaveVault() error = %v", err)
	}

	// Simulate another process (or another KDBX client) touching the file
	// after rev was read.
	if err := os.WriteFile(path, []byte("tampered by another process"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := SaveVault(path, doc, rev, 0); !errors.Is(err, ErrExternalModification) {
		t.Errorf("second SaveVault() error = %v, want %v", err, ErrExternalModification)
	}
}

func TestSaveVaultSkipsRevisionCheckForNewVault(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	// The zero FileRevision (Path == "") must not be treated as "path
	// existed and was empty" — there is nothing on disk yet to conflict
	// with.
	if _, err := SaveVault(path, doc, filesystem.FileRevision{}, 0); err != nil {
		t.Fatalf("SaveVault() on brand-new path error = %v", err)
	}
}

func TestSaveVaultCreatesBackupOnSecondSave(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	rev, err := SaveVault(path, doc, filesystem.FileRevision{}, 5)
	if err != nil {
		t.Fatalf("first SaveVault() error = %v", err)
	}

	entry := domain.NewEntry(doc.Vault.RootGroupID(), "Second Entry")
	if err := doc.Vault.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if _, err := SaveVault(path, doc, rev, 5); err != nil {
		t.Fatalf("second SaveVault() error = %v", err)
	}

	backups := filesystem.NewBackupStore(path, 5)
	names, err := backups.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(names) != 1 {
		t.Errorf("backup count = %d, want 1 (only the second save has a prior file to back up)", len(names))
	}
}

func TestSaveVaultSkipsBackupWhenRetentionZero(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	rev, err := SaveVault(path, doc, filesystem.FileRevision{}, 0)
	if err != nil {
		t.Fatalf("first SaveVault() error = %v", err)
	}
	if _, err := SaveVault(path, doc, rev, 0); err != nil {
		t.Fatalf("second SaveVault() error = %v", err)
	}

	backups := filesystem.NewBackupStore(path, 0)
	names, err := backups.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("backup count = %d, want 0 (retention disabled)", len(names))
	}
}

func TestSaveVaultReleasesLockOnSuccess(t *testing.T) {
	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	if _, err := SaveVault(path, doc, filesystem.FileRevision{}, 0); err != nil {
		t.Fatalf("SaveVault() error = %v", err)
	}

	lock, err := filesystem.Acquire(path)
	if err != nil {
		t.Fatalf("Acquire() after SaveVault() error = %v, want the lock to be free", err)
	}
	lock.Release()
}

func TestSaveVaultRestrictsPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	doc := newTestDoc(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")

	if _, err := SaveVault(path, doc, filesystem.FileRevision{}, 0); err != nil {
		t.Fatalf("SaveVault() error = %v", err)
	}

	ok, err := filesystem.Verify(path, filesystem.SecretFileMode)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Error("saved vault does not have SecretFileMode permissions")
	}
}
