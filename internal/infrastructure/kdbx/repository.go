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

// Open decodes a KDBX file from r with masterPassphrase, unlocks its
// protected entries, and maps the result into a Document (spec section
// 15.2's lifecycle up through "map into domain model"). Key-file
// credentials are P1 scope (spec section 6.2) — not supported yet.
//
// Open takes a plain passphrase rather than a *gokeepasslib.DBCredentials
// so that no caller above this package needs to import gokeepasslib
// itself (spec section 18.2: only the KDBX adapter talks to gokeepasslib).
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
func Open(r io.Reader, masterPassphrase string) (*Document, error) {
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(masterPassphrase)

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

// NewDocument builds a brand-new KDBX 4.1 Document (spec section 7.2) for
// vault, secured with masterPassphrase under kdf — typically
// PortableProfile(). It never touches disk; pass the result to Save.
func NewDocument(vault *domain.Vault, masterPassphrase string, kdf Argon2Profile) (*Document, error) {
	if err := kdf.Validate(); err != nil {
		return nil, err
	}

	root, err := vault.Group(vault.RootGroupID())
	if err != nil {
		return nil, err
	}

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Credentials = gokeepasslib.NewPasswordCredentials(masterPassphrase)
	db.Content.Meta.DatabaseName = root.Name

	params := db.Header.FileHeaders.KdfParameters
	params.Parallelism = uint32(kdf.Parallelism)
	params.Memory = uint64(kdf.MemoryKiB) * 1024
	params.Iterations = uint64(kdf.Iterations)

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
