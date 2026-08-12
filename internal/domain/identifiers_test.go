package domain

import "testing"

func TestNewEntryIDIsNonZeroAndUnique(t *testing.T) {
	a := NewEntryID()
	b := NewEntryID()

	if a.IsZero() {
		t.Fatal("NewEntryID() returned the zero value")
	}
	if a == b {
		t.Fatal("two calls to NewEntryID() returned the same value")
	}
}

func TestNewGroupIDIsNonZeroAndUnique(t *testing.T) {
	a := NewGroupID()
	b := NewGroupID()

	if a.IsZero() {
		t.Fatal("NewGroupID() returned the zero value")
	}
	if a == b {
		t.Fatal("two calls to NewGroupID() returned the same value")
	}
}

func TestZeroIDIsZero(t *testing.T) {
	var e EntryID
	var g GroupID
	if !e.IsZero() {
		t.Error("zero-value EntryID.IsZero() = false, want true")
	}
	if !g.IsZero() {
		t.Error("zero-value GroupID.IsZero() = false, want true")
	}
}

func TestIDStringIsHex(t *testing.T) {
	id := NewEntryID()
	s := id.String()
	if len(s) != 32 { // 16 bytes -> 32 hex characters
		t.Errorf("String() length = %d, want 32 (%q)", len(s), s)
	}
}
