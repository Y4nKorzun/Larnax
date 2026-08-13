package filesystem

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// sequentialClock returns each of times in order on successive calls,
// giving tests full control over backup timestamps without relying on
// real wall-clock timing or sleeping between Create calls.
func sequentialClock(times ...time.Time) func() time.Time {
	i := 0
	return func() time.Time {
		t := times[i]
		if i < len(times)-1 {
			i++
		}
		return t
	}
}

func TestBackupStoreCreateWritesFile(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "personal.kdbx")
	store := NewBackupStore(vaultPath, 10)

	path, err := store.Create([]byte("kdbx bytes"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(got) != "kdbx bytes" {
		t.Errorf("backup content = %q, want %q", got, "kdbx bytes")
	}
}

func TestBackupStoreCreateUsesSpecFilenameFormat(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "personal.kdbx")
	store := &BackupStore{
		vaultPath: vaultPath,
		retention: 10,
		now:       func() time.Time { return time.Date(2026, 8, 12, 18, 30, 2, 0, time.UTC) },
	}

	path, err := store.Create([]byte("x"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	want := "2026-08-12T18-30-02Z.kdbx"
	if got := filepath.Base(path); got != want {
		t.Errorf("filename = %q, want %q (spec section 16.5's own example format)", got, want)
	}
}

func TestBackupStoreCreateUsesCorrectDirectory(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "Secrets", "personal.kdbx")
	if err := os.MkdirAll(filepath.Dir(vaultPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	store := NewBackupStore(vaultPath, 10)

	path, err := store.Create([]byte("x"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	wantDir := filepath.Join(dir, "Secrets", ".personal.kdbx.backups")
	if got := filepath.Dir(path); got != wantDir {
		t.Errorf("backup directory = %q, want %q", got, wantDir)
	}
}

func TestBackupStoreDirectoryAndFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "personal.kdbx")
	store := NewBackupStore(vaultPath, 10)

	path, err := store.Create([]byte("x"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(file) error = %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0o600 {
		t.Errorf("backup file mode = %o, want 0600 (spec section 16.6)", perm)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("os.Stat(dir) error = %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("backup directory mode = %o, want 0700 (spec section 16.6)", perm)
	}
}

func TestBackupStoreListEmptyWhenNoBackupsYet(t *testing.T) {
	dir := t.TempDir()
	store := NewBackupStore(filepath.Join(dir, "personal.kdbx"), 10)

	backups, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("List() = %v, want empty", backups)
	}
}

func TestBackupStoreListSortedOldestFirst(t *testing.T) {
	dir := t.TempDir()
	times := []time.Time{
		time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
	}
	store := &BackupStore{
		vaultPath: filepath.Join(dir, "personal.kdbx"),
		retention: 10,
		now:       sequentialClock(times...),
	}

	for range times {
		if _, err := store.Create([]byte("x")); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	backups, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	want := []string{
		"2026-08-10T00-00-00Z.kdbx",
		"2026-08-11T00-00-00Z.kdbx",
		"2026-08-12T00-00-00Z.kdbx",
	}
	if len(backups) != len(want) {
		t.Fatalf("List() = %v, want %v", backups, want)
	}
	for i := range want {
		if backups[i] != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, backups[i], want[i])
		}
	}
}

func TestBackupStoreRetentionPrunesOldest(t *testing.T) {
	dir := t.TempDir()
	times := []time.Time{
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
	}
	store := &BackupStore{
		vaultPath: filepath.Join(dir, "personal.kdbx"),
		retention: 2,
		now:       sequentialClock(times...),
	}

	for range times {
		if _, err := store.Create([]byte("x")); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	backups, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	want := []string{
		"2026-08-04T00-00-00Z.kdbx",
		"2026-08-05T00-00-00Z.kdbx",
	}
	if len(backups) != len(want) {
		t.Fatalf("List() = %v, want the 2 newest: %v", backups, want)
	}
	for i := range want {
		if backups[i] != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, backups[i], want[i])
		}
	}
}

func TestBackupStorePruneHandlesNonPositiveRetention(t *testing.T) {
	for _, retention := range []int{0, -1} {
		t.Run("", func(t *testing.T) {
			dir := t.TempDir()
			times := []time.Time{
				time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
			}
			store := &BackupStore{
				vaultPath: filepath.Join(dir, "personal.kdbx"),
				retention: retention,
				now:       sequentialClock(times...),
			}

			for range times {
				if _, err := store.Create([]byte("x")); err != nil {
					t.Fatalf("Create() with retention=%d error = %v", retention, err)
				}
			}

			backups, err := store.List()
			if err != nil {
				t.Fatalf("List() with retention=%d error = %v", retention, err)
			}
			if len(backups) != 0 {
				t.Errorf("List() with retention=%d = %v, want empty (keep none)", retention, backups)
			}
		})
	}
}
