package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

// defaultGeneratorPassphraseWords is spec section 10.5's own site
// passphrase example ("lantern-copper-museum-orbit-falcon-river") word
// count — spec states no separate fixed default the way section 7.3
// does for the master passphrase, so this screen's starting point is
// that example, not the master-passphrase wizard's 8.
const defaultGeneratorPassphraseWords = 6

// GeneratorModel is spec section 10.6's password/passphrase generator
// screen. r/y/c/p/Esc act immediately — there is no separate "apply"
// step for regenerating or switching type, matching spec's own example
// keybinding list exactly. Every regenerate uses fresh randomness (spec
// 10.5: "each press of Generate uses new cryptographic randomness") —
// GeneratePassword/GeneratePassphrase both draw from src anew each call,
// nothing here caches or reuses a previous value.
type GeneratorModel struct {
	src random.Source

	PassphraseMode   bool
	PasswordPolicy   application.PasswordPolicy
	PassphrasePolicy application.PassphrasePolicy

	Value    domain.Secret
	Strength application.GeneratedStrength
	Err      error

	// UseInEntry/CopyOnly/Cancelled record which of [y]/[c]/[Esc] the
	// user picked. What each one actually does with Value — inserting it
	// into an entry field, copying it to the clipboard — is a parent
	// screen's job (the entry editor and clipboard wiring, not built
	// yet); this model only records the intent.
	UseInEntry bool
	CopyOnly   bool
	Cancelled  bool
}

// NewGeneratorModel returns a GeneratorModel already showing a freshly
// generated character password under spec section 10.2's default policy.
func NewGeneratorModel(src random.Source) GeneratorModel {
	m := GeneratorModel{
		src:              src,
		PasswordPolicy:   application.DefaultPasswordPolicy(),
		PassphrasePolicy: application.PassphrasePolicy{WordCount: defaultGeneratorPassphraseWords},
	}
	return m.regenerate()
}

func (m GeneratorModel) Init() tea.Cmd { return nil }

func (m GeneratorModel) Update(msg tea.Msg) (GeneratorModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Code == tea.KeyEscape {
		m.Cancelled = true
		return m, nil
	}

	switch key.Text {
	case "r":
		return m.regenerate(), nil
	case "p":
		m.PassphraseMode = !m.PassphraseMode
		return m.regenerate(), nil
	case "y":
		m.UseInEntry = true
	case "c":
		m.CopyOnly = true
	}
	return m, nil
}

func (m GeneratorModel) regenerate() GeneratorModel {
	if m.PassphraseMode {
		m.Value, m.Err = application.GeneratePassphrase(m.src, m.PassphrasePolicy)
		if m.Err == nil {
			m.Strength = application.EstimatePassphraseStrength(m.PassphrasePolicy)
		}
		return m
	}
	m.Value, m.Err = application.GeneratePassword(m.src, m.PasswordPolicy)
	if m.Err == nil {
		m.Strength = application.EstimatePasswordStrength(m.PasswordPolicy)
	}
	return m
}

// View renders spec section 10.6's layout, with the type-specific field
// block spec's example shows for a character password, or the
// passphrase equivalent when PassphraseMode is set, followed by spec
// 10.7's strength line.
func (m GeneratorModel) View() string {
	var b strings.Builder
	b.WriteString("Password Generator\n\n")

	if m.PassphraseMode {
		fmt.Fprintf(&b, "Type:       Passphrase\nWords:      %d\n\n", m.PassphrasePolicy.WordCount)
	} else {
		p := m.PasswordPolicy
		fmt.Fprintf(&b,
			"Type:       Character password\nLength:     %d\nLowercase:  %s\nUppercase:  %s\nDigits:     %s\nSymbols:    %s\nAmbiguous:  %s\n\n",
			p.Length, yesNo(p.Lowercase), yesNo(p.Uppercase), yesNo(p.Digits), yesNo(p.Symbols), allowAvoid(p.AvoidAmbiguous),
		)
	}

	switch {
	case m.Err != nil:
		fmt.Fprintf(&b, "(generation failed: %v)\n\n", m.Err)
	case m.Value != nil:
		_ = m.Value.Reveal(func(v []byte) error {
			fmt.Fprintf(&b, "%s\n\n", v)
			return nil
		})
		fmt.Fprintf(&b, "Estimated strength: %s\nReason: %s\n\n", m.Strength.Level, m.Strength.Reason)
	}

	switchLabel := "Switch to passphrase"
	if m.PassphraseMode {
		switchLabel = "Switch to character password"
	}
	fmt.Fprintf(&b, "[r] Regenerate\n[y] Use in entry\n[c] Copy without saving\n[p] %s\n[Esc] Cancel\n", switchLabel)
	return b.String()
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func allowAvoid(avoid bool) string {
	if avoid {
		return "avoid"
	}
	return "allow"
}
