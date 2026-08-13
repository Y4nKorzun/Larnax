//go:build darwin

package clipboard

import (
	"bytes"
	"context"
	"os/exec"
)

// DarwinClipboard implements Clipboard via macOS's pbcopy/pbpaste — the
// same mechanism every terminal-based macOS clipboard tool uses. No CGo,
// no external Go dependency: pbcopy and pbpaste ship with every macOS
// install.
type DarwinClipboard struct{}

// WriteText writes value to the system pasteboard via pbcopy. An empty
// value is a valid write (an empty clipboard), not a special case —
// Clear below is just WriteText(ctx, nil).
func (DarwinClipboard) WriteText(ctx context.Context, value []byte) error {
	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = bytes.NewReader(value)
	return cmd.Run()
}

// ReadText reads the system pasteboard's current text content via
// pbpaste.
func (DarwinClipboard) ReadText(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "pbpaste")
	return cmd.Output()
}

// Clear sets the system pasteboard to empty. Callers needing spec section
// 11.3's "only clear if we still own it" behavior use ClearIfOwned
// (secure_copy.go) instead of calling this directly.
func (c DarwinClipboard) Clear(ctx context.Context) error {
	return c.WriteText(ctx, nil)
}

// platformClipboard backs New/Available (available.go) on macOS: a
// DarwinClipboard is available exactly when both pbcopy and pbpaste are
// on PATH, which is true on every real macOS install but worth checking
// rather than assuming.
func platformClipboard() (Clipboard, bool) {
	if _, err := exec.LookPath("pbcopy"); err != nil {
		return nil, false
	}
	if _, err := exec.LookPath("pbpaste"); err != nil {
		return nil, false
	}
	return DarwinClipboard{}, true
}
