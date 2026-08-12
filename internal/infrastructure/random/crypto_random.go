// Package random provides the project's sole source of randomness for
// values that must not be predictable: generated passwords, passphrases,
// and identifiers. Spec section 10.1 forbids math/rand, a timestamp seed,
// or any hand-rolled PRNG for these values.
package random

import (
	"crypto/rand"
	"math/big"
)

// Source provides cryptographically secure randomness. Production code
// uses CryptoSource; tests can substitute a different implementation to
// make generator output reproducible, without weakening the production
// path (spec section 18.6).
type Source interface {
	// Intn returns a uniform random integer in [0, n). It panics if n <= 0.
	Intn(n int) int

	// Shuffle randomizes the order of a length-n sequence via swap,
	// following Fisher-Yates (spec section 10.3).
	Shuffle(n int, swap func(i, j int))
}

// CryptoSource is a Source backed by crypto/rand.
type CryptoSource struct{}

// Intn returns a uniform random integer in [0, n) using crypto/rand.Int,
// which performs rejection sampling internally — this is the "crypto/rand.Int
// with an upper bound" approach spec section 10.3 requires, rather than
// taking a random byte modulo n, which would introduce modulo bias.
func (CryptoSource) Intn(n int) int {
	if n <= 0 {
		panic("random: Intn called with n <= 0")
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// crypto/rand failing indicates the OS entropy source is broken.
		// There is no safe fallback (spec section 10.1 forbids math/rand),
		// so there is nothing left to do but stop.
		panic("random: crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

// Shuffle implements Fisher-Yates (spec section 10.3, step 3): for each
// position from the end down to 1, swap it with a uniformly chosen earlier
// or equal position. Every permutation of a length-n sequence is equally
// likely.
func (s CryptoSource) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		j := s.Intn(i + 1)
		swap(i, j)
	}
}
