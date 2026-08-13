package application

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/filesystem"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

// VaultService is spec section 18.2's single application-layer object the
// TUI talks to for everything vault-related:
//
//	TUI -> Application use case -> Domain model -> KDBX adapter -> gokeepasslib
//
// so no screen ever needs to hold a *kdbx.Document, let alone reach into
// gokeepasslib directly — the forbidden shape spec 18.2 names explicitly
// ("type Model struct { Database *gokeepasslib.Database }"). Vault
// returns the one thing above that line the TUI is allowed to hold.
//
// VaultService holds exactly one open session. A second vault open at
// once (spec section 6.2's P1 "multiple vaults or fast switch") needs a
// second VaultService, not a field added to this one.
type VaultService struct {
	document *kdbx.Document
	stack    *CommandStack
	readOnly bool
	path     string
	revision filesystem.FileRevision
}

// ErrVaultNotOpen is returned by every VaultService method except
// New/Open when no vault has been created or opened yet.
var ErrVaultNotOpen = errors.New("application: no vault is open")

// ErrReadOnly is returned by any mutating method when the open session
// is read-only (spec section 15.4).
var ErrReadOnly = errors.New("application: vault is open read-only")

// ErrNoSavePath is returned by Save when the session has no path yet —
// a brand-new vault from New that hasn't been given one via SaveAs.
// Spec section 7.1's "kdbx-tui new <path>" form supplies one up front;
// the no-argument wizard flow picks one later, at first save.
var ErrNoSavePath = errors.New("application: vault has no save path yet")

// IsOpen reports whether a session is currently open.
func (s *VaultService) IsOpen() bool {
	return s.document != nil
}

// ReadOnly reports whether the open session is read-only.
func (s *VaultService) ReadOnly() bool {
	return s.readOnly
}

// Path returns the session's current save path, or "" if none is set
// yet (see ErrNoSavePath).
func (s *VaultService) Path() string {
	return s.path
}

// Vault returns the open session's domain model. It panics if no session
// is open — callers are expected to check IsOpen (or handle the error
// from whichever New/Open call was supposed to open one) first, the same
// way indexing past the end of a slice is a programmer error, not a
// runtime condition to recover from.
func (s *VaultService) Vault() *domain.Vault {
	if s.document == nil {
		panic("application: Vault() called with no vault open")
	}
	return s.document.Vault
}

// New starts a brand-new vault session (spec section 7), replacing
// whatever session, if any, was open before. Prompting the user to save
// unsaved changes in that previous session first (spec 17.2 step 1) is a
// TUI-layer decision — New does not make it.
func (s *VaultService) New(name, masterPassphrase string) error {
	doc, err := CreateVault(name, masterPassphrase)
	if err != nil {
		return err
	}
	s.reset(doc, false, "", filesystem.FileRevision{})
	return nil
}

// Open opens an existing KDBX file read from r with masterPassphrase,
// tracked at path for Save's later external-modification check (spec
// 16.3), replacing whatever session, if any, was open before. Check
// ReadOnly afterward for spec section 15.4's fallback outcome.
func (s *VaultService) Open(r io.Reader, path, masterPassphrase string) error {
	opened, err := OpenVault(r, masterPassphrase)
	if err != nil {
		return err
	}

	var rev filesystem.FileRevision
	if !opened.ReadOnly {
		rev, err = filesystem.ReadFileRevision(path)
		if err != nil {
			return err
		}
	}
	s.reset(opened.Document, opened.ReadOnly, path, rev)
	return nil
}

func (s *VaultService) reset(doc *kdbx.Document, readOnly bool, path string, rev filesystem.FileRevision) {
	s.document = doc
	s.stack = &CommandStack{}
	s.readOnly = readOnly
	s.path = path
	s.revision = rev
}

// Save writes the session to Path using spec section 16's full safe-save
// sequence (SaveVault), refusing outright if the session is read-only or
// has no path yet.
func (s *VaultService) Save(retention int) error {
	if s.document == nil {
		return ErrVaultNotOpen
	}
	if s.readOnly {
		return ErrReadOnly
	}
	if s.path == "" {
		return ErrNoSavePath
	}

	rev, err := SaveVault(s.path, s.document, s.revision, retention)
	if err != nil {
		return err
	}
	s.revision = rev
	return nil
}

// SaveAs sets the session's path (spec section 8.4's `:save-as <path>`,
// or a brand-new vault's first save) and saves to it, treating path as
// having no prior known revision to conflict with.
func (s *VaultService) SaveAs(path string, retention int) error {
	s.path = path
	s.revision = filesystem.FileRevision{}
	return s.Save(retention)
}

// Lock clears the session's in-memory secrets and undo history in place
// (spec section 17.2 steps 3-4) without discarding the session —
// Vault()'s structure (groups, entry IDs, non-secret fields) stays
// addressable, so the TUI can still show spec 17.2 step 7's "path and
// non-secret navigation state" while locked.
func (s *VaultService) Lock() error {
	if s.document == nil {
		return ErrVaultNotOpen
	}
	LockVault(s.document.Vault, s.stack)
	return nil
}

// AddEntry adds entry to the session via its undo stack.
func (s *VaultService) AddEntry(entry domain.Entry) error {
	if err := s.requireWritable(); err != nil {
		return err
	}
	return s.stack.Do(s.document.Vault, NewAddEntryCommand(entry))
}

// UpdateEntry replaces an entry's fields via the session's undo stack.
func (s *VaultService) UpdateEntry(before, after domain.Entry) error {
	if err := s.requireWritable(); err != nil {
		return err
	}
	return s.stack.Do(s.document.Vault, NewUpdateEntryCommand(before, after))
}

// DeleteEntry removes entry via the session's undo stack.
func (s *VaultService) DeleteEntry(entry domain.Entry) error {
	if err := s.requireWritable(); err != nil {
		return err
	}
	return s.stack.Do(s.document.Vault, NewDeleteEntryCommand(entry))
}

// Undo reverses the most recently applied change, if any.
func (s *VaultService) Undo() (bool, error) {
	if s.document == nil {
		return false, ErrVaultNotOpen
	}
	return s.stack.Undo(s.document.Vault)
}

// Redo reapplies the most recently undone change, if any.
func (s *VaultService) Redo() (bool, error) {
	if s.document == nil {
		return false, ErrVaultNotOpen
	}
	return s.stack.Redo(s.document.Vault)
}

// CopyField copies one of entry's fields to the clipboard.
func (s *VaultService) CopyField(ctx context.Context, cb clipboard.Clipboard, entry domain.Entry, field FieldName) ([sha256.Size]byte, error) {
	if s.document == nil {
		return [sha256.Size]byte{}, ErrVaultNotOpen
	}
	return CopyField(ctx, cb, entry, field)
}

// Import applies plan under parent via the session's undo stack (spec
// section 13.7).
func (s *VaultService) Import(parent domain.GroupID, plan ImportPlan) (ImportResult, error) {
	if err := s.requireWritable(); err != nil {
		return ImportResult{}, err
	}
	return ImportEntries(s.document.Vault, s.stack, parent, plan)
}

func (s *VaultService) requireWritable() error {
	if s.document == nil {
		return ErrVaultNotOpen
	}
	if s.readOnly {
		return ErrReadOnly
	}
	return nil
}
