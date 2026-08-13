package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/application"
)

// Only --version and --doctor are exercised here — the interactive path
// (runInteractive, see its own doc comment) needs a real terminal and
// would hang a test process, so it is deliberately not called from any
// test in this file.

func TestRunVersionPrintsVersionAndReturnsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() = %d, want 0; stderr: %s", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != application.Version {
		t.Errorf("stdout = %q, want %q", got, application.Version)
	}
}

func TestRunDoctorPrintsReportAndReturnsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--doctor"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() = %d, want 0; stderr: %s", code, stderr.String())
	}
	for _, want := range []string{"Application: " + application.Version, "KDBX read: supported", "Network capability: unused"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout missing %q; full output:\n%s", want, stdout.String())
		}
	}
}

func TestRunUnknownFlagReturnsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)

	if code == 0 {
		t.Error("run() = 0 for an unknown flag, want non-zero")
	}
}

func TestRunVersionTakesPrecedenceOverDoctor(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version", "--doctor"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() = %d, want 0; stderr: %s", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != application.Version {
		t.Errorf("stdout = %q, want just the version (version checked first)", got)
	}
}
