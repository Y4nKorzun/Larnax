package components

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// PlainInputModel is a single-line text input for non-secret values —
// a vault save path, an entry title, a username — the same shape as
// SecureInputModel but showing the typed text directly, since none of
// those need masking the way a passphrase does (spec section 7.7 only
// requires masking the master passphrase; titles and usernames are
// routinely visible in the browser layout, spec section 8.1).
type PlainInputModel struct {
	value string
}

// NewPlainInputModel returns an empty PlainInputModel.
func NewPlainInputModel() PlainInputModel {
	return PlainInputModel{}
}

func (m PlainInputModel) Init() tea.Cmd { return nil }

// Update appends a typed printable character, or removes the last rune
// on backspace. Every other key is ignored — a screen embedding this
// component handles Enter/Esc itself, the same convention
// SecureInputModel uses.
func (m PlainInputModel) Update(msg tea.Msg) (PlainInputModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Code == tea.KeyBackspace {
		if len(m.value) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.value)
			m.value = m.value[:len(m.value)-size]
		}
		return m, nil
	}

	if key.Text != "" {
		m.value += key.Text
	}
	return m, nil
}

// Value returns what has been typed so far.
func (m PlainInputModel) Value() string {
	return m.value
}

// SetValue replaces the current value outright, e.g. to pre-fill a
// suggested default.
func (m PlainInputModel) SetValue(v string) PlainInputModel {
	m.value = v
	return m
}

func (m PlainInputModel) View() string {
	return "> " + m.value
}
