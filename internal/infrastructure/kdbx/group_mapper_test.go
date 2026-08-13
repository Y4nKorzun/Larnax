package kdbx

import (
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestGroupToGKPMapsFields(t *testing.T) {
	g := domain.NewGroup(domain.NewGroupID(), "Personal")
	g.Notes = "family accounts"

	gg := groupToGKP(g)

	if [16]byte(gg.UUID) != [16]byte(g.ID) {
		t.Errorf("UUID = %x, want %x", gg.UUID, g.ID)
	}
	if gg.Name != g.Name {
		t.Errorf("Name = %q, want %q", gg.Name, g.Name)
	}
	if gg.Notes != g.Notes {
		t.Errorf("Notes = %q, want %q", gg.Notes, g.Notes)
	}
}

func TestGroupFromGKPRoundTripsFields(t *testing.T) {
	original := domain.NewGroup(domain.NewGroupID(), "Work")
	original.Notes = "job accounts"

	got := groupFromGKP(groupToGKP(original))

	if got.ID != original.ID {
		t.Errorf("ID = %x, want %x", got.ID, original.ID)
	}
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Notes != original.Notes {
		t.Errorf("Notes = %q, want %q", got.Notes, original.Notes)
	}
	if got.ParentID != nil {
		t.Errorf("ParentID = %v, want nil (decoder.go's job to set)", got.ParentID)
	}
}
