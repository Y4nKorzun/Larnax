package googlecsv

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// passkeyHeaderAliases names header columns that mark a file as a
// passkey export rather than a traditional password export — Google
// Password Manager exports passkeys separately from passwords, so in
// practice this matches at the whole-file level, not per row. Like
// headerAliases in parser.go, this list is deliberately permissive
// rather than pinned to one exact export format this code can't
// independently verify against a live export from this environment
// (spec section 13.4's own caveat about renamed/reordered columns
// applies here too) — it matches on general WebAuthn/FIDO2 terminology
// (credential_id, public key) rather than a single assumed column name.
var passkeyHeaderAliases = map[string]bool{
	"credential_id":         true,
	"credentialid":          true,
	"passkey":               true,
	"passkey_id":            true,
	"public_key_credential": true,
	"public_key":            true,
}

// DetectPasskeyHeader reports whether header looks like a passkey export
// rather than a password export. Spec section 13.8: "if the export or
// the selected file contains unsupported credential types," the
// application must say so — this is the detection half of that
// requirement; PasskeyWarning below is the message half.
func DetectPasskeyHeader(header []string) bool {
	for _, col := range header {
		if passkeyHeaderAliases[strings.ToLower(strings.TrimSpace(col))] {
			return true
		}
	}
	return false
}

// countDataRows counts the remaining rows r can read, tolerating (by
// simply not counting) any row that fails to parse — this function only
// answers "how many passkeys are in this file," a best-effort count for
// spec section 13.8's warning message, not an import that needs every
// row to be individually well-formed.
func countDataRows(r *csv.Reader) int {
	count := 0
	for {
		if _, err := r.Read(); err != nil {
			return count
		}
		count++
	}
}

// PasskeyWarning formats spec section 13.8's required message for count
// detected-but-unimportable passkeys. Callers should only show this when
// count > 0.
func PasskeyWarning(count int) string {
	plural := "s"
	if count == 1 {
		plural = ""
	}
	return fmt.Sprintf(
		"Unsupported credentials detected: %d passkey%s.\nThey were not imported and remain outside this vault.",
		count, plural,
	)
}
