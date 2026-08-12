package domain

import (
	"crypto/rand"
	"fmt"
)

// EntryID and GroupID are opaque, independently random identifiers. They are
// generated with crypto/rand rather than derived from any counter or
// timestamp, so they carry no ordering or provenance information.

type EntryID [16]byte

type GroupID [16]byte

func NewEntryID() EntryID {
	return EntryID(newID())
}

func NewGroupID() GroupID {
	return GroupID(newID())
}

func newID() [16]byte {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		// crypto/rand failing indicates the OS entropy source is broken.
		// There is no safe fallback (spec section 10.1 forbids math/rand),
		// so there is nothing left to do but stop.
		panic("domain: crypto/rand unavailable: " + err.Error())
	}
	return id
}

func (id EntryID) IsZero() bool { return id == EntryID{} }

func (id GroupID) IsZero() bool { return id == GroupID{} }

func (id EntryID) String() string { return fmt.Sprintf("%x", [16]byte(id)) }

func (id GroupID) String() string { return fmt.Sprintf("%x", [16]byte(id)) }
