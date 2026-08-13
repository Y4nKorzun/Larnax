package application

import (
	"testing"
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func entryExpiringAt(t *testing.T, at time.Time) domain.Entry {
	t.Helper()
	e := domain.NewEntry(domain.NewGroupID(), "test")
	e.ExpiresAt = &at
	return e
}

func TestIsExpiredNilNeverExpires(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "test")
	if IsExpired(e, time.Now()) {
		t.Error("IsExpired() = true for an entry with no ExpiresAt")
	}
}

func TestIsExpiredPastDateIsExpired(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now.Add(-time.Hour))
	if !IsExpired(e, now) {
		t.Error("IsExpired() = false for an entry that expired an hour ago")
	}
}

func TestIsExpiredFutureDateIsNotExpired(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now.Add(time.Hour))
	if IsExpired(e, now) {
		t.Error("IsExpired() = true for an entry expiring an hour from now")
	}
}

func TestIsExpiredExactlyNowIsNotExpired(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now)
	if IsExpired(e, now) {
		t.Error("IsExpired() = true at the exact expiry instant, want false (strict Before)")
	}
}

func TestExpiresWithinNilNeverTrue(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "test")
	if ExpiresWithin(e, time.Now(), 30*24*time.Hour) {
		t.Error("ExpiresWithin() = true for an entry with no ExpiresAt")
	}
}

func TestExpiresWithinFutureInsideWindow(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now.Add(3*24*time.Hour))
	if !ExpiresWithin(e, now, 7*24*time.Hour) {
		t.Error("ExpiresWithin() = false for an entry expiring in 3 days within a 7-day window")
	}
}

func TestExpiresWithinFutureOutsideWindow(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now.Add(30*24*time.Hour))
	if ExpiresWithin(e, now, 7*24*time.Hour) {
		t.Error("ExpiresWithin() = true for an entry expiring in 30 days within a 7-day window")
	}
}

func TestExpiresWithinAlreadyExpiredStillTrue(t *testing.T) {
	now := time.Now()
	e := entryExpiringAt(t, now.Add(-30*24*time.Hour))
	if !ExpiresWithin(e, now, 7*24*time.Hour) {
		t.Error("ExpiresWithin() = false for an already-expired entry, want true")
	}
}
