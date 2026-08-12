package tui

import "testing"

func TestKeySequenceNonPrefixKeyPassesThrough(t *testing.T) {
	s := &KeySequence{}
	action, resolved := s.Feed('j')
	if resolved {
		t.Errorf("Feed('j') resolved = true, want false")
	}
	if action != "" {
		t.Errorf("Feed('j') action = %q, want empty", action)
	}
	if s.Pending() {
		t.Error("Pending() = true after a non-prefix key, want false")
	}
}

func TestKeySequenceCompletesTwoKeySequences(t *testing.T) {
	cases := []struct {
		name       string
		first      rune
		second     rune
		wantAction Action
	}{
		{"gg", 'g', 'g', ActionFirstItem},
		{"dd", 'd', 'd', ActionDelete},
		{"yu", 'y', 'u', ActionCopyUsername},
		{"yp", 'y', 'p', ActionCopyPassword},
		{"yU", 'y', 'U', ActionCopyURL},
		{"yn", 'y', 'n', ActionCopyNotes},
		{"yt", 'y', 't', ActionCopyTOTP},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &KeySequence{}

			action, resolved := s.Feed(c.first)
			if resolved {
				t.Fatalf("Feed(%q) resolved = true, want false (still pending)", c.first)
			}
			if !s.Pending() {
				t.Fatalf("Pending() = false after Feed(%q), want true", c.first)
			}

			action, resolved = s.Feed(c.second)
			if !resolved {
				t.Fatalf("Feed(%q) resolved = false, want true", c.second)
			}
			if action != c.wantAction {
				t.Errorf("action = %q, want %q", action, c.wantAction)
			}
			if s.Pending() {
				t.Error("Pending() = true after a completed sequence, want false")
			}
		})
	}
}

func TestKeySequenceAbortsOnUnrecognizedCompletion(t *testing.T) {
	s := &KeySequence{}
	s.Feed('g') // pending

	action, resolved := s.Feed('x') // "gx" is not a sequence, "x" is not a prefix
	if resolved {
		t.Errorf("Feed('x') resolved = true, want false")
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
	if s.Pending() {
		t.Error("Pending() = true after an aborted sequence, want false")
	}
}

func TestKeySequenceAbortedCompletionThatIsAPrefixStartsNewSequence(t *testing.T) {
	s := &KeySequence{}

	if _, resolved := s.Feed('d'); resolved {
		t.Fatal("Feed('d') resolved = true, want false")
	}

	// "dy" is not a sequence, but 'y' is itself a valid prefix, so it
	// should start a fresh pending sequence rather than being dropped.
	action, resolved := s.Feed('y')
	if resolved {
		t.Fatalf("Feed('y') resolved = true, want false")
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
	if !s.Pending() {
		t.Fatal("Pending() = false after 'dy', want true (fresh 'y' prefix)")
	}

	action, resolved = s.Feed('p')
	if !resolved {
		t.Fatal("Feed('p') resolved = false, want true")
	}
	if action != ActionCopyPassword {
		t.Errorf("action = %q, want %q", action, ActionCopyPassword)
	}
}

func TestKeySequenceResetClearsPending(t *testing.T) {
	s := &KeySequence{}
	s.Feed('g')
	if !s.Pending() {
		t.Fatal("Pending() = false after Feed('g'), want true")
	}

	s.Reset()
	if s.Pending() {
		t.Fatal("Pending() = true after Reset(), want false")
	}

	// A lone second 'g' must not complete "gg": Reset() should have
	// discarded the first 'g' entirely, so this is a fresh prefix, not a
	// completion.
	action, resolved := s.Feed('g')
	if resolved {
		t.Error("Feed('g') after Reset() resolved = true, want false (Reset did not clear state)")
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
	if !s.Pending() {
		t.Error("Pending() = false after Feed('g'), want true")
	}
}

func TestKeySequenceIndependentInstancesDoNotShareState(t *testing.T) {
	a := &KeySequence{}
	b := &KeySequence{}

	a.Feed('g')
	if b.Pending() {
		t.Error("a second KeySequence instance observed state from another instance")
	}
}
