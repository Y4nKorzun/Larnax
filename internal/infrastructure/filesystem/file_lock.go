package filesystem

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockInfo is the diagnostic information spec section 16.4 requires a
// write lock to carry — and nothing beyond that: no secrets are ever
// written to a lock file.
//
// AcquiredAt stands in for spec's literal "process start time": Go's
// standard library has no portable, cross-platform way to read a
// process's own start time without OS-specific code (/proc on Linux,
// sysctl on macOS, GetProcessTimes on Windows), so this records when the
// lock itself was taken instead. That serves the same diagnostic purpose
// spec describes — showing how old a lock is — and is arguably more
// directly relevant to this specific lock's staleness than the whole
// process's uptime would be, for a long-running process that opens
// multiple vaults over its lifetime.
type LockInfo struct {
	PID           int
	Hostname      string
	AcquiredAt    time.Time
	VaultPathHash [sha256.Size]byte
}

// ErrAlreadyLocked indicates a vault already has an active write lock.
// Spec section 16.4's UI response to this is "[o] Open read-only /
// [r] Retry lock / [q] Quit" — deliberately not an automatic stale-lock
// override, so Acquire does not attempt to detect or break a stale lock
// on its own; that choice belongs to the user.
var ErrAlreadyLocked = errors.New("filesystem: vault is already locked for writing")

// AlreadyLockedError wraps ErrAlreadyLocked with the diagnostic info the
// current lock holder recorded, so a caller can show spec section 16.4's
// prompt with real PID/hostname detail rather than a generic message.
type AlreadyLockedError struct {
	Info LockInfo
}

func (e *AlreadyLockedError) Error() string {
	return fmt.Sprintf("filesystem: vault is already locked for writing by pid %d on %s", e.Info.PID, e.Info.Hostname)
}

func (e *AlreadyLockedError) Unwrap() error {
	return ErrAlreadyLocked
}

// FileLock is a held write lock, returned by Acquire.
type FileLock struct {
	path string
	Info LockInfo
}

func lockPathFor(vaultPath string) string {
	return filepath.Join(filepath.Dir(vaultPath), "."+filepath.Base(vaultPath)+".lock")
}

// Acquire takes the write lock for vaultPath (spec section 16.4). The
// lock file is created with O_EXCL, which is atomic: two processes racing
// to acquire the same lock can never both succeed, regardless of
// platform, since this relies only on Go's standard library rather than
// OS-specific locking syscalls.
//
// If the lock is already held, Acquire returns an *AlreadyLockedError
// carrying the current holder's LockInfo.
func Acquire(vaultPath string) (*FileLock, error) {
	path := lockPathFor(vaultPath)

	info := LockInfo{
		PID:           os.Getpid(),
		Hostname:      currentHostname(),
		AcquiredAt:    time.Now(),
		VaultPathHash: sha256.Sum256([]byte(vaultPath)),
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, SecretFileMode)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			existing, readErr := readLockInfo(path)
			if readErr != nil {
				return nil, fmt.Errorf("%w (and failed to read existing lock info: %v)", ErrAlreadyLocked, readErr)
			}
			return nil, &AlreadyLockedError{Info: existing}
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(info); err != nil {
		os.Remove(path)
		return nil, err
	}

	return &FileLock{path: path, Info: info}, nil
}

// Release removes the lock file. Only the holder that successfully
// Acquire()'d it should call this.
func (l *FileLock) Release() error {
	return os.Remove(l.path)
}

func currentHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

func readLockInfo(path string) (LockInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LockInfo{}, err
	}
	var info LockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return LockInfo{}, err
	}
	return info, nil
}
