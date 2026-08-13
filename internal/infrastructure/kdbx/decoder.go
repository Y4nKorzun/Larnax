package kdbx

import (
	"errors"

	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// ErrEmptyDatabase is returned when a Database has no root group to build
// a Vault from. Every valid KDBX file has exactly one (gokeepasslib's own
// NewDatabase and NewRootData both pre-populate it), so encountering none
// means db was never decoded or was constructed by hand incorrectly.
var ErrEmptyDatabase = errors.New("kdbx: database has no root group")

// VaultFromDatabase converts a decoded, already-unlocked Database (spec
// section 15.2's decode -> unlock -> map lifecycle) into a domain.Vault.
//
// Callers must call db.UnlockProtectedEntries() first. This function has
// no crypto dependency and cannot detect a still-locked database — it
// will simply copy ciphertext into Password and custom field values
// instead of failing loudly.
func VaultFromDatabase(db *gokeepasslib.Database) (*domain.Vault, error) {
	if db == nil || db.Content == nil || db.Content.Root == nil || len(db.Content.Root.Groups) == 0 {
		return nil, ErrEmptyDatabase
	}

	rootGKP := db.Content.Root.Groups[0]
	root := groupFromGKP(rootGKP)

	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		return nil, err
	}

	if err := addChildren(vault, rootGKP, root.ID); err != nil {
		return nil, err
	}
	return vault, nil
}

// addChildren recursively adds gg's child groups and entries into vault
// under parentID, mirroring gg's position in the source tree. It is
// called once per group node, so it only ever needs to look one level
// down (gg.Groups, gg.Entries) before recursing.
func addChildren(vault *domain.Vault, gg gokeepasslib.Group, parentID domain.GroupID) error {
	for _, childGKP := range gg.Groups {
		child := groupFromGKP(childGKP)
		child.ParentID = &parentID
		if err := vault.AddGroup(child); err != nil {
			return err
		}
		if err := addChildren(vault, childGKP, child.ID); err != nil {
			return err
		}
	}

	for _, entryGKP := range gg.Entries {
		entry := entryFromGKP(entryGKP)
		entry.ParentGroup = parentID
		if err := vault.AddEntry(entry); err != nil {
			return err
		}
	}

	return nil
}
