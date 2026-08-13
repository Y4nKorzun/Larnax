package tui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

const appTestPassword = "correct horse battery staple test only"

func updateApp(t *testing.T, m AppModel, msg tea.Msg) (AppModel, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	got, ok := next.(AppModel)
	if !ok {
		t.Fatalf("Update() returned %T, want AppModel", next)
	}
	return got, cmd
}

func TestAppModelStartsAtWelcome(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelWelcomeQuitReturnsQuitCmd(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	_, cmd := updateApp(t, m, tea.KeyPressMsg{Text: "q"})

	if cmd == nil {
		t.Fatal("Update() for 'q' returned a nil Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("Cmd() did not resolve to tea.QuitMsg")
	}
}

func TestAppModelNewVaultFlowThroughRecoveryVerified(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "n"})
	if m.screen != screenCreateVault {
		t.Fatalf("screen = %v, want screenCreateVault", m.screen)
	}

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "8"})
	if m.screen != screenRecovery {
		t.Fatalf("screen = %v, want screenRecovery", m.screen)
	}

	// The recovery screen starts by showing the phrase (spec 7.6): Enter
	// confirms "I wrote it down" and only then does the challenge begin.
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})

	words := m.recovery.words
	positions := m.recovery.positions
	for _, pos := range positions {
		for _, r := range words[pos-1] {
			m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
		}
		m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	}

	if m.screen != screenDone {
		t.Fatalf("screen = %v, want screenDone", m.screen)
	}
	if !m.recovery.Verified {
		t.Error("recovery.Verified = false, want true")
	}
}

func TestAppModelCreateVaultCancelReturnsToWelcome(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "n"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelOpenWithNoPathGoesToNeedsPath(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})

	if m.screen != screenNeedsPath {
		t.Errorf("screen = %v, want screenNeedsPath", m.screen)
	}
}

func TestAppModelNeedsPathQuits(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	_, cmd := updateApp(t, m, tea.KeyPressMsg{Text: "q"})

	if cmd == nil {
		t.Fatal("Update() for 'q' returned a nil Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("Cmd() did not resolve to tea.QuitMsg")
	}
}

func TestAppModelNeedsPathEscReturnsToWelcome(t *testing.T) {
	m := NewAppModel(random.CryptoSource{}, &application.VaultService{}, "")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelOpenWithPathUnlocksSuccessfully(t *testing.T) {
	var setup application.VaultService
	if err := setup.New("My Vault", appTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := setup.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, path)

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	if m.screen != screenUnlock {
		t.Fatalf("screen = %v, want screenUnlock", m.screen)
	}

	for _, r := range appTestPassword {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Update() for Enter returned a nil Cmd")
	}
	m, _ = updateApp(t, m, cmd())

	if m.screen != screenDone {
		t.Fatalf("screen = %v, want screenDone; ErrMessage = %q", m.screen, m.unlock.ErrMessage)
	}
	if !m.unlock.Unlocked {
		t.Error("unlock.Unlocked = false, want true")
	}
}

func TestAppModelUnlockCancelReturnsToWelcome(t *testing.T) {
	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, "/nonexistent/vault.kdbx")

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}
