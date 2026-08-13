package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var testTitles = []string{"GitHub", "GitLab", "Bitbucket"}

func TestEntryListModelDownMovesCursor(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})

	if m.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", m.Cursor)
	}
}

func TestEntryListModelDownStopsAtLastItem(t *testing.T) {
	m := NewEntryListModel(testTitles)
	for range 10 {
		m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	}
	if m.Cursor != len(testTitles)-1 {
		t.Errorf("Cursor = %d, want %d", m.Cursor, len(testTitles)-1)
	}
}

func TestEntryListModelUpStopsAtFirstItem(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "k"})

	if m.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", m.Cursor)
	}
}

func TestEntryListModelUpMovesCursorBack(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "k"})

	if m.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", m.Cursor)
	}
}

func TestEntryListModelGKeyJumpsToLast(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "G"})

	if m.Cursor != len(testTitles)-1 {
		t.Errorf("Cursor = %d, want %d", m.Cursor, len(testTitles)-1)
	}
}

func TestEntryListModelFirstAndLastMethods(t *testing.T) {
	m := NewEntryListModel(testTitles).Last()
	if m.Cursor != len(testTitles)-1 {
		t.Errorf("Last(): Cursor = %d, want %d", m.Cursor, len(testTitles)-1)
	}
	m = m.First()
	if m.Cursor != 0 {
		t.Errorf("First(): Cursor = %d, want 0", m.Cursor)
	}
}

func TestEntryListModelDownOnEmptyListIsNoop(t *testing.T) {
	m := NewEntryListModel(nil)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})

	if m.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", m.Cursor)
	}
}

func TestEntryListModelSelectedOnEmptyListReturnsFalse(t *testing.T) {
	m := NewEntryListModel(nil)
	if _, ok := m.Selected(); ok {
		t.Error("Selected() ok = true for an empty list")
	}
}

func TestEntryListModelSelectedReturnsCurrentTitle(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})

	got, ok := m.Selected()
	if !ok {
		t.Fatal("Selected() ok = false")
	}
	if got != "GitLab" {
		t.Errorf("Selected() = %q, want %q", got, "GitLab")
	}
}

func TestEntryListModelViewShowsCursorMarker(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.KeyPressMsg{Text: "j"})

	view := m.View()
	if !strings.Contains(view, "> GitLab") {
		t.Errorf("View() missing cursor marker on GitLab; full view:\n%s", view)
	}
	if !strings.Contains(view, "  GitHub") {
		t.Errorf("View() missing unmarked GitHub; full view:\n%s", view)
	}
}

func TestEntryListModelViewEmptyList(t *testing.T) {
	m := NewEntryListModel(nil)
	if got := m.View(); got != "(no entries)\n" {
		t.Errorf("View() = %q, want %q", got, "(no entries)\n")
	}
}

func TestEntryListModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewEntryListModel(testTitles)
	m, _ = m.Update(tea.QuitMsg{})

	if m.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", m.Cursor)
	}
}
