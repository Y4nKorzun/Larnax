package domain

// Group is a node in the vault's tree used to organize entries. ParentID is
// nil only for a vault's root group.
type Group struct {
	ID       GroupID
	ParentID *GroupID
	Name     string
	Notes    string
}

// NewGroup creates a Group with a fresh ID under parent.
func NewGroup(parent GroupID, name string) Group {
	return Group{ID: NewGroupID(), ParentID: &parent, Name: name}
}
