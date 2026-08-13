package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
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

func newTestAppModel(unlockPath string) AppModel {
	return NewAppModel(random.CryptoSource{}, &application.VaultService{}, unlockPath, nil)
}

func TestAppModelStartsAtWelcome(t *testing.T) {
	m := newTestAppModel("")
	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelWelcomeQuitReturnsQuitCmd(t *testing.T) {
	m := newTestAppModel("")
	_, cmd := updateApp(t, m, tea.KeyPressMsg{Text: "q"})

	if cmd == nil {
		t.Fatal("Update() for 'q' returned a nil Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("Cmd() did not resolve to tea.QuitMsg")
	}
}

// TestAppModelFullCreateFlowReachesBrowserAndPersistsToDisk exercises
// spec 26.1's acceptance script end to end through AppModel: new vault ->
// pick strength -> confirm recovery -> choose a save path -> land on the
// browser screen with a real KDBX file on disk.
func TestAppModelFullCreateFlowReachesBrowserAndPersistsToDisk(t *testing.T) {
	m := newTestAppModel("")

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

	if m.screen != screenSavePath {
		t.Fatalf("screen = %v, want screenSavePath; recovery.Verified = %v", m.screen, m.recovery.Verified)
	}

	path := filepath.Join(t.TempDir(), "personal.kdbx")
	for _, r := range path {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.screen != screenBrowser {
		t.Fatalf("screen = %v, want screenBrowser; saveErr = %v", m.screen, m.saveErr)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("vault file not found on disk: %v", err)
	}
}

func TestAppModelCreateVaultCancelReturnsToWelcome(t *testing.T) {
	m := newTestAppModel("")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "n"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelOpenWithNoPathGoesToNeedsPath(t *testing.T) {
	m := newTestAppModel("")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})

	if m.screen != screenNeedsPath {
		t.Errorf("screen = %v, want screenNeedsPath", m.screen)
	}
}

func TestAppModelNeedsPathQuits(t *testing.T) {
	m := newTestAppModel("")
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
	m := newTestAppModel("")
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelOpenWithPathUnlocksIntoBrowser(t *testing.T) {
	var setup application.VaultService
	if err := setup.New("My Vault", appTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	entry := domain.NewEntry(setup.Vault().RootGroupID(), "GitHub")
	if err := setup.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := setup.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, path, nil)

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

	if m.screen != screenBrowser {
		t.Fatalf("screen = %v, want screenBrowser; ErrMessage = %q", m.screen, m.unlock.ErrMessage)
	}
	if len(m.browser.list.Titles) != 1 || m.browser.list.Titles[0] != "GitHub" {
		t.Errorf("browser.list.Titles = %v, want [GitHub]", m.browser.list.Titles)
	}
}

func TestAppModelUnlockCancelReturnsToWelcome(t *testing.T) {
	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, "/nonexistent/vault.kdbx", nil)

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome", m.screen)
	}
}

func TestAppModelBrowserLockReturnsToWelcome(t *testing.T) {
	var setup application.VaultService
	if err := setup.New("My Vault", appTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := setup.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, path, nil)
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	for _, r := range appTestPassword {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = updateApp(t, m, cmd())
	if m.screen != screenBrowser {
		t.Fatalf("setup: screen = %v, want screenBrowser", m.screen)
	}

	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: " "})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "l"})

	if m.screen != screenWelcome {
		t.Errorf("screen = %v, want screenWelcome after lock", m.screen)
	}
}

func TestAppModelBrowserQuitReturnsQuitCmd(t *testing.T) {
	var setup application.VaultService
	if err := setup.New("My Vault", appTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := setup.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, path, nil)
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "o"})
	for _, r := range appTestPassword {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, cmd := updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = updateApp(t, m, cmd())
	if m.screen != screenBrowser {
		t.Fatalf("setup: screen = %v, want screenBrowser", m.screen)
	}

	_, cmd = updateApp(t, m, tea.KeyPressMsg{Text: "q"})
	if cmd == nil {
		t.Fatal("Update() for 'q' returned a nil Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("Cmd() did not resolve to tea.QuitMsg")
	}
}
