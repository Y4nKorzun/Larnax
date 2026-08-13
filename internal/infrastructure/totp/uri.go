package totp

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidScheme   = errors.New("totp: otpauth URI must use the otpauth scheme")
	ErrUnsupportedType = errors.New("totp: only otpauth://totp/... URIs are supported")
	ErrMissingSecret   = errors.New("totp: otpauth URI is missing the secret parameter")
	ErrInvalidSecret   = errors.New("totp: secret is not valid Base32")
)

// URIParams is the result of parsing an otpauth:// URI (spec section
// 14.1). Only the totp type is supported; hotp (counter-based) is out of
// scope for this application.
type URIParams struct {
	Label  string // the path component, conventionally "Issuer:AccountName"
	Issuer string
	Params Params // this package's Generate() input
}

// ParseURI parses an otpauth://totp/... URI — the de facto standard
// Google Authenticator key URI format — into URIParams. It performs only
// syntax-level parsing (Base32 decoding, integer parsing of digits/
// period); semantic validation of the resulting Params (is Digits 6 or 8,
// is Algorithm supported) is Generate's job, not duplicated here.
//
// ParseURI never panics on malformed input, returning an error instead —
// spec section 24.2 names this exact parser as a fuzz target with that
// invariant (see FuzzParseURI).
func ParseURI(raw string) (URIParams, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return URIParams{}, fmt.Errorf("totp: %w", err)
	}
	if u.Scheme != "otpauth" {
		return URIParams{}, ErrInvalidScheme
	}
	if u.Host != "totp" {
		return URIParams{}, ErrUnsupportedType
	}

	label, err := url.PathUnescape(strings.TrimPrefix(u.Path, "/"))
	if err != nil {
		return URIParams{}, fmt.Errorf("totp: invalid label encoding: %w", err)
	}

	query := u.Query()

	secretRaw := query.Get("secret")
	if secretRaw == "" {
		return URIParams{}, ErrMissingSecret
	}
	secret, err := decodeBase32Secret(secretRaw)
	if err != nil {
		return URIParams{}, fmt.Errorf("%w: %v", ErrInvalidSecret, err)
	}

	algorithm := AlgorithmSHA1
	if a := query.Get("algorithm"); a != "" {
		algorithm = Algorithm(strings.ToUpper(a))
	}

	digits := 6
	if d := query.Get("digits"); d != "" {
		n, err := strconv.Atoi(d)
		if err != nil {
			return URIParams{}, fmt.Errorf("totp: invalid digits parameter: %w", err)
		}
		digits = n
	}

	period := 30 * time.Second
	if p := query.Get("period"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return URIParams{}, fmt.Errorf("totp: invalid period parameter: %w", err)
		}
		period = time.Duration(n) * time.Second
	}

	issuer := query.Get("issuer")
	if issuer == "" {
		// Fall back to the "Issuer:Account" label convention only when
		// the query parameter is absent; when both are present, the
		// explicit query parameter wins.
		if idx := strings.Index(label, ":"); idx >= 0 {
			issuer = strings.TrimSpace(label[:idx])
		}
	}

	return URIParams{
		Label:  label,
		Issuer: issuer,
		Params: Params{
			Secret:    secret,
			Digits:    digits,
			Period:    period,
			Algorithm: algorithm,
		},
	}, nil
}

// decodeBase32Secret decodes s as Base32, tolerating the unpadded form
// most otpauth URIs actually use in practice — Go's base32.StdEncoding
// requires correctly padded input, so padding is restored first if
// missing.
func decodeBase32Secret(s string) ([]byte, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if m := len(s) % 8; m != 0 {
		s += strings.Repeat("=", 8-m)
	}
	return base32.StdEncoding.DecodeString(s)
}
