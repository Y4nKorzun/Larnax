package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/tui/components"
)

type addEntryStep int

const (
	addEntryStepTitle addEntryStep = iota
	addEntryStepUsername
	addEntryStepDone
)

// AddEntryModel is a minimal spec section 9.2 "add entry" flow: title,
// then username, then an automatically generated password under spec
// section 10.2's default policy — spec 26.1's acceptance script's
// "generate a 24-character password" step, made automatic here rather
// than routing through the full GeneratorModel screen, which stays
// available separately for anyone who wants a different policy. Esc at
// either text step cancels without adding anything.
type AddEntryModel struct {
	src    random.Source
	parent domain.GroupID

	step     addEntryStep
	title    components.PlainInputModel
	username components.PlainInputModel

	Entry     domain.Entry
	Done      bool
	Cancelled bool
	Err       error
}

// NewAddEntryModel returns an AddEntryModel that will add its entry
// under parent once finished, drawing the generated password from src.
func NewAddEntryModel(src random.Source, parent domain.GroupID) AddEntryModel {
	return AddEntryModel{
		src:      src,
		parent:   parent,
		title:    components.NewPlainInputModel(),
		username: components.NewPlainInputModel(),
	}
}

func (m AddEntryModel) Init() tea.Cmd { return nil }

func (m AddEntryModel) Update(msg tea.Msg) (AddEntryModel, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Code == tea.KeyEscape {
		m.Cancelled = true
		return m, nil
	}

	switch m.step {
	case addEntryStepTitle:
		if key.Code == tea.KeyEnter {
			if strings.TrimSpace(m.title.Value()) == "" {
				return m, nil // spec 9.2 needs a title; keep waiting for one
			}
			m.step = addEntryStepUsername
			return m, nil
		}
		m.title, _ = m.title.Update(msg)
		return m, nil

	case addEntryStepUsername:
		if key.Code == tea.KeyEnter {
			return m.finish()
		}
		m.username, _ = m.username.Update(msg)
		return m, nil

	default:
		return m, nil
	}
}

func (m AddEntryModel) finish() (AddEntryModel, tea.Cmd) {
	password, err := application.GeneratePassword(m.src, application.DefaultPasswordPolicy())
	if err != nil {
		m.Err = err
		return m, nil
	}

	entry := domain.NewEntry(m.parent, m.title.Value())
	entry.Username = m.username.Value()
	entry.Password = password

	m.Entry = entry
	m.Done = true
	m.step = addEntryStepDone
	return m, nil
}

func (m AddEntryModel) View() string {
	switch m.step {
	case addEntryStepTitle:
		return "Add entry\n\nTitle:\n" + m.title.View() + "\n"
	case addEntryStepUsername:
		return "Add entry\n\nTitle: " + m.title.Value() + "\nUsername:\n" + m.username.View() + "\n"
	default:
		return "Entry added.\n"
	}
}
