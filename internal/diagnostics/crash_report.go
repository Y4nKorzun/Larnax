package diagnostics

// CrashReportsEnabled reports whether spec section 22.3 currently allows
// this application to ever write a local crash dump: "if safe redaction
// cannot be guaranteed, the crash dump feature is not enabled."
//
// Redact and RedactSecret (redaction.go) are explicitly documented as
// defense-in-depth, not a guarantee — they only catch a secret value a
// caller already knows to pass in, not every way one could end up in an
// arbitrary panic message or stack trace a crash handler would capture.
// Closing that gap is exactly the precondition spec 22.3 sets before
// crash dumps may turn on, and it hasn't been closed, so this returns
// false.
//
// There is deliberately no parameter and no way to force this true: spec
// 22.3 treats "cannot guarantee it" as a hard no for the whole feature,
// not a per-call choice a caller gets to override.
func CrashReportsEnabled() bool {
	return false
}

// CrashReportsDisabledReason explains CrashReportsEnabled's current
// value, for :doctor or a settings screen to show instead of an
// unexplained "off."
func CrashReportsDisabledReason() string {
	return "crash dumps are disabled: safe redaction of secrets from an arbitrary panic or stack trace cannot yet be guaranteed (spec section 22.3)"
}
