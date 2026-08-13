package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func typePlain(m PlainInputModel, s string) PlainInputModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	return m
}

func TestPlainInputModelTypingAppends(t *testing.T) {
	m := NewPlainInputModel()
	m = typePlain(m, "personal.kdbx")

	if m.Value() != "personal.kdbx" {
		t.Errorf("Value() = %q, want %q", m.Value(), "personal.kdbx")
	}
	if m.View() != "> personal.kdbx" {
		t.Errorf("View() = %q, want %q", m.View(), "> personal.kdbx")
	}
}

func TestPlainInputModelBackspaceRemovesLastRune(t *testing.T) {
	m := NewPlainInputModel()
	m = typePlain(m, "abc")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Value() != "ab" {
		t.Errorf("Value() = %q, want %q", m.Value(), "ab")
	}
}

func TestPlainInputModelBackspaceOnMultiByteRune(t *testing.T) {
	m := NewPlainInputModel()
	m = typePlain(m, "п")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Value() != "" {
		t.Errorf("Value() = %q, want empty", m.Value())
	}
}

func TestPlainInputModelBackspaceOnEmptyIsNoop(t *testing.T) {
	m := NewPlainInputModel()
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Value() != "" {
		t.Errorf("Value() = %q, want empty", m.Value())
	}
}

func TestPlainInputModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewPlainInputModel()
	m, _ = m.Update(tea.QuitMsg{})

	if m.Value() != "" {
		t.Errorf("Value() = %q, want empty", m.Value())
	}
}

func TestPlainInputModelSetValue(t *testing.T) {
	m := NewPlainInputModel().SetValue("default.kdbx")
	if m.Value() != "default.kdbx" {
		t.Errorf("Value() = %q, want %q", m.Value(), "default.kdbx")
	}
}
