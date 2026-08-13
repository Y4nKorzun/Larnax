package kdbx

import "errors"

// Argon2Profile configures the Argon2id key derivation used when creating
// a new vault (spec section 7.2). Parameters are shown to the user in
// advanced settings and must never change silently on save — spec is
// explicit that changing them requires a backup first.
type Argon2Profile struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

// PortableProfile is spec section 7.2's starting profile: 64 MiB memory,
// 3 iterations, parallelism 2. Spec calls this a starting point, not a
// permanent constant — it must be validated against the weakest supported
// desktop and mobile client before a stable release, and any future
// auto-calibration mode (targeting ~500-1000ms, per spec) must never
// produce a file the user's weakest target device can't open.
func PortableProfile() Argon2Profile {
	return Argon2Profile{
		MemoryKiB:   64 * 1024,
		Iterations:  3,
		Parallelism: 2,
	}
}

// These lower bounds come directly from RFC 9106 section 3.1's parameter
// constraints: parallelism p from 1 to 2^24-1, memory m from 8*p to
// 2^32-1 KiB, iterations t from 1 to 2^32-1. The upper bounds need no
// explicit check here: MemoryKiB and Iterations are uint32, so they
// cannot exceed 2^32-1, and Parallelism is uint8, so it cannot exceed
// 255 — comfortably under 2^24-1 and a sane ceiling for a password KDF
// regardless.
var (
	ErrParallelismTooLow = errors.New("kdbx: argon2id parallelism must be at least 1")
	ErrMemoryTooLow      = errors.New("kdbx: argon2id memory must be at least 8 KiB per degree of parallelism (RFC 9106)")
	ErrIterationsTooLow  = errors.New("kdbx: argon2id iterations must be at least 1")
)

func (p Argon2Profile) Validate() error {
	if p.Parallelism < 1 {
		return ErrParallelismTooLow
	}
	if p.MemoryKiB < 8*uint32(p.Parallelism) {
		return ErrMemoryTooLow
	}
	if p.Iterations < 1 {
		return ErrIterationsTooLow
	}
	return nil
}
