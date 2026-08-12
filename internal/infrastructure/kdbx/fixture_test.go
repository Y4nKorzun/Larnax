package kdbx

import (
	"os"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
)

// TestDecodeGenuineKeePassFixture decodes a KDBX file that a real KeePass
// client produced (see testdata/kdbx/keepass/README.md for provenance),
// rather than a file gokeepasslib wrote itself. A library can pass its own
// round-trip test (see roundtrip_test.go) while still misreading a
// differently shaped file from another writer, so this is a distinct check.
func TestDecodeGenuineKeePassFixture(t *testing.T) {
	const fixturePath = "../../../testdata/kdbx/keepass/kdbx4-example.kdbx"
	const fixturePassword = "abcdefg12345678"

	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("os.Open() failed: %v", err)
	}
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(fixturePassword)
	if err := gokeepasslib.NewDecoder(file).Decode(db); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		t.Fatalf("UnlockProtectedEntries() failed: %v", err)
	}

	root := db.Content.Root.Groups[0]
	if got := len(root.Groups); got < 2 {
		t.Fatalf("expected at least 2 groups under root, got %d", got)
	}

	firstGroup := root.Groups[0]
	if got := len(firstGroup.Entries); got < 2 {
		t.Fatalf("expected at least 2 entries in first group, got %d", got)
	}

	if got := firstGroup.Entries[0].GetPassword(); got != "Password" {
		t.Errorf("first entry password = %q, want %q", got, "Password")
	}
	if got := firstGroup.Entries[1].GetPassword(); got != "AnotherPassword" {
		t.Errorf("second entry password = %q, want %q", got, "AnotherPassword")
	}

	secondGroup := root.Groups[1]
	if got := len(secondGroup.Entries); got < 1 {
		t.Fatalf("expected at least 1 entry in second group, got %d", got)
	}
	binaryRefs := secondGroup.Entries[0].Binaries
	if got := len(binaryRefs); got < 1 {
		t.Fatalf("expected the entry to have at least 1 binary attachment, got %d", got)
	}

	binary := db.FindBinary(binaryRefs[0].Value.ID)
	if binary == nil {
		t.Fatal("FindBinary() returned nil for a referenced binary ID")
	}
	content, err := binary.GetContentString()
	if err != nil {
		t.Fatalf("GetContentString() failed: %v", err)
	}
	if content != "Hello world" {
		t.Errorf("binary attachment content = %q, want %q", content, "Hello world")
	}
}
