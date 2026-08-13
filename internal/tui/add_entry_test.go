package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

func typeIntoAddEntry(m AddEntryModel, s string) AddEntryModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	return m
}

func TestAddEntryModelFullFlowGeneratesPassword(t *testing.T) {
	parent := domain.NewGroupID()
	m := NewAddEntryModel(random.CryptoSource{}, parent)

	m = typeIntoAddEntry(m, "GitHub")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.step != addEntryStepUsername {
		t.Fatalf("step = %v, want addEntryStepUsername", m.step)
	}

	m = typeIntoAddEntry(m, "octocat")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.Done {
		t.Fatalf("Done = false, want true; Err = %v", m.Err)
	}
	if m.Entry.Title != "GitHub" {
		t.Errorf("Entry.Title = %q, want %q", m.Entry.Title, "GitHub")
	}
	if m.Entry.Username != "octocat" {
		t.Errorf("Entry.Username = %q, want %q", m.Entry.Username, "octocat")
	}
	if m.Entry.ParentGroup != parent {
		t.Errorf("Entry.ParentGroup = %x, want %x", m.Entry.ParentGroup, parent)
	}

	var length int
	if err := m.Entry.Password.Reveal(func(v []byte) error { length = len(v); return nil }); err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	if length != 24 {
		t.Errorf("generated password length = %d, want 24", length)
	}
}

func TestAddEntryModelEmptyTitleBlocksAdvance(t *testing.T) {
	m := NewAddEntryModel(random.CryptoSource{}, domain.NewGroupID())
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.step != addEntryStepTitle {
		t.Errorf("step = %v, want addEntryStepTitle (empty title must not advance)", m.step)
	}
}

func TestAddEntryModelEscCancelsFromTitle(t *testing.T) {
	m := NewAddEntryModel(random.CryptoSource{}, domain.NewGroupID())
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true")
	}
}

func TestAddEntryModelEscCancelsFromUsername(t *testing.T) {
	m := NewAddEntryModel(random.CryptoSource{}, domain.NewGroupID())
	m = typeIntoAddEntry(m, "GitHub")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true")
	}
}

func TestAddEntryModelAllowsEmptyUsername(t *testing.T) {
	m := NewAddEntryModel(random.CryptoSource{}, domain.NewGroupID())
	m = typeIntoAddEntry(m, "GitHub")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // empty username, straight to Enter

	if !m.Done {
		t.Fatalf("Done = false, want true; Err = %v", m.Err)
	}
	if m.Entry.Username != "" {
		t.Errorf("Entry.Username = %q, want empty", m.Entry.Username)
	}
}
