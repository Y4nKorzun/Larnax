package application

import (
	"math"
	"testing"
)

// TestEstimateGeneratedStrengthMatchesSpecPassphraseTable cross-checks
// against spec section 7.4's own published entropy table for the EFF Long
// Wordlist (7776 words): 4 words ~51.7 bits, 6 ~77.5, 8 ~103.4, 12 ~155.1.
func TestEstimateGeneratedStrengthMatchesSpecPassphraseTable(t *testing.T) {
	cases := []struct {
		words    int
		wantBits float64
	}{
		{4, 51.7},
		{6, 77.5},
		{8, 103.4},
		{12, 155.1},
	}

	for _, c := range cases {
		got := EstimateGeneratedStrength(c.words, 7776, "word", "wordlist")
		if diff := math.Abs(got.Bits - c.wantBits); diff > 0.05 {
			t.Errorf("EstimateGeneratedStrength(%d, 7776, ...).Bits = %.4f, want ~%.1f (diff %.4f)",
				c.words, got.Bits, c.wantBits, diff)
		}
	}
}

func TestEstimateGeneratedStrengthReasonText(t *testing.T) {
	cases := []struct {
		name       string
		unitCount  int
		spaceSize  int
		unitNoun   string
		spaceNoun  string
		wantReason string
	}{
		{
			name:       "plural characters",
			unitCount:  24,
			spaceSize:  94,
			unitNoun:   "character",
			spaceNoun:  "alphabet",
			wantReason: "24 independently chosen characters from a 94-character alphabet.",
		},
		{
			name:       "plural words",
			unitCount:  8,
			spaceSize:  7776,
			unitNoun:   "word",
			spaceNoun:  "wordlist",
			wantReason: "8 independently chosen words from a 7776-word wordlist.",
		},
		{
			name:       "singular unit has no trailing s",
			unitCount:  1,
			spaceSize:  26,
			unitNoun:   "character",
			spaceNoun:  "alphabet",
			wantReason: "1 independently chosen character from a 26-character alphabet.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EstimateGeneratedStrength(c.unitCount, c.spaceSize, c.unitNoun, c.spaceNoun)
			if got.Reason != c.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, c.wantReason)
			}
		})
	}
}

func TestStrengthLevelTiers(t *testing.T) {
	cases := []struct {
		bits float64
		want StrengthLevel
	}{
		{0, StrengthLow},
		{39.9, StrengthLow},
		{40, StrengthMedium},
		{74.9, StrengthMedium},
		{75, StrengthHigh},
		{127.9, StrengthHigh},
		{128, StrengthVeryHigh},
		{256, StrengthVeryHigh},
	}
	for _, c := range cases {
		if got := strengthLevel(c.bits); got != c.want {
			t.Errorf("strengthLevel(%.1f) = %q, want %q", c.bits, got, c.want)
		}
	}
}

func TestEstimatePasswordStrengthMatchesActualAlphabet(t *testing.T) {
	policy := DefaultPasswordPolicy() // length 24, all four classes

	got := EstimatePasswordStrength(policy)

	const combinedAlphabetSize = 26 + 26 + 10 + 32 // lower + upper + digits + symbols
	wantBits := float64(policy.Length) * math.Log2(float64(combinedAlphabetSize))
	if diff := math.Abs(got.Bits - wantBits); diff > 0.01 {
		t.Errorf("Bits = %.4f, want %.4f (diff %.4f)", got.Bits, wantBits, diff)
	}
	if got.Level != StrengthVeryHigh {
		t.Errorf("Level = %q, want %q for a 24-character 94-alphabet password", got.Level, StrengthVeryHigh)
	}
}

func TestEstimatePasswordStrengthSingleClass(t *testing.T) {
	policy := PasswordPolicy{Length: 12, Digits: true}

	got := EstimatePasswordStrength(policy)

	wantBits := 12 * math.Log2(10)
	if diff := math.Abs(got.Bits - wantBits); diff > 0.01 {
		t.Errorf("Bits = %.4f, want %.4f", got.Bits, wantBits)
	}
}

func TestEstimatePasswordStrengthReflectsAvoidAmbiguous(t *testing.T) {
	full := PasswordPolicy{Length: 20, Lowercase: true}
	reduced := PasswordPolicy{Length: 20, Lowercase: true, AvoidAmbiguous: true}

	fullStrength := EstimatePasswordStrength(full)
	reducedStrength := EstimatePasswordStrength(reduced)

	if reducedStrength.Bits >= fullStrength.Bits {
		t.Errorf("AvoidAmbiguous strength (%.2f bits) should be lower than the full alphabet (%.2f bits)",
			reducedStrength.Bits, fullStrength.Bits)
	}
}
