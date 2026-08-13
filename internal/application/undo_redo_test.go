package application

import (
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func readSecret(t *testing.T, s domain.Secret) ([]byte, error) {
	t.Helper()
	var got []byte
	err := s.Reveal(func(value []byte) error {
		got = append(got, value...)
		return nil
	})
	return got, err
}

func TestAddEntryCommandApplyAndUndo(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	cmd := NewAddEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if _, err := v.Entry(entry.ID); err != nil {
		t.Fatalf("Entry() after Apply error = %v", err)
	}

	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if _, err := v.Entry(entry.ID); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Errorf("Entry() after Undo error = %v, want %v", err, domain.ErrEntryNotFound)
	}
}

func TestDeleteEntryCommandApplyAndUndo(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	cmd := NewDeleteEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if _, err := v.Entry(entry.ID); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Errorf("Entry() after Apply error = %v, want %v", err, domain.ErrEntryNotFound)
	}

	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	got, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() after Undo error = %v", err)
	}
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub")
	}
}

func TestUpdateEntryCommandApplyAndUndo(t *testing.T) {
	v := domain.NewVault("test")
	before := domain.NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(before); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	after := before
	after.Title = "GitHub Renamed"
	cmd := NewUpdateEntryCommand(before, after)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	got, _ := v.Entry(before.ID)
	if got.Title != "GitHub Renamed" {
		t.Errorf("Title after Apply = %q, want %q", got.Title, "GitHub Renamed")
	}

	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	got, _ = v.Entry(before.ID)
	if got.Title != "GitHub" {
		t.Errorf("Title after Undo = %q, want %q", got.Title, "GitHub")
	}
}

func TestMoveEntryCommandApplyAndUndo(t *testing.T) {
	v := domain.NewVault("test")
	work := domain.NewGroup(v.RootGroupID(), "Work")
	if err := v.AddGroup(work); err != nil {
		t.Fatalf("AddGroup() error = %v", err)
	}
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	cmd := NewMoveEntryCommand(entry.ID, v.RootGroupID(), work.ID)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	got, _ := v.Entry(entry.ID)
	if got.ParentGroup != work.ID {
		t.Errorf("ParentGroup after Apply = %v, want %v", got.ParentGroup, work.ID)
	}

	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	got, _ = v.Entry(entry.ID)
	if got.ParentGroup != v.RootGroupID() {
		t.Errorf("ParentGroup after Undo = %v, want %v", got.ParentGroup, v.RootGroupID())
	}
}

func TestCommandStackDoUndoRedoCycle(t *testing.T) {
	v := domain.NewVault("test")
	entryA := domain.NewEntry(v.RootGroupID(), "A")
	entryB := domain.NewEntry(v.RootGroupID(), "B")
	stack := &CommandStack{}

	if err := stack.Do(v, NewAddEntryCommand(entryA)); err != nil {
		t.Fatalf("Do(A) error = %v", err)
	}
	if err := stack.Do(v, NewAddEntryCommand(entryB)); err != nil {
		t.Fatalf("Do(B) error = %v", err)
	}
	if len(v.EntriesIn(v.RootGroupID())) != 2 {
		t.Fatalf("entries after two Do() = %d, want 2", len(v.EntriesIn(v.RootGroupID())))
	}

	if ok, err := stack.Undo(v); !ok || err != nil {
		t.Fatalf("Undo() = %v, %v, want true, nil", ok, err)
	}
	if len(v.EntriesIn(v.RootGroupID())) != 1 {
		t.Fatalf("entries after first Undo() = %d, want 1", len(v.EntriesIn(v.RootGroupID())))
	}

	if ok, err := stack.Undo(v); !ok || err != nil {
		t.Fatalf("Undo() = %v, %v, want true, nil", ok, err)
	}
	if len(v.EntriesIn(v.RootGroupID())) != 0 {
		t.Fatalf("entries after second Undo() = %d, want 0", len(v.EntriesIn(v.RootGroupID())))
	}

	if ok, err := stack.Redo(v); !ok || err != nil {
		t.Fatalf("Redo() = %v, %v, want true, nil", ok, err)
	}
	if _, err := v.Entry(entryA.ID); err != nil {
		t.Errorf("Entry(A) after Redo() error = %v", err)
	}

	if ok, err := stack.Redo(v); !ok || err != nil {
		t.Fatalf("Redo() = %v, %v, want true, nil", ok, err)
	}
	if _, err := v.Entry(entryB.ID); err != nil {
		t.Errorf("Entry(B) after Redo() error = %v", err)
	}
}

func TestCommandStackUndoOnEmptyStackReturnsFalse(t *testing.T) {
	v := domain.NewVault("test")
	stack := &CommandStack{}
	ok, err := stack.Undo(v)
	if err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if ok {
		t.Error("Undo() = true, want false (nothing to undo)")
	}
}

func TestCommandStackRedoOnEmptyStackReturnsFalse(t *testing.T) {
	v := domain.NewVault("test")
	stack := &CommandStack{}
	ok, err := stack.Redo(v)
	if err != nil {
		t.Fatalf("Redo() error = %v", err)
	}
	if ok {
		t.Error("Redo() = true, want false (nothing to redo)")
	}
}

