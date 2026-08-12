// Package clipboard provides secret-aware clipboard access (spec section
// 11). Platform adapters implementing Clipboard for macOS, Windows, Linux
// X11, and Linux Wayland are a separate, later piece of work — each needs
// its own integration tests (spec section 24.5) before being trusted here.
package clipboard

import "context"

// Clipboard is the port each platform adapter implements (spec section
// 11.6). The TUI never calls a platform clipboard API directly.
type Clipboard interface {
	ReadText(ctx context.Context) ([]byte, error)
	WriteText(ctx context.Context, value []byte) error
	Clear(ctx context.Context) error
}
