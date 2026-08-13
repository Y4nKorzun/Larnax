package domain

import (
	"errors"
	"testing"
)

func TestNewVaultHasRootGroup(t *testing.T) {
	v := NewVault("test vault")

	root, err := v.Group(v.RootGroupID())
	if err != nil {
		t.Fatalf("Group(RootGroupID()) error = %v", err)
	}
	if root.Name != "test vault" {
		t.Errorf("root.Name = %q, want %q", root.Name, "test vault")
	}
	if root.ParentID != nil {
		t.Errorf("root.ParentID = %v, want nil", root.ParentID)
	}
}

func TestAddGroupRequiresExistingParent(t *testing.T) {
	v := NewVault("test vault")
	group := NewGroup(NewGroupID(), "Personal") // parent not added to v

	if err := v.AddGroup(group); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("AddGroup() with unknown parent error = %v, want %v", err, ErrGroupNotFound)
	}
}

func TestAddGroupRejectsNilParent(t *testing.T) {
	v := NewVault("test vault")
	group := Group{ID: NewGroupID(), Name: "second root?"}

	if err := v.AddGroup(group); !errors.Is(err, ErrGroupMustHaveParent) {
		t.Errorf("AddGroup() with nil parent error = %v, want %v", err, ErrGroupMustHaveParent)
	}
}

func TestAddAndLookupGroup(t *testing.T) {
	v := NewVault("test vault")
	group := NewGroup(v.RootGroupID(), "Personal")

	if err := v.AddGroup(group); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}

	got, err := v.Group(group.ID)
	if err != nil {
		t.Fatalf("Group() error = %v", err)
	}
	if got.Name != "Personal" {
		t.Errorf("Name = %q, want %q", got.Name, "Personal")
	}

	children := v.ChildGroups(v.RootGroupID())
	if len(children) != 1 || children[0].ID != group.ID {
		t.Errorf("ChildGroups(root) = %+v, want [%+v]", children, group)
	}
}

func TestAddEntryRequiresExistingGroup(t *testing.T) {
	v := NewVault("test vault")
	entry := NewEntry(NewGroupID(), "GitHub") // group not added to v

	if err := v.AddEntry(entry); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("AddEntry() with unknown group error = %v, want %v", err, ErrGroupNotFound)
	}
}

func TestAddLookupAndRemoveEntry(t *testing.T) {
	v := NewVault("test vault")
	entry := NewEntry(v.RootGroupID(), "GitHub")

	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	got, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub")
	}

	inRoot := v.EntriesIn(v.RootGroupID())
	if len(inRoot) != 1 || inRoot[0].ID != entry.ID {
		t.Errorf("EntriesIn(root) = %+v, want [%+v]", inRoot, entry)
	}

	if err := v.RemoveEntry(entry.ID); err != nil {
		t.Fatalf("RemoveEntry() error = %v", err)
	}
	if _, err := v.Entry(entry.ID); !errors.Is(err, ErrEntryNotFound) {
		t.Errorf("Entry() after removal error = %v, want %v", err, ErrEntryNotFound)
	}
}

func TestUpdateEntryReplacesStoredValue(t *testing.T) {
	v := NewVault("test vault")
	entry := NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	updated := entry
	updated.Title = "GitHub (renamed)"
	if err := v.UpdateEntry(updated); err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}

	got, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if got.Title != "GitHub (renamed)" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub (renamed)")
	}
}

func TestUpdateEntryRequiresExistingEntry(t *testing.T) {
	v := NewVault("test vault")
	unknown := NewEntry(v.RootGroupID(), "GitHub")

	if err := v.UpdateEntry(unknown); !errors.Is(err, ErrEntryNotFound) {
		t.Errorf("UpdateEntry() error = %v, want %v", err, ErrEntryNotFound)
	}
}

func TestUpdateEntryRequiresExistingGroup(t *testing.T) {
	v := NewVault("test vault")
	entry := NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	moved := entry
	moved.ParentGroup = NewGroupID() // a group that doesn't exist in v
	if err := v.UpdateEntry(moved); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("UpdateEntry() error = %v, want %v", err, ErrGroupNotFound)
	}
}

func TestRemoveUnknownEntry(t *testing.T) {
	v := NewVault("test vault")
	if err := v.RemoveEntry(NewEntryID()); !errors.Is(err, ErrEntryNotFound) {
		t.Errorf("RemoveEntry() error = %v, want %v", err, ErrEntryNotFound)
	}
}

