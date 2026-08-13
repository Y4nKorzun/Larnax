package application

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// fakeClipboard is an in-memory clipboard.Clipboard: there is no real OS
// backend yet (see internal/infrastructure/clipboard/clipboard.go), so
// this is what CopyField's own unit tests use instead.
type fakeClipboard struct {
	content []byte
}

func (f *fakeClipboard) ReadText(context.Context) ([]byte, error) {
	return f.content, nil
}

func (f *fakeClipboard) WriteText(_ context.Context, value []byte) error {
	f.content = value
	return nil
}

func (f *fakeClipboard) Clear(context.Context) error {
	f.content = nil
	return nil
}

func sampleCopyEntry() domain.Entry {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	e.Username = "octocat"
	e.Password = domain.NewSecretFromString("hunter2")
	e.URL = "https://github.com"
	return e
}

func TestCopyFieldWritesUsername(t *testing.T) {
	cb := &fakeClipboard{}
	if _, err := CopyField(context.Background(), cb, sampleCopyEntry(), FieldUsername); err != nil {
		t.Fatalf("CopyField() error = %v", err)
	}
	if string(cb.content) != "octocat" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "octocat")
	}
}

func TestCopyFieldWritesURL(t *testing.T) {
	cb := &fakeClipboard{}
	if _, err := CopyField(context.Background(), cb, sampleCopyEntry(), FieldURL); err != nil {
		t.Fatalf("CopyField() error = %v", err)
	}
	if string(cb.content) != "https://github.com" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "https://github.com")
	}
}

func TestCopyFieldWritesPasswordAndReturnsMatchingHash(t *testing.T) {
	cb := &fakeClipboard{}
	hash, err := CopyField(context.Background(), cb, sampleCopyEntry(), FieldPassword)
	if err != nil {
		t.Fatalf("CopyField() error = %v", err)
	}
	if string(cb.content) != "hunter2" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "hunter2")
	}
	if want := sha256.Sum256([]byte("hunter2")); hash != want {
		t.Errorf("ownership hash = %x, want %x", hash, want)
	}
}

func TestCopyFieldRejectsUnknownField(t *testing.T) {
	cb := &fakeClipboard{}
	if _, err := CopyField(context.Background(), cb, sampleCopyEntry(), FieldName(99)); !errors.Is(err, ErrUnknownField) {
		t.Errorf("CopyField() error = %v, want %v", err, ErrUnknownField)
	}
}

func TestCopyFieldPropagatesClearedPasswordSecret(t *testing.T) {
	cb := &fakeClipboard{}
	entry := sampleCopyEntry()
	entry.Password.Clear()

	if _, err := CopyField(context.Background(), cb, entry, FieldPassword); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("CopyField() error = %v, want %v", err, domain.ErrSecretCleared)
	}
}
