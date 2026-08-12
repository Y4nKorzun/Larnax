package terminal

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSanitizeRemovesANSIColorEscape(t *testing.T) {
	got := Sanitize("\x1b[31mRED\x1b[0m")
	if strings.ContainsRune(got, 0x1b) {
		t.Fatalf("Sanitize() output still contains ESC: %q", got)
	}
	want := "[31mRED[0m"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizeRemovesOSCTitleInjection(t *testing.T) {
	// ESC ] 0 ; <title> BEL sets the terminal window title on many emulators.
	got := Sanitize("\x1b]0;pwned\x07Title")
	if strings.ContainsAny(got, "\x1b\x07") {
		t.Fatalf("Sanitize() output still contains ESC or BEL: %q", got)
	}
	want := "]0;pwnedTitle"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizeStripsC1Controls(t *testing.T) {
	// U+009B is CSI in the single-byte C1 form some terminals still honor.
	got := Sanitize("before31mafter")
	want := "before31mafter"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizePreservesTabAndNewline(t *testing.T) {
	in := "line1\tindented\nline2"
	if got := Sanitize(in); got != in {
		t.Errorf("Sanitize(%q) = %q, want unchanged", in, got)
	}
}

func TestSanitizeNormalizesCRLF(t *testing.T) {
	got := Sanitize("line1\r\nline2")
	want := "line1\nline2"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizeConvertsBareCRToNewline(t *testing.T) {
	got := Sanitize("line1\rline2")
	want := "line1\nline2"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizeStripsDEL(t *testing.T) {
	got := Sanitize("abc\x7fdef")
	want := "abcdef"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizeDropsInvalidUTF8(t *testing.T) {
	got := Sanitize(string([]byte{'A', 0xff, 'B'}))
	want := "AB"
	if got != want {
		t.Errorf("Sanitize() = %q, want %q", got, want)
	}
}

func TestSanitizePreservesOrdinaryUnicode(t *testing.T) {
	in := "héllo wörld 日本語 🎉"
	if got := Sanitize(in); got != in {
		t.Errorf("Sanitize(%q) = %q, want unchanged", in, got)
	}
}

// FuzzSanitize encodes spec section 24.2's invariant for this exact
// function: the sanitizer must never let a terminal escape through, and
// must never panic, no matter what bytes a KDBX or CSV field contains.
func FuzzSanitize(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"\x1b[31mred\x1b[0m",
		"\x1b]0;title\x07",
		string([]byte{0xff, 0xfe, 0x00}),
		"line1\r\nline2\rline3",
		"日本語 emoji 🎉",
		"\x9b31m",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		out := Sanitize(s)

		if !utf8.ValidString(out) {
			t.Fatalf("Sanitize(%q) output is not valid UTF-8: %q", s, out)
		}

		for _, r := range out {
			switch {
			case r == '\t' || r == '\n':
				// explicitly allowed
			case r < 0x20, r == 0x7f:
				t.Fatalf("Sanitize(%q) output contains C0/DEL %#U: %q", s, r, out)
			case r >= 0x80 && r <= 0x9f:
				t.Fatalf("Sanitize(%q) output contains C1 control %#U: %q", s, r, out)
			}
		}
	})
}
