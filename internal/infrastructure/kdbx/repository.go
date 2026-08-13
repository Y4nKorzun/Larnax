package kdbx

import (
	"io"

	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// Document couples a decoded Vault with the *gokeepasslib.Database it was
// decoded from. Save needs the Database back, not just the Vault: KDBX
// settings the domain model has no field for at all — header, KDF
// parameters, cipher, compression — live only there, and Save must carry
// them through unchanged rather than resetting them to a library default
// on every write (spec section 15.3 forbids silent version/setting
// changes).
type Document struct {
	Vault    *domain.Vault
	Database *gokeepasslib.Database
}

// Open decodes a KDBX file from r with credentials, unlocks its protected
// entries, and maps the result into a Document (spec section 15.2's
// lifecycle up through "map into domain model").
//
// Open does not consult feature_detector.go: deciding whether a decoded
// file is safe to ever write back — and falling back to read-only when
// it is not (spec section 15.4) — is application-layer policy, not this
// package's job. Every entry/group field this package's domain model has
// no slot for (History, Binaries, CustomData, AutoType, icons, and more)
// is silently dropped the moment Save reconstructs the tree from
// Document.Vault, whether or not the caller ever touched it. Callers
// MUST run DetectUnsupportedFeatures against Document.Database before
// trusting Save with a file that didn't originate from this package.
func Open(r io.Reader, credentials *gokeepasslib.DBCredentials) (*Document, error) {
	db := gokeepasslib.NewDatabase()
	db.Credentials = credentials

	if err := gokeepasslib.NewDecoder(r).Decode(db); err != nil {
		return nil, err
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, err
	}

	vault, err := VaultFromDatabase(db)
	if err != nil {
		return nil, err
	}

	return &Document{Vault: vault, Database: db}, nil
}

// Save maps doc.Vault back into doc.Database's group tree, locks
// protected entries, and encodes the result to w (spec section 15.2's
// "map back -> lock -> encode"). doc.Database keeps whatever header, KDF,
// and cipher settings it already had — from Open, or from a caller that
// built one directly for a brand-new vault — untouched by this function.
func Save(w io.Writer, doc *Document) error {
	root, err := RootGroupFromVault(doc.Vault)
	if err != nil {
		return err
	}
	doc.Database.Content.Root.Groups = []gokeepasslib.Group{root}

	if err := doc.Database.LockProtectedEntries(); err != nil {
		return err
	}
	return gokeepasslib.NewEncoder(w).Encode(doc.Database)
}
