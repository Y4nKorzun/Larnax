package kdbx

import (
	"testing"
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func sampleEntry() domain.Entry {
	e := domain.NewEntry(domain.NewGroupID(), "GitHub")
	e.Username = "octocat"
	e.Password = domain.NewSecretFromString("s3cr3t-P@ssw0rd")
	e.URL = "https://github.com"
	e.Notes = "created by mapper_test"
	e.Tags = []string{"work", "dev"}
	e.CustomFields = []domain.Field{
		{Name: "Recovery Code", Value: domain.NewSecretFromString("abc123")},
	}
	return e
}

func TestEntryToGKPMapsStandardFields(t *testing.T) {
	e := sampleEntry()
	ge, err := entryToGKP(e)
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}

	if got := ge.GetTitle(); got != e.Title {
		t.Errorf("Title = %q, want %q", got, e.Title)
	}
	if got := ge.GetContent(fieldUserName); got != e.Username {
		t.Errorf("UserName = %q, want %q", got, e.Username)
	}
	if got := ge.GetPassword(); got != "s3cr3t-P@ssw0rd" {
		t.Errorf("Password = %q, want %q", got, "s3cr3t-P@ssw0rd")
	}
	if got := ge.GetContent(fieldURL); got != e.URL {
		t.Errorf("URL = %q, want %q", got, e.URL)
	}
	if got := ge.GetContent(fieldNotes); got != e.Notes {
		t.Errorf("Notes = %q, want %q", got, e.Notes)
	}
	if gokeepUUID, want := [16]byte(ge.UUID), [16]byte(e.ID); gokeepUUID != want {
		t.Errorf("UUID = %x, want %x", gokeepUUID, want)
	}
}

func TestEntryToGKPMarksPasswordProtected(t *testing.T) {
	ge, err := entryToGKP(sampleEntry())
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}
	pw := ge.Get(fieldPassword)
	if pw == nil {
		t.Fatal("Password value missing")
	}
	if !pw.Value.Protected.Bool {
		t.Error("Password value not marked Protected")
	}
}

func TestEntryToGKPMarksCustomFieldsProtected(t *testing.T) {
	ge, err := entryToGKP(sampleEntry())
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}
	custom := ge.Get("Recovery Code")
	if custom == nil {
		t.Fatal("custom field missing")
	}
	if custom.Value.Content != "abc123" {
		t.Errorf("custom field content = %q, want %q", custom.Value.Content, "abc123")
	}
	if !custom.Value.Protected.Bool {
		t.Error("custom field not marked Protected")
	}
}

func TestEntryToGKPJoinsTags(t *testing.T) {
	ge, err := entryToGKP(sampleEntry())
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}
	if ge.Tags != "work;dev" {
		t.Errorf("Tags = %q, want %q", ge.Tags, "work;dev")
	}
}

func TestEntryToGKPWritesExpiry(t *testing.T) {
	e := sampleEntry()
	expiry := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	e.ExpiresAt = &expiry

	ge, err := entryToGKP(e)
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}
	if !ge.Times.Expires.Bool {
		t.Error("Expires = false, want true")
	}
	if ge.Times.ExpiryTime == nil || !ge.Times.ExpiryTime.Time.Equal(expiry) {
		t.Errorf("ExpiryTime = %v, want %v", ge.Times.ExpiryTime, expiry)
	}
}

func TestEntryFromGKPRoundTripsStandardFields(t *testing.T) {
	original := sampleEntry()
	ge, err := entryToGKP(original)
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}

	got := entryFromGKP(ge)

	if got.ID != original.ID {
		t.Errorf("ID = %x, want %x", got.ID, original.ID)
	}
	if got.Title != original.Title {
		t.Errorf("Title = %q, want %q", got.Title, original.Title)
	}
	if got.Username != original.Username {
		t.Errorf("Username = %q, want %q", got.Username, original.Username)
	}
	gotPassword, err := revealString(got.Password)
	if err != nil {
		t.Fatalf("revealing decoded password: %v", err)
	}
	if wantPassword, _ := revealString(original.Password); gotPassword != wantPassword {
		t.Errorf("Password = %q, want %q", gotPassword, wantPassword)
	}
	if got.URL != original.URL {
		t.Errorf("URL = %q, want %q", got.URL, original.URL)
	}
	if got.Notes != original.Notes {
		t.Errorf("Notes = %q, want %q", got.Notes, original.Notes)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "work" || got.Tags[1] != "dev" {
		t.Errorf("Tags = %v, want [work dev]", got.Tags)
	}
}

func TestEntryFromGKPSkipsStandardFieldsAsCustom(t *testing.T) {
	ge, err := entryToGKP(sampleEntry())
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}

	got := entryFromGKP(ge)

	if len(got.CustomFields) != 1 {
		t.Fatalf("CustomFields = %d entries, want 1: %+v", len(got.CustomFields), got.CustomFields)
	}
	if got.CustomFields[0].Name != "Recovery Code" {
		t.Errorf("custom field name = %q, want %q", got.CustomFields[0].Name, "Recovery Code")
	}
	content, err := revealString(got.CustomFields[0].Value)
	if err != nil {
		t.Fatalf("revealing custom field: %v", err)
	}
	if content != "abc123" {
		t.Errorf("custom field content = %q, want %q", content, "abc123")
	}
}

func TestEntryFromGKPParsesExpiry(t *testing.T) {
	e := sampleEntry()
	expiry := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	e.ExpiresAt = &expiry
	ge, err := entryToGKP(e)
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}

	got := entryFromGKP(ge)

	if got.ExpiresAt == nil {
		t.Fatal("ExpiresAt = nil, want non-nil")
	}
	if !got.ExpiresAt.Equal(expiry) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, expiry)
	}
}

func TestEntryFromGKPNoExpiryLeavesNilExpiresAt(t *testing.T) {
	ge, err := entryToGKP(sampleEntry())
	if err != nil {
		t.Fatalf("entryToGKP() error = %v", err)
	}

	got := entryFromGKP(ge)

	if got.ExpiresAt != nil {
		t.Errorf("ExpiresAt = %v, want nil", got.ExpiresAt)
	}
}

func TestSplitTagsEmptyReturnsNil(t *testing.T) {
	if got := splitTags(""); got != nil {
		t.Errorf("splitTags(\"\") = %v, want nil", got)
	}
}
