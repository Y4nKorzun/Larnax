package components

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// ConfirmModel is spec section 8.2's CONFIRM mode for a dangerous action
// (delete, overwrite, import, KDF change). Spec section 9.5 gives the
// concrete pattern this follows: the user must type an exact phrase
// back, not just press y/n —
//
//	Permanently delete "GitHub Work" and its history?
//	Type: delete
//
// A typo'd or partial answer never confirms — Confirmed is recomputed on
// every keystroke, not only checked once on Enter, so it can never go
// stale relative to what Input() currently shows.
type ConfirmModel struct {
	Prompt         string
	RequiredPhrase string
	input          string

	Confirmed bool
	Cancelled bool
}

// NewConfirmModel returns a ConfirmModel asking prompt, requiring the
// user to type requiredPhrase exactly to confirm.
func NewConfirmModel(prompt, requiredPhrase string) ConfirmModel {
	return ConfirmModel{Prompt: prompt, RequiredPhrase: requiredPhrase}
}

func (m ConfirmModel) Init() tea.Cmd { return nil }

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch key.Code {
	case tea.KeyEscape:
		m.Cancelled = true
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.input)
			m.input = m.input[:len(m.input)-size]
		}
		m.Confirmed = m.input == m.RequiredPhrase
		return m, nil
	}

	if key.Text != "" {
		m.input += key.Text
	}
	m.Confirmed = m.input == m.RequiredPhrase
	return m, nil
}

// Input returns what the user has typed so far.
func (m ConfirmModel) Input() string {
	return m.input
}

func (m ConfirmModel) View() string {
	return strings.Join([]string{m.Prompt, "Type: " + m.RequiredPhrase, "> " + m.input}, "\n")
}
