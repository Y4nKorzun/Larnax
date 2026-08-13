// Command kdbx-tui is this application's entry point — spec section
// 18.7's "composition root": the one place that constructs a
// *application.VaultService and hands it to the TUI layer. Per spec
// section 18.2, nothing below this file is allowed to do that wiring
// itself.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/Y4nKorzun/Larnax/internal/application"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/clipboard"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/random"
	"github.com/Y4nKorzun/Larnax/internal/tui"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main's testable body. Every side effect — which writer output
// goes to, whether a real Bubble Tea program starts — is a parameter or
// an explicit call below, never a bare os.Stdout/os.Stdin reference, so
// --version and --doctor can be exercised from an automated test without
// a real terminal.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("kdbx-tui", flag.ContinueOnError)
	fs.SetOutput(stderr)
	version := fs.Bool("version", false, "print the application version and exit")
	doctor := fs.Bool("doctor", false, "print spec section 22.2's environment/capability report and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	switch {
	case *version:
		fmt.Fprintln(stdout, application.Version)
		return 0
	case *doctor:
		fmt.Fprintln(stdout, application.BuildDoctorReport(application.Version, nil).String())
		return 0
	default:
		return runInteractive(fs.Args(), stderr)
	}
}

// runInteractive starts the actual Bubble Tea program: spec section
// 7.1's welcome screen onward, with a single *application.VaultService
// for the whole run (spec 6.2's "multiple vaults or fast switch" is P1
// scope this constructs no support for yet). positional[0], if present,
// is the vault path spec 7.1's "kdbx-tui <path>" form supplies up front.
//
// This needs a real terminal to run meaningfully — there is no way to
// exercise it from an automated test the way --version/--doctor above
// can be; cmd/kdbx-tui's own tests instead confirm the binary builds and
// that the non-interactive flags work, and leave this path to manual
// verification.
func runInteractive(positional []string, stderr io.Writer) int {
	var unlockPath string
	if len(positional) > 0 {
		unlockPath = positional[0]
	}

	cb, _ := clipboard.New() // nil, ok=false on a platform with no adapter yet; AppModel handles that

	var service application.VaultService
	model := tui.NewAppModel(random.CryptoSource{}, &service, unlockPath, cb)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintln(stderr, "kdbx-tui:", err)
		return 1
	}
	return 0
}
