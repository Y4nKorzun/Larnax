package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

// TestScenario261CreateVaultAddEntrySaveAndReopen drives AppModel through
// spec section 26.1's full acceptance script — the definition of "a
// minimal usable version" this goal is chasing:
//
//  1. start the app (NewAppModel);
//  2. create personal.kdbx;
//  3. pick 8 random words;
//  4. write down and confirm the recovery phrase;
//  5. add a GitHub entry;
//  6. enter a username;
//  7. generate a 24-character password;
//  8. save the vault;
//  9. "close" it (this test just drops the in-memory VaultService);
//  10. reopen the same file with a brand-new, independent VaultService
//     and the exact master passphrase the wizard generated, and confirm
//     the entry and its generated password come back correctly.
//
// Step 10 stands in for spec's "open it in KeePassXC" — this repo has no
// KeePassXC to shell out to, but the KDBX bytes this test's reopen reads
// are produced by the same kdbx.Save/Open pair already proven against a
// genuine KeePass-generated fixture elsewhere (kdbx/fixture_test.go,
// kdbx/repository_test.go), so an independent reopen through this
// package's own repository layer is the strongest check available here.
func TestScenario261CreateVaultAddEntrySaveAndReopen(t *testing.T) {
	var service application.VaultService
	m := NewAppModel(random.CryptoSource{}, &service, "", nil)

	// 1-2: start, choose "New vault".
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "n"})

	// 3: pick 8 words (spec 7.3's default/recommended strength).
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "8"})
	if m.screen != screenRecovery {
		t.Fatalf("screen = %v, want screenRecovery", m.screen)
	}

	// Capture the generated passphrase now, in plaintext, purely for this
	// test's own later independent-reopen step — nothing in the
	// production code path does this; SavePath.updateSavePath only ever
	// reveals it internally to hand to VaultService.New.
	var masterPassphrase string
	if err := m.createVault.Generated.Phrase.Reveal(func(v []byte) error {
		masterPassphrase = string(v)
		return nil
	}); err != nil {
		t.Fatalf("revealing generated passphrase: %v", err)
	}
	words := m.createVault.Generated.Words

	// 4: confirm the recovery challenge.
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	positions := m.recovery.positions
	for _, pos := range positions {
		for _, r := range words[pos-1] {
			m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
		}
		m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if m.screen != screenSavePath {
		t.Fatalf("screen = %v, want screenSavePath", m.screen)
	}

	// 2 (path) + 8 (this is where SaveAs actually happens): name the file
	// and save.
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	for _, r := range path {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.screen != screenBrowser {
		t.Fatalf("screen = %v, want screenBrowser; saveErr = %v", m.screen, m.saveErr)
	}

	// 5-7: add the GitHub entry with a username and a generated password.
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "a"})
	for _, r := range "GitHub" {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	for _, r := range "octocat" {
		m, _ = updateApp(t, m, tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = updateApp(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(service.Vault().AllEntries()) != 1 {
		t.Fatalf("vault has %d entries after adding GitHub, want 1", len(service.Vault().AllEntries()))
	}

	// 8: save.
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: " "})
	m, _ = updateApp(t, m, tea.KeyPressMsg{Text: "s"})
	if m.browser.StatusMessage != "" {
		t.Fatalf("Save reported an error: %s", m.browser.StatusMessage)
	}

	// 9: "close" — nothing left to do but stop using service; the
	// independent reopen below never touches it again.

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("vault file not found on disk: %v", err)
	}

	// 10: reopen independently and verify.
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	var reopened application.VaultService
	if err := reopened.Open(f, path, masterPassphrase); err != nil {
		t.Fatalf("independent reopen failed: %v", err)
	}
	if reopened.ReadOnly() {
		t.Fatal("reopened vault is read-only, want writable")
	}

	entries := reopened.Vault().AllEntries()
	if len(entries) != 1 {
		t.Fatalf("reopened vault has %d entries, want 1", len(entries))
	}
	got := entries[0]
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub")
	}
	if got.Username != "octocat" {
		t.Errorf("Username = %q, want %q", got.Username, "octocat")
	}

	var passwordLen int
	if err := got.Password.Reveal(func(v []byte) error {
		passwordLen = len(v)
		return nil
	}); err != nil {
		t.Fatalf("revealing reopened password: %v", err)
	}
	if passwordLen != 24 {
		t.Errorf("password length = %d, want 24 (spec 26.1's script)", passwordLen)
	}
}
