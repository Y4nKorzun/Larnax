package kdbx

import (
	"errors"
	"testing"
)

func TestPortableProfileMatchesSpecValues(t *testing.T) {
	p := PortableProfile()
	if p.MemoryKiB != 64*1024 {
		t.Errorf("MemoryKiB = %d, want %d (64 MiB)", p.MemoryKiB, 64*1024)
	}
	if p.Iterations != 3 {
		t.Errorf("Iterations = %d, want 3", p.Iterations)
	}
	if p.Parallelism != 2 {
		t.Errorf("Parallelism = %d, want 2", p.Parallelism)
	}
}

func TestPortableProfileIsValid(t *testing.T) {
	if err := PortableProfile().Validate(); err != nil {
		t.Errorf("PortableProfile().Validate() = %v, want nil", err)
	}
}

func TestValidateRejectsZeroParallelism(t *testing.T) {
	p := Argon2Profile{MemoryKiB: 1024, Iterations: 1, Parallelism: 0}
	if err := p.Validate(); !errors.Is(err, ErrParallelismTooLow) {
		t.Errorf("Validate() = %v, want %v", err, ErrParallelismTooLow)
	}
}

func TestValidateRejectsZeroIterations(t *testing.T) {
	p := Argon2Profile{MemoryKiB: 1024, Iterations: 0, Parallelism: 1}
	if err := p.Validate(); !errors.Is(err, ErrIterationsTooLow) {
		t.Errorf("Validate() = %v, want %v", err, ErrIterationsTooLow)
	}
}

func TestValidateRejectsMemoryBelowRFCMinimum(t *testing.T) {
	// RFC 9106: memory must be at least 8*parallelism KiB. 4 lanes needs
	// at least 32 KiB; 16 is below that.
	p := Argon2Profile{MemoryKiB: 16, Iterations: 1, Parallelism: 4}
	if err := p.Validate(); !errors.Is(err, ErrMemoryTooLow) {
		t.Errorf("Validate() = %v, want %v", err, ErrMemoryTooLow)
	}
}

func TestValidateAcceptsExactRFCMemoryBoundary(t *testing.T) {
	// RFC 9106 states the range as inclusive ("from 8*p"), so exactly
	// 8*parallelism must be accepted, not rejected as one-too-low.
	p := Argon2Profile{MemoryKiB: 32, Iterations: 1, Parallelism: 4}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil (memory exactly equals 8*parallelism)", err)
	}
}

func TestValidateAcceptsAbsoluteRFCMinimums(t *testing.T) {
	p := Argon2Profile{MemoryKiB: 8, Iterations: 1, Parallelism: 1}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil (RFC 9106's absolute minimum parameters)", err)
	}
}
