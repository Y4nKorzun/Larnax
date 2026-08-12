package tui

// Action identifies a NORMAL-mode action resolved from a completed
// multi-key sequence (spec section 8.3). It is deliberately not the same
// type as Command: Action comes from raw keystrokes, Command from COMMAND
// mode text.
type Action string

const (
	ActionFirstItem    Action = "first-item"    // gg
	ActionDelete       Action = "delete"        // dd
	ActionCopyUsername Action = "copy-username" // yu
	ActionCopyPassword Action = "copy-password" // yp
	ActionCopyURL      Action = "copy-url"      // yU
	ActionCopyNotes    Action = "copy-notes"    // yn
	ActionCopyTOTP     Action = "copy-totp"     // yt
)

var sequenceActions = map[[2]rune]Action{
	{'g', 'g'}: ActionFirstItem,
	{'d', 'd'}: ActionDelete,
	{'y', 'u'}: ActionCopyUsername,
	{'y', 'p'}: ActionCopyPassword,
	{'y', 'U'}: ActionCopyURL,
	{'y', 'n'}: ActionCopyNotes,
	{'y', 't'}: ActionCopyTOTP,
}

// KeySequence recognizes the two-key vim-like sequences from spec section
// 8.3 (gg, dd, yu, yp, yU, yn, yt). None of spec's sequences are longer
// than two keys, so it holds at most one pending prefix rune.
//
// KeySequence only detects sequences — it has no opinion on single keys
// that never participate in one (j, k, h, l, G, Enter, ...). Feed reports
// Pending()==false and resolved==false for those, and the caller dispatches
// them through its own single-key keymap.
type KeySequence struct {
	pending rune
}

// Feed processes one key press and returns:
//   - (action, true) if key completed a recognized sequence — the machine
//     returns to idle;
//   - ("", false) with Pending() now true if key started or continued
//     toward a recognized sequence;
//   - ("", false) with Pending() now false if key doesn't participate in
//     any sequence, or completed an unrecognized one.
//
// An unrecognized completion aborts the pending prefix and consumes the
// completing key — it does not fall through as a standalone action —
// except when that key is itself a valid prefix, in which case a new
// sequence starts from it. For example "dy" is not a sequence: it aborts
// the pending "d", but the "y" still begins a fresh sequence, so "dyp"
// resolves to ActionCopyPassword via "yp".
func (s *KeySequence) Feed(key rune) (action Action, resolved bool) {
	if s.pending == 0 {
		if isPrefixKey(key) {
			s.pending = key
		}
		return "", false
	}

	prefix := s.pending
	s.pending = 0

	if act, ok := sequenceActions[[2]rune{prefix, key}]; ok {
		return act, true
	}

	if isPrefixKey(key) {
		s.pending = key
	}
	return "", false
}

// Pending reports whether the machine is waiting for a second key to
// complete a sequence.
func (s *KeySequence) Pending() bool {
	return s.pending != 0
}

// Reset aborts any pending sequence, e.g. on Esc or a mode change the TUI
// layer detects on its own.
func (s *KeySequence) Reset() {
	s.pending = 0
}

func isPrefixKey(key rune) bool {
	switch key {
	case 'g', 'd', 'y':
		return true
	default:
		return false
	}
}
