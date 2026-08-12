package application

import (
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
)

func revealBytes(t *testing.T, s domain.Secret) []byte {
	t.Helper()
	var got []byte
	if err := s.Reveal(func(value []byte) error {
		got = append(got, value...)
		return nil
	}); err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	return got
}

func TestGeneratePasswordDefaultPolicyLength(t *testing.T) {
	src := random.CryptoSource{}
	secret, err := GeneratePassword(src, DefaultPasswordPolicy())
	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}

	got := revealBytes(t, secret)
	if len(got) != 24 {
		t.Errorf("len(password) = %d, want 24", len(got))
	}
}

func TestGeneratePasswordSatisfiesPolicyAcrossManyRuns(t *testing.T) {
	src := random.CryptoSource{}
	policies := []PasswordPolicy{
		DefaultPasswordPolicy(),
		{Length: 8, Digits: true, Symbols: true},
		{Length: 4, Lowercase: true, Uppercase: true, Digits: true, Symbols: true},
		{Length: 40, Lowercase: true, AvoidAmbiguous: true},
	}

	for _, policy := range policies {
		for i := 0; i < 200; i++ {
			secret, err := GeneratePassword(src, policy)
			if err != nil {
				t.Fatalf("GeneratePassword(%+v) error = %v", policy, err)
			}
			got := revealBytes(t, secret)
			if !SatisfiesPasswordPolicy(got, policy) {
				t.Fatalf("password %q does not satisfy policy %+v", got, policy)
			}
		}
	}
}

func TestGeneratePasswordAvoidsAmbiguousCharacters(t *testing.T) {
	src := random.CryptoSource{}
	policy := PasswordPolicy{
		Length:         64,
		Lowercase:      true,
		Uppercase:      true,
		Digits:         true,
		AvoidAmbiguous: true,
	}

	for i := 0; i < 100; i++ {
		secret, err := GeneratePassword(src, policy)
		if err != nil {
			t.Fatalf("GeneratePassword() error = %v", err)
		}
		got := revealBytes(t, secret)
		for _, b := range got {
			if containsAny([]byte{b}, ambiguousChars) {
				t.Fatalf("password %q contains ambiguous character %q", got, b)
			}
		}
	}
}

func TestGeneratePasswordVariesAcrossCalls(t *testing.T) {
	src := random.CryptoSource{}
	policy := DefaultPasswordPolicy()

	first := revealBytes(t, mustGenerate(t, src, policy))
	second := revealBytes(t, mustGenerate(t, src, policy))

	if string(first) == string(second) {
		t.Errorf("two independent GeneratePassword() calls returned the same value: %q", first)
	}
}

func TestGeneratePasswordRejectsNoClassesSelected(t *testing.T) {
	src := random.CryptoSource{}
	_, err := GeneratePassword(src, PasswordPolicy{Length: 10})
	if !errors.Is(err, ErrNoCharacterClassSelected) {
		t.Errorf("GeneratePassword() error = %v, want %v", err, ErrNoCharacterClassSelected)
	}
}

func TestGeneratePasswordRejectsTooShortLength(t *testing.T) {
	src := random.CryptoSource{}
	policy := PasswordPolicy{
		Length:    2,
		Lowercase: true,
		Uppercase: true,
		Digits:    true,
		Symbols:   true,
	}
	_, err := GeneratePassword(src, policy)
	if !errors.Is(err, ErrPasswordLengthTooShort) {
		t.Errorf("GeneratePassword() error = %v, want %v", err, ErrPasswordLengthTooShort)
	}
}

func TestSatisfiesPasswordPolicyRejectsWrongLength(t *testing.T) {
	policy := PasswordPolicy{Length: 5, Lowercase: true}
	if SatisfiesPasswordPolicy([]byte("abcd"), policy) {
		t.Error("SatisfiesPasswordPolicy() = true for a password of the wrong length")
	}
}

func TestSatisfiesPasswordPolicyRejectsMissingClass(t *testing.T) {
	policy := PasswordPolicy{Length: 4, Lowercase: true, Digits: true}
	if SatisfiesPasswordPolicy([]byte("abcd"), policy) {
		t.Error("SatisfiesPasswordPolicy() = true for a password missing a required digit")
	}
}

func TestSatisfiesPasswordPolicyAcceptsValidPassword(t *testing.T) {
	policy := PasswordPolicy{Length: 4, Lowercase: true, Digits: true}
	if !SatisfiesPasswordPolicy([]byte("ab12"), policy) {
		t.Error("SatisfiesPasswordPolicy() = false for a password that satisfies the policy")
	}
}

func mustGenerate(t *testing.T, src random.Source, policy PasswordPolicy) domain.Secret {
	t.Helper()
	secret, err := GeneratePassword(src, policy)
	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}
	return secret
}
