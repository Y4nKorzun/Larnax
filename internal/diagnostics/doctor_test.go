package diagnostics

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewReportFillsRuntimeFields(t *testing.T) {
	r := NewReport("0.1.0", true, true, "supported for detected feature set", "acceptable")

	if r.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", r.GoVersion, runtime.Version())
	}
	if r.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", r.OS, runtime.GOOS)
	}
	if r.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", r.Arch, runtime.GOARCH)
	}
	if r.AppVersion != "0.1.0" {
		t.Errorf("AppVersion = %q, want %q", r.AppVersion, "0.1.0")
	}
}

func TestReportStringMatchesSpecLayout(t *testing.T) {
	r := Report{
		AppVersion:           "0.1.0",
		GoVersion:            "go1.26.5",
		OS:                   "darwin",
		Arch:                 "arm64",
		TerminalDetected:     true,
		ClipboardAvailable:   true,
		KDBXWriteNote:        "supported for detected feature set",
		FileLockAvailable:    true,
		VaultPermissionsNote: "acceptable",
	}

	got := r.String()
	for _, want := range []string{
		"Application: 0.1.0",
		"Go: go1.26.5",
		"OS: darwin/arm64",
		"Terminal: detected",
		"Clipboard: available",
		"KDBX read: supported",
		"KDBX write: supported for detected feature set",
		"File lock: available",
		"Vault permissions: acceptable",
		"Network capability: unused",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Report.String() missing %q; full output:\n%s", want, got)
		}
	}
}

func TestReportStringLabelsUnavailableCapabilities(t *testing.T) {
	r := Report{ClipboardAvailable: false, FileLockAvailable: false, TerminalDetected: false}
	got := r.String()

	for _, want := range []string{"Terminal: not detected", "Clipboard: unavailable", "File lock: unavailable"} {
		if !strings.Contains(got, want) {
			t.Errorf("Report.String() missing %q; full output:\n%s", want, got)
		}
	}
}

func TestReportStringAlwaysShowsKDBXReadSupportedAndNetworkUnused(t *testing.T) {
	got := Report{}.String()
	if !strings.Contains(got, "KDBX read: supported") {
		t.Error(`Report{}.String() missing "KDBX read: supported"`)
	}
	if !strings.Contains(got, "Network capability: unused") {
		t.Error(`Report{}.String() missing "Network capability: unused"`)
	}
}

func TestIsTerminalFalseForRegularFile(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "not-a-terminal"))
	if err != nil {
		t.Fatalf("os.Create() error = %v", err)
	}
	defer f.Close()

	if isTerminal(f) {
		t.Error("isTerminal() = true for a regular file, want false")
	}
}
