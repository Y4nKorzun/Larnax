package kdbx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

const testMasterPassword = "correct horse battery staple test only"

func plainValue(key, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func protectedValue(key, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: w.NewBoolWrapper(true)},
	}
}

// TestRoundTripPreservesGroupEntryAndProtectedPassword is Milestone 0's
// riskiest-assumption check: gokeepasslib/v3 must be able to create a KDBX 4.1
// database in memory, lock+encode it to disk, and decode+unlock it again with
// every field intact, following the lifecycle spec section 15.2 requires
// (decode -> unlock -> ... -> lock -> encode).
func TestRoundTripPreservesGroupEntryAndProtectedPassword(t *testing.T) {
	const (
		childGroupName = "Personal"
		entryTitle     = "GitHub"
		entryUsername  = "octocat"
		entryPassword  = "s3cr3t-P@ssw0rd-value"
		entryURL       = "https://github.com"
		entryNotes     = "created by kdbx-tui round-trip test"
	)

	// --- build the in-memory database ---
	db := gokeepasslib.NewDatabase(
		gokeepasslib.WithDatabaseKDBXVersion41(),
	)
	db.Credentials = gokeepasslib.NewPasswordCredentials(testMasterPassword)
	db.Content.Meta.DatabaseName = "kdbx-tui roundtrip fixture"

	// NewDatabase() pre-populates Content.Root.Groups[0] with a group named
	// "NewDatabase" containing one sample entry ("Sample Entry"). Reuse that
	// group as our root and clear its sample entry so the test owns exactly
	// the entries it asserts on.
	rootGroup := &db.Content.Root.Groups[0]
	rootGroup.Entries = nil

	childGroup := gokeepasslib.NewGroup()
	childGroup.Name = childGroupName
	originalChildGroupUUID := childGroup.UUID

	entry := gokeepasslib.NewEntry()
	entry.Values = append(entry.Values,
		plainValue("Title", entryTitle),
		plainValue("UserName", entryUsername),
		protectedValue("Password", entryPassword),
		plainValue("URL", entryURL),
		plainValue("Notes", entryNotes),
	)
	originalEntryUUID := entry.UUID

	childGroup.Entries = append(childGroup.Entries, entry)
	rootGroup.Groups = append(rootGroup.Groups, childGroup)

	// --- lock + encode to a real temp file, mirroring an actual save ---
	if err := db.LockProtectedEntries(); err != nil {
		t.Fatalf("LockProtectedEntries() failed: %v", err)
	}

	path := filepath.Join(t.TempDir(), "roundtrip.kdbx")

	wf, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	if err := gokeepasslib.NewEncoder(wf).Encode(db); err != nil {
		wf.Close()
		t.Fatalf("Encode() failed: %v", err)
	}
	if err := wf.Close(); err != nil {
		t.Fatalf("closing written file failed: %v", err)
	}

	// --- decode with a fresh Database/Decoder, mirroring a real reopen ---
	rf, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() failed: %v", err)
	}
	defer rf.Close()

	decoded := gokeepasslib.NewDatabase()
	decoded.Credentials = gokeepasslib.NewPasswordCredentials(testMasterPassword)
	if err := gokeepasslib.NewDecoder(rf).Decode(decoded); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if err := decoded.UnlockProtectedEntries(); err != nil {
		t.Fatalf("UnlockProtectedEntries() failed: %v", err)
	}

	// --- assert exact preservation ---
	if !decoded.Header.IsKdbx41() {
		t.Fatalf(
			"expected KDBX 4.1, got version %d.%d",
			decoded.Header.Signature.MajorVersion,
			decoded.Header.Signature.MinorVersion,
		)
	}

	if decoded.Content.Meta.DatabaseName != "kdbx-tui roundtrip fixture" {
		t.Errorf("DatabaseName = %q, want %q",
			decoded.Content.Meta.DatabaseName, "kdbx-tui roundtrip fixture")
	}

	if got := len(decoded.Content.Root.Groups); got != 1 {
		t.Fatalf("expected 1 root group, got %d", got)
	}
	decodedRoot := decoded.Content.Root.Groups[0]

	if got := len(decodedRoot.Groups); got != 1 {
		t.Fatalf("expected 1 child group under root, got %d", got)
	}
	decodedChild := decodedRoot.Groups[0]

	if decodedChild.Name != childGroupName {
		t.Errorf("child group name = %q, want %q", decodedChild.Name, childGroupName)
	}
	if !decodedChild.UUID.Compare(originalChildGroupUUID) {
		t.Errorf("child group UUID changed: got %x, want %x",
			decodedChild.UUID, originalChildGroupUUID)
	}

	if got := len(decodedChild.Entries); got != 1 {
		t.Fatalf("expected 1 entry in child group, got %d", got)
	}
	decodedEntry := decodedChild.Entries[0]

	if !decodedEntry.UUID.Compare(originalEntryUUID) {
		t.Errorf("entry UUID changed: got %x, want %x", decodedEntry.UUID, originalEntryUUID)
	}
	if got := decodedEntry.GetTitle(); got != entryTitle {
		t.Errorf("Title = %q, want %q", got, entryTitle)
	}
	if got := decodedEntry.GetContent("UserName"); got != entryUsername {
		t.Errorf("UserName = %q, want %q", got, entryUsername)
	}
	if got := decodedEntry.GetPassword(); got != entryPassword {
		t.Errorf("Password = %q, want %q", got, entryPassword)
	}
	if got := decodedEntry.GetContent("URL"); got != entryURL {
		t.Errorf("URL = %q, want %q", got, entryURL)
	}
	if got := decodedEntry.GetContent("Notes"); got != entryNotes {
		t.Errorf("Notes = %q, want %q", got, entryNotes)
	}

	// The password must have gone through the encrypted-at-rest path, not
	// merely survived as accidental plaintext.
	pwValue := decodedEntry.Get("Password")
	if pwValue == nil {
		t.Fatal("Password value missing after decode")
	}
	if !pwValue.Value.Protected.Bool {
		t.Errorf("Password value lost its Protected=true flag after round trip")
	}
}
