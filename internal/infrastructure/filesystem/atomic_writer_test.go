package filesystem

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteCreatesFileWithContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")

	if err := Write(path, []byte("kdbx bytes"), AtomicWriteOptions{}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(got) != "kdbx bytes" {
		t.Errorf("content = %q, want %q", got, "kdbx bytes")
	}
}

func TestWriteReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := Write(path, []byte("new content"), AtomicWriteOptions{}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(got) != "new content" {
		t.Errorf("content = %q, want %q", got, "new content")
	}
}

func TestWriteLeavesNoTempFileBehindOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personal.kdbx")

	if err := Write(path, []byte("x"), AtomicWriteOptions{}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "personal.kdbx" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory contents = %v, want only [personal.kdbx]", names)
	}
}

func TestWriteSetsRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	path := filepath.Join(t.TempDir(), "personal.kdbx")
	if err := Write(path, []byte("x"), AtomicWriteOptions{}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != SecretFileMode {
		t.Errorf("mode = %o, want %o", perm, SecretFileMode)
	}
}

func TestWriteCallsVerifyWithWrittenContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	var gotVerify []byte

	err := Write(path, []byte("kdbx bytes"), AtomicWriteOptions{
		Verify: func(data []byte) error {
			gotVerify = append([]byte(nil), data...)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if string(gotVerify) != "kdbx bytes" {
		t.Errorf("Verify received %q, want %q", gotVerify, "kdbx bytes")
	}
}

// TestWriteAbortsOnVerifyFailureLeavesOriginalIntact is the core safety
// property spec section 16.1-16.2 exists for: a failed save must never
// touch the existing file.
func TestWriteAbortsOnVerifyFailureLeavesOriginalIntact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personal.kdbx")
	original := []byte("original kdbx bytes")
	if err := os.WriteFile(path, original, SecretFileMode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	sentinel := errors.New("structure check failed")
	err := Write(path, []byte("corrupted or wrong new data"), AtomicWriteOptions{
		Verify: func(data []byte) error { return sentinel },
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Write() error = %v, want to wrap %v", err, sentinel)
	}

	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("os.ReadFile() error = %v", readErr)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("original file was modified: got %q, want unchanged %q", got, original)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory contents = %v, want only the original file (temp file not cleaned up)", names)
	}
}

func TestWriteAbortsOnVerifyFailureForNewFileCreatesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personal.kdbx")

	err := Write(path, []byte("data"), AtomicWriteOptions{
		Verify: func(data []byte) error { return errors.New("bad structure") },
	})
	if err == nil {
		t.Fatal("Write() error = nil, want an error")
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("target file exists after a failed first save, want it never created")
	}
}

func TestWriteCallsBackupWithOriginalContentBeforeRename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	original := []byte("original content")
	if err := os.WriteFile(path, original, SecretFileMode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	var gotBackup []byte
	err := Write(path, []byte("new content"), AtomicWriteOptions{
		Backup: func(originalData []byte) error {
			gotBackup = append([]byte(nil), originalData...)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !bytes.Equal(gotBackup, original) {
		t.Errorf("Backup received %q, want the original content %q", gotBackup, original)
	}
}

func TestWriteSkipsBackupWhenNoOriginalExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	called := false

	err := Write(path, []byte("data"), AtomicWriteOptions{
		Backup: func(originalData []byte) error {
			called = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if called {
		t.Error("Backup was called for a brand-new file with no original to back up")
	}
}

func TestWriteAbortsOnBackupFailureLeavesOriginalIntact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	original := []byte("original content")
	if err := os.WriteFile(path, original, SecretFileMode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	sentinel := errors.New("disk full during backup")
	err := Write(path, []byte("new content"), AtomicWriteOptions{
		Backup: func(originalData []byte) error { return sentinel },
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Write() error = %v, want to wrap %v", err, sentinel)
	}

	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("os.ReadFile() error = %v", readErr)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("original file was modified despite backup failure: got %q, want %q", got, original)
	}
}

func TestWriteRunsVerifyBeforeBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	if err := os.WriteFile(path, []byte("original"), SecretFileMode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	var order []string
	err := Write(path, []byte("new"), AtomicWriteOptions{
		Verify: func(data []byte) error {
			order = append(order, "verify")
			return nil
		},
		Backup: func(originalData []byte) error {
			order = append(order, "backup")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(order) != 2 || order[0] != "verify" || order[1] != "backup" {
		t.Errorf("call order = %v, want [verify backup] (spec section 16.2 steps 7 then 8)", order)
	}
}

func TestWriteWithoutOptionsSucceeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personal.kdbx")
	if err := Write(path, []byte("data"), AtomicWriteOptions{}); err != nil {
		t.Fatalf("Write() with nil Verify/Backup error = %v", err)
	}
}
