package totp

import (
	"bytes"
	"testing"
	"time"
)

func TestFieldNameIsOtp(t *testing.T) {
	if FieldName != "otp" {
		t.Errorf("FieldName = %q, want %q", FieldName, "otp")
	}
}

func TestBuildURIThenParseURIRoundTrips(t *testing.T) {
	params := Params{
		Secret:    []byte("this is a test secret!!"),
		Digits:    6,
		Period:    30 * time.Second,
		Algorithm: AlgorithmSHA1,
	}

	raw := BuildURI("GitHub:octocat", "GitHub", params)
	got, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI(%q) error = %v", raw, err)
	}

	if got.Label != "GitHub:octocat" {
		t.Errorf("Label = %q, want %q", got.Label, "GitHub:octocat")
	}
	if got.Issuer != "GitHub" {
		t.Errorf("Issuer = %q, want %q", got.Issuer, "GitHub")
	}
	if !bytes.Equal(got.Params.Secret, params.Secret) {
		t.Errorf("Secret = %x, want %x", got.Params.Secret, params.Secret)
	}
	if got.Params.Digits != params.Digits {
		t.Errorf("Digits = %d, want %d", got.Params.Digits, params.Digits)
	}
	if got.Params.Period != params.Period {
		t.Errorf("Period = %v, want %v", got.Params.Period, params.Period)
	}
	if got.Params.Algorithm != params.Algorithm {
		t.Errorf("Algorithm = %q, want %q", got.Params.Algorithm, params.Algorithm)
	}
}

func TestBuildURIThenParseURIRoundTripsNonDefaultParams(t *testing.T) {
	params := Params{
		Secret:    []byte("another test secret!!!!"),
		Digits:    8,
		Period:    60 * time.Second,
		Algorithm: AlgorithmSHA512,
	}

	raw := BuildURI("account-only-label", "", params)
	got, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI(%q) error = %v", raw, err)
	}

	if !bytes.Equal(got.Params.Secret, params.Secret) {
		t.Errorf("Secret = %x, want %x", got.Params.Secret, params.Secret)
	}
	if got.Params.Digits != 8 {
		t.Errorf("Digits = %d, want 8", got.Params.Digits)
	}
	if got.Params.Period != 60*time.Second {
		t.Errorf("Period = %v, want 60s", got.Params.Period)
	}
	if got.Params.Algorithm != AlgorithmSHA512 {
		t.Errorf("Algorithm = %q, want %q", got.Params.Algorithm, AlgorithmSHA512)
	}
}

func TestBuildURIThenGenerateProducesAValidCode(t *testing.T) {
	params := Params{
		Secret:    []byte("yet another test secret!"),
		Digits:    6,
		Period:    30 * time.Second,
		Algorithm: AlgorithmSHA1,
	}
	raw := BuildURI("svc:user", "svc", params)

	parsed, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	code, err := Generate(parsed.Params, time.Now())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(code) != 6 {
		t.Errorf("len(code) = %d, want 6", len(code))
	}
}
