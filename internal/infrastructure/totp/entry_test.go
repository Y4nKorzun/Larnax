package totp

import (
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestEntryURIAbsentReturnsFalse(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	uri, present, err := EntryURI(e)
	if err != nil {
		t.Fatalf("EntryURI() error = %v", err)
	}
	if present {
		t.Errorf("present = true, want false (uri = %q)", uri)
	}
}

func TestSetEntryURIThenEntryURIRoundTrips(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	const uri = "otpauth://totp/GitHub:octocat?secret=JBSWY3DPEHPK3PXP&issuer=GitHub"

	updated := SetEntryURI(e, uri)

	got, present, err := EntryURI(updated)
	if err != nil {
		t.Fatalf("EntryURI() error = %v", err)
	}
	if !present {
		t.Fatal("present = false, want true")
	}
	if got != uri {
		t.Errorf("EntryURI() = %q, want %q", got, uri)
	}
}

func TestSetEntryURIDoesNotMutateOriginal(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	SetEntryURI(e, "otpauth://totp/x?secret=AAAA")

	if _, present, _ := EntryURI(e); present {
		t.Error("original entry gained a TOTP field after SetEntryURI() on a copy")
	}
}

func TestSetEntryURIReplacesRatherThanDuplicates(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	e = SetEntryURI(e, "otpauth://totp/x?secret=AAAA")
	e = SetEntryURI(e, "otpauth://totp/x?secret=BBBB")

	count := 0
	for _, f := range e.CustomFields {
		if f.Name == FieldName {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("entry has %d TOTP fields, want 1", count)
	}

	got, _, err := EntryURI(e)
	if err != nil {
		t.Fatalf("EntryURI() error = %v", err)
	}
	if got != "otpauth://totp/x?secret=BBBB" {
		t.Errorf("EntryURI() = %q, want the second value", got)
	}
}

func TestSetEntryURIPreservesOtherCustomFields(t *testing.T) {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	e.CustomFields = []domain.Field{{Name: "Recovery Code", Value: domain.NewSecretFromString("abc123")}}

	e = SetEntryURI(e, "otpauth://totp/x?secret=AAAA")

	if len(e.CustomFields) != 2 {
		t.Fatalf("CustomFields = %d, want 2", len(e.CustomFields))
	}
	found := false
	for _, f := range e.CustomFields {
		if f.Name == "Recovery Code" {
			found = true
		}
	}
	if !found {
		t.Error("Recovery Code field was dropped by SetEntryURI()")
	}
}
