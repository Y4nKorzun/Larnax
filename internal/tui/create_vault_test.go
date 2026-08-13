package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

func pressDigit(m CreateVaultModel, d string) CreateVaultModel {
	m, _ = m.Update(tea.KeyPressMsg{Text: d})
	return m
}

func TestCreateVaultModelSixSubmitsImmediately(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m = pressDigit(m, "6")

	if !m.Chosen {
		t.Fatalf("Chosen = false, want true; Err = %v", m.Err)
	}
	if len(m.Generated.Words) != 6 {
		t.Errorf("len(Words) = %d, want 6", len(m.Generated.Words))
	}
}

func TestCreateVaultModelEightSubmitsImmediately(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m = pressDigit(m, "8")

	if !m.Chosen {
		t.Fatalf("Chosen = false, want true; Err = %v", m.Err)
	}
	if len(m.Generated.Words) != 8 {
		t.Errorf("len(Words) = %d, want 8", len(m.Generated.Words))
	}
}

func TestCreateVaultModelTwelveRequiresEnter(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m = pressDigit(m, "1")
	if m.Chosen {
		t.Fatal("Chosen = true after just \"1\", want false")
	}
	m = pressDigit(m, "2")
	if m.Chosen {
		t.Fatal("Chosen = true after \"12\" without Enter, want false")
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.Chosen {
		t.Fatalf("Chosen = false after Enter, want true; Err = %v", m.Err)
	}
	if len(m.Generated.Words) != 12 {
		t.Errorf("len(Words) = %d, want 12", len(m.Generated.Words))
	}
}

func TestCreateVaultModelInvalidDigitDoesNotSubmit(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m = pressDigit(m, "9")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Chosen {
		t.Error("Chosen = true for an invalid digit, want false")
	}
}

func TestCreateVaultModelBackspaceEditsBuffer(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m = pressDigit(m, "1")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	m = pressDigit(m, "2")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Chosen {
		t.Error("Chosen = true for buffer \"2\" (backspace should have dropped the \"1\"), want false")
	}
}

func TestCreateVaultModelEscCancels(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if !m.Cancelled {
		t.Error("Cancelled = false, want true after Esc")
	}
}

func TestCreateVaultModelIgnoresNonKeyMsg(t *testing.T) {
	m := NewCreateVaultModel(random.CryptoSource{})
	m, _ = m.Update(tea.QuitMsg{})

	if m.Chosen || m.Cancelled {
		t.Errorf("non-key Msg mutated the model: %+v", m)
	}
}
