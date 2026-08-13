package components

import (
	"testing"
	"time"
)

func TestStatusLineViewMatchesSpecExample(t *testing.T) {
	m := StatusLineModel{
		Mode:               "NORMAL",
		GroupPath:          "/Personal",
		EntryCount:         3,
		Modified:           true,
		ClipboardRemaining: 11 * time.Second,
		LockState:          LockStateUnlocked,
	}

	want := "NORMAL   /Personal   3 entries   modified   clipboard: 11s   unlocked"
	if got := m.View(); got != want {
		t.Errorf("View() = %q, want %q", got, want)
	}
}

func TestStatusLineOmitsModifiedWhenClean(t *testing.T) {
	m := StatusLineModel{Mode: "NORMAL", GroupPath: "/", EntryCount: 0}
	if got := m.View(); got != "NORMAL   /   0 entries" {
		t.Errorf("View() = %q, want no \"modified\" segment", got)
	}
}

func TestStatusLineOmitsClipboardWhenZero(t *testing.T) {
	m := StatusLineModel{Mode: "NORMAL", GroupPath: "/", EntryCount: 1, ClipboardRemaining: 0}
	if got := m.View(); got != "NORMAL   /   1 entries" {
		t.Errorf("View() = %q, want no clipboard segment", got)
	}
}

func TestStatusLineOmitsLockStateWhenEmpty(t *testing.T) {
	m := StatusLineModel{Mode: "NORMAL", GroupPath: "/", EntryCount: 1}
	if got := m.View(); got != "NORMAL   /   1 entries" {
		t.Errorf("View() = %q, want no lock-state segment", got)
	}
}

func TestStatusLineShowsReadOnlyState(t *testing.T) {
	m := StatusLineModel{Mode: "NORMAL", GroupPath: "/", EntryCount: 1, LockState: LockStateReadOnly}
	if got := m.View(); got != "NORMAL   /   1 entries   read-only" {
		t.Errorf("View() = %q, want it to end with %q", got, "read-only")
	}
}

func TestStatusLineClipboardRoundsToWholeSeconds(t *testing.T) {
	m := StatusLineModel{Mode: "NORMAL", GroupPath: "/", EntryCount: 0, ClipboardRemaining: 10500 * time.Millisecond}
	if got := m.View(); got != "NORMAL   /   0 entries   clipboard: 11s" {
		t.Errorf("View() = %q, want clipboard rounded to 11s", got)
	}
}
