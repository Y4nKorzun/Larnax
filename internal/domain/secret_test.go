package domain

import (
	"bytes"
	"errors"
	"testing"
)

func TestSecretRevealExposesValue(t *testing.T) {
	s := NewSecret([]byte("hunter2"))

	var got []byte
	err := s.Reveal(func(value []byte) error {
		got = append(got, value...)
		return nil
	})
	if err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	if !bytes.Equal(got, []byte("hunter2")) {
		t.Errorf("Reveal() value = %q, want %q", got, "hunter2")
	}
}

func TestSecretRevealPropagatesCallbackError(t *testing.T) {
	s := NewSecret([]byte("hunter2"))
	sentinel := errors.New("boom")

	err := s.Reveal(func(value []byte) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Reveal() error = %v, want %v", err, sentinel)
	}
}

func TestSecretClearZeroesAndBlocksReveal(t *testing.T) {
	// raw is inspected directly after Clear() to verify the backing array
	// itself was zeroed, not just that a "cleared" flag was set.
	raw := []byte("hunter2")
	s := NewSecret(raw)

	s.Clear()

	for i, b := range raw {
		if b != 0 {
			t.Fatalf("byte %d not zeroed after Clear(): %v", i, raw)
		}
	}

	err := s.Reveal(func(value []byte) error { return nil })
	if !errors.Is(err, ErrSecretCleared) {
		t.Errorf("Reveal() after Clear() error = %v, want %v", err, ErrSecretCleared)
	}
}

func TestSecretClearIsIdempotent(t *testing.T) {
	s := NewSecret([]byte("hunter2"))
	s.Clear()
	s.Clear() // must not panic
}

func TestSecretCloneIsIndependent(t *testing.T) {
	s := NewSecret([]byte("hunter2"))
	clone := s.Clone()

	clone.Clear()

	err := s.Reveal(func(value []byte) error {
		if !bytes.Equal(value, []byte("hunter2")) {
			t.Errorf("original value = %q, want %q", value, "hunter2")
		}
		return nil
	})
	if err != nil {
		t.Errorf("original Reveal() after clone cleared: error = %v, want nil", err)
	}
}

func TestSecretCloneOfClearedSecretIsCleared(t *testing.T) {
	s := NewSecret([]byte("hunter2"))
	s.Clear()

	clone := s.Clone()
	err := clone.Reveal(func(value []byte) error { return nil })
	if !errors.Is(err, ErrSecretCleared) {
		t.Errorf("Reveal() on clone of cleared secret error = %v, want %v", err, ErrSecretCleared)
	}
}

func TestNewSecretFromString(t *testing.T) {
	s := NewSecretFromString("hunter2")
	err := s.Reveal(func(value []byte) error {
		if string(value) != "hunter2" {
			t.Errorf("value = %q, want %q", value, "hunter2")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
}
