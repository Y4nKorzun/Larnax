package tui

import "testing"

var allModes = []Mode{NORMAL, INSERT, SEARCH, COMMAND, CONFIRM, LOCKED}

func TestLockedIsReachableFromEveryMode(t *testing.T) {
	for _, m := range allModes {
		if !CanTransition(m, LOCKED) {
			t.Errorf("CanTransition(%s, LOCKED) = false, want true", m)
		}
	}
}

func TestLockedOnlyLeadsToNormalOrStaysLocked(t *testing.T) {
	for _, next := range allModes {
		got := CanTransition(LOCKED, next)
		want := next == NORMAL || next == LOCKED
		if got != want {
			t.Errorf("CanTransition(LOCKED, %s) = %v, want %v", next, got, want)
		}
	}
}

func TestInsertOnlyLeadsToNormalOrLocked(t *testing.T) {
	for _, next := range allModes {
		got := CanTransition(INSERT, next)
		want := next == NORMAL || next == LOCKED
		if got != want {
			t.Errorf("CanTransition(INSERT, %s) = %v, want %v", next, got, want)
		}
	}
}

func TestNormalReachesEveryMode(t *testing.T) {
	for _, next := range allModes {
		if !CanTransition(NORMAL, next) {
			t.Errorf("CanTransition(NORMAL, %s) = false, want true", next)
		}
	}
}

func TestCommandCanLockDirectly(t *testing.T) {
	if !CanTransition(COMMAND, LOCKED) {
		t.Error("CanTransition(COMMAND, LOCKED) = false, want true (:lock is a command)")
	}
}

func TestSearchCannotReachCommandDirectly(t *testing.T) {
	if CanTransition(SEARCH, COMMAND) {
		t.Error("CanTransition(SEARCH, COMMAND) = true, want false")
	}
}

func TestConfirmCannotReachInsertDirectly(t *testing.T) {
	if CanTransition(CONFIRM, INSERT) {
		t.Error("CanTransition(CONFIRM, INSERT) = true, want false")
	}
}

func TestModeStringKnownValues(t *testing.T) {
	cases := map[Mode]string{
		NORMAL:  "NORMAL",
		INSERT:  "INSERT",
		SEARCH:  "SEARCH",
		COMMAND: "COMMAND",
		CONFIRM: "CONFIRM",
		LOCKED:  "LOCKED",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", m, got, want)
		}
	}
}

func TestModeStringUnknownValue(t *testing.T) {
	if got := Mode(99).String(); got != "UNKNOWN" {
		t.Errorf("Mode(99).String() = %q, want %q", got, "UNKNOWN")
	}
}