func TestCommandStackDoClearsRedoHistory(t *testing.T) {
	v := domain.NewVault("test")
	entryA := domain.NewEntry(v.RootGroupID(), "A")
	entryB := domain.NewEntry(v.RootGroupID(), "B")
	stack := &CommandStack{}

	if err := stack.Do(v, NewAddEntryCommand(entryA)); err != nil {
		t.Fatalf("Do(A) error = %v", err)
	}
	if _, err := stack.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if len(stack.redo) != 1 {
		t.Fatalf("len(redo) = %d, want 1 before the new Do()", len(stack.redo))
	}

	if err := stack.Do(v, NewAddEntryCommand(entryB)); err != nil {
		t.Fatalf("Do(B) error = %v", err)
	}
	if len(stack.redo) != 0 {
		t.Errorf("len(redo) = %d, want 0 (a new edit must discard old redo history)", len(stack.redo))
	}

	// entryA must not have come back: its AddEntryCommand was discarded,
	// not redone.
	if _, err := v.Entry(entryA.ID); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Errorf("Entry(A) error = %v, want %v (A's redo was discarded, not replayed)", err, domain.ErrEntryNotFound)
	}
}

// TestAddEntryCommandClearWhileAppliedDoesNotCorruptLiveSecret is the
// critical safety property: an AddEntryCommand still sitting on the undo
// stack (applied) must not have Clear() touch the Secret the live vault
// entry is actively using.
func TestAddEntryCommandClearWhileAppliedDoesNotCorruptLiveSecret(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecret([]byte("hunter2"))
	cmd := NewAddEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	cmd.Clear() // must be a no-op while applied

	live, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	got, err := readSecret(t, live.Password)
	if err != nil {
		t.Fatalf("Reveal() after Clear() while applied error = %v, want nil (live secret must survive)", err)
	}
	if string(got) != "hunter2" {
		t.Errorf("password = %q, want %q", got, "hunter2")
	}
}

func TestAddEntryCommandClearAfterUndoClearsSecret(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecret([]byte("hunter2"))
	cmd := NewAddEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}

	cmd.Clear() // now safe: nothing in the vault references this entry

	if _, err := readSecret(t, entry.Password); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() after Clear() error = %v, want %v", err, domain.ErrSecretCleared)
	}
}

// TestDeleteEntryCommandClearAfterUndoDoesNotCorruptLiveSecret mirrors the
// Add case for the opposite command: after Undo restores the entry to the
// vault, Clear() must leave its secret alone.
func TestDeleteEntryCommandClearAfterUndoDoesNotCorruptLiveSecret(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecret([]byte("hunter2"))
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	cmd := NewDeleteEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if err := cmd.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}

	cmd.Clear() // must be a no-op: the entry is live again after Undo

	live, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	got, err := readSecret(t, live.Password)
	if err != nil {
		t.Fatalf("Reveal() after Clear() error = %v, want nil (live secret must survive)", err)
	}
	if string(got) != "hunter2" {
		t.Errorf("password = %q, want %q", got, "hunter2")
	}
}

func TestDeleteEntryCommandClearWhileAppliedClearsSecret(t *testing.T) {
	v := domain.NewVault("test")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecret([]byte("hunter2"))
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	cmd := NewDeleteEntryCommand(entry)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	cmd.Clear() // safe: the vault no longer holds this entry at all

	if _, err := readSecret(t, entry.Password); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() after Clear() error = %v, want %v", err, domain.ErrSecretCleared)
	}
}

// TestUpdateEntryCommandClearRespectsWhicheverSideIsLive is the same
// safety property for the trickiest command: it holds two Entry values,
// and only one of them is ever live in the vault at a time.
func TestUpdateEntryCommandClearRespectsWhicheverSideIsLive(t *testing.T) {
	v := domain.NewVault("test")
	before := domain.NewEntry(v.RootGroupID(), "GitHub")
	before.Password = domain.NewSecret([]byte("old-password"))
	if err := v.AddEntry(before); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	after := before
	after.Password = domain.NewSecret([]byte("new-password"))
	cmd := NewUpdateEntryCommand(before, after)

	if err := cmd.Apply(v); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	cmd.Clear() // "after" is live now; only "before" should be cleared

	live, _ := v.Entry(before.ID)
	got, err := readSecret(t, live.Password)
	if err != nil {
		t.Fatalf("Reveal() on live (after-Apply) secret error = %v, want nil", err)
	}
	if string(got) != "new-password" {
		t.Errorf("live password = %q, want %q", got, "new-password")
	}
	if _, err := readSecret(t, before.Password); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() on inactive \"before\" secret error = %v, want %v", err, domain.ErrSecretCleared)
	}
}

func TestCommandStackClearEmptiesBothStacks(t *testing.T) {
	v := domain.NewVault("test")
	entryA := domain.NewEntry(v.RootGroupID(), "A")
	entryB := domain.NewEntry(v.RootGroupID(), "B")
	stack := &CommandStack{}

	if err := stack.Do(v, NewAddEntryCommand(entryA)); err != nil {
		t.Fatalf("Do(A) error = %v", err)
	}
	if err := stack.Do(v, NewAddEntryCommand(entryB)); err != nil {
		t.Fatalf("Do(B) error = %v", err)
	}
	if _, err := stack.Undo(v); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}

	stack.Clear()

	if len(stack.undo) != 0 || len(stack.redo) != 0 {
		t.Errorf("after Clear(): len(undo)=%d len(redo)=%d, want 0, 0", len(stack.undo), len(stack.redo))
	}
}
