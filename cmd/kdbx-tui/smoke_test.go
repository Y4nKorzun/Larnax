package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/application"
)

// TestBinarySmokeVersionAndDoctor builds the actual kdbx-tui binary and
// runs it as a real subprocess — a different, stronger check than
// main_test.go's in-process run() calls: it proves `go build
// ./cmd/kdbx-tui` really produces a working executable that a shell can
// invoke, not just that run()'s Go-level logic is correct in isolation.
func TestBinarySmokeVersionAndDoctor(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "kdbx-tui")

	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/kdbx-tui failed: %v\n%s", err, out)
	}

	versionOut, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("%s --version failed: %v\n%s", bin, err, versionOut)
	}
	if got := strings.TrimSpace(string(versionOut)); got != application.Version {
		t.Errorf("--version output = %q, want %q", got, application.Version)
	}

	doctorOut, err := exec.Command(bin, "--doctor").CombinedOutput()
	if err != nil {
		t.Fatalf("%s --doctor failed: %v\n%s", bin, err, doctorOut)
	}
	for _, want := range []string{
		"Application: " + application.Version,
		"KDBX read: supported",
		"Network capability: unused",
	} {
		if !strings.Contains(string(doctorOut), want) {
			t.Errorf("--doctor output missing %q; full output:\n%s", want, doctorOut)
		}
	}
}

// TestBinarySmokeUnknownFlagExitsNonZero confirms the real subprocess's
// exit code, not just run()'s in-process return value.
func TestBinarySmokeUnknownFlagExitsNonZero(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "kdbx-tui")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/kdbx-tui failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "--this-flag-does-not-exist")
	err := cmd.Run()
	if err == nil {
		t.Error("process exited 0 for an unknown flag, want non-zero")
	}
}
