package tui

import (
	"context"
	"crypto/sha256"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
)

// Intent and *CompletedMsg types are spec section 19.2's
// KeyMsg -> Intent -> Application -> ...CompletedMsg chain made concrete:
// a screen's Update translates a raw key into an Intent, then returns one
// of the Cmd functions below instead of calling application.VaultService
// inline — the actual vault work runs as a Bubble Tea command, off the
// Update call itself, and its result comes back as a *CompletedMsg on a
// later Update call.

// CopyFieldIntent is the intent behind yu/yp/yU (spec section 8.3):
// "copy this entry's field to the clipboard."
type CopyFieldIntent struct {
	Entry domain.Entry
	Field application.FieldName
}

// CopyCompletedMsg is CopyFieldIntent's result.
type CopyCompletedMsg struct {
	Field         application.FieldName
	OwnershipHash [sha256.Size]byte
	Err           error
}

// CopyFieldCmd returns a tea.Cmd that acts on a CopyFieldIntent via
// service and cb, resolving to a CopyCompletedMsg.
func CopyFieldCmd(ctx context.Context, service *application.VaultService, cb clipboard.Clipboard, intent CopyFieldIntent) tea.Cmd {
	return func() tea.Msg {
		hash, err := service.CopyField(ctx, cb, intent.Entry, intent.Field)
		return CopyCompletedMsg{Field: intent.Field, OwnershipHash: hash, Err: err}
	}
}

// SaveIntent is <Leader>s / :w (spec sections 8.3, 8.4): "save the vault."
type SaveIntent struct {
	Retention int
}

// SaveCompletedMsg is SaveIntent's result.
type SaveCompletedMsg struct {
	Err error
}

// SaveCmd returns a tea.Cmd that acts on a SaveIntent via service,
// resolving to a SaveCompletedMsg.
func SaveCmd(service *application.VaultService, intent SaveIntent) tea.Cmd {
	return func() tea.Msg {
		return SaveCompletedMsg{Err: service.Save(intent.Retention)}
	}
}

// LockIntent is <Leader>l / :lock (spec sections 8.3, 8.4): "lock the
// vault."
type LockIntent struct{}

// LockCompletedMsg is LockIntent's result.
type LockCompletedMsg struct {
	Err error
}

// LockCmd returns a tea.Cmd that acts on a LockIntent via service,
// resolving to a LockCompletedMsg.
func LockCmd(service *application.VaultService, _ LockIntent) tea.Cmd {
	return func() tea.Msg {
		return LockCompletedMsg{Err: service.Lock()}
	}
}
