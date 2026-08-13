package tui

import (
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

var recoveryTestWords = []string{"velvet", "orbit", "cactus", "lantern", "walnut", "engine", "harbor", "rabbit"}

func TestRecoveryModelDisplaysNumberedWords(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 2)
	view := m.View()

	for i, w := range recoveryTestWords {
		want := strconv.Itoa(i+1) + ". " + w
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; full view:\n%s", want, view)
		}
	}
}

func TestRecoveryModelEnterAdvancesToChallengeWithCorrectPositionCount(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 3)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.phase != recoveryPhaseChallenge {
		t.Fatalf("phase = %v, want recoveryPhaseChallenge", m.phase)
	}
	if len(m.positions) != 3 {
		t.Errorf("len(positions) = %d, want 3", len(m.positions))
	}
}

func TestRecoveryModelChallengeViewDoesNotLeakWordList(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 2)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	view := m.View()
	if !strings.HasPrefix(view, "Word #") {
		t.Errorf("View() during challenge = %q, want it to start with \"Word #\"", view)
	}
	for _, w := range recoveryTestWords {
		if strings.Contains(view, w) {
			t.Errorf("View() during challenge leaked word %q: %q", w, view)
		}
	}
}

func TestRecoveryModelCorrectAnswersVerify(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 2)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	positions := m.positions

	for _, pos := range positions {
		for _, r := range recoveryTestWords[pos-1] {
			m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
		}
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}

	if m.phase != recoveryPhaseDone {
		t.Fatalf("phase = %v, want recoveryPhaseDone", m.phase)
	}
	if !m.Verified {
		t.Error("Verified = false, want true for correct answers")
	}
}

func TestRecoveryModelWrongAnswerFailsVerification(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 1)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	for _, r := range "definitely-wrong" {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(r)})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Verified {
		t.Error("Verified = true for a wrong answer")
	}
}

func TestRecoveryModelBackspaceEditsCurrentAnswer(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 1)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m, _ = m.Update(tea.KeyPressMsg{Text: "x"})
	m, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m.buffer != "x" {
		t.Errorf("buffer = %q, want %q", m.buffer, "x")
	}
}

func TestRecoveryModelEscCancelsFromDisplay(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 1)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true")
	}
}

func TestRecoveryModelEscCancelsFromChallenge(t *testing.T) {
	m := NewRecoveryModel(recoveryTestWords, random.CryptoSource{}, 1)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true")
	}
}
