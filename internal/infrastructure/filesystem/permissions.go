package filesystem

import (
	"errors"
	"os"
	"runtime"
)

// File and directory modes spec section 16.6 requires on Unix.
const (
	SecretFileMode = 0o600 // vault, temp, and backup files
	SecretDirMode  = 0o700 // backup directory
)

// ErrPermissionsNotGuaranteed indicates the platform cannot provide the
// same access restriction Unix permission bits do. Spec section 16.6 is
// explicit: on Windows the application uses a "best-effort restricted
// ACL," and if it cannot guarantee the expected permissions, it must show
// a warning rather than make a false claim. Go's standard library only
// exposes Chmod, which on Windows can merely toggle the read-only
// attribute — it cannot express an owner-only ACL — so this package
// reports that limitation honestly instead of silently proceeding as if
// SecretFileMode/SecretDirMode had taken effect the way they do on Unix.
var ErrPermissionsNotGuaranteed = errors.New("filesystem: file permissions could not be guaranteed on this platform")

// Restrict applies mode to path. On Unix this fully restricts access to
// the owner, as mode itself specifies. On Windows it still attempts
// os.Chmod (which can toggle the read-only attribute) but always returns
// ErrPermissionsNotGuaranteed afterward, since Go's stdlib cannot set a
// real restrictive ACL there.
func Restrict(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err != nil {
		return err
	}
	if !permissionsAreMeaningful(runtime.GOOS) {
		return ErrPermissionsNotGuaranteed
	}
	return nil
}

// Verify reports whether path's current permissions already equal mode.
// On Windows it returns (false, ErrPermissionsNotGuaranteed): Stat's mode
// bits there don't reflect real ACL-based access control, so reporting a
// definitive true/false from them would itself be the kind of false claim
// spec section 16.6 forbids.
func Verify(path string, mode os.FileMode) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !permissionsAreMeaningful(runtime.GOOS) {
		return false, ErrPermissionsNotGuaranteed
	}
	return info.Mode().Perm() == mode, nil
}

// permissionsAreMeaningful reports whether goos's permission bits provide
// a real owner-only access restriction the way Unix's do. Extracted from
// Restrict/Verify's platform check so both branches are directly
// unit-testable without needing to actually run this suite on Windows.
func permissionsAreMeaningful(goos string) bool {
	return goos != "windows"
}
