package application

import (
	"fmt"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// Command is a single undoable change to a domain.Vault (spec section
// 20.1). Each command knows how to apply itself, how to undo itself, what
// non-secret fields are safe to show in a summary, and — per spec's own
// requirement — which secret buffers to release once the command is
// dropped from the stack for good.
type Command interface {
	Apply(v *domain.Vault) error
	Undo(v *domain.Vault) error
	Summary() string
	// Clear best-effort releases any secret buffers this command retains,
	// once it can no longer be applied or undone again. It must never
	// clear a buffer that the vault's current live entry still shares —
	// each command below tracks which side of its before/after state is
	// currently reflected in the vault so Clear only ever touches the
	// inactive copy.
	Clear()
}

func clearEntrySecret(e domain.Entry) {
	if e.Password != nil {
		e.Password.Clear()
	}
	for _, f := range e.CustomFields {
		if f.Value != nil {
			f.Value.Clear()
		}
	}
}

// AddEntryCommand adds entry to the vault.
type AddEntryCommand struct {
	entry   domain.Entry
	applied bool
}

func NewAddEntryCommand(entry domain.Entry) *AddEntryCommand {
	return &AddEntryCommand{entry: entry}
}

func (c *AddEntryCommand) Apply(v *domain.Vault) error {
	if err := v.AddEntry(c.entry); err != nil {
		return err
	}
	c.applied = true
	return nil
}

func (c *AddEntryCommand) Undo(v *domain.Vault) error {
	if err := v.RemoveEntry(c.entry.ID); err != nil {
		return err
	}
	c.applied = false
	return nil
}

func (c *AddEntryCommand) Summary() string {
	return fmt.Sprintf("add entry %q", c.entry.Title)
}

// Clear only releases the entry's secret while it is NOT applied: once
// undone, this command's copy is the only surviving reference to it.
// While applied, the same Entry value is live in the vault, and clearing
// it here would corrupt that.
func (c *AddEntryCommand) Clear() {
	if !c.applied {
		clearEntrySecret(c.entry)
	}
}

// DeleteEntryCommand removes an entry from the vault. entry must be
// captured before deletion (e.g. via domain.Vault.Entry) since Undo needs
// its full value to restore it.
type DeleteEntryCommand struct {
	entry   domain.Entry
	applied bool
}

func NewDeleteEntryCommand(entry domain.Entry) *DeleteEntryCommand {
	return &DeleteEntryCommand{entry: entry}
}

func (c *DeleteEntryCommand) Apply(v *domain.Vault) error {
	if err := v.RemoveEntry(c.entry.ID); err != nil {
		return err
	}
	c.applied = true
	return nil
}

func (c *DeleteEntryCommand) Undo(v *domain.Vault) error {
	if err := v.AddEntry(c.entry); err != nil {
		return err
	}
	c.applied = false
	return nil
}

func (c *DeleteEntryCommand) Summary() string {
	return fmt.Sprintf("delete entry %q", c.entry.Title)
}

// Clear only releases the entry's secret while applied (deleted from the
// vault): that is the only state in which this command's copy isn't also
// the live vault entry sharing the same buffer.
func (c *DeleteEntryCommand) Clear() {
	if c.applied {
		clearEntrySecret(c.entry)
	}
}

// UpdateEntryCommand replaces an entry's fields. before and after must
// share the same ID.
type UpdateEntryCommand struct {
	before  domain.Entry
	after   domain.Entry
	applied bool
}

func NewUpdateEntryCommand(before, after domain.Entry) *UpdateEntryCommand {
	return &UpdateEntryCommand{before: before, after: after}
}

func (c *UpdateEntryCommand) Apply(v *domain.Vault) error {
	if err := v.UpdateEntry(c.after); err != nil {
		return err
	}
	c.applied = true
	return nil
}

func (c *UpdateEntryCommand) Undo(v *domain.Vault) error {
	if err := v.UpdateEntry(c.before); err != nil {
		return err
	}
	c.applied = false
	return nil
}

func (c *UpdateEntryCommand) Summary() string {
	return fmt.Sprintf("update entry %q", c.after.Title)
}

// Clear releases whichever of before/after is NOT currently reflected in
// the vault. Clearing the active side would corrupt live vault data that
// shares the same Secret.
func (c *UpdateEntryCommand) Clear() {
	if c.applied {
		clearEntrySecret(c.before)
	} else {
		clearEntrySecret(c.after)
	}
}

// MoveEntryCommand reassigns an entry to a different group. It holds no
// secrets, so Clear is a no-op.
type MoveEntryCommand struct {
	entryID   domain.EntryID
	fromGroup domain.GroupID
	toGroup   domain.GroupID
}

func NewMoveEntryCommand(entryID domain.EntryID, fromGroup, toGroup domain.GroupID) *MoveEntryCommand {
	return &MoveEntryCommand{entryID: entryID, fromGroup: fromGroup, toGroup: toGroup}
}

func (c *MoveEntryCommand) Apply(v *domain.Vault) error {
	return v.MoveEntry(c.entryID, c.toGroup)
}

func (c *MoveEntryCommand) Undo(v *domain.Vault) error {
	return v.MoveEntry(c.entryID, c.fromGroup)
}

func (c *MoveEntryCommand) Summary() string {
	return fmt.Sprintf("move entry %s", c.entryID)
}

func (c *MoveEntryCommand) Clear() {}

// CommandStack is the in-memory undo/redo stack (spec section 20.1). It is
// never persisted — spec section 20.2 forbids writing an unencrypted
// change log next to the KDBX file — so the stack simply disappears when
// the process ends; there is no separate save/load for it.
type CommandStack struct {
	undo []Command
	redo []Command
}

// Do applies cmd to v and pushes it onto the undo stack. Any commands
// still on the redo stack are cleared and discarded first — the usual
// rule that a new edit invalidates old redo history, since those
// commands' before/after state no longer describes a path back to the
// vault's new state.
func (s *CommandStack) Do(v *domain.Vault, cmd Command) error {
	if err := cmd.Apply(v); err != nil {
		return err
	}
	for _, discarded := range s.redo {
		discarded.Clear()
	}
	s.redo = nil
	s.undo = append(s.undo, cmd)
	return nil
}

// Undo reverses the most recently applied command, moving it to the redo
// stack. It reports false if there is nothing to undo.
func (s *CommandStack) Undo(v *domain.Vault) (bool, error) {
	if len(s.undo) == 0 {
		return false, nil
	}
	cmd := s.undo[len(s.undo)-1]
	if err := cmd.Undo(v); err != nil {
		return false, err
	}
	s.undo = s.undo[:len(s.undo)-1]
	s.redo = append(s.redo, cmd)
	return true, nil
}

// Redo reapplies the most recently undone command, moving it back to the
// undo stack. It reports false if there is nothing to redo.
func (s *CommandStack) Redo(v *domain.Vault) (bool, error) {
	if len(s.redo) == 0 {
		return false, nil
	}
	cmd := s.redo[len(s.redo)-1]
	if err := cmd.Apply(v); err != nil {
		return false, err
	}
	s.redo = s.redo[:len(s.redo)-1]
	s.undo = append(s.undo, cmd)
	return true, nil
}

// Clear discards the entire stack, best-effort releasing every command's
// secret buffers. Call this on lock (spec section 17.2's "clear temporary
// editor buffers") or before the process exits.
func (s *CommandStack) Clear() {
	for _, cmd := range s.undo {
		cmd.Clear()
	}
	for _, cmd := range s.redo {
		cmd.Clear()
	}
	s.undo = nil
	s.redo = nil
}
