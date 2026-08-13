package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func typeConfirm(m ConfirmModel, s string) ConfirmModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	return m
}

func TestConfirmModelExactMatchConfirms(t *testing.T) {
	m := NewConfirmModel(`Permanently delete "GitHub Work" and its history?`, "delete")
	m = typeConfirm(m, "delete")

	if !m.Confirmed {
		t.Error("Confirmed = false, want true for an exact match")
	}
}

func TestConfirmModelPartialMatchDoesNotConfirm(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m = typeConfirm(m, "del")

	if m.Confirmed {
		t.Error("Confirmed = true for a partial match")
	}
}

func TestConfirmModelWrongPhraseDoesNotConfirm(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m = typeConfirm(m, "delete!")

	if m.Confirmed {
		t.Error("Confirmed = true for an incorrect phrase")
	}
}

func TestConfirmModelBackspaceUnconfirms(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m = typeConfirm(m, "delete")
	if !m.Confirmed {
		t.Fatal("setup: Confirmed = false after typing the exact phrase")
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.Confirmed {
		t.Error("Confirmed = true after backspace broke the exact match")
	}
	if m.Input() != "delet" {
		t.Errorf("Input() = %q, want %q", m.Input(), "delet")
	}
}

func TestConfirmModelBackspaceOnEmptyIsNoop(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Input() != "" {
		t.Errorf("Input() = %q, want empty", m.Input())
	}
}

func TestConfirmModelEscCancels(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true after Esc")
	}
}

func TestConfirmModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewConfirmModel("prompt", "delete")
	m, _ = m.Update(tea.QuitMsg{})

	if m.Input() != "" || m.Confirmed || m.Cancelled {
		t.Errorf("non-key Msg mutated the model: %+v", m)
	}
}

func TestConfirmModelViewContainsPromptAndRequiredPhrase(t *testing.T) {
	m := NewConfirmModel(`Permanently delete "GitHub Work" and its history?`, "delete")
	m = typeConfirm(m, "del")

	view := m.View()
	for _, want := range []string{`Permanently delete "GitHub Work" and its history?`, "Type: delete", "> del"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; full view:\n%s", want, view)
		}
	}
}
