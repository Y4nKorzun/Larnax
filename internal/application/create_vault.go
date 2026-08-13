package application

import (
	"errors"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

// ErrEmptyMasterPassphrase is returned by CreateVault when passphrase is
// empty. Spec section 7 has no "no encryption" mode to fall back to.
var ErrEmptyMasterPassphrase = errors.New("application: master passphrase must not be empty")

// CreateVault builds a brand-new vault named name, secured with
// masterPassphrase under the portable Argon2id profile (spec section
// 7.2). It returns a kdbx.Document ready for kdbx.Save — CreateVault only
// builds the in-memory structure, it never touches disk.
func CreateVault(name, masterPassphrase string) (*kdbx.Document, error) {
	if masterPassphrase == "" {
		return nil, ErrEmptyMasterPassphrase
	}

	root := domain.Group{ID: domain.NewGroupID(), Name: name}
	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		return nil, err
	}

	return kdbx.NewDocument(vault, masterPassphrase, kdbx.PortableProfile())
}
