package googlecsv

import "testing"

func TestNormalizeOriginStripsPathQueryFragment(t *testing.T) {
	got := normalizeOrigin("https://github.com/login?next=/settings#top")
	want := "https://github.com"
	if got != want {
		t.Errorf("normalizeOrigin() = %q, want %q", got, want)
	}
}

func TestNormalizeOriginLowercasesSchemeAndHost(t *testing.T) {
	got := normalizeOrigin("HTTPS://GitHub.COM")
	want := "https://github.com"
	if got != want {
		t.Errorf("normalizeOrigin() = %q, want %q", got, want)
	}
}

func TestNormalizeOriginStripsWWW(t *testing.T) {
	got := normalizeOrigin("https://www.github.com")
	want := "https://github.com"
	if got != want {
		t.Errorf("normalizeOrigin() = %q, want %q", got, want)
	}
}

func TestNormalizeOriginDropsDefaultPort(t *testing.T) {
	cases := []string{
		"https://github.com:443",
		"http://github.com:80",
	}
	wants := []string{
		"https://github.com",
		"http://github.com",
	}
	for i, in := range cases {
		if got := normalizeOrigin(in); got != wants[i] {
			t.Errorf("normalizeOrigin(%q) = %q, want %q", in, got, wants[i])
		}
	}
}

func TestNormalizeOriginKeepsNonDefaultPort(t *testing.T) {
	got := normalizeOrigin("https://github.com:8443")
	want := "https://github.com:8443"
	if got != want {
		t.Errorf("normalizeOrigin() = %q, want %q", got, want)
	}
}

func TestNormalizeOriginDefaultsMissingScheme(t *testing.T) {
	got := normalizeOrigin("github.com")
	want := "https://github.com"
	if got != want {
		t.Errorf("normalizeOrigin() = %q, want %q", got, want)
	}
}

func TestNormalizeOriginEmptyInput(t *testing.T) {
	if got := normalizeOrigin(""); got != "" {
		t.Errorf("normalizeOrigin(\"\") = %q, want empty", got)
	}
	if got := normalizeOrigin("   "); got != "" {
		t.Errorf("normalizeOrigin(whitespace) = %q, want empty", got)
	}
}

func TestNormalizeOriginInvalidInput(t *testing.T) {
	// A control character makes this an invalid URL per net/url.
	if got := normalizeOrigin("https://exa\x7fmple.com"); got != "" {
		t.Errorf("normalizeOrigin(invalid) = %q, want empty", got)
	}
}

// TestNormalizeOriginTreatsEquivalentURLsIdentically is the actual property
// spec section 13.6 needs: URLs a human would consider "the same site" must
// normalize to the same value, since that's what makes duplicate detection
// work at all.
func TestNormalizeOriginTreatsEquivalentURLsIdentically(t *testing.T) {
	equivalents := []string{
		"https://github.com",
		"https://www.github.com",
		"https://GitHub.com/login",
		"https://www.github.com:443/settings?tab=security",
		"github.com",
		"www.github.com",
	}

	want := normalizeOrigin(equivalents[0])
	for _, in := range equivalents[1:] {
		if got := normalizeOrigin(in); got != want {
			t.Errorf("normalizeOrigin(%q) = %q, want %q (same origin as %q)",
				in, got, want, equivalents[0])
		}
	}
}

func TestNormalizeOriginDistinguishesDifferentOrigins(t *testing.T) {
	a := normalizeOrigin("https://github.com")
	b := normalizeOrigin("https://gitlab.com")
	if a == b {
		t.Errorf("normalizeOrigin() collapsed distinct hosts to the same value: %q", a)
	}

	c := normalizeOrigin("http://github.com")
	if a == c {
		t.Errorf("normalizeOrigin() collapsed http and https to the same value: %q", a)
	}
}

func TestNormalizeUsernameTrimsAndLowercases(t *testing.T) {
	got := normalizeUsername("  User@Example.com  ")
	want := "user@example.com"
	if got != want {
		t.Errorf("normalizeUsername() = %q, want %q", got, want)
	}
}
