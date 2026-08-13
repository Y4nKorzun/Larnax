package application

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
)

// FieldName identifies which of an entry's fields CopyField copies. Spec
// section 11.2 defines exactly these three for P0 (yu/yp/yU); yt (TOTP)
// is P1 scope and has no FieldName here yet.
type FieldName int

const (
	FieldUsername FieldName = iota
	FieldPassword
	FieldURL
)

// ErrUnknownField is returned by CopyField for a FieldName outside the
// three this package currently defines.
var ErrUnknownField = errors.New("application: unknown field")

// CopyField writes one field of entry to the clipboard via cb and returns
// the ownership hash clipboard.ClearIfOwned needs to expire it safely —
// after the configured timeout (spec section 11.4), but only if the
// clipboard still holds exactly this value. Scheduling that later call is
// the caller's responsibility, same division as clipboard.SecureCopy
// itself.
func CopyField(ctx context.Context, cb clipboard.Clipboard, entry domain.Entry, field FieldName) ([sha256.Size]byte, error) {
	switch field {
	case FieldUsername:
		return clipboard.SecureCopy(ctx, cb, []byte(entry.Username))
	case FieldURL:
		return clipboard.SecureCopy(ctx, cb, []byte(entry.URL))
	case FieldPassword:
		var hash [sha256.Size]byte
		err := entry.Password.Reveal(func(value []byte) error {
			var err error
			hash, err = clipboard.SecureCopy(ctx, cb, value)
			return err
		})
		if err != nil {
			return [sha256.Size]byte{}, err
		}
		return hash, nil
	default:
		return [sha256.Size]byte{}, ErrUnknownField
	}
}
