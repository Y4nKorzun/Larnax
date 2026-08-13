package clipboard

// New returns a Clipboard implementation for the current platform, and
// whether one is actually available — spec section 22.2's :doctor
// "Clipboard: available/unavailable" check, and the first stop for any
// caller (like copy_field.go's eventual TUI wiring) that just wants a
// working Clipboard without caring which OS it's running on.
//
// Platform-specific detection lives in platformClipboard, one
// implementation per build-tagged file — pbcopy_darwin.go today, with
// Windows and Linux X11/Wayland adapters still separate future work
// (spec sections 6.2, 24.5).
func New() (Clipboard, bool) {
	return platformClipboard()
}

// Available is New without the Clipboard value, for a caller that only
// needs the yes/no.
func Available() bool {
	_, ok := New()
	return ok
}
