package diagnostics

import "testing"

func TestCrashReportsDisabledByDefault(t *testing.T) {
	if CrashReportsEnabled() {
		t.Error("CrashReportsEnabled() = true, want false (spec 22.3: not enabled until safe redaction is guaranteed)")
	}
}

func TestCrashReportsDisabledReasonIsNonEmpty(t *testing.T) {
	if CrashReportsDisabledReason() == "" {
		t.Error("CrashReportsDisabledReason() = \"\", want a non-empty explanation")
	}
}
