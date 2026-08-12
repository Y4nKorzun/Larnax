package tui

import (
	"errors"
	"testing"
)

func TestParseCommandNoArgCommands(t *testing.T) {
	cases := []struct {
		line string
		want Command
	}{
		{"w", WriteCommand{}},
		{"q", QuitCommand{}},
		{"wq", WriteQuitCommand{}},
		{"lock", LockCommand{}},
		{"generate", GenerateCommand{}},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			got, err := ParseCommand(c.line)
			if err != nil {
				t.Fatalf("ParseCommand(%q) error = %v", c.line, err)
			}
			if got != c.want {
				t.Errorf("ParseCommand(%q) = %#v, want %#v", c.line, got, c.want)
			}
		})
	}
}

func TestParseCommandNoArgCommandsRejectArguments(t *testing.T) {
	cases := []string{"w extra", "q now", "wq later", "lock hard", "generate now"}
	for _, line := range cases {
		t.Run(line, func(t *testing.T) {
			_, err := ParseCommand(line)
			if !errors.Is(err, ErrUnexpectedArgument) {
				t.Errorf("ParseCommand(%q) error = %v, want %v", line, err, ErrUnexpectedArgument)
			}
		})
	}
}

func TestParseCommandHelp(t *testing.T) {
	cases := []struct {
		line string
		want HelpCommand
	}{
		{"help", HelpCommand{}},
		{"help yp", HelpCommand{Topic: "yp"}},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			got, err := ParseCommand(c.line)
			if err != nil {
				t.Fatalf("ParseCommand(%q) error = %v", c.line, err)
			}
			if got != c.want {
				t.Errorf("ParseCommand(%q) = %#v, want %#v", c.line, got, c.want)
			}
		})
	}
}

func TestParseCommandPathCommands(t *testing.T) {
	cases := []struct {
		line string
		want Command
	}{
		{"open ~/Secrets/personal.kdbx", OpenCommand{Path: "~/Secrets/personal.kdbx"}},
		{"new ~/Secrets/personal.kdbx", NewCommand{Path: "~/Secrets/personal.kdbx"}},
		{"save-as ~/Secrets/copy.kdbx", SaveAsCommand{Path: "~/Secrets/copy.kdbx"}},
		// Paths may contain spaces (spec section 13.2's own example does).
		{"open ~/My Secrets/personal.kdbx", OpenCommand{Path: "~/My Secrets/personal.kdbx"}},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			got, err := ParseCommand(c.line)
			if err != nil {
				t.Fatalf("ParseCommand(%q) error = %v", c.line, err)
			}
			if got != c.want {
				t.Errorf("ParseCommand(%q) = %#v, want %#v", c.line, got, c.want)
			}
		})
	}
}

func TestParseCommandSearch(t *testing.T) {
	cases := []struct {
		line string
		want SearchCommand
	}{
		{"search github", SearchCommand{Query: "github"}},
		{"search my github login", SearchCommand{Query: "my github login"}},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			got, err := ParseCommand(c.line)
			if err != nil {
				t.Fatalf("ParseCommand(%q) error = %v", c.line, err)
			}
			if got != c.want {
				t.Errorf("ParseCommand(%q) = %#v, want %#v", c.line, got, c.want)
			}
		})
	}
}

func TestParseCommandImportGoogle(t *testing.T) {
	// The literal example from spec section 13.2, including its unescaped space.
	line := "import google ~/Downloads/Google Passwords.csv"
	want := ImportGoogleCommand{Path: "~/Downloads/Google Passwords.csv"}

	got, err := ParseCommand(line)
	if err != nil {
		t.Fatalf("ParseCommand(%q) error = %v", line, err)
	}
	if got != want {
		t.Errorf("ParseCommand(%q) = %#v, want %#v", line, got, want)
	}
}

func TestParseCommandImportRejectsUnknownSource(t *testing.T) {
	_, err := ParseCommand("import bing ~/x.csv")
	if !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("ParseCommand() error = %v, want %v", err, ErrUnknownCommand)
	}
}

func TestParseCommandImportRequiresPath(t *testing.T) {
	_, err := ParseCommand("import google")
	if !errors.Is(err, ErrMissingArgument) {
		t.Errorf("ParseCommand() error = %v, want %v", err, ErrMissingArgument)
	}
}

func TestParseCommandSet(t *testing.T) {
	line := "set clipboard-timeout 15s"
	want := SetCommand{Key: "clipboard-timeout", Value: "15s"}

	got, err := ParseCommand(line)
	if err != nil {
		t.Fatalf("ParseCommand(%q) error = %v", line, err)
	}
	if got != want {
		t.Errorf("ParseCommand(%q) = %#v, want %#v", line, got, want)
	}
}

func TestParseCommandSetRequiresKeyAndValue(t *testing.T) {
	cases := []string{"set", "set clipboard-timeout"}
	for _, line := range cases {
		t.Run(line, func(t *testing.T) {
			_, err := ParseCommand(line)
			if !errors.Is(err, ErrMissingArgument) {
				t.Errorf("ParseCommand(%q) error = %v, want %v", line, err, ErrMissingArgument)
			}
		})
	}
}

func TestParseCommandEmptyInput(t *testing.T) {
	cases := []string{"", "   ", "\t"}
	for _, line := range cases {
		_, err := ParseCommand(line)
		if !errors.Is(err, ErrEmptyCommand) {
			t.Errorf("ParseCommand(%q) error = %v, want %v", line, err, ErrEmptyCommand)
		}
	}
}

func TestParseCommandUnknown(t *testing.T) {
	_, err := ParseCommand("frobnicate")
	if !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("ParseCommand() error = %v, want %v", err, ErrUnknownCommand)
	}
}

func TestParseCommandMissingPathArguments(t *testing.T) {
	cases := []string{"open", "new", "save-as", "search"}
	for _, line := range cases {
		t.Run(line, func(t *testing.T) {
			_, err := ParseCommand(line)
			if !errors.Is(err, ErrMissingArgument) {
				t.Errorf("ParseCommand(%q) error = %v, want %v", line, err, ErrMissingArgument)
			}
		})
	}
}

func TestParseCommandToleratesSurroundingWhitespace(t *testing.T) {
	got, err := ParseCommand("   open   ~/x.kdbx  ")
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	want := OpenCommand{Path: "~/x.kdbx"}
	if got != want {
		t.Errorf("ParseCommand() = %#v, want %#v", got, want)
	}
}
