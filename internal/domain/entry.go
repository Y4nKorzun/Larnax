package domain

import "time"

// Entry is a single credential record, independent of how it is encoded in
// a KDBX file (spec section 18.4).
type Entry struct {
	ID           EntryID
	ParentGroup  GroupID
	Title        string
	Username     string
	Password     Secret
	URL          string
	Notes        string
	Tags         []string
	CustomFields []Field
	CreatedAt    time.Time
	ModifiedAt   time.Time
	ExpiresAt    *time.Time
}

// NewEntry creates an Entry inside parent with a fresh ID and an empty (but
// non-nil) Password, so Password.Reveal never panics on a freshly
// constructed Entry.
func NewEntry(parent GroupID, title string) Entry {
	now := time.Now()
	return Entry{
		ID:          NewEntryID(),
		ParentGroup: parent,
		Title:       title,
		Password:    NewSecret(nil),
		CreatedAt:   now,
		ModifiedAt:  now,
	}
}
