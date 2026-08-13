package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func typeText(m SecureInputModel, s string) SecureInputModel {
	m, _ = m.Update(tea.KeyPressMsg{Text: s})
	return m
}

func TestSecureInputModelTypingAppendsAndMasks(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "h")
	m = typeText(m, "i")

	if got := m.View(); got != "> ••" {
		t.Errorf("View() = %q, want %q", got, "> ••")
	}
	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2", m.Len())
	}
}

func TestSecureInputModelViewNeverContainsTypedText(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "s")
	m = typeText(m, "3")
	m = typeText(m, "c")
	m = typeText(m, "r")
	m = typeText(m, "3")
	m = typeText(m, "t")

	if got := m.View(); got != "> ••••••" {
		t.Errorf("View() = %q, contains typed text or wrong mask count", got)
	}
}

func TestSecureInputModelBackspaceRemovesLastRune(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "a")
	m = typeText(m, "b")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if got := m.View(); got != "> •" {
		t.Errorf("View() = %q, want %q", got, "> •")
	}
	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1", m.Len())
	}
}

func TestSecureInputModelBackspaceOnMultiByteRune(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "п") // 2-byte UTF-8 rune
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after removing a multi-byte rune", m.Len())
	}
	v := m.Value()
	if len(v) != 0 {
		t.Errorf("Value() = %q, want empty", v)
	}
}

func TestSecureInputModelBackspaceOnEmptyIsNoop(t *testing.T) {
	m := NewSecureInputModel()
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
	if cmd != nil {
		t.Error("Update() returned a non-nil Cmd")
	}
}

func TestSecureInputModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewSecureInputModel()
	m, _ = m.Update(tea.QuitMsg{})

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
}

func TestSecureInputModelIgnoresKeysWithNoText(t *testing.T) {
	m := NewSecureInputModel()
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for a key with no Text", m.Len())
	}
}

func TestSecureInputModelValueReturnsBytesAndClearsModel(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "secret")

	v := m.Value()
	if string(v) != "secret" {
		t.Errorf("Value() = %q, want %q", v, "secret")
	}
	if m.Len() != 0 {
		t.Errorf("Len() after Value() = %d, want 0", m.Len())
	}
	if got := m.View(); got != "> " {
		t.Errorf("View() after Value() = %q, want %q", got, "> ")
	}
}

func TestSecureInputModelReset(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "secret")
	m.Reset()

	if m.Len() != 0 {
		t.Errorf("Len() after Reset() = %d, want 0", m.Len())
	}
}

func TestSecureInputModelLenCountsRunesNotBytes(t *testing.T) {
	m := NewSecureInputModel()
	m = typeText(m, "пароль") // 6 runes, 12 bytes in UTF-8

	if m.Len() != 6 {
		t.Errorf("Len() = %d, want 6", m.Len())
	}
}
