package application

import (
	"errors"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

// MasterPassphraseStrength is one of the choices spec section 7.3's
// normal create-vault wizard offers.
type MasterPassphraseStrength int

const (
	MasterPassphraseStrong      MasterPassphraseStrength = 6  // "strong"
	MasterPassphraseRecommended MasterPassphraseStrength = 8  // "recommended" — spec's default
	MasterPassphraseMaximum     MasterPassphraseStrength = 12 // "maximum"
)

// DefaultMasterPassphraseStrength is spec section 7.3's stated default.
const DefaultMasterPassphraseStrength = MasterPassphraseRecommended

// ErrUnsafeMasterPassphraseStrength is returned by GenerateMasterPassphrase
// for a strength spec section 7.3 says the normal wizard must never offer.
// Four words is spec section 10.5's per-site passphrase length — enough to
// protect one site, not the master passphrase that protects every secret
// in the vault at once. GenerateMasterPassphraseUnsafe exists separately,
// and named accordingly, so a caller can't reach a four-word master
// passphrase by accident.
var ErrUnsafeMasterPassphraseStrength = errors.New(
	"application: master passphrase strength must be 6, 8, or 12 in the normal wizard; use GenerateMasterPassphraseUnsafe with an explicit warning instead",
)

// GenerateMasterPassphrase generates a master passphrase at one of spec
// section 7.3's three normal-wizard strengths. Any other value —
// including 4, which spec explicitly keeps out of the normal wizard — is
// rejected with ErrUnsafeMasterPassphraseStrength.
func GenerateMasterPassphrase(src random.Source, strength MasterPassphraseStrength) (domain.Secret, error) {
	switch strength {
	case MasterPassphraseStrong, MasterPassphraseRecommended, MasterPassphraseMaximum:
		return GeneratePassphrase(src, PassphrasePolicy{WordCount: int(strength)})
	default:
		return nil, ErrUnsafeMasterPassphraseStrength
	}
}

// GenerateMasterPassphraseUnsafe generates a master passphrase at
// wordCount words with none of GenerateMasterPassphrase's strength
// restriction — spec section 7.3's "advanced/unsafe mode with an explicit
// warning" escape hatch. Showing that warning is the caller's
// responsibility; this function only removes the strength floor, not
// GeneratePassphrase's own requirement that wordCount be at least 1.
func GenerateMasterPassphraseUnsafe(src random.Source, wordCount int) (domain.Secret, error) {
	return GeneratePassphrase(src, PassphrasePolicy{WordCount: wordCount})
}
