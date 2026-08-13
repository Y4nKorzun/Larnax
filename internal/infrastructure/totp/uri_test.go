package totp

import (
	"bytes"
	"encoding/base32"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestParseURIValidTOTP(t *testing.T) {
	secretBytes := []byte("test-secret-bytes")
	secret := base32.StdEncoding.EncodeToString(secretBytes)

	raw := "otpauth://totp/GitHub:alice@example.com?secret=" + secret +
		"&issuer=GitHub&algorithm=SHA256&digits=8&period=60"

	got, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}

	if got.Label != "GitHub:alice@example.com" {
		t.Errorf("Label = %q, want %q", got.Label, "GitHub:alice@example.com")
	}
	if got.Issuer != "GitHub" {
		t.Errorf("Issuer = %q, want %q", got.Issuer, "GitHub")
	}
	if !bytes.Equal(got.Params.Secret, secretBytes) {
		t.Errorf("Secret = %q, want %q", got.Params.Secret, secretBytes)
	}
	if got.Params.Algorithm != AlgorithmSHA256 {
		t.Errorf("Algorithm = %q, want %q", got.Params.Algorithm, AlgorithmSHA256)
	}
	if got.Params.Digits != 8 {
		t.Errorf("Digits = %d, want 8", got.Params.Digits)
	}
	if got.Params.Period != 60*time.Second {
		t.Errorf("Period = %v, want 60s", got.Params.Period)
	}
}

func TestParseURIDefaultsWhenParamsOmitted(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("abc"))
	raw := "otpauth://totp/Example:bob?secret=" + secret

	got, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if got.Params.Algorithm != AlgorithmSHA1 {
		t.Errorf("Algorithm = %q, want default %q", got.Params.Algorithm, AlgorithmSHA1)
	}
	if got.Params.Digits != 6 {
		t.Errorf("Digits = %d, want default 6", got.Params.Digits)
	}
	if got.Params.Period != 30*time.Second {
		t.Errorf("Period = %v, want default 30s", got.Params.Period)
	}
}

func TestParseURIRejectsNonOtpauthScheme(t *testing.T) {
	_, err := ParseURI("https://example.com/totp?secret=ABC")
	if !errors.Is(err, ErrInvalidScheme) {
		t.Errorf("ParseURI() error = %v, want %v", err, ErrInvalidScheme)
	}
}

func TestParseURIRejectsNonTotpType(t *testing.T) {
	_, err := ParseURI("otpauth://hotp/Example:bob?secret=ABC&counter=0")
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("ParseURI() error = %v, want %v", err, ErrUnsupportedType)
	}
}

func TestParseURIRejectsMissingSecret(t *testing.T) {
	_, err := ParseURI("otpauth://totp/Example:bob?issuer=Example")
	if !errors.Is(err, ErrMissingSecret) {
		t.Errorf("ParseURI() error = %v, want %v", err, ErrMissingSecret)
	}
}

func TestParseURIRejectsInvalidBase32Secret(t *testing.T) {
	_, err := ParseURI("otpauth://totp/Example:bob?secret=not-valid-base32!!!")
	if !errors.Is(err, ErrInvalidSecret) {
		t.Errorf("ParseURI() error = %v, want %v", err, ErrInvalidSecret)
	}
}

func TestParseURIHandlesUnpaddedBase32Secret(t *testing.T) {
	original := []byte("hello world!") // 12 bytes: std Base32 needs padding
	padded := base32.StdEncoding.EncodeToString(original)
	unpadded := strings.TrimRight(padded, "=")
	if unpadded == padded {
		t.Fatal("test setup invariant violated: expected padding to actually be present")
	}

	got, err := ParseURI("otpauth://totp/Example:bob?secret=" + unpadded)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if !bytes.Equal(got.Params.Secret, original) {
		t.Errorf("Secret = %q, want %q", got.Params.Secret, original)
	}
}

func TestParseURIIssuerFallsBackToLabelPrefix(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("abc"))
	got, err := ParseURI("otpauth://totp/GitHub:alice@example.com?secret=" + secret)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if got.Issuer != "GitHub" {
		t.Errorf("Issuer = %q, want %q (from label prefix)", got.Issuer, "GitHub")
	}
}

func TestParseURIQueryIssuerTakesPrecedenceOverLabel(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("abc"))
	got, err := ParseURI("otpauth://totp/OldName:alice@example.com?secret=" + secret + "&issuer=NewName")
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if got.Issuer != "NewName" {
		t.Errorf("Issuer = %q, want %q (query parameter takes precedence)", got.Issuer, "NewName")
	}
}

func TestParseURILabelIsURLDecoded(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("abc"))
	raw := "otpauth://totp/" + url.PathEscape("My Bank:alice@example.com") + "?secret=" + secret

	got, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if got.Label != "My Bank:alice@example.com" {
		t.Errorf("Label = %q, want %q", got.Label, "My Bank:alice@example.com")
	}
}

// FuzzParseURI is spec section 24.2's named fuzz target: the parser must
// never panic on a malformed otpauth:// URI, regardless of what error (if
// any) it returns.
func FuzzParseURI(f *testing.F) {
	seeds := []string{
		"",
		"otpauth://totp/Example:bob?secret=JBSWY3DPEHPK3PXP",
		"otpauth://hotp/Example:bob?secret=ABC&counter=1",
		"https://example.com",
		"otpauth://totp/",
		"otpauth://totp/Example?secret=",
		"otpauth://totp/Example?secret=%zz",
		"otpauth://totp/Example?secret=ABC&digits=abc",
		"otpauth://totp/Example?secret=ABC&period=-30",
		"otpauth://totp/Example?secret=ABC&algorithm=" + strings.Repeat("A", 10000),
		"not a uri at all\x00\x01\x02",
		"otpauth://totp/日本語?secret=ABC",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ParseURI(%q) panicked: %v", raw, r)
			}
		}()
		_, _ = ParseURI(raw)
	})
}
