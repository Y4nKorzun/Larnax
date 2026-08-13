package application

import (
	"errors"
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random/wordlist"
)

var ErrPassphraseWordCountTooLow = errors.New("application: passphrase word count must be at least 1")

// PassphrasePolicy configures the passphrase generator (spec sections 7.4
// for the master passphrase and 10.5 for per-site passphrases — both use
// this same mechanism; spec section 10.5 is explicit that the app must
// never derive one from the other, so nothing here retains state between
// calls).
type PassphrasePolicy struct {
	WordCount int
}

// GeneratePassphrase builds a passphrase from policy.WordCount
// independently, uniformly randomly chosen words from the EFF Long
// Wordlist, joined with hyphens per spec section 7.5's canonical
// representation (e.g.
// "velvet-orbit-cactus-lantern-walnut-engine-harbor-rabbit"). No hidden
// normalization is applied: spec section 7.5 requires the exact generated
// string to be what gets shown, stored, and re-typed, so this does not
// lowercase, trim, or otherwise touch the joined result beyond the join
// itself.
func GeneratePassphrase(src random.Source, policy PassphrasePolicy) (domain.Secret, error) {
	if policy.WordCount < 1 {
		return nil, ErrPassphraseWordCountTooLow
	}

	return domain.NewSecretFromString(strings.Join(drawWords(src, policy.WordCount), "-")), nil
}

// drawWords picks n independent, uniformly random words from the EFF
// Long Wordlist. Shared by GeneratePassphrase above and
// GenerateMasterPassphraseWithWords (master_passphrase.go), which needs
// the individual words themselves rather than only the joined result.
func drawWords(src random.Source, n int) []string {
	words := make([]string, n)
	for i := range words {
		words[i] = wordlist.EFFLarge[src.Intn(len(wordlist.EFFLarge))]
	}
	return words
}
