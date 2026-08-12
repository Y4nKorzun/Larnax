// Package terminal makes untrusted text (from a KDBX file or an imported
// CSV) safe to write to a terminal the user's shell, multiplexer, or
// emulator is attached to.
package terminal

import "strings"

// Sanitize is a required security gate, not a cosmetic feature (spec
// section 8.6): no field from KDBX or CSV may reach the screen as trusted
// ANSI. It removes:
//   - invalid UTF-8 byte sequences;
//   - ESC (0x1B) and every other C0 control character except tab and
//     newline, so no ANSI/OSC/CSI escape sequence can be introduced;
//   - DEL (0x7F);
//   - C1 control characters (0x80-0x9F), which some terminals treat as
//     escape introducers even without a literal ESC byte.
//
// Tab and newline survive, matching spec section 8.6's "safe line breaks
// and tab" allowance for notes. CRLF and bare CR both become LF: a bare
// carriage return moves the cursor to the start of the line without
// advancing it, letting already-printed text be overwritten in place, so
// it is treated as the line break it almost always was meant to be rather
// than passed through unchanged.
func Sanitize(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n':
			b.WriteRune(r)
		case r < 0x20, r == 0x7f:
			// remaining C0 control characters (ESC, BEL, backspace, ...) and DEL
		case r >= 0x80 && r <= 0x9f:
			// C1 control characters
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