func TestMoveEntryBetweenGroups(t *testing.T) {
	v := NewVault("test vault")
	work := NewGroup(v.RootGroupID(), "Work")
	if err := v.AddGroup(work); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}
	entry := NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	if err := v.MoveEntry(entry.ID, work.ID); err != nil {
		t.Fatalf("MoveEntry() error = %v", err)
	}

	moved, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if moved.ParentGroup != work.ID {
		t.Errorf("ParentGroup = %v, want %v", moved.ParentGroup, work.ID)
	}
	if len(v.EntriesIn(v.RootGroupID())) != 0 {
		t.Error("EntriesIn(root) still has entries after move")
	}
	if len(v.EntriesIn(work.ID)) != 1 {
		t.Errorf("EntriesIn(work) = %d entries, want 1", len(v.EntriesIn(work.ID)))
	}
}

func TestMoveEntryToUnknownGroupFails(t *testing.T) {
	v := NewVault("test vault")
	entry := NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	if err := v.MoveEntry(entry.ID, NewGroupID()); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("MoveEntry() to unknown group error = %v, want %v", err, ErrGroupNotFound)
	}
}

func TestMoveRootGroupFails(t *testing.T) {
	v := NewVault("test vault")
	work := NewGroup(v.RootGroupID(), "Work")
	if err := v.AddGroup(work); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}

	if err := v.MoveGroup(v.RootGroupID(), work.ID); !errors.Is(err, ErrCannotMoveRootGroup) {
		t.Errorf("MoveGroup(root, ...) error = %v, want %v", err, ErrCannotMoveRootGroup)
	}
}

func TestMoveGroupUnderItselfFails(t *testing.T) {
	v := NewVault("test vault")
	work := NewGroup(v.RootGroupID(), "Work")
	if err := v.AddGroup(work); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}

	if err := v.MoveGroup(work.ID, work.ID); !errors.Is(err, ErrCyclicGroupMove) {
		t.Errorf("MoveGroup(work, work) error = %v, want %v", err, ErrCyclicGroupMove)
	}
}

func TestMoveGroupUnderOwnDescendantFails(t *testing.T) {
	v := NewVault("test vault")
	parent := NewGroup(v.RootGroupID(), "Parent")
	if err := v.AddGroup(parent); err != nil {
		t.Fatalf("AddGroup(parent) error = %v", err)
	}
	child := NewGroup(parent.ID, "Child")
	if err := v.AddGroup(child); err != nil {
		t.Fatalf("AddGroup(child) error = %v", err)
	}

	if err := v.MoveGroup(parent.ID, child.ID); !errors.Is(err, ErrCyclicGroupMove) {
		t.Errorf("MoveGroup(parent, child) error = %v, want %v", err, ErrCyclicGroupMove)
	}

	// The tree must be unchanged after a rejected move.
	got, err := v.Group(parent.ID)
	if err != nil {
		t.Fatalf("Group(parent) error = %v", err)
	}
	if *got.ParentID != v.RootGroupID() {
		t.Errorf("parent.ParentID = %v, want %v (unchanged)", *got.ParentID, v.RootGroupID())
	}
}

func TestMoveGroupToUnknownParentFails(t *testing.T) {
	v := NewVault("test vault")
	work := NewGroup(v.RootGroupID(), "Work")
	if err := v.AddGroup(work); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}

	if err := v.MoveGroup(work.ID, NewGroupID()); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("MoveGroup() to unknown parent error = %v, want %v", err, ErrGroupNotFound)
	}
}

func TestMoveGroupSucceeds(t *testing.T) {
	v := NewVault("test vault")
	workA := NewGroup(v.RootGroupID(), "Work A")
	workB := NewGroup(v.RootGroupID(), "Work B")
	if err := v.AddGroup(workA); err != nil {
		t.Fatalf("AddGroup(workA) error = %v", err)
	}
	if err := v.AddGroup(workB); err != nil {
		t.Fatalf("AddGroup(workB) error = %v", err)
	}
	child := NewGroup(workA.ID, "Child")
	if err := v.AddGroup(child); err != nil {
		t.Fatalf("AddGroup(child) error = %v", err)
	}

	if err := v.MoveGroup(child.ID, workB.ID); err != nil {
		t.Fatalf("MoveGroup() error = %v", err)
	}

	got, err := v.Group(child.ID)
	if err != nil {
		t.Fatalf("Group(child) error = %v", err)
	}
	if *got.ParentID != workB.ID {
		t.Errorf("*ParentID = %v, want %v", *got.ParentID, workB.ID)
	}
	if len(v.ChildGroups(workA.ID)) != 0 {
		t.Error("ChildGroups(workA) still has children after move")
	}
	if len(v.ChildGroups(workB.ID)) != 1 {
		t.Errorf("ChildGroups(workB) = %d, want 1", len(v.ChildGroups(workB.ID)))
	}
}
