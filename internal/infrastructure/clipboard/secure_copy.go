package clipboard

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
)

// SecureCopy writes secret to the clipboard and returns a SHA-256 hash of
// exactly what was written. The caller holds onto this hash and later
// passes it to ClearIfOwned — typically after the configured clipboard
// timeout (spec section 11.4) — so the clipboard is only cleared if it
// still holds this exact value.
//
// Scheduling the delayed ClearIfOwned call is the caller's responsibility;
// this function only performs the write and computes the ownership hash.
func SecureCopy(ctx context.Context, cb Clipboard, secret []byte) ([sha256.Size]byte, error) {
	if err := cb.WriteText(ctx, secret); err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(secret), nil
}

// ClearIfOwned implements spec section 11.3's clipboard-clear algorithm.
// The naive "sleep then clear" approach can delete text the user copied
// *after* the secret, so instead: read the current clipboard, hash it, and
// clear only if that hash matches ownershipHash (the value SecureCopy
// returned for the write this call is trying to expire). If the user
// copied something else in the meantime, its hash won't match and
// ClearIfOwned leaves the clipboard alone.
//
// The hash comparison uses crypto/subtle for a constant-time compare (spec
// section 11.3), and the locally read clipboard bytes are zeroed
// best-effort before returning.
func ClearIfOwned(ctx context.Context, cb Clipboard, ownershipHash [sha256.Size]byte) error {
	current, err := cb.ReadText(ctx)
	if err != nil {
		return err
	}
	defer zero(current)

	currentHash := sha256.Sum256(current)
	if subtle.ConstantTimeCompare(currentHash[:], ownershipHash[:]) != 1 {
		return nil
	}
	return cb.Clear(ctx)
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
