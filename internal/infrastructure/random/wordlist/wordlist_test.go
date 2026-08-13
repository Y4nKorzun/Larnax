package wordlist

import "testing"

func TestEFFLargeHasExactCount(t *testing.T) {
	if got := len(EFFLarge); got != 7776 {
		t.Errorf("len(EFFLarge) = %d, want 7776", got)
	}
}

func TestEFFLargeHasNoDuplicates(t *testing.T) {
	seen := make(map[string]bool, len(EFFLarge))
	for _, w := range EFFLarge {
		if seen[w] {
			t.Fatalf("duplicate word in EFFLarge: %q", w)
		}
		seen[w] = true
	}
}

func TestEFFLargeWordsAreNonEmpty(t *testing.T) {
	for i, w := range EFFLarge {
		if w == "" {
			t.Fatalf("EFFLarge[%d] is empty", i)
		}
	}
}

func TestEFFLargeContainsKnownWords(t *testing.T) {
	// velvet/cactus/lantern/walnut/harbor are all drawn from spec section
	// 7.5's example passphrase; "orbit" from that same example is
	// deliberately not included here — it does not actually appear in the
	// real EFF Long Wordlist, so spec's illustrative text and the actual
	// source data disagree on that one word.
	known := []string{"abacus", "zoom", "velvet", "cactus", "lantern", "walnut", "harbor"}
	set := make(map[string]bool, len(EFFLarge))
	for _, w := range EFFLarge {
		set[w] = true
	}
	for _, w := range known {
		if !set[w] {
			t.Errorf("expected EFFLarge to contain %q", w)
		}
	}
}

func TestParseWordlistRejectsMalformedLine(t *testing.T) {
	_, err := parseWordlist("11111\tabacus\nnotabinaryfield\n")
	if err == nil {
		t.Fatal("parseWordlist() error = nil, want an error for a line missing the tab-separated dice roll")
	}
}

func TestParseWordlistHandlesTrailingNewline(t *testing.T) {
	words, err := parseWordlist("11111\tabacus\n11112\tabdomen\n")
	if err != nil {
		t.Fatalf("parseWordlist() error = %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("len(words) = %d, want 2 (trailing newline must not produce a spurious empty word)", len(words))
	}
	if words[0] != "abacus" || words[1] != "abdomen" {
		t.Errorf("words = %v, want [abacus abdomen]", words)
	}
}
