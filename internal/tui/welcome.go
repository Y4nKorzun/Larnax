package tui

import tea "charm.land/bubbletea/v2"

// WelcomeChoice is what the user picked on WelcomeModel.
type WelcomeChoice int

const (
	WelcomeChoiceNone WelcomeChoice = iota
	WelcomeChoiceNew
	WelcomeChoiceOpen
	WelcomeChoiceQuit
)

// WelcomeModel is spec section 7.1's very first screen:
//
//	Create or open a vault
//
//	[n] New vault
//	[o] Open vault
//	[q] Quit
//
// It is the first real tea.Model in this codebase — proof the tui
// package can actually drive a Bubble Tea program, not just recognize
// keystrokes and hold state in the abstract. It deliberately holds no
// vault data (spec section 19.2: UI components receive a presentation
// model and emit intents, they don't own vault data) — there is no vault
// yet to not-own at this screen.
type WelcomeModel struct {
	Choice WelcomeChoice
}

func (m WelcomeModel) Init() tea.Cmd {
	return nil
}

// Update resolves n/o/q directly rather than through Keymap/ApplyOverrides:
// this screen exists before any vault is open, so there is nothing yet
// for a user override to rebind spec 21.3's way — Keymap's job starts at
// the browser screen, a later piece of work.
func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "n":
		m.Choice = WelcomeChoiceNew
		return m, tea.Quit
	case "o":
		m.Choice = WelcomeChoiceOpen
		return m, tea.Quit
	case "q":
		m.Choice = WelcomeChoiceQuit
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m WelcomeModel) View() tea.View {
	return tea.NewView("Create or open a vault\n\n[n] New vault\n[o] Open vault\n[q] Quit\n")
}
