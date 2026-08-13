package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/importers/googlecsv"
)

const vaultServiceTestPassword = "correct horse battery staple test only"

func TestVaultServiceMethodsRequireOpenSession(t *testing.T) {
	var s VaultService

	if s.IsOpen() {
		t.Error("IsOpen() = true before New/Open")
	}
	if err := s.Save(0); !errors.Is(err, ErrVaultNotOpen) {
		t.Errorf("Save() error = %v, want %v", err, ErrVaultNotOpen)
	}
	if err := s.Lock(); !errors.Is(err, ErrVaultNotOpen) {
		t.Errorf("Lock() error = %v, want %v", err, ErrVaultNotOpen)
	}
	if err := s.AddEntry(domain.Entry{}); !errors.Is(err, ErrVaultNotOpen) {
		t.Errorf("AddEntry() error = %v, want %v", err, ErrVaultNotOpen)
	}
	if _, err := s.Undo(); !errors.Is(err, ErrVaultNotOpen) {
		t.Errorf("Undo() error = %v, want %v", err, ErrVaultNotOpen)
	}
}

func TestVaultServiceVaultPanicsWhenNotOpen(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Vault() with no session open did not panic")
		}
	}()
	var s VaultService
	s.Vault()
}

func TestVaultServiceNewStartsWritableSession(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !s.IsOpen() {
		t.Error("IsOpen() = false after New()")
	}
	if s.ReadOnly() {
		t.Error("ReadOnly() = true after New()")
	}
	if s.Path() != "" {
		t.Errorf("Path() = %q, want empty before any save", s.Path())
	}
}

func TestVaultServiceSaveWithoutPathFails(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := s.Save(0); !errors.Is(err, ErrNoSavePath) {
		t.Errorf("Save() error = %v, want %v", err, ErrNoSavePath)
	}
}

func TestVaultServiceSaveAsThenSaveRoundTrips(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := s.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	if s.Path() != path {
		t.Errorf("Path() = %q, want %q", s.Path(), path)
	}

	entry := domain.NewEntry(s.Vault().RootGroupID(), "GitHub")
	if err := s.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if err := s.Save(0); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	var reopened VaultService
	if err := reopened.Open(f, path, vaultServiceTestPassword); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if reopened.ReadOnly() {
		t.Error("ReadOnly() = true for a clean, freshly-saved vault")
	}
	if _, err := reopened.Vault().Entry(entry.ID); err != nil {
		t.Errorf("Entry() error = %v", err)
	}
}

func TestVaultServiceOpenFixtureWithAttachmentIsReadOnlyAndRejectsWrites(t *testing.T) {
	const fixturePath = "../../testdata/kdbx/keepass/kdbx4-example.kdbx"
	const fixturePassword = "abcdefg12345678"

	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	var s VaultService
	if err := s.Open(f, fixturePath, fixturePassword); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !s.ReadOnly() {
		t.Fatal("ReadOnly() = false, want true")
	}

	entry := domain.NewEntry(s.Vault().RootGroupID(), "Should not be allowed")
	if err := s.AddEntry(entry); !errors.Is(err, ErrReadOnly) {
		t.Errorf("AddEntry() error = %v, want %v", err, ErrReadOnly)
	}
	if err := s.Save(0); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Save() error = %v, want %v", err, ErrReadOnly)
	}
	if _, err := s.Import(s.Vault().RootGroupID(), ImportPlan{}); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Import() error = %v, want %v", err, ErrReadOnly)
	}
}

func TestVaultServiceUndoRedo(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	entry := domain.NewEntry(s.Vault().RootGroupID(), "GitHub")
	if err := s.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if ok, err := s.Undo(); err != nil || !ok {
		t.Fatalf("Undo() = (%v, %v), want (true, nil)", ok, err)
	}
	if _, err := s.Vault().Entry(entry.ID); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Errorf("Entry() after Undo() error = %v, want %v", err, domain.ErrEntryNotFound)
	}
	if ok, err := s.Redo(); err != nil || !ok {
		t.Fatalf("Redo() = (%v, %v), want (true, nil)", ok, err)
	}
	if _, err := s.Vault().Entry(entry.ID); err != nil {
		t.Errorf("Entry() after Redo() error = %v", err)
	}
}

func TestVaultServiceLockClearsSecretsAndHistory(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	entry := domain.NewEntry(s.Vault().RootGroupID(), "GitHub")
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := s.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	if err := s.Lock(); err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	got, err := s.Vault().Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if err := got.Password.Reveal(func([]byte) error { return nil }); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() after Lock() error = %v, want %v", err, domain.ErrSecretCleared)
	}
	if ok, _ := s.Undo(); ok {
		t.Error("Undo() succeeded after Lock(), want history cleared")
	}
}

func TestVaultServiceCopyField(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	entry := domain.NewEntry(s.Vault().RootGroupID(), "GitHub")
	entry.Username = "octocat"

	cb := &fakeClipboard{}
	if _, err := s.CopyField(context.Background(), cb, entry, FieldUsername); err != nil {
		t.Fatalf("CopyField() error = %v", err)
	}
	if string(cb.content) != "octocat" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "octocat")
	}
}

func TestVaultServiceImport(t *testing.T) {
	var s VaultService
	if err := s.New("My Vault", vaultServiceTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	plan := ImportPlan{New: []googlecsv.ImportedEntry{{
		Title:    "GitHub",
		Password: domain.NewSecretFromString("hunter2"),
	}}}
	result, err := s.Import(s.Vault().RootGroupID(), plan)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("Imported = %d, want 1", result.Imported)
	}
	if len(s.Vault().AllEntries()) != 1 {
		t.Errorf("vault has %d entries, want 1", len(s.Vault().AllEntries()))
	}
}
