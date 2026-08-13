package application

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/filesystem"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

// ErrExternalModification indicates path changed on disk since lastKnown
// was read (spec section 16.3). SaveVault refuses to overwrite it —
// spec is explicit the application must not auto-resolve this with
// last-write-wins; the caller has to re-read and decide.
var ErrExternalModification = errors.New("application: vault file was modified externally since it was last read")

// SaveVault writes doc to path using spec section 16's full safe-save
// sequence: acquire the write lock, recheck the on-disk revision against
// lastKnown, encode and write atomically with verification, back up the
// previous contents, and restrict permissions — releasing the lock
// before returning either way.
//
// Pass the zero filesystem.FileRevision as lastKnown for a vault that has
// never been saved to path before; SaveVault then skips the revision
// check, since there is nothing on disk yet to have diverged from.
//
// retention configures backups (spec section 16.5): retention <= 0 means
// backups are disabled and no backup file is written at all, rather than
// writing one and immediately pruning it back out.
//
// On success it returns the FileRevision of what was just written, for
// the caller to pass as lastKnown on the next SaveVault call.
func SaveVault(path string, doc *kdbx.Document, lastKnown filesystem.FileRevision, retention int) (filesystem.FileRevision, error) {
	lock, err := filesystem.Acquire(path)
	if err != nil {
		return filesystem.FileRevision{}, err
	}
	defer lock.Release()

	if lastKnown.Path != "" {
		changed, _, err := filesystem.Recheck(lastKnown)
		if err != nil {
			return filesystem.FileRevision{}, err
		}
		if changed {
			return filesystem.FileRevision{}, ErrExternalModification
		}
	}

	var buf bytes.Buffer
	if err := kdbx.Save(&buf, doc); err != nil {
		return filesystem.FileRevision{}, err
	}
	data := buf.Bytes()

	opts := filesystem.AtomicWriteOptions{
		Verify: func(reread []byte) error {
			if !bytes.Equal(reread, data) {
				return fmt.Errorf("application: rewritten vault does not match what was written")
			}
			return nil
		},
	}
	if retention > 0 {
		backups := filesystem.NewBackupStore(path, retention)
		opts.Backup = func(original []byte) error {
			_, err := backups.Create(original)
			return err
		}
	}

	if err := filesystem.Write(path, data, opts); err != nil {
		return filesystem.FileRevision{}, err
	}

	if err := filesystem.Restrict(path, filesystem.SecretFileMode); err != nil &&
		!errors.Is(err, filesystem.ErrPermissionsNotGuaranteed) {
		return filesystem.FileRevision{}, err
	}

	return filesystem.ReadFileRevision(path)
}
