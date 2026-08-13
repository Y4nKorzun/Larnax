package tui

import "errors"

// Additional Action values for spec section 8.3's single-key and
// <Leader>-prefixed bindings. ActionFirstItem, ActionDelete,
// ActionCopyUsername, ActionCopyPassword, ActionCopyURL, ActionCopyNotes,
// ActionCopyTOTP, and ActionOpenGenerator already exist in
// key_sequence.go — those are the entries DefaultKeymap shares with
// KeySequence's own two-key recognizer.
const (
	ActionDown             Action = "down"               // j
	ActionUp               Action = "up"                 // k
	ActionParent           Action = "parent"             // h
	ActionOpenSelected     Action = "open-selected"      // l, Enter
	ActionLastItem         Action = "last-item"          // G
	ActionHalfPageDown     Action = "half-page-down"     // Ctrl+d
	ActionHalfPageUp       Action = "half-page-up"       // Ctrl+u
	ActionNextPanel        Action = "next-panel"         // Tab
	ActionPrevPanel        Action = "prev-panel"         // Shift+Tab
	ActionSearchStart      Action = "search-start"       // /
	ActionSearchNext       Action = "search-next"        // n
	ActionSearchPrev       Action = "search-prev"        // N
	ActionCancel           Action = "cancel"             // Esc
	ActionFuzzyFinder      Action = "fuzzy-finder"       // <Leader>f
	ActionAddEntry         Action = "add-entry"          // a
	ActionAddGroup         Action = "add-group"          // A
	ActionEditEntry        Action = "edit-entry"         // e
	ActionUndo             Action = "undo"               // u
	ActionRedo             Action = "redo"               // Ctrl+r
	ActionRevealPassword   Action = "reveal-password"    // r
	ActionLockVault        Action = "lock-vault"         // <Leader>l
	ActionSave             Action = "save"               // <Leader>s
	ActionOpenAnotherVault Action = "open-another-vault" // <Leader>o
	ActionCommandMode      Action = "command-mode"       // :
	ActionQuit             Action = "quit"               // q
)

// Keymap maps a canonical key string — spec section 8.3's own notation,
// e.g. "j", "gg", "Ctrl+d", "<Leader>l" — to the Action it triggers.
// Translating an actual terminal key event into one of these strings is
// Bubble Tea integration work this package doesn't do yet; Keymap is the
// framework-independent policy layer underneath that.
type Keymap map[string]Action

// DefaultKeymap returns spec section 8.3's full keybinding table.
func DefaultKeymap() Keymap {
	return Keymap{
		// Навигация
		"j":         ActionDown,
		"k":         ActionUp,
		"h":         ActionParent,
		"l":         ActionOpenSelected,
		"Enter":     ActionOpenSelected,
		"gg":        ActionFirstItem,
		"G":         ActionLastItem,
		"Ctrl+d":    ActionHalfPageDown,
		"Ctrl+u":    ActionHalfPageUp,
		"Tab":       ActionNextPanel,
		"Shift+Tab": ActionPrevPanel,

		// Поиск
		"/":         ActionSearchStart,
		"n":         ActionSearchNext,
		"N":         ActionSearchPrev,
		"Esc":       ActionCancel,
		"<Leader>f": ActionFuzzyFinder,

		// Записи
		"a":      ActionAddEntry,
		"A":      ActionAddGroup,
		"e":      ActionEditEntry,
		"dd":     ActionDelete,
		"u":      ActionUndo,
		"Ctrl+r": ActionRedo,
		"r":      ActionRevealPassword,
		"gp":     ActionOpenGenerator,

		// Копирование
		"yu": ActionCopyUsername,
		"yp": ActionCopyPassword,
		"yU": ActionCopyURL,
		"yn": ActionCopyNotes,
		"yt": ActionCopyTOTP,

		// Vault
		"<Leader>l": ActionLockVault,
		"<Leader>s": ActionSave,
		"<Leader>o": ActionOpenAnotherVault,
		":":         ActionCommandMode,
		"q":         ActionQuit,
	}
}

// ErrNoQuitBinding indicates the requested keymap would leave no key
// bound to ActionQuit. Spec section 21.3 requires an override to always
// "leave an emergency way to exit" — ApplyOverrides refuses to produce a
// keymap like that rather than silently trapping the user.
var ErrNoQuitBinding = errors.New("tui: keymap must keep at least one binding for quit")

// ErrNoLockBinding is ApplyOverrides' analogous check for ActionLockVault
// — spec 21.3 groups lock together with quit as bindings that must never
// disappear without an explicit warning.
var ErrNoLockBinding = errors.New("tui: keymap must keep at least one binding for lock")

// ApplyOverrides returns a new Keymap with overrides layered on top of
// base — a plain key-by-key merge, later entries winning, same as
// spec section 21.3 describes rebinding to work. base is left untouched.
//
// It only enforces the one concrete guarantee spec 21.3 states outright:
// the result must still have a binding for quit and for lock. Spec's
// broader "checks conflicts" and "shows the effective keymap in :help"
// both need a raw, unparsed override source (to point at exactly what the
// user typed) and a :help renderer — neither exists yet, so they are not
// implemented here.
func ApplyOverrides(base Keymap, overrides map[string]Action) (Keymap, error) {
	merged := make(Keymap, len(base)+len(overrides))
	for key, action := range base {
		merged[key] = action
	}
	for key, action := range overrides {
		merged[key] = action
	}

	if !merged.hasBinding(ActionQuit) {
		return nil, ErrNoQuitBinding
	}
	if !merged.hasBinding(ActionLockVault) {
		return nil, ErrNoLockBinding
	}
	return merged, nil
}

func (km Keymap) hasBinding(action Action) bool {
	for _, a := range km {
		if a == action {
			return true
		}
	}
	return false
}
