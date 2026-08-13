package tui

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
)

const messagesTestPassword = "correct horse battery staple test only"

// fakeClipboard is an in-memory clipboard.Clipboard — there is no real OS
// backend available in every test environment (only the darwin adapter
// exists so far), matching the fake application's own copy_field_test.go
// uses.
type fakeClipboard struct {
	content []byte
}

func (f *fakeClipboard) ReadText(context.Context) ([]byte, error) { return f.content, nil }
func (f *fakeClipboard) WriteText(_ context.Context, value []byte) error {
	f.content = value
	return nil
}
func (f *fakeClipboard) Clear(context.Context) error { f.content = nil; return nil }

func newOpenService(t *testing.T) *application.VaultService {
	t.Helper()
	var s application.VaultService
	if err := s.New("My Vault", messagesTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return &s
}

func TestCopyFieldCmdReturnsCompletedMsg(t *testing.T) {
	service := newOpenService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Username = "octocat"

	cb := &fakeClipboard{}
	cmd := CopyFieldCmd(context.Background(), service, cb, CopyFieldIntent{Entry: entry, Field: application.FieldUsername})

	msg := cmd()
	completed, ok := msg.(CopyCompletedMsg)
	if !ok {
		t.Fatalf("Cmd() returned %T, want CopyCompletedMsg", msg)
	}
	if completed.Err != nil {
		t.Errorf("Err = %v, want nil", completed.Err)
	}
	if completed.Field != application.FieldUsername {
		t.Errorf("Field = %v, want %v", completed.Field, application.FieldUsername)
	}
	if string(cb.content) != "octocat" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "octocat")
	}
}

func TestSaveCmdReturnsCompletedMsg(t *testing.T) {
	service := newOpenService(t)
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := service.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	msg := SaveCmd(service, SaveIntent{Retention: 0})()
	completed, ok := msg.(SaveCompletedMsg)
	if !ok {
		t.Fatalf("Cmd() returned %T, want SaveCompletedMsg", msg)
	}
	if completed.Err != nil {
		t.Errorf("Err = %v, want nil", completed.Err)
	}
}

func TestSaveCmdPropagatesNoSavePathError(t *testing.T) {
	service := newOpenService(t) // never given a path via SaveAs

	msg := SaveCmd(service, SaveIntent{Retention: 0})()
	completed, ok := msg.(SaveCompletedMsg)
	if !ok {
		t.Fatalf("Cmd() returned %T, want SaveCompletedMsg", msg)
	}
	if !errors.Is(completed.Err, application.ErrNoSavePath) {
		t.Errorf("Err = %v, want %v", completed.Err, application.ErrNoSavePath)
	}
}

func TestLockCmdReturnsCompletedMsgAndClearsSecrets(t *testing.T) {
	service := newOpenService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	msg := LockCmd(service, LockIntent{})()
	completed, ok := msg.(LockCompletedMsg)
	if !ok {
		t.Fatalf("Cmd() returned %T, want LockCompletedMsg", msg)
	}
	if completed.Err != nil {
		t.Errorf("Err = %v, want nil", completed.Err)
	}

	got, err := service.Vault().Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if err := got.Password.Reveal(func([]byte) error { return nil }); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() after LockCmd error = %v, want %v", err, domain.ErrSecretCleared)
	}
}
