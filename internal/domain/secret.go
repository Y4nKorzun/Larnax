package domain

// Secret provides controlled access to sensitive in-memory data such as
// passwords, so they never leak into fmt/%v output, JSON encoding, or
// accidental struct copies as plain strings.
//
// Per spec section 5.3, this is best-effort memory hygiene: Go's garbage
// collector, immutable strings, and third-party library copies mean no
// implementation can guarantee a secret is unrecoverable from process
// memory. Secret minimizes the lifetime and number of copies; it does not
// promise more than that.
type Secret interface {
	// Reveal invokes fn with the secret's raw bytes. The slice is only
	// valid for the duration of fn and must not be retained by the caller.
	Reveal(fn func(value []byte) error) error

	// Clone returns an independent copy. Clearing one does not affect the
	// other.
	Clone() Secret

	// Clear best-effort zeroes the underlying memory. After Clear, Reveal
	// returns ErrSecretCleared.
	Clear()
}

type byteSecret struct {
	value   []byte
	cleared bool
}

// NewSecret wraps value in a Secret. Ownership of the backing array
// transfers to the returned Secret; the caller must not read or mutate
// value afterwards.
func NewSecret(value []byte) Secret {
	return &byteSecret{value: value}
}

// NewSecretFromString wraps s in a Secret. Because Go strings are immutable
// and may have been copied by the runtime, the original string's backing
// memory cannot be cleared and may persist until garbage collected — prefer
// NewSecret with a []byte you control when that matters.
func NewSecretFromString(s string) Secret {
	return NewSecret([]byte(s))
}

func (s *byteSecret) Reveal(fn func(value []byte) error) error {
	if s.cleared {
		return ErrSecretCleared
	}
	return fn(s.value)
}

func (s *byteSecret) Clone() Secret {
	if s.cleared {
		return &byteSecret{cleared: true}
	}
	clone := make([]byte, len(s.value))
	copy(clone, s.value)
	return &byteSecret{value: clone}
}

func (s *byteSecret) Clear() {
	if s.cleared {
		return
	}
	for i := range s.value {
		s.value[i] = 0
	}
	s.value = nil
	s.cleared = true
}
