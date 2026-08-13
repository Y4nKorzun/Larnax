package application

import (
	"errors"
	"sort"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

// Spec section 7.6: after the generated passphrase is shown once, the
// wizard asks the user to retype a few random word positions
// ("Word #3: ________", "Word #7: ________") before vault creation
// proceeds — proof the user actually wrote it down correctly, not just
// that the screen briefly displayed it.

var (
	ErrChallengeCountTooLow           = errors.New("application: recovery challenge count must be at least 1")
	ErrChallengeCountExceedsWordCount = errors.New("application: recovery challenge count cannot exceed the passphrase's word count")
	ErrRecoveryPositionOutOfRange     = errors.New("application: recovery answer position is out of range")
)

// ChooseRecoveryChallenge picks count distinct, uniformly random
// 1-indexed word positions out of wordCount total, returned in ascending
// order for display. Using src (the same crypto/rand-backed source
// everything else security-relevant in this package uses) rather than
// e.g. always asking for the first few positions matters here for the
// same reason spec section 10.1 forbids predictable randomness elsewhere:
// a fixed challenge would let an attacker who saw only part of a written
// note know exactly which words they still need.
func ChooseRecoveryChallenge(src random.Source, wordCount, count int) ([]int, error) {
	if count < 1 {
		return nil, ErrChallengeCountTooLow
	}
	if count > wordCount {
		return nil, ErrChallengeCountExceedsWordCount
	}

	indices := make([]int, wordCount)
	for i := range indices {
		indices[i] = i
	}
	src.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })

	positions := make([]int, count)
	for i := 0; i < count; i++ {
		positions[i] = indices[i] + 1 // spec's display is 1-indexed ("Word #3:")
	}
	sort.Ints(positions)
	return positions, nil
}

// VerifyRecoveryAnswers reports whether every answer (1-indexed word
// position -> what the user typed) exactly matches words at that
// position. The comparison is exact and case-sensitive: spec section 7.5
// forbids hidden normalization of the passphrase, and that applies just
// as much to checking it back as to generating it — a wrong answer must
// not be silently accepted as "close enough."
//
// It returns an error only for a caller mistake (a position outside
// [1, len(words)]), distinct from ok == false, which means the answers
// were well-formed but at least one word was wrong.
func VerifyRecoveryAnswers(words []string, answers map[int]string) (ok bool, err error) {
	for position, answer := range answers {
		if position < 1 || position > len(words) {
			return false, ErrRecoveryPositionOutOfRange
		}
		if words[position-1] != answer {
			return false, nil
		}
	}
	return true, nil
}
