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

func TestGeneratePassphraseWordCount(t *testing.T) {
	src := random.CryptoSource{}
	secret, err := GeneratePassphrase(src, PassphrasePolicy{WordCount: 8})
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}

	got := revealBytes(t, secret)
	words := strings.Split(string(got), "-")
	if len(words) != 8 {
		t.Errorf("word count = %d, want 8 (passphrase: %q)", len(words), got)
	}
}

func TestGeneratePassphraseWordsAreFromWordlist(t *testing.T) {
	src := random.CryptoSource{}
	set := make(map[string]bool, len(wordlist.EFFLarge))
	for _, w := range wordlist.EFFLarge {
		set[w] = true
	}

	secret, err := GeneratePassphrase(src, PassphrasePolicy{WordCount: 12})
	if err != nil {
		t.Fatalf("GeneratePassphrase() error = %v", err)
	}
	got := revealBytes(t, secret)

	for _, w := range strings.Split(string(got), "-") {
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

func TestGeneratePassphraseSingleWordHasNoHyphen(t *testing.T) {
	src := random.CryptoSource{}
	secret, err := GeneratePassphrase(src, PassphrasePolicy{WordCount: 1})
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
