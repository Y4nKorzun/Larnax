package application

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

const (
	lowercaseAlphabet = "abcdefghijklmnopqrstuvwxyz"
	uppercaseAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitsAlphabet    = "0123456789"
	symbolsAlphabet   = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	// ambiguousChars are characters commonly confused with each other in
	// most terminal fonts (spec section 10.2's "Avoid ambiguous" option).
	ambiguousChars = "0O1lI"
)

var (
	ErrNoCharacterClassSelected = errors.New("application: password policy selects no character classes")
	ErrPasswordLengthTooShort   = errors.New("application: password length is shorter than the number of selected character classes")
)

// PasswordPolicy configures the character password generator (spec section
// 10.2).
type PasswordPolicy struct {
	Length         int
	Lowercase      bool
	Uppercase      bool
	Digits         bool
	Symbols        bool
	AvoidAmbiguous bool
}

// DefaultPasswordPolicy matches spec section 10.2's default profile.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		Length:    24,
		Lowercase: true,
		Uppercase: true,
		Digits:    true,
		Symbols:   true,
	}
}

// GeneratePassword produces a password satisfying policy, following the
// algorithm in spec section 10.3:
//  1. pick one character from each selected class, so every selected class
//     is guaranteed to be represented regardless of what the later random
//     fill produces;
//  2. fill the remaining positions from the combined alphabet;
//  3. Fisher-Yates shuffle so the guaranteed-class characters from step 1
//     aren't predictably in the first N positions;
//  4. verify the result actually satisfies the policy before returning it.
//
// src supplies all randomness; production callers use random.CryptoSource.
func GeneratePassword(src random.Source, policy PasswordPolicy) (domain.Secret, error) {
	classes := classAlphabets(policy)
	if len(classes) == 0 {
		return nil, ErrNoCharacterClassSelected
	}
	if policy.Length < len(classes) {
		return nil, ErrPasswordLengthTooShort
	}

	result := make([]byte, policy.Length)

	for i, alphabet := range classes {
		result[i] = alphabet[src.Intn(len(alphabet))]
	}

	full := strings.Join(classes, "")
	for i := len(classes); i < policy.Length; i++ {
		result[i] = full[src.Intn(len(full))]
	}

	src.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	if !SatisfiesPasswordPolicy(result, policy) {
		// Should be unreachable given the construction above; guards
		// against a future refactor silently breaking the
		// guaranteed-class-representation property.
		return nil, fmt.Errorf("application: generated password failed policy verification")
	}

	return domain.NewSecret(result), nil
}

// SatisfiesPasswordPolicy reports whether password is exactly policy.Length
// bytes and contains at least one character from every class policy
// selects (spec section 10.3, step 4).
func SatisfiesPasswordPolicy(password []byte, policy PasswordPolicy) bool {
	if len(password) != policy.Length {
		return false
	}
	for _, class := range []struct {
		enabled  bool
		alphabet string
	}{
		{policy.Lowercase, filterAmbiguous(lowercaseAlphabet, policy.AvoidAmbiguous)},
		{policy.Uppercase, filterAmbiguous(uppercaseAlphabet, policy.AvoidAmbiguous)},
		{policy.Digits, filterAmbiguous(digitsAlphabet, policy.AvoidAmbiguous)},
		{policy.Symbols, symbolsAlphabet},
	} {
		if class.enabled && !containsAny(password, class.alphabet) {
			return false
		}
	}
	return true
}

// classAlphabets returns the alphabet for each character class policy
// selects, in a fixed order (lowercase, uppercase, digits, symbols).
func classAlphabets(policy PasswordPolicy) []string {
	var classes []string
	if policy.Lowercase {
		classes = append(classes, filterAmbiguous(lowercaseAlphabet, policy.AvoidAmbiguous))
	}
	if policy.Uppercase {
		classes = append(classes, filterAmbiguous(uppercaseAlphabet, policy.AvoidAmbiguous))
	}
	if policy.Digits {
		classes = append(classes, filterAmbiguous(digitsAlphabet, policy.AvoidAmbiguous))
	}
	if policy.Symbols {
		classes = append(classes, symbolsAlphabet)
	}
	return classes
}

func filterAmbiguous(alphabet string, avoid bool) string {
	if !avoid {
		return alphabet
	}
	var b strings.Builder
	for _, r := range alphabet {
		if !strings.ContainsRune(ambiguousChars, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func containsAny(password []byte, alphabet string) bool {
	for _, b := range password {
		if strings.IndexByte(alphabet, b) >= 0 {
			return true
		}
	}
	return false
}
