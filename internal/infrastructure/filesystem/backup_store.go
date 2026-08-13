package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// backupTimeFormat matches spec section 16.5's own example filename
// exactly ("2026-08-12T18-30-02Z.kdbx"): hyphens instead of colons in the
// time-of-day, since colons are not safe in filenames on every filesystem.
const backupTimeFormat = "2006-01-02T15-04-05Z"

// BackupStore manages timestamped KDBX backup snapshots for one vault
// file, per spec section 16.5: create a backup before each successful
// replacement, keep only the most recent Retention of them, and never
// name a backup after entry content — only a timestamp appears in the
// filename.
type BackupStore struct {
	vaultPath string
	retention int
	now       func() time.Time
}

// NewBackupStore creates a BackupStore for the vault at vaultPath, keeping
// at most retention backups.
func NewBackupStore(vaultPath string, retention int) *BackupStore {
	return &BackupStore{vaultPath: vaultPath, retention: retention, now: time.Now}
}

// dir is the backup directory for this vault: a hidden sibling directory,
// e.g. "personal.kdbx" -> ".personal.kdbx.backups" (spec section 16.5's
// own example).
func (s *BackupStore) dir() string {
	return filepath.Join(filepath.Dir(s.vaultPath), "."+filepath.Base(s.vaultPath)+".backups")
}

// Create writes content as a new timestamped backup, creating the backup
// directory if needed (0700, spec section 16.6) with the backup file
// itself at 0600 (same as the vault). It then prunes backups older than
// the most recent Retention. Pruning removes files outright; per spec
// section 16.5 this must never be described to a user as a "secure
// erase" — it is an ordinary delete.
func (s *BackupStore) Create(content []byte) (string, error) {
	dir := s.dir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	name := s.now().UTC().Format(backupTimeFormat) + ".kdbx"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return "", err
	}

	if err := s.prune(); err != nil {
		return path, err
	}
	return path, nil
}

// List returns this vault's backup filenames (not full paths), sorted
// oldest first. Lexicographic sort matches chronological order here
// because backupTimeFormat is fixed-width. Returns (nil, nil) if no
// backup has ever been created for this vault.
func (s *BackupStore) List() ([]string, error) {
	entries, err := os.ReadDir(s.dir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// prune deletes the oldest backups beyond retention. A negative retention
// is treated as 0 (keep none) rather than underflowing the slice bound
// that would otherwise compute.
func (s *BackupStore) prune() error {
	backups, err := s.List()
	if err != nil {
		return err
	}

	retention := s.retention
	if retention < 0 {
		retention = 0
	}
	if len(backups) <= retention {
		return nil
	}

	for _, name := range backups[:len(backups)-retention] {
		if err := os.Remove(filepath.Join(s.dir(), name)); err != nil {
			return err
		}
	}
	return nil
}
