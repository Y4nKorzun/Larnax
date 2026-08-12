// Package tui will eventually hold the Bubble Tea application (spec section
// 19), but this file has no framework dependency: it only parses COMMAND
// mode input (spec section 8.4) into typed values.
package tui

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrEmptyCommand       = errors.New("tui: empty command")
	ErrUnknownCommand     = errors.New("tui: unknown command")
	ErrMissingArgument    = errors.New("tui: missing required argument")
	ErrUnexpectedArgument = errors.New("tui: unexpected argument")
)

// Command is a parsed COMMAND-mode command, typed so dispatch code can use
// a type switch instead of comparing strings. ParseCommand only parses —
// executing a command is Application's job (spec section 18.2's
// TUI -> Application use case -> ... chain).
type Command interface {
	isCommand()
}

type (
	WriteCommand        struct{}
	QuitCommand         struct{}
	WriteQuitCommand    struct{}
	LockCommand         struct{}
	GenerateCommand     struct{}
	HelpCommand         struct{ Topic string }
	OpenCommand         struct{ Path string }
	NewCommand          struct{ Path string }
	SaveAsCommand       struct{ Path string }
	SearchCommand       struct{ Query string }
	ImportGoogleCommand struct{ Path string }
	SetCommand          struct{ Key, Value string }
)

func (WriteCommand) isCommand()        {}
func (QuitCommand) isCommand()         {}
func (WriteQuitCommand) isCommand()    {}
func (LockCommand) isCommand()         {}
func (GenerateCommand) isCommand()     {}
func (HelpCommand) isCommand()         {}
func (OpenCommand) isCommand()         {}
func (NewCommand) isCommand()          {}
func (SaveAsCommand) isCommand()       {}
func (SearchCommand) isCommand()       {}
func (ImportGoogleCommand) isCommand() {}
func (SetCommand) isCommand()          {}

// ParseCommand parses the text a user typed in COMMAND mode after the
// leading ':', which the caller strips before calling this (spec section
// 8.4 lists the exact command set).
//
// Commands that take a path or query use the entire trimmed remainder of
// the line as a single argument rather than splitting on every space, since
// paths and search queries may themselves contain spaces — spec section
// 13.2's own example, ":import google ~/Downloads/Google Passwords.csv",
// has an unescaped space in the path.
func ParseCommand(line string) (Command, error) {
	name, rest := splitCommandLine(line)
	if name == "" {
		return nil, ErrEmptyCommand
	}

	switch name {
	case "w":
		return noRest(WriteCommand{}, rest)
	case "q":
		return noRest(QuitCommand{}, rest)
	case "wq":
		return noRest(WriteQuitCommand{}, rest)
	case "lock":
		return noRest(LockCommand{}, rest)
	case "generate":
		return noRest(GenerateCommand{}, rest)
	case "help":
		return HelpCommand{Topic: rest}, nil
	case "open":
		if rest == "" {
			return nil, fmt.Errorf("%w: path", ErrMissingArgument)
		}
		return OpenCommand{Path: rest}, nil
	case "new":
		if rest == "" {
			return nil, fmt.Errorf("%w: path", ErrMissingArgument)
		}
		return NewCommand{Path: rest}, nil
	case "save-as":
		if rest == "" {
			return nil, fmt.Errorf("%w: path", ErrMissingArgument)
		}
		return SaveAsCommand{Path: rest}, nil
	case "search":
		if rest == "" {
			return nil, fmt.Errorf("%w: query", ErrMissingArgument)
		}
		return SearchCommand{Query: rest}, nil
	case "import":
		sub, path := splitCommandLine(rest)
		if sub != "google" {
			return nil, fmt.Errorf("%w: import %q", ErrUnknownCommand, sub)
		}
		if path == "" {
			return nil, fmt.Errorf("%w: path", ErrMissingArgument)
		}
		return ImportGoogleCommand{Path: path}, nil
	case "set":
		key, value := splitCommandLine(rest)
		if key == "" || value == "" {
			return nil, fmt.Errorf("%w: set requires a key and a value", ErrMissingArgument)
		}
		return SetCommand{Key: key, Value: value}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownCommand, name)
	}
}

// splitCommandLine splits line into its first whitespace-delimited token
// and the raw remainder, with leading/trailing whitespace trimmed from
// both. Internal whitespace within the remainder is preserved verbatim.
func splitCommandLine(line string) (name, rest string) {
	line = strings.TrimSpace(line)
	i := strings.IndexFunc(line, unicode.IsSpace)
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimSpace(line[i:])
}

func noRest(cmd Command, rest string) (Command, error) {
	if rest != "" {
		return nil, fmt.Errorf("%w: %q", ErrUnexpectedArgument, rest)
	}
	return cmd, nil
}
