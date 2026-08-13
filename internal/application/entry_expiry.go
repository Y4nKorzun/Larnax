package application

import (
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// Entry expiration is P1 scope (spec section 9.1: P0 works with
// Title/Username/Password/URL/Notes/group; P1 adds a full editor for
// tags, custom fields, history, and expiration). These two helpers exist
// ahead of that editor because domain.Entry.ExpiresAt already has to
// round-trip through KDBX (see internal/infrastructure/kdbx/mapper.go),
// so a browser screen wanting to flag an expiring entry needs no new
// domain field, just this small bit of policy over the one already there.

// IsExpired reports whether entry's ExpiresAt has passed as of now. An
// entry with no ExpiresAt (nil) never expires. The exact instant
// ExpiresAt == now is treated as not yet expired — Before is a strict
// comparison — matching how a deadline of "end of day" conventionally
// still holds through that exact moment.
func IsExpired(entry domain.Entry, now time.Time) bool {
	return entry.ExpiresAt != nil && entry.ExpiresAt.Before(now)
}

// ExpiresWithin reports whether entry expires within window of now.
// An entry that has already expired counts as within any window — this
// stays true the day after expiry, rather than flipping back to false
// once the original deadline is in the past.
func ExpiresWithin(entry domain.Entry, now time.Time, window time.Duration) bool {
	if entry.ExpiresAt == nil {
		return false
	}
	return entry.ExpiresAt.Before(now.Add(window))
}
