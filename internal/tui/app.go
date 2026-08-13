package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/config"
	clipboardpkg "github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/tui/components"
)

type appScreen int

const (
	screenWelcome appScreen = iota
	screenCreateVault
	screenRecovery
	screenSavePath
	screenUnlock
	screenBrowser
	screenNeedsPath
	screenDone
)

// defaultRootGroupName names a newly created vault's root group. Spec's
// wizard (section 7.1-7.3) never prompts for one separately from the
// save path, so this fixed name stands in until a real name-entry step
// exists.
const defaultRootGroupName = "Passwords"

// defaultBackupRetention is used until config is actually loaded from
// disk (internal/config/loader.go exists, but nothing yet reads
// config.toml at startup) — config.Default().Backups matches spec
// section 21.1's own example values, so this is that default, not an
// invented number.
var defaultBackupRetention = application.BackupRetention(config.Default().Backups)

// recoveryChallengeCount is how many word positions RecoveryModel
// challenges after a passphrase is generated — spec section 7.6's own
// example challenges 2 positions ("Word #3", "Word #7") for an 8-word
// phrase, so 2 is this app's fixed count too, capped at the actual word
// count for the (currently unreachable through the normal wizard, but
// not impossible) case of a shorter phrase.
func recoveryChallengeCount(wordCount int) int {
	if wordCount < 2 {
		return wordCount
	}
	return 2
}

// AppModel is the top-level tea.Model composing every screen this
// codebase has: Welcome -> (n) CreateVault -> Recovery -> SavePath ->
// Browser, and Welcome -> (o) Unlock -> Browser when a path is already
// known (spec section 7.1's "kdbx-tui <path>" CLI form — NewAppModel's
// path parameter). It is spec section 28's Milestone 1/2 minimum, not
// yet the full read-only explorer: there is no group navigation (one
// flat entry list per vault), no path-entry screen for "open" without a
// path already known (screenNeedsPath is an honest stub, not a real file
// picker), and no entry edit/delete.
type AppModel struct {
	screen appScreen

	welcome     WelcomeModel
	createVault CreateVaultModel
	recovery    RecoveryModel
	savePath    components.PlainInputModel
	saveErr     error
	unlock      UnlockModel
	browser     BrowserModel

	src        random.Source
	service    *application.VaultService
	clipboard  clipboardpkg.Clipboard
	retention  int
	unlockPath string
}

// NewAppModel returns an AppModel starting at the welcome screen.
// unlockPath is used if the user picks "Open vault" — pass "" if none is
// known yet (the no-argument wizard case), which routes to
// screenNeedsPath instead of a real file picker that doesn't exist yet.
// cb may be nil (no clipboard available on this platform yet — spec
// section 24.5's other OS adapters aren't built) — copy actions then
// report an error instead of attempting one.
func NewAppModel(src random.Source, service *application.VaultService, unlockPath string, cb clipboardpkg.Clipboard) AppModel {
	return AppModel{
		src:        src,
		service:    service,
		unlockPath: unlockPath,
		clipboard:  cb,
		retention:  defaultBackupRetention,
	}
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWelcome:
		return m.updateWelcome(msg)
	case screenCreateVault:
		return m.updateCreateVault(msg)
	case screenRecovery:
		return m.updateRecovery(msg)
	case screenSavePath:
		return m.updateSavePath(msg)
	case screenUnlock:
		return m.updateUnlock(msg)
	case screenBrowser:
		return m.updateBrowser(msg)
	case screenNeedsPath, screenDone:
		return m.updateTerminalScreen(msg)
	default:
		return m, nil
	}
}

func (m AppModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.welcome.Update(msg)
	m.welcome = next.(WelcomeModel)

	switch m.welcome.Choice {
	case WelcomeChoiceNew:
		m.welcome.Choice = WelcomeChoiceNone
		m.createVault = NewCreateVaultModel(m.src)
		m.screen = screenCreateVault
		return m, nil
	case WelcomeChoiceOpen:
		m.welcome.Choice = WelcomeChoiceNone
		if m.unlockPath == "" {
			m.screen = screenNeedsPath
			return m, nil
		}
		m.unlock = NewUnlockModel(m.unlockPath, m.service)
		m.screen = screenUnlock
		return m, nil
	case WelcomeChoiceQuit:
		return m, tea.Quit
	default:
		return m, cmd
	}
}

