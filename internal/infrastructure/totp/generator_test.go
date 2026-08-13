package totp

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// rfc6238Secret reproduces the exact test secrets from RFC 6238 Appendix
// B's reference Java implementation, which builds them from repetitions of
// the ASCII digits "1234567890" ("Seed for HMAC-SHA1 - 20 bytes",
// "Seed for HMAC-SHA256 - 32 bytes", "Seed for HMAC-SHA512 - 64 bytes").
func rfc6238Secret(length int) []byte {
	return []byte(strings.Repeat("1234567890", (length+9)/10)[:length])
}

// TestGenerateMatchesRFC6238TestVectors checks every value in RFC 6238
// Appendix B's published Table 1 (8-digit codes, 30-second period,
// T0 = 0), across all three algorithms and every listed timestamp.
func TestGenerateMatchesRFC6238TestVectors(t *testing.T) {
	sha1Secret := rfc6238Secret(20)
	sha256Secret := rfc6238Secret(32)
	sha512Secret := rfc6238Secret(64)

	cases := []struct {
		unixSeconds int64
		algorithm   Algorithm
		secret      []byte
		want        string
	}{
		{59, AlgorithmSHA1, sha1Secret, "94287082"},
		{59, AlgorithmSHA256, sha256Secret, "46119246"},
		{59, AlgorithmSHA512, sha512Secret, "90693936"},

		{1111111109, AlgorithmSHA1, sha1Secret, "07081804"},
		{1111111109, AlgorithmSHA256, sha256Secret, "68084774"},
		{1111111109, AlgorithmSHA512, sha512Secret, "25091201"},

		{1111111111, AlgorithmSHA1, sha1Secret, "14050471"},
		{1111111111, AlgorithmSHA256, sha256Secret, "67062674"},
		{1111111111, AlgorithmSHA512, sha512Secret, "99943326"},

		{1234567890, AlgorithmSHA1, sha1Secret, "89005924"},
		{1234567890, AlgorithmSHA256, sha256Secret, "91819424"},
		{1234567890, AlgorithmSHA512, sha512Secret, "93441116"},

		{2000000000, AlgorithmSHA1, sha1Secret, "69279037"},
		{2000000000, AlgorithmSHA256, sha256Secret, "90698825"},
		{2000000000, AlgorithmSHA512, sha512Secret, "38618901"},

		{20000000000, AlgorithmSHA1, sha1Secret, "65353130"},
		{20000000000, AlgorithmSHA256, sha256Secret, "77737706"},
		{20000000000, AlgorithmSHA512, sha512Secret, "47863826"},
	}

	for _, c := range cases {
		params := Params{
			Secret:    c.secret,
			Digits:    8,
			Period:    30 * time.Second,
			Algorithm: c.algorithm,
		}
		got, err := Generate(params, time.Unix(c.unixSeconds, 0).UTC())
		if err != nil {
			t.Fatalf("Generate(%d, %s) error = %v", c.unixSeconds, c.algorithm, err)
		}
		if got != c.want {
			t.Errorf("Generate(%d, %s) = %q, want %q", c.unixSeconds, c.algorithm, got, c.want)
		}
	}
}

func TestGenerateDefaultsPeriodTo30Seconds(t *testing.T) {
	secret := rfc6238Secret(20)
	withDefault, err := Generate(Params{Secret: secret, Digits: 8, Algorithm: AlgorithmSHA1}, time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if withDefault != "94287082" {
		t.Errorf("Generate() with zero Period = %q, want %q (should default to 30s)", withDefault, "94287082")
	}
}

func TestGenerateProducesSameCodeThroughoutOnePeriod(t *testing.T) {
	secret := rfc6238Secret(20)
	params := Params{Secret: secret, Digits: 6, Period: 30 * time.Second, Algorithm: AlgorithmSHA1}

	start, err := Generate(params, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	// 1700000000 is not a multiple of 30; find the current window bounds.
	var windowStart int64 = (1700000000 / 30) * 30
	end, err := Generate(params, time.Unix(windowStart+29, 0).UTC())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if start != end {
		t.Errorf("code changed within the same 30s window: %q vs %q", start, end)
	}

	next, err := Generate(params, time.Unix(windowStart+30, 0).UTC())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if next == start {
		t.Errorf("code did not change across a window boundary: %q", next)
	}
}

func TestGenerateSixDigitsIsZeroPadded(t *testing.T) {
	// A 6-digit truncation can legitimately start with one or more
	// zeros; the result must still be exactly 6 characters, not a
	// shorter unpadded number.
	secret := rfc6238Secret(20)
	got, err := Generate(Params{Secret: secret, Digits: 6, Algorithm: AlgorithmSHA1}, time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(got) != 6 {
		t.Errorf("len(code) = %d, want 6 (code = %q)", len(got), got)
	}
	// The 8-digit RFC vector at the same instant is 94287082, so the
	// 6-digit truncation must be its low 6 digits.
	if got != "287082" {
		t.Errorf("Generate(digits=6) = %q, want %q (low 6 digits of the 8-digit RFC vector)", got, "287082")
	}
}

func TestGenerateRejectsEmptySecret(t *testing.T) {
	_, err := Generate(Params{Secret: nil, Digits: 6, Algorithm: AlgorithmSHA1}, time.Now())
	if !errors.Is(err, ErrEmptySecret) {
		t.Errorf("Generate() error = %v, want %v", err, ErrEmptySecret)
	}
}

func TestGenerateRejectsInvalidDigits(t *testing.T) {
	secret := rfc6238Secret(20)
	for _, digits := range []int{0, 5, 7, 9} {
		_, err := Generate(Params{Secret: secret, Digits: digits, Algorithm: AlgorithmSHA1}, time.Now())
		if err == nil {
			t.Errorf("Generate(digits=%d) error = nil, want an error", digits)
		}
	}
}

func TestGenerateRejectsUnknownAlgorithm(t *testing.T) {
	secret := rfc6238Secret(20)
	_, err := Generate(Params{Secret: secret, Digits: 6, Algorithm: "MD5"}, time.Now())
	if err == nil {
		t.Error("Generate() error = nil, want an error for an unsupported algorithm")
	}
}
