package diagnostics

import (
	"fmt"
	"os"
	"runtime"
)

// Report is spec section 22.2's `:doctor` output — an environment and
// capability snapshot. It never touches vault content: every field here
// is an environment fact or a yes/no capability, not entry data.
type Report struct {
	AppVersion           string
	GoVersion            string
	OS                   string
	Arch                 string
	TerminalDetected     bool
	ClipboardAvailable   bool
	KDBXWriteNote        string
	FileLockAvailable    bool
	VaultPermissionsNote string
}

// NewReport assembles a Report. appVersion is a build-time value this
// package has no way to know on its own. clipboardAvailable,
// fileLockAvailable, kdbxWriteNote, and vaultPermissionsNote come from
// probing the relevant infrastructure package directly — clipboard,
// filesystem, kdbx's feature detector — which this package deliberately
// stays decoupled from, the same separation logger.go keeps from the
// packages it logs about. GoVersion, OS, Arch, and TerminalDetected only
// need the standard library, so they're filled in here.
func NewReport(appVersion string, clipboardAvailable, fileLockAvailable bool, kdbxWriteNote, vaultPermissionsNote string) Report {
	return Report{
		AppVersion:           appVersion,
		GoVersion:            runtime.Version(),
		OS:                   runtime.GOOS,
		Arch:                 runtime.GOARCH,
		TerminalDetected:     isTerminal(os.Stdout),
		ClipboardAvailable:   clipboardAvailable,
		KDBXWriteNote:        kdbxWriteNote,
		FileLockAvailable:    fileLockAvailable,
		VaultPermissionsNote: vaultPermissionsNote,
	}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// String renders r matching spec section 22.2's example layout.
//
// KDBX read is always "supported": gokeepasslib decodes every KDBX
// version this application targets (spec 15.3), so there is nothing to
// probe there the way write support needs (a given file's feature set may
// force read-only). Network capability is always "unused": spec sections
// 19.2 and 29 make zero required network an architectural constant of
// this application, not a runtime condition to detect.
func (r Report) String() string {
	return fmt.Sprintf(
		"Application: %s\n"+
			"Go: %s\n"+
			"OS: %s/%s\n"+
			"Terminal: %s\n"+
			"Clipboard: %s\n"+
			"KDBX read: supported\n"+
			"KDBX write: %s\n"+
			"File lock: %s\n"+
			"Vault permissions: %s\n"+
			"Network capability: unused\n",
		r.AppVersion,
		r.GoVersion,
		r.OS, r.Arch,
		boolLabel(r.TerminalDetected, "detected", "not detected"),
		boolLabel(r.ClipboardAvailable, "available", "unavailable"),
		r.KDBXWriteNote,
		boolLabel(r.FileLockAvailable, "available", "unavailable"),
		r.VaultPermissionsNote,
	)
}

func boolLabel(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}
