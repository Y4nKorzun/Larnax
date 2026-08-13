package application

import (
	"bytes"
	"os"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

const openVaultTestPassword = "correct horse battery staple test only"

func TestOpenVaultCleanVaultIsWritable(t *testing.T) {
	doc, err := CreateVault("My Vault", openVaultTestPassword)
	if err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}

	var buf bytes.Buffer
	if err := kdbx.Save(&buf, doc); err != nil {
		t.Fatalf("kdbx.Save() error = %v", err)
	}

	opened, err := OpenVault(&buf, openVaultTestPassword)
	if err != nil {
		t.Fatalf("OpenVault() error = %v", err)
	}
	if opened.ReadOnly {
		t.Errorf("ReadOnly = true, want false: Unsupported = %v", opened.Unsupported)
	}
	if len(opened.Unsupported) != 0 {
		t.Errorf("Unsupported = %v, want empty", opened.Unsupported)
	}
}

// TestOpenVaultFixtureWithAttachmentIsReadOnly uses a KDBX file a real
// KeePass client produced (see testdata/kdbx/keepass/README.md), whose
// second group has an entry with a binary attachment — a construct
// domain.Entry has no field for at all. Spec section 15.4: opening it
// must fall back to read-only rather than silently drop that attachment
// on the next save.
func TestOpenVaultFixtureWithAttachmentIsReadOnly(t *testing.T) {
	const fixturePath = "../../testdata/kdbx/keepass/kdbx4-example.kdbx"
	const fixturePassword = "abcdefg12345678"

	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	opened, err := OpenVault(f, fixturePassword)
	if err != nil {
		t.Fatalf("OpenVault() error = %v", err)
	}
	if !opened.ReadOnly {
		t.Fatal("ReadOnly = false, want true (fixture has a binary attachment)")
	}

	found := false
	for _, feature := range opened.Unsupported {
		if feature == kdbx.FeatureAttachments {
			found = true
		}
	}
	if !found {
		t.Errorf("Unsupported = %v, want it to include %q", opened.Unsupported, kdbx.FeatureAttachments)
	}
}

func TestOpenVaultPropagatesWrongPasswordError(t *testing.T) {
	doc, err := CreateVault("My Vault", openVaultTestPassword)
	if err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}

	var buf bytes.Buffer
	if err := kdbx.Save(&buf, doc); err != nil {
		t.Fatalf("kdbx.Save() error = %v", err)
	}

	if _, err := OpenVault(&buf, "wrong password"); err == nil {
		t.Error("OpenVault() with wrong password succeeded, want error")
	}
}
