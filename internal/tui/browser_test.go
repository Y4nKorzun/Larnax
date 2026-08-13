package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

const browserTestPassword = "correct horse battery staple test only"

func newBrowserTestService(t *testing.T) *application.VaultService {
	t.Helper()
	var s application.VaultService
	if err := s.New("My Vault", browserTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return &s
}

func TestBrowserModelListsExistingEntries(t *testing.T) {
	service := newBrowserTestService(t)
	e1 := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	e2 := domain.NewEntry(service.Vault().RootGroupID(), "GitLab")
	if err := service.AddEntry(e1); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if err := service.AddEntry(e2); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	if len(m.list.Titles) != 2 {
		t.Fatalf("len(Titles) = %d, want 2", len(m.list.Titles))
	}
}

func TestBrowserModelAKeyOpensAddEntryOverlay(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "a"})

	if m.overlay != browserOverlayAddEntry {
		t.Errorf("overlay = %v, want browserOverlayAddEntry", m.overlay)
	}
}

func TestBrowserModelAddEntryFlowAddsToVaultAndRefreshesList(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	m, _ = m.Update(tea.KeyPressMsg{Text: "a"})
	for _, r := range "GitHub" {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	for _, r := range "octocat" {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.overlay != browserOverlayNone {
		t.Errorf("overlay = %v, want browserOverlayNone after finishing", m.overlay)
	}
	if len(service.Vault().AllEntries()) != 1 {
		t.Fatalf("vault has %d entries, want 1", len(service.Vault().AllEntries()))
	}
	if len(m.list.Titles) != 1 || m.list.Titles[0] != "GitHub" {
		t.Errorf("list.Titles = %v, want [GitHub]", m.list.Titles)
	}
}

func TestBrowserModelCopyPasswordSetsIntent(t *testing.T) {
	service := newBrowserTestService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "p"})

	if m.CopyIntent == nil {
		t.Fatal("CopyIntent = nil, want set after yp")
	}
	if m.CopyIntent.Field != application.FieldPassword {
		t.Errorf("CopyIntent.Field = %v, want %v", m.CopyIntent.Field, application.FieldPassword)
	}
	if m.CopyIntent.Entry.ID != entry.ID {
		t.Errorf("CopyIntent.Entry.ID = %x, want %x", m.CopyIntent.Entry.ID, entry.ID)
	}
}

func TestBrowserModelCopyUsernameSetsIntent(t *testing.T) {
	service := newBrowserTestService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Username = "octocat"
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "u"})

	if m.CopyIntent == nil {
		t.Fatal("CopyIntent = nil, want set after yu")
	}
	if m.CopyIntent.Field != application.FieldUsername {
		t.Errorf("CopyIntent.Field = %v, want %v", m.CopyIntent.Field, application.FieldUsername)
	}
}

func TestBrowserModelCopyIntentClearsNextUpdate(t *testing.T) {
	service := newBrowserTestService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "u"})
	if m.CopyIntent == nil {
		t.Fatal("setup: CopyIntent = nil")
	}

	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	if m.CopyIntent != nil {
		t.Error("CopyIntent still set after an unrelated Update")
	}
}

func TestBrowserModelLeaderSSetsSave(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	m, _ = m.Update(tea.KeyPressMsg{Text: " "})
	m, _ = m.Update(tea.KeyPressMsg{Text: "s"})

	if !m.Save {
		t.Error("Save = false, want true after <Leader>s")
	}
}

func TestBrowserModelLeaderLSetsLock(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	m, _ = m.Update(tea.KeyPressMsg{Text: " "})
	m, _ = m.Update(tea.KeyPressMsg{Text: "l"})

	if !m.Lock {
		t.Error("Lock = false, want true after <Leader>l")
	}
}

func TestBrowserModelQSetsQuit(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "q"})

	if !m.Quit {
		t.Error("Quit = false, want true")
	}
}

func TestBrowserModelJKMovesCursor(t *testing.T) {
	service := newBrowserTestService(t)
	for _, title := range []string{"GitHub", "GitLab"} {
		if err := service.AddEntry(domain.NewEntry(service.Vault().RootGroupID(), title)); err != nil {
			t.Fatalf("AddEntry() error = %v", err)
		}
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})

	if m.list.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", m.list.Cursor)
	}
}

func TestBrowserModelGGMovesToFirst(t *testing.T) {
	service := newBrowserTestService(t)
	for _, title := range []string{"GitHub", "GitLab", "Bitbucket"} {
		if err := service.AddEntry(domain.NewEntry(service.Vault().RootGroupID(), title)); err != nil {
			t.Fatalf("AddEntry() error = %v", err)
		}
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	if m.list.Cursor != 2 {
		t.Fatalf("setup: Cursor = %d, want 2", m.list.Cursor)
	}

	m, _ = m.Update(tea.KeyPressMsg{Text: "g"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "g"})
	if m.list.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0 after gg", m.list.Cursor)
	}
}

func TestBrowserModelEscResetsPendingSequenceAndLeader(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m, _ = m.Update(tea.KeyPressMsg{Text: "p"})

	if m.CopyIntent != nil {
		t.Error("CopyIntent set after Esc broke the yp sequence")
	}
}

func TestBrowserModelStatusLineShowsEntryCount(t *testing.T) {
	service := newBrowserTestService(t)
	if err := service.AddEntry(domain.NewEntry(service.Vault().RootGroupID(), "GitHub")); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	m := NewBrowserModel(service, random.CryptoSource{})

	if got := m.StatusLine(); got == "" {
		t.Error("StatusLine() is empty")
	}
}

func TestBrowserModelSelectedEntryReturnsCurrentEntry(t *testing.T) {
	service := newBrowserTestService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Username = "octocat"
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	got, ok := m.SelectedEntry()
	if !ok {
		t.Fatal("SelectedEntry() ok = false")
	}
	if got.ID != entry.ID {
		t.Errorf("SelectedEntry().ID = %x, want %x", got.ID, entry.ID)
	}
}

func TestBrowserModelSelectedEntryOnEmptyVaultReturnsFalse(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	if _, ok := m.SelectedEntry(); ok {
		t.Error("SelectedEntry() ok = true for an empty vault")
	}
}

func TestBrowserModelDetailViewShowsFieldsWithMaskedPassword(t *testing.T) {
	service := newBrowserTestService(t)
	entry := domain.NewEntry(service.Vault().RootGroupID(), "GitHub")
	entry.Username = "octocat"
	entry.URL = "https://github.com"
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := service.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	m := NewBrowserModel(service, random.CryptoSource{})
	view := m.View()

	for _, want := range []string{"Title:    GitHub", "Username: octocat", "URL:      https://github.com", passwordMask} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; full view:\n%s", want, view)
		}
	}
	if strings.Contains(view, "hunter2") {
		t.Error("View() leaked the plaintext password")
	}
}

func TestBrowserModelDetailViewNoSelection(t *testing.T) {
	service := newBrowserTestService(t)
	m := NewBrowserModel(service, random.CryptoSource{})

	if got := m.detailView(); got != "(no entry selected)\n" {
		t.Errorf("detailView() = %q, want %q", got, "(no entry selected)\n")
	}
}
