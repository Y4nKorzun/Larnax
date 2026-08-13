// Package components holds the small, reusable Bubble Tea sub-models
// screens compose — spec section 18.3's tui/components — starting with
// SecureInputModel.
package components

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// SecureInputModel is a masked text input for a secret the user types
// interactively — the master passphrase (spec section 7.7) is never
// accepted any other way: no --password flag, no environment variable.
// Every typed character renders as "•" (spec 7.7's own example format),
// never the actual text.
//
// SecureInputModel holds the typed bytes itself rather than wrapping
// them in a domain.Secret while still being edited — Secret's contract
// is for a value that is done being mutated, not one gaining a byte per
// keystroke. Value hands the caller ownership of the raw bytes once
// entry is done, ready to wrap in a domain.Secret; Reset zeroes them if
// input is abandoned instead (e.g. the user cancels).
type SecureInputModel struct {
	value []byte
}

// NewSecureInputModel returns an empty SecureInputModel.
func NewSecureInputModel() SecureInputModel {
	return SecureInputModel{}
}

func (m SecureInputModel) Init() tea.Cmd { return nil }

// Update appends a typed printable character, or removes the last one on
// backspace. Every other key is ignored: a screen embedding this
// component handles Enter/Esc itself, since "commit" and "cancel" mean
// different things on different screens (e.g. the create wizard versus
// the unlock screen).
func (m SecureInputModel) Update(msg tea.Msg) (SecureInputModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Code == tea.KeyBackspace {
		if len(m.value) > 0 {
			_, size := utf8.DecodeLastRune(m.value)
			for i := len(m.value) - size; i < len(m.value); i++ {
				m.value[i] = 0
			}
			m.value = m.value[:len(m.value)-size]
		}
		return m, nil
	}

	if key.Text != "" {
		m.value = append(m.value, []byte(key.Text)...)
	}
	return m, nil
}

// View renders the masked field per spec section 7.7's example format.
func (m SecureInputModel) View() string {
	return "> " + strings.Repeat("•", utf8.RuneCount(m.value))
}

// Len reports how many runes have been typed so far. This is safe to
// show or log — spec section 22.1 forbids logging the passphrase itself,
// not its length — unlike Value.
func (m SecureInputModel) Len() int {
	return utf8.RuneCount(m.value)
}

// Value returns the raw typed bytes and clears the model's own copy, so
// the caller becomes the sole owner — the same ownership handoff
// domain.NewSecret documents for a byte slice.
func (m *SecureInputModel) Value() []byte {
	v := m.value
	m.value = nil
	return v
}

// Reset zeroes and discards whatever has been typed so far, e.g. when
// the user cancels input.
func (m *SecureInputModel) Reset() {
	for i := range m.value {
		m.value[i] = 0
	}
	m.value = nil
}
