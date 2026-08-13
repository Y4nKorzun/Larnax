package diagnostics

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"
)

// Logger is spec section 22.1's fixed allowlist of what this application
// may log, expressed as a typed API rather than a free-form
// message-plus-args logger: there is no method here that accepts a
// title, username, URL, password, note, TOTP secret, raw KDBX XML, CSV
// row, clipboard content, or full filesystem path — every one of those
// is on spec's forbidden-to-log list, and a typed method simply has
// nowhere to put it. Redact/RedactSecret (redaction.go) are
// defense-in-depth for a value that reaches a log line anyway despite
// this; they are not a substitute for the allowlist.
type Logger struct {
	slog *slog.Logger
}

// New wraps base behind spec section 22.1's allowlist. Passing nil uses
// slog.Default().
func New(base *slog.Logger) *Logger {
	if base == nil {
		base = slog.Default()
	}
	return &Logger{slog: base}
}

// HashPath returns a short, non-reversible hex digest of path, for
// operations that must identify *which* vault they touched without
// logging its real filesystem path (spec section 22.1). 16 hex characters
// is plenty to tell a handful of vaults apart in one log stream; it is
// not meant to stand in anywhere a real cryptographic hash target would.
func HashPath(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:8])
}

// Operation logs a vault operation (open/save/import) identified by
// pathHash (see HashPath), not by path. errCategory is a short, fixed
// label the caller chose (e.g. "decrypt-failed"), or "" for success —
// never a raw error message. Spec section 22.1 allows logging an error
// *category*, not error content, and for good reason: Go's own
// filesystem errors routinely embed the full path spec forbids logging
// (a *fs.PathError's Error() string includes it), so passing err.Error()
// straight through here would leak exactly what this type exists to
// prevent.
func (l *Logger) Operation(op, pathHash, errCategory string) {
	if errCategory != "" {
		l.slog.Error("vault operation failed", "op", op, "path_hash", pathHash, "category", errCategory)
		return
	}
	l.slog.Info("vault operation", "op", op, "path_hash", pathHash)
}

// Version logs the application and runtime version info spec sections
// 22.1 and 22.2 (:doctor) both allow.
func (l *Logger) Version(appVersion, goVersion, os, arch string) {
	l.slog.Info("version", "app", appVersion, "go", goVersion, "os", os, "arch", arch)
}

// KDBXVersion logs the detected file format version of an opened vault.
func (l *Logger) KDBXVersion(version string) {
	l.slog.Info("kdbx version detected", "version", version)
}

// FeatureFlags logs which optional features are currently active.
func (l *Logger) FeatureFlags(flags map[string]bool) {
	args := make([]any, 0, len(flags)*2)
	for name, enabled := range flags {
		args = append(args, name, enabled)
	}
	l.slog.Info("feature flags", args...)
}

// EntryCounts logs how many entries and groups a vault holds — a count,
// never the entries or groups themselves.
func (l *Logger) EntryCounts(entries, groups int) {
	l.slog.Info("vault contents", "entries", entries, "groups", groups)
}

// Timing logs how long operation took, with no secret-bearing detail
// attached.
func (l *Logger) Timing(operation string, d time.Duration) {
	l.slog.Info("timing", "operation", operation, "duration_ms", d.Milliseconds())
}

// ErrorCategory logs a categorized failure that isn't tied to a specific
// vault operation (see Operation for that case). Like Operation,
// category must be a short fixed label, never a raw error message.
func (l *Logger) ErrorCategory(category string) {
	l.slog.Error("error", "category", category)
}
