package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/tui/components"
)

// unlockErrorMessage is spec section 7.8's deliberately generic failure
// text. Spec is explicit this must never distinguish a wrong passphrase
// from a damaged file or a bad key file — so this is the one fixed
// string every failed attempt produces, never the underlying error's own
// text, which could leak which case actually happened.
const unlockErrorMessage = "Unable to unlock the vault.\nThe credentials may be incorrect, or the file may be damaged."

// UnlockModel is spec section 7.7's unlock screen:
//
//	Unlock personal.kdbx
//
//	Master passphrase:
//	> •••••••••••••••••••••••••••••••••••••••
//
// The passphrase is only ever accepted interactively through this
// screen's SecureInputModel — spec 7.7 forbids a --password flag or a
// KDBX_PASSWORD-style environment variable, and UnlockModel simply has
// no field or code path that could accept one that way.
type UnlockModel struct {
	Path    string
	Input   components.SecureInputModel
	service *application.VaultService

	// Unlocked is true once an UnlockAttemptedMsg reports success.
	Unlocked bool
	// Cancelled is true once the user presses Esc — spec has no stated
	// behavior for backing out of this screen, so this only records the
	// intent; a parent model decides what Cancelled actually leads to
	// (e.g. back to the welcome screen).
	Cancelled bool
	// ErrMessage is unlockErrorMessage after a failed attempt, or "" if
	// none has failed yet.
	ErrMessage string
}

// NewUnlockModel returns an UnlockModel for the vault at path, which
// will use service to attempt each unlock.
func NewUnlockModel(path string, service *application.VaultService) UnlockModel {
	return UnlockModel{Path: path, Input: components.NewSecureInputModel(), service: service}
}

func (m UnlockModel) Init() tea.Cmd { return nil }

// UnlockAttemptedMsg is the result of trying to open Path with whatever
// passphrase the user had typed when they pressed Enter.
type UnlockAttemptedMsg struct {
	Err error
}

func (m UnlockModel) Update(msg tea.Msg) (UnlockModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UnlockAttemptedMsg:
		if msg.Err != nil {
			m.ErrMessage = unlockErrorMessage
			return m, nil
		}
		m.Unlocked = true
		m.ErrMessage = ""
		return m, nil

	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEnter:
			// Evaluated in two steps, not "return m, m.attemptUnlock()":
			// attemptUnlock mutates m.Input in place (it takes ownership
			// of and clears the field), and that mutation must happen
			// before m is captured as the first return value, not left to
			// Go's return-expression evaluation order to get right.
			cmd := m.attemptUnlock()
			return m, cmd
		case tea.KeyEscape:
			m.Cancelled = true
			return m, nil
		}
		var cmd tea.Cmd
		m.Input, cmd = m.Input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// attemptUnlock takes ownership of whatever is currently in m.Input
// (clearing the field, same as SecureInputModel.Value always does) and
// returns a Cmd that opens Path with it.
//
// The []byte is converted to a string only once, right before calling
// service.Open — VaultService.Open takes a plain passphrase string, and
// unlike a []byte a Go string can never be zeroed afterward. That is the
// same tradeoff internal/infrastructure/kdbx/mapper.go's revealString
// already documents; it is not introduced here, only inherited.
func (m *UnlockModel) attemptUnlock() tea.Cmd {
	path := m.Path
	service := m.service
	passphrase := m.Input.Value()

	return func() tea.Msg {
		defer zeroBytes(passphrase)

		f, err := os.Open(path)
		if err != nil {
			return UnlockAttemptedMsg{Err: err}
		}
		defer f.Close()

		return UnlockAttemptedMsg{Err: service.Open(f, path, string(passphrase))}
	}
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
