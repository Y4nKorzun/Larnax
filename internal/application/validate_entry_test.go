package application

import (
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestValidateEntryRejectsEmptyTitle(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "")
	if err := ValidateEntry(e); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("ValidateEntry() error = %v, want %v", err, ErrEmptyTitle)
	}
}

func TestValidateEntryRejectsWhitespaceOnlyTitle(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "   ")
	if err := ValidateEntry(e); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("ValidateEntry() error = %v, want %v", err, ErrEmptyTitle)
	}
}

func TestValidateEntryAcceptsMinimalValidEntry(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	if err := ValidateEntry(e); err != nil {
		t.Errorf("ValidateEntry() error = %v, want nil", err)
	}
}

func TestValidateEntryAllowsEmptyOptionalFields(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	// Username, URL, Notes, Password all left at their zero values.
	if err := ValidateEntry(e); err != nil {
		t.Errorf("ValidateEntry() error = %v, want nil (only Title is required)", err)
	}
}
