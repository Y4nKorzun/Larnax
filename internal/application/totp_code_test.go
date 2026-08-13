package application

import (
	"errors"
	"testing"
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/totp"
)

func entryWithTOTP(t *testing.T) domain.Entry {
	t.Helper()
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	params := totp.Params{
		Secret:    []byte("test secret for totp code!"),
		Digits:    6,
		Period:    30 * time.Second,
		Algorithm: totp.AlgorithmSHA1,
	}
	return totp.SetEntryURI(e, totp.BuildURI("GitHub:octocat", "GitHub", params))
}

func TestGenerateTOTPCodeReturnsErrNoTOTPField(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	if _, err := GenerateTOTPCode(e, time.Now()); !errors.Is(err, ErrNoTOTPField) {
		t.Errorf("GenerateTOTPCode() error = %v, want %v", err, ErrNoTOTPField)
	}
}

func TestGenerateTOTPCodeMatchesDirectGenerate(t *testing.T) {
	e := entryWithTOTP(t)
	now := time.Unix(1_700_000_000, 0)

	got, err := GenerateTOTPCode(e, now)
	if err != nil {
		t.Fatalf("GenerateTOTPCode() error = %v", err)
	}

	uri, _, err := totp.EntryURI(e)
	if err != nil {
		t.Fatalf("totp.EntryURI() error = %v", err)
	}
	parsed, err := totp.ParseURI(uri)
	if err != nil {
		t.Fatalf("totp.ParseURI() error = %v", err)
	}
	want, err := totp.Generate(parsed.Params, now)
	if err != nil {
		t.Fatalf("totp.Generate() error = %v", err)
	}

	if got != want {
		t.Errorf("GenerateTOTPCode() = %q, want %q", got, want)
	}
}

func TestTOTPTimeRemainingErrNoTOTPField(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	if _, err := TOTPTimeRemaining(e, time.Now()); !errors.Is(err, ErrNoTOTPField) {
		t.Errorf("TOTPTimeRemaining() error = %v, want %v", err, ErrNoTOTPField)
	}
}

func TestTOTPTimeRemainingAtPeriodBoundaryIsFullPeriod(t *testing.T) {
	e := entryWithTOTP(t)
	// 1_700_000_000 is not a multiple of 30; pick the next one.
	var boundary int64 = (1_700_000_000/30 + 1) * 30
	now := time.Unix(boundary, 0)

	remaining, err := TOTPTimeRemaining(e, now)
	if err != nil {
		t.Fatalf("TOTPTimeRemaining() error = %v", err)
	}
	if remaining != 30*time.Second {
		t.Errorf("remaining = %v, want 30s at a period boundary", remaining)
	}
}

func TestTOTPTimeRemainingDecreasesOverASecond(t *testing.T) {
	e := entryWithTOTP(t)
	var boundary int64 = (1_700_000_000/30 + 1) * 30
	first, err := TOTPTimeRemaining(e, time.Unix(boundary+5, 0))
	if err != nil {
		t.Fatalf("TOTPTimeRemaining() error = %v", err)
	}
	second, err := TOTPTimeRemaining(e, time.Unix(boundary+6, 0))
	if err != nil {
		t.Fatalf("TOTPTimeRemaining() error = %v", err)
	}
	if first-second != time.Second {
		t.Errorf("remaining dropped by %v over one second, want exactly 1s (first=%v second=%v)", first-second, first, second)
	}
}

func TestTOTPTimeRemainingWithinPeriodBounds(t *testing.T) {
	e := entryWithTOTP(t)
	remaining, err := TOTPTimeRemaining(e, time.Unix(1_700_000_017, 0))
	if err != nil {
		t.Fatalf("TOTPTimeRemaining() error = %v", err)
	}
	if remaining <= 0 || remaining > 30*time.Second {
		t.Errorf("remaining = %v, want in (0, 30s]", remaining)
	}
}
