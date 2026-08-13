package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Y4nKorzun/Larnax/internal/diagnostics"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/filesystem"
)

// BuildDoctorReport assembles spec section 22.2's :doctor report by
// actually probing the environment — diagnostics.NewReport itself
// deliberately leaves that composition to a caller, so the diagnostics
// package can stay decoupled from clipboard/filesystem/kdbx (see its own
// doc comment). This is that caller.
//
// vault may be nil, or open or not: :doctor works before any vault has
// been opened (spec's example output has no vault-specific detail beyond
// the two lines this function tailors when one is open).
func BuildDoctorReport(appVersion string, vault *VaultService) diagnostics.Report {
	return diagnostics.NewReport(
		appVersion,
		clipboard.Available(),
		fileLockAvailable(),
		kdbxWriteNote(vault),
		vaultPermissionsNote(vault),
	)
}

// fileLockAvailable probes whether this process can actually acquire and
// release a write lock (spec section 16.4), against a scratch path in
// the OS temp directory rather than any real vault — :doctor can run
// before a vault is open at all.
func fileLockAvailable() bool {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("kdbx-tui-doctor-%d", os.Getpid()))
	lock, err := filesystem.Acquire(path)
	if err != nil {
		return false
	}
	_ = lock.Release()
	return true
}

func kdbxWriteNote(vault *VaultService) string {
	if vault == nil || !vault.IsOpen() {
		return "supported (no vault open to check feature compatibility)"
	}
	if vault.ReadOnly() {
		return "unavailable — unsupported KDBX features forced read-only (spec 15.4)"
	}
	return "supported for detected feature set"
}

func vaultPermissionsNote(vault *VaultService) string {
	if vault == nil || !vault.IsOpen() || vault.Path() == "" {
		return "not applicable (no vault open)"
	}

	ok, err := filesystem.Verify(vault.Path(), filesystem.SecretFileMode)
	switch {
	case errors.Is(err, filesystem.ErrPermissionsNotGuaranteed):
		return "not guaranteed on this platform"
	case err != nil:
		return "could not be checked"
	case ok:
		return "acceptable"
	default:
		return "too permissive"
	}
}
