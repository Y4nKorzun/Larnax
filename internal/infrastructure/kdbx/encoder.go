package kdbx

import (
	"github.com/tobischo/gokeepasslib/v3"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// RootGroupFromVault converts vault's flat map + parent-pointer structure
// back into gokeepasslib's nested Group tree, ready to be assigned to a
// Database's Content.Root.Groups (spec section 15.2's "map back" step,
// the inverse of decoder.go's VaultFromDatabase). The caller still owns
// calling Database.LockProtectedEntries before encoding — this function
// only builds the tree, it does not touch a Database.
//
// It returns an error only if entryToGKP fails to reveal an already
// cleared secret — Vault's own invariants (spec section 18.4, enforced by
// AddGroup/AddEntry/MoveGroup/MoveEntry) already guarantee every parent
// reference in vault resolves, so no other failure mode exists here.
func RootGroupFromVault(vault *domain.Vault) (gokeepasslib.Group, error) {
	return buildGroup(vault, vault.RootGroupID())
}

// buildGroup converts the group id into a gokeepasslib.Group, then
// recurses into its direct entries and child groups.
func buildGroup(vault *domain.Vault, id domain.GroupID) (gokeepasslib.Group, error) {
	g, err := vault.Group(id)
	if err != nil {
		return gokeepasslib.Group{}, err
	}
	gg := groupToGKP(g)

	for _, entry := range vault.EntriesIn(id) {
		ge, err := entryToGKP(entry)
		if err != nil {
			return gokeepasslib.Group{}, err
		}
		gg.Entries = append(gg.Entries, ge)
	}

	for _, child := range vault.ChildGroups(id) {
		childGKP, err := buildGroup(vault, child.ID)
		if err != nil {
			return gokeepasslib.Group{}, err
		}
		gg.Groups = append(gg.Groups, childGKP)
	}

	return gg, nil
}
