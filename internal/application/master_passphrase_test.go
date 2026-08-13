package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random/wordlist"
)

// fixedIndexSource always returns the same wordlist index, so a test can
// predict the exact passphrase GeneratePassphrase produces instead of
// trying to recover word count from the joined string — the EFF Long
// Wordlist itself contains a handful of hyphenated entries ("drop-down",
// "t-shirt", ...), so splitting the result on "-" to count words is not
// reliable.
type fixedIndexSource int

func (f fixedIndexSource) Intn(int) int              { return int(f) }
func (fixedIndexSource) Shuffle(int, func(i, j int)) {}

// nonHyphenatedWordIndex returns the index of some wordlist entry that
// itself contains no hyphen, so tests using fixedIndexSource can build an
// unambiguous expected passphrase string.
func nonHyphenatedWordIndex(t *testing.T) int {
	t.Helper()
	for i, w := range wordlist.EFFLarge {
		if !strings.Contains(w, "-") {
			return i
		}
	}
	t.Fatal("wordlist has no non-hyphenated word")
	return 0
}

func TestGenerateMasterPassphraseAcceptsNormalStrengths(t *testing.T) {
	idx := nonHyphenatedWordIndex(t)
	word := wordlist.EFFLarge[idx]
	src := fixedIndexSource(idx)

	for _, strength := range []MasterPassphraseStrength{
		MasterPassphraseStrong, MasterPassphraseRecommended, MasterPassphraseMaximum,
	} {
		secret, err := GenerateMasterPassphrase(src, strength)
		if err != nil {
			t.Fatalf("GenerateMasterPassphrase(%d) error = %v", strength, err)
		}
		want := strings.Repeat(word+"-", int(strength)-1) + word
		if got := string(revealBytes(t, secret)); got != want {
			t.Errorf("GenerateMasterPassphrase(%d) = %q, want %q", strength, got, want)
		}
	}
}

func TestGenerateMasterPassphraseRejectsFourWords(t *testing.T) {
	if _, err := GenerateMasterPassphrase(random.CryptoSource{}, 4); !errors.Is(err, ErrUnsafeMasterPassphraseStrength) {
		t.Errorf("GenerateMasterPassphrase(4) error = %v, want %v", err, ErrUnsafeMasterPassphraseStrength)
	}
}

func TestGenerateMasterPassphraseRejectsArbitraryStrength(t *testing.T) {
	if _, err := GenerateMasterPassphrase(random.CryptoSource{}, 7); !errors.Is(err, ErrUnsafeMasterPassphraseStrength) {
		t.Errorf("GenerateMasterPassphrase(7) error = %v, want %v", err, ErrUnsafeMasterPassphraseStrength)
	}
}

func TestDefaultMasterPassphraseStrengthIsEight(t *testing.T) {
	if DefaultMasterPassphraseStrength != 8 {
		t.Errorf("DefaultMasterPassphraseStrength = %d, want 8", DefaultMasterPassphraseStrength)
	}
}

func TestGenerateMasterPassphraseUnsafeAllowsFourWords(t *testing.T) {
	idx := nonHyphenatedWordIndex(t)
	word := wordlist.EFFLarge[idx]

	secret, err := GenerateMasterPassphraseUnsafe(fixedIndexSource(idx), 4)
	if err != nil {
		t.Fatalf("GenerateMasterPassphraseUnsafe(4) error = %v", err)
	}
	want := strings.Repeat(word+"-", 3) + word
	if got := string(revealBytes(t, secret)); got != want {
		t.Errorf("GenerateMasterPassphraseUnsafe(4) = %q, want %q", got, want)
	}
}

func TestGenerateMasterPassphraseUnsafeRejectsZero(t *testing.T) {
	if _, err := GenerateMasterPassphraseUnsafe(random.CryptoSource{}, 0); !errors.Is(err, ErrPassphraseWordCountTooLow) {
		t.Errorf("GenerateMasterPassphraseUnsafe(0) error = %v, want %v", err, ErrPassphraseWordCountTooLow)
	}
}
