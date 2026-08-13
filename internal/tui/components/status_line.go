package components

import (
	"fmt"
	"strings"
	"time"
)

// Lock states StatusLineModel.LockState shows, matching the "unlocked"
// spec section 8.1's own layout example uses, plus the two other states
// this application can be in: read-only (spec 15.4) and locked (spec
// 8.2's LOCKED mode).
const (
	LockStateUnlocked = "unlocked"
	LockStateReadOnly = "read-only"
	LockStateLocked   = "locked"
)

// StatusLineModel is spec section 8.1's status bar along the bottom of
// the main layout:
//
//	NORMAL   /Personal   3 entries   modified   clipboard: 11s   unlocked
//
// Unlike SecureInputModel and ConfirmModel, StatusLineModel has no
// Update: nothing about it responds to a keystroke directly. A parent
// model sets its fields as other state changes — mode switches, an edit
// happens, the clipboard countdown ticks.
type StatusLineModel struct {
	Mode               string // e.g. "NORMAL", matching tui.Mode.String()
	GroupPath          string // e.g. "/Personal"
	EntryCount         int
	Modified           bool
	ClipboardRemaining time.Duration // 0 means "not shown"
	LockState          string        // "" means "not shown"
}

// View renders the fields spec 8.1's example shows, each only when it
// has something to say: Modified only when true, clipboard only when
// ClipboardRemaining is positive, LockState only when set.
func (m StatusLineModel) View() string {
	parts := []string{m.Mode, m.GroupPath, fmt.Sprintf("%d entries", m.EntryCount)}
	if m.Modified {
		parts = append(parts, "modified")
	}
	if m.ClipboardRemaining > 0 {
		seconds := int(m.ClipboardRemaining.Round(time.Second) / time.Second)
		parts = append(parts, fmt.Sprintf("clipboard: %ds", seconds))
	}
	if m.LockState != "" {
		parts = append(parts, m.LockState)
	}
	return strings.Join(parts, "   ")
}
