package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/tui/components"
)

// passwordMask is what the detail view shows for a hidden password
// (spec section 8.5: hidden by default). It is a fixed width rather than
// sized to the actual password's length, so the detail view doesn't leak
// how long a password is just by looking at it.
const passwordMask = "••••••••"

// revealDuration is spec section 8.5's own example ("shows it for a
// limited time, e.g. 5 seconds"). config.SecurityConfig.RevealTimeout is
// the configurable version of this same value; nothing yet loads
// config.toml at startup to wire it through, so this stays the fixed
// spec example until that exists.
const revealDuration = 5 * time.Second

// revealTickMsg drives the "Auto-hide in Ns" countdown: while an entry's
// password is revealed, BrowserModel reschedules one of these every
// second so View() gets called again and re-evaluates isRevealed against
// the current time, purely to force a redraw — the actual hide decision
// never depends on receiving a tick, only on m.now() versus
// revealedUntil, so a dropped or delayed tick can't extend a reveal.
type revealTickMsg struct{}

type browserOverlay int

const (
	browserOverlayNone browserOverlay = iota
	browserOverlayAddEntry
)

// BrowserModel is the minimal spec section 8 main screen reached after a
// vault is unlocked or created: browse entries, add one, copy fields,
// save, lock. It is not yet spec 8.1's full three-panel group/entry/
// detail layout — there is only one flat entry list per vault so far (no
// group navigation), and no entry detail/edit view beyond the list
// itself and StatusLineModel.
//
// BrowserModel never calls application.VaultService's mutating or
// clipboard methods itself, only records what the user asked for
// (CopyIntent, Save, Lock, Quit) — a parent (AppModel) is the one that
// turns those into the actual Cmd, the same Intent/Cmd split
// messages.go's doc comment describes for spec 19.2's chain.
type BrowserModel struct {
	service *application.VaultService
	src     random.Source

	list components.EntryListModel
	ids  []domain.EntryID // parallel to list.Titles

	keys          KeySequence
	leaderPending bool

	overlay  browserOverlay
	addEntry AddEntryModel

	revealedEntryID domain.EntryID
	revealedUntil   time.Time
	now             func() time.Time

	CopyIntent    *CopyFieldIntent
	Save          bool
	Lock          bool
	Quit          bool
	StatusMessage string
}

// NewBrowserModel returns a BrowserModel over service's currently open
// vault.
func NewBrowserModel(service *application.VaultService, src random.Source) BrowserModel {
	m := BrowserModel{service: service, src: src, now: time.Now}
	return m.refresh()
}

// refresh rebuilds the entry list from the vault's current state. Call
// after any mutation — adding an entry, undo/redo — so the list never
// drifts from what VaultService actually holds.
func (m BrowserModel) refresh() BrowserModel {
	entries := m.service.Vault().AllEntries()
	titles := make([]string, len(entries))
	ids := make([]domain.EntryID, len(entries))
	for i, e := range entries {
		titles[i] = e.Title
		ids[i] = e.ID
	}

	cursor := m.list.Cursor
	m.list = components.NewEntryListModel(titles)
	if cursor < len(titles) {
		m.list.Cursor = cursor
	}
	m.ids = ids
	return m
}

// StatusLine renders spec 8.1's status bar for the current state.
func (m BrowserModel) StatusLine() string {
	lockState := components.LockStateUnlocked
	if m.service.ReadOnly() {
		lockState = components.LockStateReadOnly
	}
	return components.StatusLineModel{
		Mode:       "NORMAL",
		GroupPath:  "/",
		EntryCount: len(m.list.Titles),
		LockState:  lockState,
	}.View()
}

func (m BrowserModel) Init() tea.Cmd { return nil }

func (m BrowserModel) Update(msg tea.Msg) (BrowserModel, tea.Cmd) {
	m.CopyIntent = nil
	m.Save = false
	m.Lock = false
	m.Quit = false

	if m.overlay == browserOverlayAddEntry {
		return m.updateAddEntryOverlay(msg)
	}

	if _, ok := msg.(revealTickMsg); ok {
		return m, m.rescheduleRevealTick()
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Code == tea.KeyEscape {
		m.keys.Reset()
		m.leaderPending = false
		return m, nil
	}

	if m.leaderPending {
		m.leaderPending = false
		return m.handleLeaderKey(key)
	}
	if key.Text == " " {
		m.leaderPending = true
		return m, nil
	}

	if key.Text != "" {
		if action, resolved := m.keys.Feed([]rune(key.Text)[0]); resolved {
			return m.handleAction(action)
		}
		if m.keys.Pending() {
			return m, nil
		}
	}

	switch key.Text {
	case "a":
		m.overlay = browserOverlayAddEntry
		m.addEntry = NewAddEntryModel(m.src, m.service.Vault().RootGroupID())
		return m, nil
	case "q":
		m.Quit = true
		return m, nil
	case "r":
		return m.reveal()
	}

	m.list, _ = m.list.Update(msg)
	return m, nil
}

// reveal shows the selected entry's password for revealDuration (spec
// section 8.5). It is keyed to the specific entry ID, not just a time
// window: navigating to a different entry within that window must never
// inherit the reveal — isRevealed checks both.
func (m BrowserModel) reveal() (BrowserModel, tea.Cmd) {
	id, ok := m.selectedID()
	if !ok {
		return m, nil
	}
	m.revealedEntryID = id
	m.revealedUntil = m.now().Add(revealDuration)
	return m, m.rescheduleRevealTick()
}

// isRevealed reports whether the currently selected entry's password
// should be shown in plaintext right now.
func (m BrowserModel) isRevealed() bool {
	if m.revealedUntil.IsZero() || !m.now().Before(m.revealedUntil) {
		return false
	}
	id, ok := m.selectedID()
	return ok && id == m.revealedEntryID
}

// rescheduleRevealTick returns a Cmd that sends another revealTickMsg
// in a second, or nil once nothing is revealed anymore — so the tick
// train stops on its own instead of ticking forever in the background.
func (m BrowserModel) rescheduleRevealTick() tea.Cmd {
	if !m.isRevealed() {
		return nil
	}
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return revealTickMsg{} })
}