func (m AppModel) updateCreateVault(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.createVault, _ = m.createVault.Update(msg)

	if m.createVault.Cancelled {
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	}
	if m.createVault.Chosen {
		words := m.createVault.Generated.Words
		m.recovery = NewRecoveryModel(words, m.src, recoveryChallengeCount(len(words)))
		m.screen = screenRecovery
	}
	return m, nil
}

func (m AppModel) updateRecovery(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.recovery, _ = m.recovery.Update(msg)

	if m.recovery.Cancelled {
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	}
	if m.recovery.phase == recoveryPhaseDone && m.recovery.Verified {
		m.savePath = components.NewPlainInputModel()
		m.screen = screenSavePath
	}
	return m, nil
}

func (m AppModel) updateSavePath(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Code == tea.KeyEscape {
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	}
	if key.Code != tea.KeyEnter {
		m.savePath, _ = m.savePath.Update(msg)
		return m, nil
	}

	path := m.savePath.Value()
	if path == "" {
		return m, nil
	}

	var passphrase string
	if err := m.createVault.Generated.Phrase.Reveal(func(v []byte) error {
		passphrase = string(v)
		return nil
	}); err != nil {
		m.saveErr = err
		return m, nil
	}

	if err := m.service.New(defaultRootGroupName, passphrase); err != nil {
		m.saveErr = err
		return m, nil
	}
	if err := m.service.SaveAs(path, m.retention); err != nil {
		m.saveErr = err
		return m, nil
	}

	m.saveErr = nil
	m.browser = NewBrowserModel(m.service, m.src)
	m.screen = screenBrowser
	return m, nil
}

func (m AppModel) updateUnlock(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.unlock.Update(msg)
	m.unlock = next

	if m.unlock.Cancelled {
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	}
	if m.unlock.Unlocked {
		m.browser = NewBrowserModel(m.service, m.src)
		m.screen = screenBrowser
		return m, nil
	}
	return m, cmd
}

func (m AppModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.browser, _ = m.browser.Update(msg)

	if m.browser.Quit {
		return m, tea.Quit
	}
	if m.browser.Lock {
		if err := m.service.Lock(); err != nil {
			m.browser.StatusMessage = err.Error()
		}
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	}
	if m.browser.Save {
		if err := m.service.Save(m.retention); err != nil {
			m.browser.StatusMessage = err.Error()
		}
		return m, nil
	}
	if m.browser.CopyIntent != nil {
		if m.clipboard == nil {
			m.browser.StatusMessage = "no clipboard available on this platform yet"
			return m, nil
		}
		return m, CopyFieldCmd(context.Background(), m.service, m.clipboard, *m.browser.CopyIntent)
	}
	return m, nil
}

// updateTerminalScreen handles screenNeedsPath and screenDone, neither of
// which has a real next step yet: q quits, Esc goes back to the welcome
// screen.
func (m AppModel) updateTerminalScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Text == "q":
		return m, tea.Quit
	case key.Code == tea.KeyEscape:
		m.screen = screenWelcome
		m.welcome = WelcomeModel{}
		return m, nil
	default:
		return m, nil
	}
}

func (m AppModel) View() tea.View {
	switch m.screen {
	case screenWelcome:
		return m.welcome.View()
	case screenCreateVault:
		return tea.NewView(m.createVault.View())
	case screenRecovery:
		return tea.NewView(m.recovery.View())
	case screenSavePath:
		view := "Save vault as:\n" + m.savePath.View() + "\n"
		if m.saveErr != nil {
			view += "\n" + m.saveErr.Error() + "\n"
		}
		return tea.NewView(view)
	case screenUnlock:
		view := "Unlock " + m.unlock.Path + "\n\nMaster passphrase:\n" + m.unlock.Input.View() + "\n"
		if m.unlock.ErrMessage != "" {
			view += "\n" + m.unlock.ErrMessage + "\n"
		}
		return tea.NewView(view)
	case screenBrowser:
		return tea.NewView(m.browser.View())
	case screenNeedsPath:
		return tea.NewView("Opening a vault without a path is not supported yet — pass the file as a command-line argument.\n\n[q] Quit\n")
	case screenDone:
		return tea.NewView(m.doneView())
	default:
		return tea.NewView("")
	}
}

func (m AppModel) doneView() string {
	return "Recovery failed — the typed words did not match.\n"
}
