package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

func revealGeneratorValue(t *testing.T, s domain.Secret) string {
	t.Helper()
	var v string
	if err := s.Reveal(func(b []byte) error { v = string(b); return nil }); err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	return v
}

func TestNewGeneratorModelStartsWithCharacterPassword(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	if m.PassphraseMode {
		t.Error("PassphraseMode = true, want false")
	}
	if m.Err != nil {
		t.Fatalf("Err = %v", m.Err)
	}
	if len(revealGeneratorValue(t, m.Value)) != m.PasswordPolicy.Length {
		t.Errorf("generated value length = %d, want %d", len(revealGeneratorValue(t, m.Value)), m.PasswordPolicy.Length)
	}
}

func TestGeneratorModelRegenerateProducesDifferentValue(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	first := revealGeneratorValue(t, m.Value)

	m, _ = m.Update(tea.KeyPressMsg{Text: "r"})
	second := revealGeneratorValue(t, m.Value)

	if first == second {
		t.Error("regenerate produced the same value twice")
	}
}

func TestGeneratorModelSwitchToPassphraseAndBack(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})

	m, _ = m.Update(tea.KeyPressMsg{Text: "p"})
	if !m.PassphraseMode {
		t.Fatal("PassphraseMode = false after 'p', want true")
	}
	if !strings.Contains(revealGeneratorValue(t, m.Value), "-") {
		t.Errorf("passphrase value %q does not look hyphen-joined", revealGeneratorValue(t, m.Value))
	}

	m, _ = m.Update(tea.KeyPressMsg{Text: "p"})
	if m.PassphraseMode {
		t.Error("PassphraseMode = true after a second 'p', want false")
	}
}

func TestGeneratorModelYSetsUseInEntry(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})

	if !m.UseInEntry {
		t.Error("UseInEntry = false, want true")
	}
}

func TestGeneratorModelCSetsCopyOnly(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Text: "c"})

	if !m.CopyOnly {
		t.Error("CopyOnly = false, want true")
	}
}

func TestGeneratorModelEscCancels(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true")
	}
}

func TestGeneratorModelViewShowsGeneratedValueAndSpecLabels(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	view := m.View()

	for _, want := range []string{
		"Password Generator", "Type:       Character password", "Length:     24",
		revealGeneratorValue(t, m.Value),
		"Estimated strength:", "[r] Regenerate", "[y] Use in entry", "[c] Copy without saving",
		"[p] Switch to passphrase", "[Esc] Cancel",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; full view:\n%s", want, view)
		}
	}
}

func TestGeneratorModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewGeneratorModel(random.CryptoSource{})
	before := revealGeneratorValue(t, m.Value)

	m, _ = m.Update(tea.QuitMsg{})

	if revealGeneratorValue(t, m.Value) != before {
		t.Error("non-key Msg regenerated the value")
	}
}
