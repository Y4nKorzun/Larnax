//go:build !darwin

package clipboard

// platformClipboard has no real implementation yet on this platform —
// spec sections 6.2 and 24.5 name Windows and Linux X11/Wayland adapters
// as separate work still to do, alongside the macOS one pbcopy_darwin.go
// already provides.
func platformClipboard() (Clipboard, bool) {
	return nil, false
}
