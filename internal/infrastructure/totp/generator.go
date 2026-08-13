// Package totp implements RFC 6238 time-based one-time passwords (spec
// section 14). otpauth:// URI parsing and KDBX storage compatibility
// (spec section 14.3, an open question per spec section 30 #8) are
// separate, later work — this package is only the code-generation
// algorithm itself.
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"time"
)

// Algorithm selects the HMAC hash function TOTP uses (spec section 14.2).
type Algorithm string

const (
	AlgorithmSHA1   Algorithm = "SHA1"
	AlgorithmSHA256 Algorithm = "SHA256"
	AlgorithmSHA512 Algorithm = "SHA512"
)

func (a Algorithm) newHash() (func() hash.Hash, error) {
	switch a {
	case AlgorithmSHA1:
		return sha1.New, nil
	case AlgorithmSHA256:
		return sha256.New, nil
	case AlgorithmSHA512:
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("totp: unsupported algorithm %q", a)
	}
}

var ErrEmptySecret = errors.New("totp: secret must not be empty")

// Params configures TOTP generation (spec section 14.2). Period defaults
// to 30 seconds (spec's stated default) when zero.
type Params struct {
	Secret    []byte
	Digits    int // 6 or 8
	Period    time.Duration
	Algorithm Algorithm
}

// Generate computes the TOTP code for at, per RFC 6238 (built on HOTP,
// RFC 4226): derive a time counter from at and Period, HMAC it with
// Secret under Algorithm, and apply RFC 4226 section 5.3's dynamic
// truncation to produce a Digits-length numeric code. T0 (the epoch TOTP
// counts from) is fixed at the Unix epoch, matching every real-world TOTP
// deployment and RFC 6238's own test vectors — spec section 14 gives no
// reason to make it configurable.
func Generate(params Params, at time.Time) (string, error) {
	if len(params.Secret) == 0 {
		return "", ErrEmptySecret
	}
	if params.Digits != 6 && params.Digits != 8 {
		return "", fmt.Errorf("totp: digits must be 6 or 8, got %d", params.Digits)
	}
	newHash, err := params.Algorithm.newHash()
	if err != nil {
		return "", err
	}

	period := params.Period
	if period <= 0 {
		period = 30 * time.Second
	}
	counter := uint64(at.Unix() / int64(period.Seconds()))

	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)

	mac := hmac.New(newHash, params.Secret)
	mac.Write(counterBytes[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	mod := uint32(1)
	for i := 0; i < params.Digits; i++ {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", params.Digits, truncated%mod), nil
}
