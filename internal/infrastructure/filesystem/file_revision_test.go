package filesystem

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}

func TestReadFileRevisionCapturesSizeAndHash(t *testing.T) {
	content := "kdbx file bytes go here"
	path := writeTempFile(t, content)

	rev, err := ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}

	if rev.Path != path {
		t.Errorf("Path = %q, want %q", rev.Path, path)
	}
	if rev.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", rev.Size, len(content))
	}
	if want := sha256.Sum256([]byte(content)); rev.Hash != want {
		t.Errorf("Hash = %x, want %x", rev.Hash, want)
	}
	if rev.ModifiedTime.IsZero() {
		t.Error("ModifiedTime is zero")
	}
}

func TestReadFileRevisionMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.kdbx")

	_, err := ReadFileRevision(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadFileRevision() error = %v, want os.ErrNotExist", err)
	}
}

func TestSameContentTrueForIdenticalBytes(t *testing.T) {
	pathA := writeTempFile(t, "identical content")
	pathB := writeTempFile(t, "identical content")

	revA, err := ReadFileRevision(pathA)
	if err != nil {
		t.Fatalf("ReadFileRevision(A) error = %v", err)
	}
	revB, err := ReadFileRevision(pathB)
	if err != nil {
		t.Fatalf("ReadFileRevision(B) error = %v", err)
	}

	if !SameContent(revA, revB) {
		t.Error("SameContent() = false for two files with identical bytes")
	}
}

func TestSameContentFalseForDifferentBytesEvenIfSameSize(t *testing.T) {
	pathA := writeTempFile(t, "AAAAAAAAAA")
	pathB := writeTempFile(t, "BBBBBBBBBB")

	revA, err := ReadFileRevision(pathA)
	if err != nil {
		t.Fatalf("ReadFileRevision(A) error = %v", err)
	}
	revB, err := ReadFileRevision(pathB)
	if err != nil {
		t.Fatalf("ReadFileRevision(B) error = %v", err)
	}

	if revA.Size != revB.Size {
		t.Fatalf("test setup invariant violated: sizes differ (%d vs %d)", revA.Size, revB.Size)
	}
	if SameContent(revA, revB) {
		t.Error("SameContent() = true for two same-size files with different bytes")
	}
}

func TestRecheckDetectsNoChange(t *testing.T) {
	path := writeTempFile(t, "original content")

	old, err := ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}

	changed, current, err := Recheck(old)
	if err != nil {
		t.Fatalf("Recheck() error = %v", err)
	}
	if changed {
		t.Error("Recheck() changed = true, want false (file untouched)")
	}
	if !SameContent(old, current) {
		t.Error("Recheck() current revision has a different hash than old, want same")
	}
}

func TestRecheckDetectsContentChange(t *testing.T) {
	path := writeTempFile(t, "original content")

	old, err := ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}

	if err := os.WriteFile(path, []byte("modified content"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	changed, current, err := Recheck(old)
	if err != nil {
		t.Fatalf("Recheck() error = %v", err)
	}
	if !changed {
		t.Error("Recheck() changed = false, want true (file was modified)")
	}
	if want := sha256.Sum256([]byte("modified content")); current.Hash != want {
		t.Errorf("current.Hash = %x, want %x", current.Hash, want)
	}
}

// TestRecheckDetectsSameSizeChange guards against an implementation that
// only compares size (or size+mtime) instead of hashing content — spec
// section 16.3 requires the hash to be authoritative for exactly this
// reason.
func TestRecheckDetectsSameSizeChange(t *testing.T) {
	path := writeTempFile(t, "AAAAAAAAAA")

	old, err := ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}

	if err := os.WriteFile(path, []byte("BBBBBBBBBB"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	changed, _, err := Recheck(old)
	if err != nil {
		t.Fatalf("Recheck() error = %v", err)
	}
	if !changed {
		t.Error("Recheck() changed = false for a same-size content change, want true")
	}
}

func TestRecheckPropagatesReadError(t *testing.T) {
	path := writeTempFile(t, "original content")

	old, err := ReadFileRevision(path)
	if err != nil {
		t.Fatalf("ReadFileRevision() error = %v", err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("os.Remove() error = %v", err)
	}

	_, _, err = Recheck(old)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Recheck() error = %v, want os.ErrNotExist", err)
	}
}
