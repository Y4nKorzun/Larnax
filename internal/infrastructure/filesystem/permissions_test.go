package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRestrictSetsUnixFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	path := filepath.Join(t.TempDir(), "secret.kdbx")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := Restrict(path, SecretFileMode); err != nil {
		t.Fatalf("Restrict() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != SecretFileMode {
		t.Errorf("mode = %o, want %o", perm, SecretFileMode)
	}
}

func TestRestrictSetsUnixDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	dir := filepath.Join(t.TempDir(), "backups")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}

	if err := Restrict(dir, SecretDirMode); err != nil {
		t.Fatalf("Restrict() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != SecretDirMode {
		t.Errorf("mode = %o, want %o", perm, SecretDirMode)
	}
}

func TestRestrictPropagatesChmodError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.kdbx")

	err := Restrict(path, SecretFileMode)
	if err == nil {
		t.Fatal("Restrict() error = nil, want an error for a nonexistent path")
	}
	if errors.Is(err, ErrPermissionsNotGuaranteed) {
		t.Error("Restrict() returned ErrPermissionsNotGuaranteed for a real Chmod failure; want the underlying error surfaced instead")
	}
}

func TestVerifyReportsMatchingPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	path := filepath.Join(t.TempDir(), "secret.kdbx")
	if err := os.WriteFile(path, []byte("x"), SecretFileMode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	match, err := Verify(path, SecretFileMode)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !match {
		t.Error("Verify() = false, want true")
	}
}

func TestVerifyReportsMismatchedPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	path := filepath.Join(t.TempDir(), "secret.kdbx")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	match, err := Verify(path, SecretFileMode)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if match {
		t.Error("Verify() = true, want false (mode is 0644, not 0600)")
	}
}

func TestVerifyPropagatesStatError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.kdbx")

	_, err := Verify(path, SecretFileMode)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Verify() error = %v, want os.ErrNotExist", err)
	}
}

func TestPermissionsAreMeaningfulByPlatform(t *testing.T) {
	cases := []struct {
		goos string
		want bool
	}{
		{"windows", false},
		{"darwin", true},
		{"linux", true},
		{"freebsd", true},
	}
	for _, c := range cases {
		if got := permissionsAreMeaningful(c.goos); got != c.want {
			t.Errorf("permissionsAreMeaningful(%q) = %v, want %v", c.goos, got, c.want)
		}
	}
}
