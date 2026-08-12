// Package filesystem provides the safe-file-handling primitives spec
// section 16 requires: detecting external modification here, with atomic
// writes, backups, and file locking as separate, later pieces of work.
package filesystem

import (
	"crypto/sha256"
	"os"
	"time"
)

// FileRevision captures everything spec section 16.3 requires to detect
// whether a file changed on disk since it was last read: path, size,
// modification time, and a cryptographic hash of the file's bytes. The
// hash is the authoritative signal — size and modification time are
// exposed for callers that want to show a summary (spec section 16.3's
// "[d] Show non-secret change summary") without claiming more precision
// than mtime actually offers (filesystems commonly have only 1-second
// mtime resolution, and some tools rewrite a file without changing its
// size).
type FileRevision struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
	Hash         [sha256.Size]byte
}

// ReadFileRevision computes the current FileRevision for path. Size and
// Hash are derived from the same read so they can never disagree with each
// other; ModifiedTime comes from a separate stat call, since it cannot be
// derived from content.
func ReadFileRevision(path string) (FileRevision, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileRevision{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return FileRevision{}, err
	}
	return FileRevision{
		Path:         path,
		Size:         int64(len(data)),
		ModifiedTime: info.ModTime(),
		Hash:         sha256.Sum256(data),
	}, nil
}

// SameContent reports whether a and b represent the same file content,
// using the cryptographic hash as the sole, authoritative signal (spec
// section 16.3) — matching size or modification time is not sufficient,
// and their absence of a match must not be treated as sufficient evidence
// of change either.
func SameContent(a, b FileRevision) bool {
	return a.Hash == b.Hash
}

// Recheck implements spec section 16.3's pre-save check: recompute the
// revision at old.Path and report whether its content differs from what
// old captured. The application must not auto-resolve a detected change
// (spec: "Приложение не выполняет автоматический last-write-wins") — that
// choice belongs to the user.
func Recheck(old FileRevision) (changed bool, current FileRevision, err error) {
	current, err = ReadFileRevision(old.Path)
	if err != nil {
		return false, FileRevision{}, err
	}
	return !SameContent(old, current), current, nil
}
