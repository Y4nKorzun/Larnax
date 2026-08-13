package diagnostics

import (
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestRedactReplacesOccurrence(t *testing.T) {
	got := Redact("user logged in as alice@example.com", "alice@example.com")
	want := "user logged in as [REDACTED]"
	if got != want {
		t.Errorf("Redact() = %q, want %q", got, want)
	}
}

func TestRedactReplacesMultipleValues(t *testing.T) {
	got := Redact("alice@example.com visited https://github.com", "alice@example.com", "https://github.com")
	want := "[REDACTED] visited [REDACTED]"
	if got != want {
		t.Errorf("Redact() = %q, want %q", got, want)
	}
}

func TestRedactReplacesAllOccurrences(t *testing.T) {
	got := Redact("retry for alice@example.com failed, retry for alice@example.com again", "alice@example.com")
	want := "retry for [REDACTED] failed, retry for [REDACTED] again"
	if got != want {
		t.Errorf("Redact() = %q, want %q", got, want)
	}
}

func TestRedactSkipsEmptyValues(t *testing.T) {
	in := "nothing sensitive here"
	got := Redact(in, "", "also not present")
	if got != in {
		t.Errorf("Redact() = %q, want unchanged %q", got, in)
	}
}

func TestRedactLeavesMessageUnchangedWhenNoMatch(t *testing.T) {
	in := "operation=open result=ok"
	got := Redact(in, "bob@example.com")
	if got != in {
		t.Errorf("Redact() = %q, want unchanged %q", got, in)
	}
}

func TestRedactSecretReplacesRevealedValue(t *testing.T) {
	secret := domain.NewSecret([]byte("hunter2"))
	got := RedactSecret("attempted password was hunter2", secret)
	want := "attempted password was [REDACTED]"
	if got != want {
		t.Errorf("RedactSecret() = %q, want %q", got, want)
	}
}

func TestRedactSecretHandlesNilSecret(t *testing.T) {
	in := "no secret here"
	got := RedactSecret(in, nil)
	if got != in {
		t.Errorf("RedactSecret(nil) = %q, want unchanged %q", got, in)
	}
}

func TestRedactSecretHandlesClearedSecret(t *testing.T) {
	secret := domain.NewSecret([]byte("hunter2"))
	secret.Clear()

	in := "attempted password was hunter2"
	got := RedactSecret(in, secret)
	if got != in {
		t.Errorf("RedactSecret(cleared) = %q, want unchanged %q (nothing left to reveal)", got, in)
	}
}

func TestRedactSecretDoesNotLeakRevealedValueBeyondRedaction(t *testing.T) {
	secret := domain.NewSecret([]byte("hunter2"))
	got := RedactSecret("password: hunter2, note: unrelated", secret)
	if got != "password: [REDACTED], note: unrelated" {
		t.Errorf("RedactSecret() = %q, want the secret gone and the rest of the message intact", got)
	}
}
