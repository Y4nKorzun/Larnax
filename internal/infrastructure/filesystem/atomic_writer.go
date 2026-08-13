package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteOptions configures Write's behavior beyond the bare
// write-and-rename.
type AtomicWriteOptions struct {
	// Verify, if non-nil, is called with the temp file's bytes reread
	// from disk (spec section 16.2 step 7: "reopen the temp file and
	// check unlock/structure") before it can replace the original. A
	// re-read from disk, not the in-memory data passed to Write, is what
	// actually catches a partial or corrupted write — e.g. disk-full
	// truncation — that a check against Write's own in-memory copy could
	// never see. An error aborts the write; the original file is left
	// untouched.
	Verify func(data []byte) error

	// Backup, if non-nil, is called with the original file's current
	// bytes (spec section 16.2 step 8: back up the source before
	// replacing it) after Verify succeeds and before the rename. Not
	// called at all if path does not exist yet — creating a brand-new
	// vault has no original to back up. An error aborts the write.
	Backup func(originalData []byte) error
}

// Write implements spec section 16.2's atomic save algorithm: write data
// to a temp file in the same directory as path (so the final rename stays
// atomic — spec is explicit the temp file must be on the same filesystem
// for this reason), fsync it, verify it, back up the original, atomically
// rename the temp file over path, and fsync the containing directory on
// platforms that support it.
//
// Write covers spec's steps 3-10: it does not itself check a
// FileRevision (step 1) or hold a FileLock (step 2) — those decisions
// belong around a Write call, using file_revision.go and file_lock.go,
// not inside it, so this function stays independently testable.
func Write(path string, data []byte, opts AtomicWriteOptions) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("filesystem: creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once the rename below succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("filesystem: writing temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("filesystem: syncing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("filesystem: closing temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, SecretFileMode); err != nil {
		return fmt.Errorf("filesystem: setting temp file permissions: %w", err)
	}

	if opts.Verify != nil {
		reread, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("filesystem: reopening temp file for verification: %w", err)
		}
		if err := opts.Verify(reread); err != nil {
			return fmt.Errorf("filesystem: temp file failed verification: %w", err)
		}
	}

	if opts.Backup != nil {
		original, err := os.ReadFile(path)
		switch {
		case err == nil:
			if err := opts.Backup(original); err != nil {
				return fmt.Errorf("filesystem: backup failed: %w", err)
			}
		case errors.Is(err, os.ErrNotExist):
			// Nothing to back up yet — this is the first save of a new vault.
		default:
			return fmt.Errorf("filesystem: reading original file for backup: %w", err)
		}
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("filesystem: renaming temp file over original: %w", err)
	}

	syncDir(dir)

	return nil
}

// syncDir best-effort fsyncs dir (spec section 16.2 step 10: "on Unix,
// where supported"). Not every platform or filesystem supports fsyncing a
// directory, so any error here is deliberately swallowed rather than
// failing an otherwise-successful save.
func syncDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()
	_ = d.Sync()
}
