package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var _ tea.Model = WelcomeModel{}

func TestWelcomeModelInitReturnsNilCmd(t *testing.T) {
	if cmd := (WelcomeModel{}).Init(); cmd != nil {
		t.Error("Init() returned a non-nil Cmd")
	}
}

func TestWelcomeModelIgnoresNonKeyMsg(t *testing.T) {
	m := WelcomeModel{}
	next, cmd := m.Update(tea.QuitMsg{})

	got := next.(WelcomeModel)
	if got.Choice != WelcomeChoiceNone {
		t.Errorf("Choice = %v, want %v", got.Choice, WelcomeChoiceNone)
	}
	if cmd != nil {
		t.Error("Update() with a non-key Msg returned a non-nil Cmd")
	}
}

func isQuitCmd(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestWelcomeModelSelectsNew(t *testing.T) {
	m := WelcomeModel{}
	next, cmd := m.Update(tea.KeyPressMsg{Text: "n"})

	got := next.(WelcomeModel)
	if got.Choice != WelcomeChoiceNew {
		t.Errorf("Choice = %v, want %v", got.Choice, WelcomeChoiceNew)
	}
	if !isQuitCmd(t, cmd) {
		t.Error("Update() for 'n' did not return tea.Quit")
	}
}

func TestWelcomeModelSelectsOpen(t *testing.T) {
	m := WelcomeModel{}
	next, cmd := m.Update(tea.KeyPressMsg{Text: "o"})

	got := next.(WelcomeModel)
	if got.Choice != WelcomeChoiceOpen {
		t.Errorf("Choice = %v, want %v", got.Choice, WelcomeChoiceOpen)
	}
	if !isQuitCmd(t, cmd) {
		t.Error("Update() for 'o' did not return tea.Quit")
	}
}

func TestWelcomeModelSelectsQuit(t *testing.T) {
	m := WelcomeModel{}
	next, cmd := m.Update(tea.KeyPressMsg{Text: "q"})

	got := next.(WelcomeModel)
	if got.Choice != WelcomeChoiceQuit {
		t.Errorf("Choice = %v, want %v", got.Choice, WelcomeChoiceQuit)
	}
	if !isQuitCmd(t, cmd) {
		t.Error("Update() for 'q' did not return tea.Quit")
	}
}

func TestWelcomeModelIgnoresUnknownKey(t *testing.T) {
	m := WelcomeModel{}
	next, cmd := m.Update(tea.KeyPressMsg{Text: "z"})

	got := next.(WelcomeModel)
	if got.Choice != WelcomeChoiceNone {
		t.Errorf("Choice = %v, want %v", got.Choice, WelcomeChoiceNone)
	}
	if cmd != nil {
		t.Error("Update() for an unbound key returned a non-nil Cmd")
	}
}

func TestWelcomeModelViewMatchesSpecLayout(t *testing.T) {
	content := (WelcomeModel{}).View().Content
	for _, want := range []string{"Create or open a vault", "[n] New vault", "[o] Open vault", "[q] Quit"} {
		if !strings.Contains(content, want) {
			t.Errorf("View content missing %q; full content:\n%s", want, content)
		}
	}
}
