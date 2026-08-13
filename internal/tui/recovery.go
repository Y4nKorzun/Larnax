package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

type recoveryPhase int

const (
	recoveryPhaseDisplay recoveryPhase = iota
	recoveryPhaseChallenge
	recoveryPhaseDone
)

// RecoveryModel is spec section 7.6's recovery flow: show the generated
// passphrase once, numbered, then challenge a few random word positions
// before vault creation proceeds — proof the user actually wrote it
// down, not just that the screen briefly displayed it. It never writes
// the passphrase to a file, never touches the clipboard, and never
// sends it anywhere — spec 7.6's own list of what this flow must not
// do — simply because no code path here does any of those things.
type RecoveryModel struct {
	words          []string
	src            random.Source
	challengeCount int

	phase     recoveryPhase
	positions []int
	answers   map[int]string
	current   int
	buffer    string

	Verified  bool
	Cancelled bool
	Err       error
}

// NewRecoveryModel returns a RecoveryModel for words (the generated
// master passphrase's individual words, e.g.
// application.GeneratedMasterPassphrase.Words), which will challenge
// challengeCount random positions once the user confirms they wrote the
// phrase down.
func NewRecoveryModel(words []string, src random.Source, challengeCount int) RecoveryModel {
	return RecoveryModel{words: words, src: src, challengeCount: challengeCount, answers: map[int]string{}}
}

func (m RecoveryModel) Init() tea.Cmd { return nil }

func (m RecoveryModel) Update(msg tea.Msg) (RecoveryModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Code == tea.KeyEscape {
		m.Cancelled = true
		return m, nil
	}

	switch m.phase {
	case recoveryPhaseDisplay:
		return m.updateDisplay(key)
	case recoveryPhaseChallenge:
		return m.updateChallenge(key)
	default:
		return m, nil
	}
}

func (m RecoveryModel) updateDisplay(key tea.KeyPressMsg) (RecoveryModel, tea.Cmd) {
	if key.Code != tea.KeyEnter {
		return m, nil
	}
	positions, err := application.ChooseRecoveryChallenge(m.src, len(m.words), m.challengeCount)
	if err != nil {
		m.Err = err
		return m, nil
	}
	m.positions = positions
	m.phase = recoveryPhaseChallenge
	return m, nil
}

func (m RecoveryModel) updateChallenge(key tea.KeyPressMsg) (RecoveryModel, tea.Cmd) {
	switch key.Code {
	case tea.KeyEnter:
		m.answers[m.positions[m.current]] = m.buffer
		m.buffer = ""
		m.current++
		if m.current < len(m.positions) {
			return m, nil
		}
		verified, err := application.VerifyRecoveryAnswers(m.words, m.answers)
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.Verified = verified
		m.phase = recoveryPhaseDone
		return m, nil

	case tea.KeyBackspace:
		if len(m.buffer) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.buffer)
			m.buffer = m.buffer[:len(m.buffer)-size]
		}
		return m, nil
	}

	if key.Text != "" {
		m.buffer += key.Text
	}
	return m, nil
}

// View renders spec 7.6's two screens: the numbered word list, then one
// "Word #N:" challenge prompt at a time. During the challenge phase the
// word list itself is never shown again — only View's own past output
// (already off-screen by then) ever displayed it.
func (m RecoveryModel) View() string {
	switch m.phase {
	case recoveryPhaseDisplay:
		var b strings.Builder
		b.WriteString("Your recovery passphrase\n\n")
		for i, w := range m.words {
			fmt.Fprintf(&b, "%d. %s\n", i+1, w)
		}
		b.WriteString("\nWrite it down on paper.\nDo not store it next to the KDBX backup.\n")
		return b.String()

	case recoveryPhaseChallenge:
		return fmt.Sprintf("Word #%d: %s", m.positions[m.current], m.buffer)

	default:
		if m.Verified {
			return "Recovery confirmed.\n"
		}
		return "Recovery answers did not match.\n"
	}
}
