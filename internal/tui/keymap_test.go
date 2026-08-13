package tui

import (
	"errors"
	"testing"
)

func TestDefaultKeymapCoversSpecTable(t *testing.T) {
	km := DefaultKeymap()

	cases := map[string]Action{
		"j":         ActionDown,
		"gg":        ActionFirstItem,
		"G":         ActionLastItem,
		"Ctrl+d":    ActionHalfPageDown,
		"/":         ActionSearchStart,
		"Esc":       ActionCancel,
		"dd":        ActionDelete,
		"gp":        ActionOpenGenerator,
		"yp":        ActionCopyPassword,
		"<Leader>l": ActionLockVault,
		":":         ActionCommandMode,
		"q":         ActionQuit,
	}
	for key, want := range cases {
		got, ok := km[key]
		if !ok {
			t.Errorf("DefaultKeymap()[%q] missing", key)
			continue
		}
		if got != want {
			t.Errorf("DefaultKeymap()[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestDefaultKeymapOpenSelectedHasTwoKeys(t *testing.T) {
	km := DefaultKeymap()
	if km["l"] != ActionOpenSelected || km["Enter"] != ActionOpenSelected {
		t.Errorf(`"l" and "Enter" = %q, %q, want both %q`, km["l"], km["Enter"], ActionOpenSelected)
	}
}

func TestDefaultKeymapHasEmergencyBindings(t *testing.T) {
	km := DefaultKeymap()
	if !km.hasBinding(ActionQuit) {
		t.Error("DefaultKeymap() has no binding for ActionQuit")
	}
	if !km.hasBinding(ActionLockVault) {
		t.Error("DefaultKeymap() has no binding for ActionLockVault")
	}
}

func TestApplyOverridesLetsLaterEntryWin(t *testing.T) {
	base := DefaultKeymap()
	overrides := map[string]Action{"j": ActionSearchStart}

	merged, err := ApplyOverrides(base, overrides)
	if err != nil {
		t.Fatalf("ApplyOverrides() error = %v", err)
	}
	if merged["j"] != ActionSearchStart {
		t.Errorf(`merged["j"] = %q, want %q`, merged["j"], ActionSearchStart)
	}
	if merged["k"] != ActionUp {
		t.Errorf(`merged["k"] = %q, want unchanged %q`, merged["k"], ActionUp)
	}
}

func TestApplyOverridesDoesNotMutateBase(t *testing.T) {
	base := DefaultKeymap()
	if _, err := ApplyOverrides(base, map[string]Action{"j": ActionSearchStart}); err != nil {
		t.Fatalf("ApplyOverrides() error = %v", err)
	}
	if base["j"] != ActionDown {
		t.Errorf(`base["j"] = %q after ApplyOverrides(), want unchanged %q`, base["j"], ActionDown)
	}
}

func TestApplyOverridesRejectsLosingAllQuitBindings(t *testing.T) {
	base := DefaultKeymap()
	// "q" is the only default binding for ActionQuit.
	overrides := map[string]Action{"q": ActionDown}

	if _, err := ApplyOverrides(base, overrides); !errors.Is(err, ErrNoQuitBinding) {
		t.Errorf("ApplyOverrides() error = %v, want %v", err, ErrNoQuitBinding)
	}
}

func TestApplyOverridesRejectsLosingAllLockBindings(t *testing.T) {
	base := DefaultKeymap()
	// "<Leader>l" is the only default binding for ActionLockVault.
	overrides := map[string]Action{"<Leader>l": ActionSave}

	if _, err := ApplyOverrides(base, overrides); !errors.Is(err, ErrNoLockBinding) {
		t.Errorf("ApplyOverrides() error = %v, want %v", err, ErrNoLockBinding)
	}
}

func TestApplyOverridesAllowsMovingQuitToADifferentKey(t *testing.T) {
	base := DefaultKeymap()
	overrides := map[string]Action{
		"q":      ActionDown,
		"Ctrl+c": ActionQuit,
	}

	merged, err := ApplyOverrides(base, overrides)
	if err != nil {
		t.Fatalf("ApplyOverrides() error = %v, want nil (quit is still bound, just to a different key)", err)
	}
	if merged["Ctrl+c"] != ActionQuit {
		t.Errorf(`merged["Ctrl+c"] = %q, want %q`, merged["Ctrl+c"], ActionQuit)
	}
}

func TestResolveKeymapConvertsActionNames(t *testing.T) {
	base := DefaultKeymap()
	raw := map[string]string{"j": "search-start"}

	merged, err := ResolveKeymap(base, raw)
	if err != nil {
		t.Fatalf("ResolveKeymap() error = %v", err)
	}
	if merged["j"] != ActionSearchStart {
		t.Errorf(`merged["j"] = %q, want %q`, merged["j"], ActionSearchStart)
	}
}

func TestResolveKeymapRejectsUnknownActionName(t *testing.T) {
	base := DefaultKeymap()
	raw := map[string]string{"j": "not-a-real-action"}

	if _, err := ResolveKeymap(base, raw); !errors.Is(err, ErrUnknownAction) {
		t.Errorf("ResolveKeymap() error = %v, want %v", err, ErrUnknownAction)
	}
}

func TestResolveKeymapStillEnforcesEmergencyBindings(t *testing.T) {
	base := DefaultKeymap()
	raw := map[string]string{"q": "down"} // renames the only quit binding away

	if _, err := ResolveKeymap(base, raw); !errors.Is(err, ErrNoQuitBinding) {
		t.Errorf("ResolveKeymap() error = %v, want %v", err, ErrNoQuitBinding)
	}
}

func TestResolveKeymapEmptyOverridesReturnsBaseEquivalent(t *testing.T) {
	base := DefaultKeymap()
	merged, err := ResolveKeymap(base, nil)
	if err != nil {
		t.Fatalf("ResolveKeymap() error = %v", err)
	}
	if len(merged) != len(base) {
		t.Errorf("len(merged) = %d, want %d", len(merged), len(base))
	}
	for k, v := range base {
		if merged[k] != v {
			t.Errorf("merged[%q] = %q, want %q", k, merged[k], v)
		}
	}
}
