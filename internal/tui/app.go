package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

type appScreen int

const (
	screenWelcome appScreen = iota
	screenCreateVault
	screenRecovery
	screenUnlock
	screenNeedsPath
	screenDone
)

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

// AppModel is the top-level tea.Model composing every screen built so
// far: Welcome -> (n) CreateVault -> Recovery, and Welcome -> (o) Unlock
// when a path is already known (spec section 7.1's "kdbx-tui <path>" CLI
// form — NewAppModel's path parameter). It is not yet Milestone 1's full
// read-only explorer: there is no browser screen, no path-entry screen
// for "open" without a path already known (screenNeedsPath is an honest
// stub for that case, not a real picker), and the create flow stops once
// the recovery challenge is verified — collecting a vault name and save
// path to actually call application.CreateVault is separate, not yet
// built work.
type AppModel struct {
	screen appScreen

	welcome     WelcomeModel
	createVault CreateVaultModel
	recovery    RecoveryModel
	unlock      UnlockModel

	src        random.Source
	service    *application.VaultService
	unlockPath string
}

// NewAppModel returns an AppModel starting at the welcome screen.
// unlockPath is used if the user picks "Open vault" — pass "" if none is
// known yet (the no-argument wizard case), which routes to
// screenNeedsPath instead of a real file picker that doesn't exist yet.
func NewAppModel(src random.Source, service *application.VaultService, unlockPath string) AppModel {
	return AppModel{src: src, service: service, unlockPath: unlockPath}
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
	case screenUnlock:
		return m.updateUnlock(msg)
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
	if m.recovery.phase == recoveryPhaseDone {
		m.screen = screenDone
	}
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
		m.screen = screenDone
		return m, nil
	}
	return m, cmd
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
	case screenUnlock:
		view := "Unlock " + m.unlock.Path + "\n\nMaster passphrase:\n" + m.unlock.Input.View() + "\n"
		if m.unlock.ErrMessage != "" {
			view += "\n" + m.unlock.ErrMessage + "\n"
		}
		return tea.NewView(view)
	case screenNeedsPath:
		return tea.NewView("Opening a vault without a path is not supported yet — pass the file as a command-line argument.\n\n[q] Quit\n")
	case screenDone:
		return tea.NewView(m.doneView())
	default:
		return tea.NewView("")
	}
}

func (m AppModel) doneView() string {
	if m.unlock.Unlocked {
		return "Vault unlocked.\n"
	}
	if m.recovery.Verified {
		return "Recovery confirmed. (Vault name/path collection is separate, not-yet-built work.)\n"
	}
	return "Recovery failed — the typed words did not match.\n"
}
