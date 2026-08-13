package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// tomlConfig mirrors spec section 21.1's exact TOML shape — snake_case
// keys, durations as strings — kept separate from Config so Config's own
// field types (time.Duration, not string) don't have to compromise for
// what one specific serialization needs.
type tomlConfig struct {
	UI        tomlUIConfig        `toml:"ui"`
	Security  tomlSecurityConfig  `toml:"security"`
	Generator tomlGeneratorConfig `toml:"generator"`
	Backups   tomlBackupsConfig   `toml:"backups"`
	// Keymap is spec section 21.3's override table — key string to
	// action name — omitted entirely from the file when empty rather
	// than written as an empty [keymap] section.
	Keymap map[string]string `toml:"keymap,omitempty"`
}

type tomlUIConfig struct {
	Leader    string `toml:"leader"`
	ShowIcons bool   `toml:"show_icons"`
	ColorMode string `toml:"color_mode"`
}

type tomlSecurityConfig struct {
	AutoLock              string `toml:"auto_lock"`
	ClipboardTimeout      string `toml:"clipboard_timeout"`
	AllowClipboardOverSSH bool   `toml:"allow_clipboard_over_ssh"`
	RevealTimeout         string `toml:"reveal_timeout"`
}

type tomlGeneratorConfig struct {
	DefaultLength  int  `toml:"default_length"`
	Lowercase      bool `toml:"lowercase"`
	Uppercase      bool `toml:"uppercase"`
	Digits         bool `toml:"digits"`
	Symbols        bool `toml:"symbols"`
	AvoidAmbiguous bool `toml:"avoid_ambiguous"`
}

type tomlBackupsConfig struct {
	Enabled   bool `toml:"enabled"`
	Retention int  `toml:"retention"`
}

// Load reads and parses the TOML config file at path (spec section
// 21.1's shape), returning an error if it is missing, malformed, or the
// parsed result fails Config.Validate.
func Load(path string) (Config, error) {
	var tc tomlConfig
	if _, err := toml.DecodeFile(path, &tc); err != nil {
		return Config{}, fmt.Errorf("config: decoding %s: %w", path, err)
	}

	cfg, err := fromToml(tc)
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes cfg to path as TOML matching spec section 21.1's shape,
// creating path's parent directory if it does not exist yet (e.g. a
// first run with no config directory at all). It does not call
// cfg.Validate itself — that is the caller's own decision point, such as
// before ever accepting a value from a settings screen, not something
// Save should silently re-enforce.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("config: creating directory for %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(toToml(cfg)); err != nil {
		return fmt.Errorf("config: encoding: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("config: writing %s: %w", path, err)
	}
	return nil
}

func toToml(c Config) tomlConfig {
	return tomlConfig{
		UI: tomlUIConfig{
			Leader:    c.UI.Leader,
			ShowIcons: c.UI.ShowIcons,
			ColorMode: c.UI.ColorMode,
		},
		Security: tomlSecurityConfig{
			AutoLock:              formatDuration(c.Security.AutoLock),
			ClipboardTimeout:      formatDuration(c.Security.ClipboardTimeout),
			AllowClipboardOverSSH: c.Security.AllowClipboardOverSSH,
			RevealTimeout:         formatDuration(c.Security.RevealTimeout),
		},
		Generator: tomlGeneratorConfig{
			DefaultLength:  c.Generator.DefaultLength,
			Lowercase:      c.Generator.Lowercase,
			Uppercase:      c.Generator.Uppercase,
			Digits:         c.Generator.Digits,
			Symbols:        c.Generator.Symbols,
			AvoidAmbiguous: c.Generator.AvoidAmbiguous,
		},
		Backups: tomlBackupsConfig{
			Enabled:   c.Backups.Enabled,
			Retention: c.Backups.Retention,
		},
		Keymap: c.Keymap,
	}
}

func fromToml(tc tomlConfig) (Config, error) {
	autoLock, err := time.ParseDuration(tc.Security.AutoLock)
	if err != nil {
		return Config{}, fmt.Errorf("config: parsing security.auto_lock %q: %w", tc.Security.AutoLock, err)
	}
	clipboardTimeout, err := time.ParseDuration(tc.Security.ClipboardTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("config: parsing security.clipboard_timeout %q: %w", tc.Security.ClipboardTimeout, err)
	}
	revealTimeout, err := time.ParseDuration(tc.Security.RevealTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("config: parsing security.reveal_timeout %q: %w", tc.Security.RevealTimeout, err)
	}

	return Config{
		UI: UIConfig{
			Leader:    tc.UI.Leader,
			ShowIcons: tc.UI.ShowIcons,
			ColorMode: tc.UI.ColorMode,
		},
		Security: SecurityConfig{
			AutoLock:              autoLock,
			ClipboardTimeout:      clipboardTimeout,
			AllowClipboardOverSSH: tc.Security.AllowClipboardOverSSH,
			RevealTimeout:         revealTimeout,
		},
		Generator: GeneratorConfig{
			DefaultLength:  tc.Generator.DefaultLength,
			Lowercase:      tc.Generator.Lowercase,
			Uppercase:      tc.Generator.Uppercase,
			Digits:         tc.Generator.Digits,
			Symbols:        tc.Generator.Symbols,
			AvoidAmbiguous: tc.Generator.AvoidAmbiguous,
		},
		Backups: BackupsConfig{
			Enabled:   tc.Backups.Enabled,
			Retention: tc.Backups.Retention,
		},
		Keymap: tc.Keymap,
	}, nil
}

// formatDuration renders d the way spec section 21.1's example config
// does — "15m", "15s" — rather than time.Duration.String()'s default,
// which renders exactly 15 minutes as "15m0s".
func formatDuration(d time.Duration) string {
	switch {
	case d == 0:
		return "0s"
	case d%time.Minute == 0:
		return fmt.Sprintf("%dm", d/time.Minute)
	case d%time.Second == 0:
		return fmt.Sprintf("%ds", d/time.Second)
	default:
		return d.String()
	}
}
