package domain

import "testing"

func TestNewGroupDefaults(t *testing.T) {
	parent := NewGroupID()
	g := NewGroup(parent, "Personal")

	if g.ID.IsZero() {
		t.Error("NewGroup().ID is zero")
	}
	if g.ParentID == nil {
		t.Fatal("ParentID is nil, want a pointer to parent")
	}
	if *g.ParentID != parent {
		t.Errorf("*ParentID = %v, want %v", *g.ParentID, parent)
	}
	if g.Name != "Personal" {
		t.Errorf("Name = %q, want %q", g.Name, "Personal")
	}
}
