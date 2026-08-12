package domain

import "testing"

func TestNewEntryDefaults(t *testing.T) {
	parent := NewGroupID()
	e := NewEntry(parent, "GitHub")

	if e.ID.IsZero() {
		t.Error("NewEntry().ID is zero")
	}
	if e.ParentGroup != parent {
		t.Errorf("ParentGroup = %v, want %v", e.ParentGroup, parent)
	}
	if e.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", e.Title, "GitHub")
	}
	if e.Password == nil {
		t.Fatal("Password is nil; Reveal would panic")
	}
	if err := e.Password.Reveal(func(value []byte) error {
		if len(value) != 0 {
			t.Errorf("default Password value = %q, want empty", value)
		}
		return nil
	}); err != nil {
		t.Errorf("Reveal() on default Password error = %v", err)
	}
	if e.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if !e.CreatedAt.Equal(e.ModifiedAt) {
		t.Errorf("CreatedAt (%v) != ModifiedAt (%v) on a freshly created entry", e.CreatedAt, e.ModifiedAt)
	}
}
