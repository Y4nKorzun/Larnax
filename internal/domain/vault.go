package domain

// Vault is the in-memory aggregate of a decrypted KDBX database: a tree of
// groups rooted at RootGroupID, and the entries each group directly
// contains.
//
// Vault enforces only structural invariants (a referenced group must exist,
// the group tree stays acyclic). It has no knowledge of KDBX encoding,
// undo/redo, or persistence — those are application/infrastructure
// concerns (spec section 18.1).
type Vault struct {
	rootGroupID GroupID
	groups      map[GroupID]Group
	entries     map[EntryID]Entry
}

// NewVault creates a Vault containing a single root group named name.
func NewVault(name string) *Vault {
	root := Group{ID: NewGroupID(), Name: name}
	return &Vault{
		rootGroupID: root.ID,
		groups:      map[GroupID]Group{root.ID: root},
		entries:     map[EntryID]Entry{},
	}
}

func (v *Vault) RootGroupID() GroupID {
	return v.rootGroupID
}

func (v *Vault) Group(id GroupID) (Group, error) {
	g, ok := v.groups[id]
	if !ok {
		return Group{}, ErrGroupNotFound
	}
	return g, nil
}

func (v *Vault) Entry(id EntryID) (Entry, error) {
	e, ok := v.entries[id]
	if !ok {
		return Entry{}, ErrEntryNotFound
	}
	return e, nil
}

// AddGroup inserts group into the vault. group.ParentID must reference an
// already-present group; only the vault's own root may have a nil parent.
func (v *Vault) AddGroup(group Group) error {
	if group.ParentID == nil {
		return ErrGroupMustHaveParent
	}
	if _, ok := v.groups[*group.ParentID]; !ok {
		return ErrGroupNotFound
	}
	v.groups[group.ID] = group
	return nil
}

// AddEntry inserts entry into the vault. entry.ParentGroup must reference
// an already-present group.
func (v *Vault) AddEntry(entry Entry) error {
	if _, ok := v.groups[entry.ParentGroup]; !ok {
		return ErrGroupNotFound
	}
	v.entries[entry.ID] = entry
	return nil
}

// UpdateEntry replaces the stored entry sharing updated.ID with updated in
// full. updated.ParentGroup must reference an existing group — use
// MoveEntry to change an entry's group along with other field updates in
// a single call.
func (v *Vault) UpdateEntry(updated Entry) error {
	if _, ok := v.entries[updated.ID]; !ok {
		return ErrEntryNotFound
	}
	if _, ok := v.groups[updated.ParentGroup]; !ok {
		return ErrGroupNotFound
	}
	v.entries[updated.ID] = updated
	return nil
}

// RemoveEntry deletes an entry outright. Recycle-bin semantics (spec
// section 9.5) are an application-layer concern layered on top of this.
func (v *Vault) RemoveEntry(id EntryID) error {
	if _, ok := v.entries[id]; !ok {
		return ErrEntryNotFound
	}
	delete(v.entries, id)
	return nil
}

// MoveEntry reassigns entry to a different group.
func (v *Vault) MoveEntry(id EntryID, newParent GroupID) error {
	entry, ok := v.entries[id]
	if !ok {
		return ErrEntryNotFound
	}
	if _, ok := v.groups[newParent]; !ok {
		return ErrGroupNotFound
	}
	entry.ParentGroup = newParent
	v.entries[id] = entry
	return nil
}

// MoveGroup reassigns group to a new parent group. It returns
// ErrCyclicGroupMove if newParent is group itself or one of group's own
// descendants, which would disconnect part of the tree from the root.
func (v *Vault) MoveGroup(id GroupID, newParent GroupID) error {
	group, ok := v.groups[id]
	if !ok {
		return ErrGroupNotFound
	}
	if group.ParentID == nil {
		return ErrCannotMoveRootGroup
	}
	if _, ok := v.groups[newParent]; !ok {
		return ErrGroupNotFound
	}
	if id == newParent || v.isDescendant(newParent, id) {
		return ErrCyclicGroupMove
	}
	group.ParentID = &newParent
	v.groups[id] = group
	return nil
}

// isDescendant reports whether candidate is a descendant of ancestor.
func (v *Vault) isDescendant(candidate, ancestor GroupID) bool {
	current := candidate
	for {
		g, ok := v.groups[current]
		if !ok || g.ParentID == nil {
			return false
		}
		if *g.ParentID == ancestor {
			return true
		}
		current = *g.ParentID
	}
}

// ChildGroups returns the direct child groups of parent, in no particular
// order.
func (v *Vault) ChildGroups(parent GroupID) []Group {
	var children []Group
	for _, g := range v.groups {
		if g.ParentID != nil && *g.ParentID == parent {
			children = append(children, g)
		}
	}
	return children
}

// EntriesIn returns the entries directly inside group, in no particular
// order.
func (v *Vault) EntriesIn(group GroupID) []Entry {
	var entries []Entry
	for _, e := range v.entries {
		if e.ParentGroup == group {
			entries = append(entries, e)
		}
	}
	return entries
}