func (m BrowserModel) handleLeaderKey(key tea.KeyPressMsg) (BrowserModel, tea.Cmd) {
	switch key.Text {
	case "s":
		m.Save = true
	case "l":
		m.Lock = true
	}
	return m, nil
}

func (m BrowserModel) handleAction(action Action) (BrowserModel, tea.Cmd) {
	switch action {
	case ActionFirstItem:
		m.list = m.list.First()
	case ActionCopyUsername, ActionCopyPassword, ActionCopyURL, ActionCopyNotes:
		m.setCopyIntent(action)
	}
	return m, nil
}

// copyableFields maps the two-key copy Actions this browser already
// supports to application's FieldName. ActionCopyNotes is deliberately
// absent: spec 11.2's yn has no application.FieldName counterpart yet
// (copy_field.go only covers username/password/URL), so it is simply
// not actionable here rather than mapped to the wrong field.
var copyableFields = map[Action]application.FieldName{
	ActionCopyUsername: application.FieldUsername,
	ActionCopyPassword: application.FieldPassword,
	ActionCopyURL:      application.FieldURL,
}

func (m *BrowserModel) setCopyIntent(action Action) {
	field, ok := copyableFields[action]
	if !ok {
		return
	}

	id, ok := m.selectedID()
	if !ok {
		return
	}
	entry, err := m.service.Vault().Entry(id)
	if err != nil {
		return
	}

	intent := CopyFieldIntent{Entry: entry, Field: field}
	m.CopyIntent = &intent
}

func (m BrowserModel) selectedID() (domain.EntryID, bool) {
	if m.list.Cursor < 0 || m.list.Cursor >= len(m.ids) {
		return domain.EntryID{}, false
	}
	return m.ids[m.list.Cursor], true
}

// SelectedEntry returns the entry at the list cursor, if any.
func (m BrowserModel) SelectedEntry() (domain.Entry, bool) {
	id, ok := m.selectedID()
	if !ok {
		return domain.Entry{}, false
	}
	entry, err := m.service.Vault().Entry(id)
	if err != nil {
		return domain.Entry{}, false
	}
	return entry, true
}

// detailView renders spec 8.1's "Details" panel for the currently
// selected entry, matching spec 8.5's reveal format
// ("Password: qA8$...xP2" / "Auto-hide in 4s") while isRevealed holds.
func (m BrowserModel) detailView() string {
	entry, ok := m.SelectedEntry()
	if !ok {
		return "(no entry selected)\n"
	}

	passwordLine := "Password: " + passwordMask
	if m.isRevealed() {
		var plain string
		_ = entry.Password.Reveal(func(v []byte) error {
			plain = string(v)
			return nil
		})
		remaining := int(m.revealedUntil.Sub(m.now()).Round(time.Second) / time.Second)
		passwordLine = fmt.Sprintf("Password: %s\nAuto-hide in %ds", plain, remaining)
	}

	return fmt.Sprintf(
		"Title:    %s\nUsername: %s\n%s\nURL:      %s\nNotes:    %s\n",
		entry.Title, entry.Username, passwordLine, entry.URL, entry.Notes,
	)
}

func (m BrowserModel) updateAddEntryOverlay(msg tea.Msg) (BrowserModel, tea.Cmd) {
	m.addEntry, _ = m.addEntry.Update(msg)

	if m.addEntry.Cancelled {
		m.overlay = browserOverlayNone
		return m, nil
	}
	if m.addEntry.Done {
		if err := m.service.AddEntry(m.addEntry.Entry); err != nil {
			m.StatusMessage = err.Error()
		}
		m.overlay = browserOverlayNone
		return m.refresh(), nil
	}
	return m, nil
}

func (m BrowserModel) View() string {
	if m.overlay == browserOverlayAddEntry {
		return m.addEntry.View()
	}
	return m.list.View() + "\n" + m.detailView() + "\n" + m.StatusLine() + "\n"
}
