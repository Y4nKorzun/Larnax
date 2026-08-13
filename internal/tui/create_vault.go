package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

// CreateVaultModel is spec section 7.3's master passphrase strength
// wizard:
//
//	Choose master passphrase strength
//
//	[6]  6 random words   — strong
//	[8]  8 random words   — recommended
//	[12] 12 random words  — maximum
//
// There is deliberately no way to reach 4 words here: spec 7.3 keeps that
// out of the normal wizard entirely (application.GenerateMasterPassphraseUnsafe
// is the separate, explicit escape hatch this screen never calls).
//
// Digit input is buffered rather than acted on key-by-key, because "12"
// is two keystrokes a terminal always sends as two separate events, not
// one. "6" and "8" submit immediately on their own digit, since neither
// could be the start of a longer valid choice; "1" only submits after
// Enter, since it might still become "12".
type CreateVaultModel struct {
	src    random.Source
	buffer string

	Chosen    bool
	Generated application.GeneratedMasterPassphrase
	Err       error
	Cancelled bool
}

// NewCreateVaultModel returns a CreateVaultModel that draws words via src.
func NewCreateVaultModel(src random.Source) CreateVaultModel {
	return CreateVaultModel{src: src}
}

func (m CreateVaultModel) Init() tea.Cmd { return nil }

func (m CreateVaultModel) Update(msg tea.Msg) (CreateVaultModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch key.Code {
	case tea.KeyEscape:
		m.Cancelled = true
		return m, nil
	case tea.KeyEnter:
		return m.submit()
	case tea.KeyBackspace:
		if len(m.buffer) > 0 {
			m.buffer = m.buffer[:len(m.buffer)-1]
		}
		return m, nil
	}

	if len(key.Text) == 1 && key.Text[0] >= '0' && key.Text[0] <= '9' {
		m.buffer += key.Text
		if m.buffer == "6" || m.buffer == "8" {
			return m.submit()
		}
	}
	return m, nil
}

// submit resolves the buffered digits into a strength and generates the
// passphrase, clearing the buffer either way. An unrecognized buffer
// (anything but "6", "8", or "12") leaves Chosen false so the user can
// keep typing — spec gives no "invalid choice" message to show, so this
// only declines to act rather than inventing one.
func (m CreateVaultModel) submit() (CreateVaultModel, tea.Cmd) {
	strength, ok := strengthFromDigits(m.buffer)
	m.buffer = ""
	if !ok {
		return m, nil
	}

	generated, err := application.GenerateMasterPassphraseWithWords(m.src, strength)
	m.Generated = generated
	m.Err = err
	m.Chosen = err == nil
	return m, nil
}

func strengthFromDigits(s string) (application.MasterPassphraseStrength, bool) {
	switch s {
	case "6":
		return application.MasterPassphraseStrong, true
	case "8":
		return application.MasterPassphraseRecommended, true
	case "12":
		return application.MasterPassphraseMaximum, true
	default:
		return 0, false
	}
}

func (m CreateVaultModel) View() string {
	return "Choose master passphrase strength\n\n" +
		"[6]  6 random words   — strong\n" +
		"[8]  8 random words   — recommended\n" +
		"[12] 12 random words  — maximum\n"
}
