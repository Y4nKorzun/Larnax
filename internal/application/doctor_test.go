package application

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
)

const doctorTestPassword = "correct horse battery staple test only"

func TestFileLockAvailableReturnsTrue(t *testing.T) {
	if !fileLockAvailable() {
		t.Error("fileLockAvailable() = false, want true in a normal test environment")
	}
}

func TestBuildDoctorReportPassesThroughAppVersionAndClipboard(t *testing.T) {
	r := BuildDoctorReport("0.1.0", nil)
	if r.AppVersion != "0.1.0" {
		t.Errorf("AppVersion = %q, want %q", r.AppVersion, "0.1.0")
	}
	if r.ClipboardAvailable != clipboard.Available() {
		t.Errorf("ClipboardAvailable = %v, want %v", r.ClipboardAvailable, clipboard.Available())
	}
}

func TestBuildDoctorReportNilVault(t *testing.T) {
	r := BuildDoctorReport("0.1.0", nil)
	if !strings.Contains(r.KDBXWriteNote, "no vault open") {
		t.Errorf("KDBXWriteNote = %q, want it to mention no vault open", r.KDBXWriteNote)
	}
	if !strings.Contains(r.VaultPermissionsNote, "not applicable") {
		t.Errorf("VaultPermissionsNote = %q, want %q", r.VaultPermissionsNote, "not applicable (no vault open)")
	}
}

func TestBuildDoctorReportUnopenedVaultService(t *testing.T) {
	var vault VaultService // zero value: never New'd or Open'd
	r := BuildDoctorReport("0.1.0", &vault)
	if !strings.Contains(r.KDBXWriteNote, "no vault open") {
		t.Errorf("KDBXWriteNote = %q, want it to mention no vault open", r.KDBXWriteNote)
	}
}

func TestBuildDoctorReportWritableVault(t *testing.T) {
	var vault VaultService
	if err := vault.New("My Vault", doctorTestPassword); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "vault.kdbx")
	if err := vault.SaveAs(path, 0); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	r := BuildDoctorReport("0.1.0", &vault)
	if r.KDBXWriteNote != "supported for detected feature set" {
		t.Errorf("KDBXWriteNote = %q, want %q", r.KDBXWriteNote, "supported for detected feature set")
	}

	if runtime.GOOS == "windows" {
		return // permission bits aren't meaningful there (filesystem.ErrPermissionsNotGuaranteed)
	}
	if r.VaultPermissionsNote != "acceptable" {
		t.Errorf("VaultPermissionsNote = %q, want %q", r.VaultPermissionsNote, "acceptable")
	}
}

func TestBuildDoctorReportReadOnlyVault(t *testing.T) {
	const fixturePath = "../../testdata/kdbx/keepass/kdbx4-example.kdbx"
	const fixturePassword = "abcdefg12345678"

	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	var vault VaultService
	if err := vault.Open(f, fixturePath, fixturePassword); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !vault.ReadOnly() {
		t.Fatal("test fixture is not read-only, precondition for this test broke")
	}

	r := BuildDoctorReport("0.1.0", &vault)
	if !strings.Contains(r.KDBXWriteNote, "unavailable") {
		t.Errorf("KDBXWriteNote = %q, want it to mention unavailable", r.KDBXWriteNote)
	}
}
