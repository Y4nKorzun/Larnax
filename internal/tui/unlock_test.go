package tui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
)

const unlockTestPassword = "correct horse battery staple test only"

func createVaultFile(t *testing.T) string {
	t.Helper()
	var s application.VaultService
	if err := s.New("My Vault", unlockTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := s.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	return path
}

func TestUnlockModelTypingUpdatesInput(t *testing.T) {
	m := NewUnlockModel("personal.kdbx", &application.VaultService{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "s"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "3"})

	if m.Input.Len() != 2 {
		t.Errorf("Input.Len() = %d, want 2", m.Input.Len())
	}
}

func TestUnlockModelEscSetsCancelled(t *testing.T) {
	m := NewUnlockModel("personal.kdbx", &application.VaultService{})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true after Esc")
	}
}

func TestUnlockModelEnterClearsInputAndReturnsCmd(t *testing.T) {
	m := NewUnlockModel("personal.kdbx", &application.VaultService{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "x"})

	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Update() for Enter returned a nil Cmd")
	}
	if m.Input.Len() != 0 {
		t.Errorf("Input.Len() after Enter = %d, want 0 (ownership taken by attemptUnlock)", m.Input.Len())
	}
}

func TestUnlockModelSuccessfulUnlock(t *testing.T) {
	path := createVaultFile(t)
	var service application.VaultService
	m := NewUnlockModel(path, &service)

	for _, r := range unlockTestPassword {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Update() for Enter returned a nil Cmd")
	}

	msg := cmd()
	m, _ = m.Update(msg)

	if !m.Unlocked {
		t.Errorf("Unlocked = false, want true; ErrMessage = %q", m.ErrMessage)
	}
	if m.ErrMessage != "" {
		t.Errorf("ErrMessage = %q, want empty", m.ErrMessage)
	}
}

func TestUnlockModelFailedUnlockShowsGenericMessage(t *testing.T) {
	path := createVaultFile(t)
	var service application.VaultService
	m := NewUnlockModel(path, &service)

	for _, r := range "wrong passphrase entirely" {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Update() for Enter returned a nil Cmd")
	}

	msg := cmd()
	m, _ = m.Update(msg)

	if m.Unlocked {
		t.Error("Unlocked = true for a wrong passphrase")
	}
	if m.ErrMessage != unlockErrorMessage {
		t.Errorf("ErrMessage = %q, want the fixed generic message", m.ErrMessage)
	}
}

func TestUnlockModelMissingFileShowsGenericMessage(t *testing.T) {
	var service application.VaultService
	m := NewUnlockModel(filepath.Join(t.TempDir(), "does-not-exist.kdbx"), &service)

	for _, r := range unlockTestPassword {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	msg := cmd()
	m, _ = m.Update(msg)

	if m.ErrMessage != unlockErrorMessage {
		t.Errorf("ErrMessage = %q, want the fixed generic message even for a missing file", m.ErrMessage)
	}
}
