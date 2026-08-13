package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

func TestChooseRecoveryChallengeReturnsDistinctInRangePositions(t *testing.T) {
	positions, err := ChooseRecoveryChallenge(random.CryptoSource{}, 8, 3)
	if err != nil {
		t.Fatalf("ChooseRecoveryChallenge() error = %v", err)
	}
	if len(positions) != 3 {
		t.Fatalf("len(positions) = %d, want 3", len(positions))
	}

	seen := map[int]bool{}
	for _, p := range positions {
		if p < 1 || p > 8 {
			t.Errorf("position %d out of range [1,8]", p)
		}
		if seen[p] {
			t.Errorf("position %d appears more than once: %v", p, positions)
		}
		seen[p] = true
	}
}

func TestChooseRecoveryChallengeIsAscending(t *testing.T) {
	positions, err := ChooseRecoveryChallenge(random.CryptoSource{}, 12, 5)
	if err != nil {
		t.Fatalf("ChooseRecoveryChallenge() error = %v", err)
	}
	for i := 1; i < len(positions); i++ {
		if positions[i-1] >= positions[i] {
			t.Errorf("positions not strictly ascending: %v", positions)
		}
	}
}

func TestChooseRecoveryChallengeAllowsFullCoverage(t *testing.T) {
	positions, err := ChooseRecoveryChallenge(random.CryptoSource{}, 6, 6)
	if err != nil {
		t.Fatalf("ChooseRecoveryChallenge() error = %v", err)
	}
	want := []int{1, 2, 3, 4, 5, 6}
	for i, p := range positions {
		if p != want[i] {
			t.Errorf("positions = %v, want %v", positions, want)
			break
		}
	}
}

func TestChooseRecoveryChallengeRejectsCountBelowOne(t *testing.T) {
	if _, err := ChooseRecoveryChallenge(random.CryptoSource{}, 8, 0); !errors.Is(err, ErrChallengeCountTooLow) {
		t.Errorf("error = %v, want %v", err, ErrChallengeCountTooLow)
	}
}

func TestChooseRecoveryChallengeRejectsCountAboveWordCount(t *testing.T) {
	if _, err := ChooseRecoveryChallenge(random.CryptoSource{}, 8, 9); !errors.Is(err, ErrChallengeCountExceedsWordCount) {
		t.Errorf("error = %v, want %v", err, ErrChallengeCountExceedsWordCount)
	}
}

func TestVerifyRecoveryAnswersAllCorrect(t *testing.T) {
	words := []string{"velvet", "orbit", "cactus", "lantern"}
	ok, err := VerifyRecoveryAnswers(words, map[int]string{2: "orbit", 4: "lantern"})
	if err != nil {
		t.Fatalf("VerifyRecoveryAnswers() error = %v", err)
	}
	if !ok {
		t.Error("VerifyRecoveryAnswers() = false, want true")
	}
}

func TestVerifyRecoveryAnswersWrongWord(t *testing.T) {
	words := []string{"velvet", "orbit", "cactus", "lantern"}
	ok, err := VerifyRecoveryAnswers(words, map[int]string{2: "wrong"})
	if err != nil {
		t.Fatalf("VerifyRecoveryAnswers() error = %v", err)
	}
	if ok {
		t.Error("VerifyRecoveryAnswers() = true, want false")
	}
}

func TestVerifyRecoveryAnswersIsCaseSensitive(t *testing.T) {
	words := []string{"velvet", "orbit"}
	ok, err := VerifyRecoveryAnswers(words, map[int]string{1: "Velvet"})
	if err != nil {
		t.Fatalf("VerifyRecoveryAnswers() error = %v", err)
	}
	if ok {
		t.Error("VerifyRecoveryAnswers() = true for a case mismatch, want false (spec 7.5 forbids normalization)")
	}
}

func TestVerifyRecoveryAnswersRejectsOutOfRangePosition(t *testing.T) {
	words := []string{"velvet", "orbit"}
	if _, err := VerifyRecoveryAnswers(words, map[int]string{5: "x"}); !errors.Is(err, ErrRecoveryPositionOutOfRange) {
		t.Errorf("error = %v, want %v", err, ErrRecoveryPositionOutOfRange)
	}
	if _, err := VerifyRecoveryAnswers(words, map[int]string{0: "x"}); !errors.Is(err, ErrRecoveryPositionOutOfRange) {
		t.Errorf("error = %v, want %v", err, ErrRecoveryPositionOutOfRange)
	}
}

func TestGenerateMasterPassphraseWithWordsMatchesJoinedPhrase(t *testing.T) {
	generated, err := GenerateMasterPassphraseWithWords(random.CryptoSource{}, MasterPassphraseRecommended)
	if err != nil {
		t.Fatalf("GenerateMasterPassphraseWithWords() error = %v", err)
	}
	if len(generated.Words) != 8 {
		t.Fatalf("len(Words) = %d, want 8", len(generated.Words))
	}

	want := strings.Join(generated.Words, "-")
	if got := string(revealBytes(t, generated.Phrase)); got != want {
		t.Errorf("Phrase = %q, want %q", got, want)
	}
}

func TestGenerateMasterPassphraseWithWordsRejectsUnsafeStrength(t *testing.T) {
	if _, err := GenerateMasterPassphraseWithWords(random.CryptoSource{}, 4); !errors.Is(err, ErrUnsafeMasterPassphraseStrength) {
		t.Errorf("error = %v, want %v", err, ErrUnsafeMasterPassphraseStrength)
	}
}

// TestFullRecoveryFlow exercises generate -> choose challenge -> verify
// answers end to end, matching how the wizard actually uses these
// together.
func TestFullRecoveryFlow(t *testing.T) {
	generated, err := GenerateMasterPassphraseWithWords(random.CryptoSource{}, MasterPassphraseStrong)
	if err != nil {
		t.Fatalf("GenerateMasterPassphraseWithWords() error = %v", err)
	}

	positions, err := ChooseRecoveryChallenge(random.CryptoSource{}, len(generated.Words), 2)
	if err != nil {
		t.Fatalf("ChooseRecoveryChallenge() error = %v", err)
	}

	correctAnswers := make(map[int]string, len(positions))
	for _, p := range positions {
		correctAnswers[p] = generated.Words[p-1]
	}
	ok, err := VerifyRecoveryAnswers(generated.Words, correctAnswers)
	if err != nil {
		t.Fatalf("VerifyRecoveryAnswers() error = %v", err)
	}
	if !ok {
		t.Error("correct answers were rejected")
	}

	wrongAnswers := make(map[int]string, len(positions))
	for _, p := range positions {
		wrongAnswers[p] = generated.Words[p-1] + "-typo"
	}
	ok, err = VerifyRecoveryAnswers(generated.Words, wrongAnswers)
	if err != nil {
		t.Fatalf("VerifyRecoveryAnswers() error = %v", err)
	}
	if ok {
		t.Error("wrong answers were accepted")
	}
}
