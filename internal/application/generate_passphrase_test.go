package application

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random/wordlist"
)

// TestGeneratePassphraseWordCount uses fixedIndexSource (defined in
// master_passphrase_test.go) rather than a real random source: the EFF
// Long Wordlist contains a handful of hyphenated entries ("drop-down",
// "t-shirt", ...), so counting words by splitting the joined result on
// "-" is unreliable with a genuinely random choice of words — it would
// make this test flaky in proportion to how often a hyphenated word gets
// drawn, instead of deterministically correct.
func TestGeneratePassphraseWordCount(t *testing.T) {
	idx := nonHyphenatedWordIndex(t)
	secret, err := GeneratePassphrase(fixedIndexSource(idx), PassphrasePolicy{WordCount: 8})
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}

	got := revealBytes(t, secret)
	words := strings.Split(string(got), "-")
	if len(words) != 8 {
		t.Errorf("word count = %d, want 8 (passphrase: %q)", len(words), got)
	}
}

// TestGeneratePassphraseWordsAreFromWordlist uses fixedIndexSource for
// the same reason TestGeneratePassphraseWordCount does: splitting the
// joined result on "-" to recover individual words is only unambiguous
// when none of the drawn words themselves contain a hyphen.
func TestGeneratePassphraseWordsAreFromWordlist(t *testing.T) {
	idx := nonHyphenatedWordIndex(t)
	word := wordlist.EFFLarge[idx]
	set := make(map[string]bool, len(wordlist.EFFLarge))
	for _, w := range wordlist.EFFLarge {
		set[w] = true
	}

	secret, err := GeneratePassphrase(fixedIndexSource(idx), PassphrasePolicy{WordCount: 12})
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}
	got := revealBytes(t, secret)

	for _, w := range strings.Split(string(got), "-") {
		if w != word {
			t.Errorf("word %q != expected fixed word %q", w, word)
		}
		if !set[w] {
			t.Errorf("word %q is not in the EFF wordlist", w)
		}
	}
}

func TestGeneratePassphraseVariesAcrossCalls(t *testing.T) {
	src := random.CryptoSource{}
	policy := PassphrasePolicy{WordCount: 6}

	first := revealBytes(t, mustGeneratePassphrase(t, src, policy))
	second := revealBytes(t, mustGeneratePassphrase(t, src, policy))

	if string(first) == string(second) {
		t.Errorf("two independent GeneratePassphrase() calls returned the same value: %q", first)
	}
}

func TestGeneratePassphraseRejectsNonPositiveWordCount(t *testing.T) {
	src := random.CryptoSource{}
	for _, n := range []int{0, -1} {
		_, err := GeneratePassphrase(src, PassphrasePolicy{WordCount: n})
		if !errors.Is(err, ErrPassphraseWordCountTooLow) {
			t.Errorf("GeneratePassphrase(WordCount: %d) error = %v, want %v", n, err, ErrPassphraseWordCountTooLow)
		}
	}
}

// TestGeneratePassphraseSingleWordHasNoHyphen uses fixedIndexSource
// rather than a real random source: a genuinely random single word could
// legitimately land on one of the wordlist's own hyphenated entries
// ("drop-down", "t-shirt", ...), which is correct WordCount:1 behavior,
// not a bug — this test is specifically about GeneratePassphrase not
// adding a hyphen of its own for a single word, so it needs a word that
// has none to begin with.
func TestGeneratePassphraseSingleWordHasNoHyphen(t *testing.T) {
	idx := nonHyphenatedWordIndex(t)
	secret, err := GeneratePassphrase(fixedIndexSource(idx), PassphrasePolicy{WordCount: 1})
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}
	got := string(revealBytes(t, secret))
	if strings.Contains(got, "-") {
		t.Errorf("single-word passphrase %q contains a hyphen", got)
	}
}

func TestEstimatePassphraseStrengthUsesActualWordlistSize(t *testing.T) {
	got := EstimatePassphraseStrength(PassphrasePolicy{WordCount: 8})

	wantBits := 8 * math.Log2(float64(len(wordlist.EFFLarge)))
	if diff := math.Abs(got.Bits - wantBits); diff > 0.0001 {
		t.Errorf("Bits = %.4f, want %.4f", got.Bits, wantBits)
	}
	if len(wordlist.EFFLarge) != 7776 {
		t.Fatalf("test assumption violated: len(wordlist.EFFLarge) = %d, want 7776", len(wordlist.EFFLarge))
	}
}

func mustGeneratePassphrase(t *testing.T, src random.Source, policy PassphrasePolicy) domain.Secret {
	t.Helper()
	secret, err := GeneratePassphrase(src, policy)
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}
	return secret
}
