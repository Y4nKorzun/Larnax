package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// EntryListModel is a minimal list of entry titles with a movable cursor
// — spec section 8.3's j/k/G navigation over the middle "Entries" panel
// of spec 8.1's layout. It only ever holds titles, never full
// domain.Entry values: spec section 19.2 is explicit that UI components
// receive a presentation model, not vault data, so a parent screen is
// responsible for mapping whatever domain.Entry it's showing down to
// just the title before handing it to this component.
//
// Update only handles the single, unambiguous keys directly (j/k/G and
// the arrow keys). "gg" (first item) is deliberately not handled here:
// it is a two-key sequence, and resolving that safely needs the
// prefix-tracking state machine tui.KeySequence already owns at a higher
// level — a parent that has already resolved "gg" calls First() directly
// instead of routing it through Update.
type EntryListModel struct {
	Titles []string
	Cursor int
}

// NewEntryListModel returns an EntryListModel over titles, cursor at the
// first item.
func NewEntryListModel(titles []string) EntryListModel {
	return EntryListModel{Titles: titles}
}

func (m EntryListModel) Init() tea.Cmd { return nil }

func (m EntryListModel) Update(msg tea.Msg) (EntryListModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Text == "j" || key.Code == tea.KeyDown:
		return m.Down(), nil
	case key.Text == "k" || key.Code == tea.KeyUp:
		return m.Up(), nil
	case key.Text == "G":
		return m.Last(), nil
	default:
		return m, nil
	}
}

// Down moves the cursor one item down, stopping at the last item.
func (m EntryListModel) Down() EntryListModel {
	if len(m.Titles) > 0 && m.Cursor < len(m.Titles)-1 {
		m.Cursor++
	}
	return m
}

// Up moves the cursor one item up, stopping at the first item.
func (m EntryListModel) Up() EntryListModel {
	if m.Cursor > 0 {
		m.Cursor--
	}
	return m
}

// First moves the cursor to the first item (spec 8.3's gg).
func (m EntryListModel) First() EntryListModel {
	m.Cursor = 0
	return m
}

// Last moves the cursor to the last item (spec 8.3's G).
func (m EntryListModel) Last() EntryListModel {
	if len(m.Titles) > 0 {
		m.Cursor = len(m.Titles) - 1
	} else {
		m.Cursor = 0
	}
	return m
}

// Selected returns the title at the cursor, or "", false if Titles is
// empty.
func (m EntryListModel) Selected() (string, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Titles) {
		return "", false
	}
	return m.Titles[m.Cursor], true
}

func (m EntryListModel) View() string {
	if len(m.Titles) == 0 {
		return "(no entries)\n"
	}
	var b strings.Builder
	for i, title := range m.Titles {
		marker := "  "
		if i == m.Cursor {
			marker = "> "
		}
		b.WriteString(marker + title + "\n")
	}
	return b.String()
}
