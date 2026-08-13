package tui

// Mode is one of the six input contexts spec section 8.2 defines. Like
// commands.go and key_sequence.go, this file has no Bubble Tea dependency
// — it only encodes which mode transitions are valid; wiring it to an
// actual running TUI is later work.
type Mode int

const (
	NORMAL Mode = iota
	INSERT
	SEARCH
	COMMAND
	CONFIRM
	LOCKED
)

func (m Mode) String() string {
	switch m {
	case NORMAL:
		return "NORMAL"
	case INSERT:
		return "INSERT"
	case SEARCH:
		return "SEARCH"
	case COMMAND:
		return "COMMAND"
	case CONFIRM:
		return "CONFIRM"
	case LOCKED:
		return "LOCKED"
	default:
		return "UNKNOWN"
	}
}

// validTransitions encodes which mode changes spec section 8.2 allows.
//
// Spec describes what each mode is for but gives no explicit transition
// table, so this is this package's own reading of it, cross-checked
// against section 8.3's keybindings and section 17's lock behavior:
//
//   - LOCKED is reachable from every mode, including itself: spec section
//     17.3 requires inactivity lock and manual lock (<Leader>l) to
//     interrupt whatever the user is doing, with no exception carved out
//     for any mode — and a lock event firing again while already LOCKED
//     (e.g. a stale inactivity timer) is a harmless no-op, not an error.
//   - LOCKED otherwise only ever leads back to NORMAL — spec section 7.8
//     has a failed unlock attempt stay in LOCKED, never fall into some
//     other mode.
//   - INSERT: "обычные клавиши печатают текст, а не запускают команды"
//     (spec 8.2) — no keybinding reaches SEARCH, COMMAND, or CONFIRM
//     while typing a field, so the only voluntary way out is back to
//     NORMAL (Esc or committing the edit).
//   - SEARCH and COMMAND are both entered from NORMAL (`/` and `:`,
//     spec 8.3) and return to it on Esc or on completing the
//     search/command.
//   - COMMAND can also lead straight to LOCKED, since `:lock` is itself
//     one of the listed commands (spec 8.4) rather than something that
//     first returns to NORMAL.
//   - CONFIRM is entered from NORMAL for a dangerous action (delete,
//     overwrite, import, KDF change — spec 8.2) and always resolves back
//     to NORMAL, whichever way the user answers.
var validTransitions = map[Mode]map[Mode]bool{
	NORMAL:  {NORMAL: true, INSERT: true, SEARCH: true, COMMAND: true, CONFIRM: true, LOCKED: true},
	INSERT:  {NORMAL: true, LOCKED: true},
	SEARCH:  {NORMAL: true, LOCKED: true},
	COMMAND: {NORMAL: true, LOCKED: true},
	CONFIRM: {NORMAL: true, LOCKED: true},
	LOCKED:  {NORMAL: true, LOCKED: true},
}

// CanTransition reports whether moving from the from mode to next is a
// valid mode change.
func CanTransition(from, next Mode) bool {
	return validTransitions[from][next]
}
