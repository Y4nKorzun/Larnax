package kdbx

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/totp"
)

// TestTOTPFieldSurvivesSaveAndOpen proves spec section 14.3's storage
// convention (totp.FieldName) round-trips through this package's actual
// file-level Open/Save — not just the in-memory mapper — with no special
// case needed anywhere in decoder.go/encoder.go/mapper.go: a TOTP URI is,
// as far as this package is concerned, just another protected custom
// field, and the existing generic custom-field handling already covers
// it.
func TestTOTPFieldSurvivesSaveAndOpen(t *testing.T) {
	const totpSecret = "JBSWY3DPEHPK3PXP"
	params := totp.Params{
		Secret:    []byte(totpSecret),
		Digits:    6,
		Period:    30 * time.Second,
		Algorithm: totp.AlgorithmSHA1,
	}
	uri := totp.BuildURI("GitHub:octocat", "GitHub", params)

	root := domain.Group{ID: domain.NewGroupID(), Name: "My Vault"}
	vault, err := domain.NewVaultFromRoot(root)
	if err != nil {
		t.Fatalf("NewVaultFromRoot() error = %v", err)
	}
	entry := domain.NewEntry(root.ID, "GitHub")
	entry = totp.SetEntryURI(entry, uri)
	if err := vault.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	doc, err := NewDocument(vault, repoTestPassword, PortableProfile())
	if err != nil {
		t.Fatalf("NewDocument() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	writeVault(t, path, doc)
	reopened := openVault(t, path, repoTestPassword)

	got, err := reopened.Vault.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}

	gotURI, present, err := totp.EntryURI(got)
	if err != nil {
		t.Fatalf("totp.EntryURI() error = %v", err)
	}
	if !present {
		t.Fatal("TOTP field did not survive the round trip")
	}
	if gotURI != uri {
		t.Errorf("TOTP URI = %q, want %q", gotURI, uri)
	}

	parsed, err := totp.ParseURI(gotURI)
	if err != nil {
		t.Fatalf("totp.ParseURI() on the round-tripped URI error = %v", err)
	}
	if _, err := totp.Generate(parsed.Params, time.Now()); err != nil {
		t.Errorf("totp.Generate() on the round-tripped params error = %v", err)
	}
}
