package filesystem

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestAcquireCreatesLockFileWithDiagnosticInfo(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "personal.kdbx")

	lock, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	if lock.Info.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", lock.Info.PID, os.Getpid())
	}
	wantHash := sha256.Sum256([]byte(vaultPath))
	if lock.Info.VaultPathHash != wantHash {
		t.Errorf("VaultPathHash = %x, want %x", lock.Info.VaultPathHash, wantHash)
	}
	if lock.Info.AcquiredAt.IsZero() {
		t.Error("AcquiredAt is zero")
	}

	lockPath := filepath.Join(filepath.Dir(vaultPath), ".personal.kdbx.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("lock file %q does not exist: %v", lockPath, err)
	}
}

func TestAcquireFailsWhenAlreadyLocked(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "personal.kdbx")

	first, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}
	defer first.Release()

	_, err = Acquire(vaultPath)
	if err == nil {
		t.Fatal("second Acquire() error = nil, want ErrAlreadyLocked")
	}
	if !errors.Is(err, ErrAlreadyLocked) {
		t.Errorf("second Acquire() error = %v, want to wrap %v", err, ErrAlreadyLocked)
	}

	var alreadyLocked *AlreadyLockedError
	if !errors.As(err, &alreadyLocked) {
		t.Fatalf("error does not unwrap to *AlreadyLockedError: %v", err)
	}
	if alreadyLocked.Info.PID != first.Info.PID {
		t.Errorf("reported PID = %d, want %d (the first lock's holder)", alreadyLocked.Info.PID, first.Info.PID)
	}
}

func TestReleaseAllowsReacquisition(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "personal.kdbx")

	lock, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	second, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("Acquire() after Release() error = %v", err)
	}
	defer second.Release()
}

func TestAcquireLockFileHasRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}

	vaultPath := filepath.Join(t.TempDir(), "personal.kdbx")
	lock, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	info, err := os.Stat(lock.path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != SecretFileMode {
		t.Errorf("lock file mode = %o, want %o", perm, SecretFileMode)
	}
}

// TestLockInfoHasNoSecretCapableFields enforces spec section 16.4's
// "секретные данные в lock-файл не записываются" structurally, via a
// denylist of the kinds that could plausibly hold a secret indirectly
// (an interface like domain.Secret, a raw []byte, or a pointer to
// something secret-shaped) rather than an allowlist of today's exact
// field types — so it keeps working if a legitimate new plain field
// (another int or string) is added later, while still catching the kind
// of field that would actually reintroduce spec's forbidden case.
func TestLockInfoHasNoSecretCapableFields(t *testing.T) {
	forbiddenKinds := map[reflect.Kind]bool{
		reflect.Interface: true,
		reflect.Slice:     true,
		reflect.Ptr:       true,
	}

	typ := reflect.TypeOf(LockInfo{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if forbiddenKinds[f.Type.Kind()] {
			t.Errorf("field %s has kind %s, which could hold a secret indirectly", f.Name, f.Type.Kind())
		}
	}
}

func TestVaultPathHashDistinguishesDifferentPaths(t *testing.T) {
	a := sha256.Sum256([]byte("/home/user/a.kdbx"))
	b := sha256.Sum256([]byte("/home/user/b.kdbx"))
	if a == b {
		t.Error("different vault paths hashed to the same value")
	}
}

func TestAcquireRecordsRecentAcquiredAt(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "personal.kdbx")
	before := time.Now()

	lock, err := Acquire(vaultPath)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	after := time.Now()
	if lock.Info.AcquiredAt.Before(before) || lock.Info.AcquiredAt.After(after) {
		t.Errorf("AcquiredAt = %v, want between %v and %v", lock.Info.AcquiredAt, before, after)
	}
}
