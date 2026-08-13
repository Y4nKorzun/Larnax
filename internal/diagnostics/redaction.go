// Package diagnostics implements spec section 22's logging discipline.
package diagnostics

import (
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

const redactionPlaceholder = "[REDACTED]"

// Redact returns message with every occurrence of each non-empty value in
// values replaced by a fixed placeholder.
//
// This is defense-in-depth, not the primary defense. The primary defense
// against spec section 22.1's forbidden-to-log list (master passphrase,
// titles, usernames, URLs, passwords, notes, TOTP, raw KDBX XML, CSV rows,
// clipboard content, full filesystem paths) is simply never constructing a
// log line from that data in the first place — logger.go's fixed allowlist
// of what may be logged. Redact exists for the case where a value from
// that list ends up interpolated into a string that reaches a logging call
// anyway, e.g. a bug that formats an error containing a path or field
// content.
//
// An empty string in values is skipped rather than passed to
// strings.ReplaceAll, which would otherwise insert the placeholder between
// every character of message.
func Redact(message string, values ...string) string {
	for _, v := range values {
		if v == "" {
			continue
		}
		message = strings.ReplaceAll(message, v, redactionPlaceholder)
	}
	return message
}

// RedactSecret is a convenience for the common case of a domain.Secret: it
// reveals secret only long enough to compute the redacted result, so
// callers never need to extract and separately manage the raw bytes
// themselves just to call Redact. A nil secret, or one that has already
// been cleared, leaves message unchanged.
func RedactSecret(message string, secret domain.Secret) string {
	if secret == nil {
		return message
	}
	_ = secret.Reveal(func(value []byte) error {
		if len(value) > 0 {
			message = strings.ReplaceAll(message, string(value), redactionPlaceholder)
		}
		return nil
	})
	return message
}
