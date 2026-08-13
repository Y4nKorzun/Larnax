// Package config defines the application's local, non-secret settings
// (spec section 21): the struct shape, defaults, and validation here,
// TOML load/save in loader.go, and the default file location in paths.go.
package config

import (
	"errors"
	"strings"
	"time"
)

// Config mirrors spec section 21.1's example TOML shape.
type Config struct {
	UI        UIConfig
	Security  SecurityConfig
	Generator GeneratorConfig
	Backups   BackupsConfig
}

type UIConfig struct {
	Leader    string
	ShowIcons bool
	ColorMode string
}

// SecurityConfig's AutoLock and ClipboardTimeout use 0 to mean "disabled",
// matching spec sections 17.1 and 11.4, both of which allow the user to
// turn the corresponding timeout off entirely.
type SecurityConfig struct {
	AutoLock              time.Duration
	ClipboardTimeout      time.Duration
	AllowClipboardOverSSH bool
	RevealTimeout         time.Duration
}

type GeneratorConfig struct {
	DefaultLength  int
	Lowercase      bool
	Uppercase      bool
	Digits         bool
	Symbols        bool
	AvoidAmbiguous bool
}

type BackupsConfig struct {
	Enabled   bool
	Retention int
}

// Default matches spec section 21.1's example values exactly.
func Default() Config {
	return Config{
		UI: UIConfig{
			Leader:    "space",
			ShowIcons: true,
			ColorMode: "auto",
		},
		Security: SecurityConfig{
			AutoLock:              15 * time.Minute,
			ClipboardTimeout:      15 * time.Second,
			AllowClipboardOverSSH: false,
			RevealTimeout:         5 * time.Second,
		},
		Generator: GeneratorConfig{
			DefaultLength: 24,
			Lowercase:     true,
			Uppercase:     true,
			Digits:        true,
			Symbols:       true,
		},
		Backups: BackupsConfig{
			Enabled:   true,
			Retention: 10,
		},
	}
}

// The AutoLock and ClipboardTimeout bounds are spec-mandated: section 17.1
// states auto-lock is 1-60 minutes or disabled, and section 11.4 states
// clipboard timeout is 5-120 seconds or disabled. The remaining bounds
// (RevealTimeout, generator length, backup retention) have no spec-stated
// range — these are this package's own basic sanity floors/ceilings, not
// spec requirements, and are documented as such below.
var (
	ErrInvalidAutoLock         = errors.New("config: auto_lock must be 0 (disabled) or between 1m and 60m")
	ErrInvalidClipboardTimeout = errors.New("config: clipboard_timeout must be 0 (disabled) or between 5s and 120s")
	ErrInvalidRevealTimeout    = errors.New("config: reveal_timeout must be positive")
	ErrInvalidGeneratorLength  = errors.New("config: generator default_length must be between 1 and 128")
	ErrInvalidBackupRetention  = errors.New("config: backups retention must be at least 1")
	ErrEmptyLeader             = errors.New("config: ui leader must not be empty")
)

// Validate reports the first constraint violation found, or nil if c is
// entirely valid. Section 21.2's list of what must never appear in config
// (master passphrase, entry passwords, TOTP secrets, key file bytes,
// recovery phrase, decrypted cache, CSV data) is enforced by Config having
// no fields capable of holding any of that — there is nothing to validate
// against, because there is nowhere to put it.
func (c Config) Validate() error {
	if c.Security.AutoLock != 0 && (c.Security.AutoLock < time.Minute || c.Security.AutoLock > 60*time.Minute) {
		return ErrInvalidAutoLock
	}
	if c.Security.ClipboardTimeout != 0 && (c.Security.ClipboardTimeout < 5*time.Second || c.Security.ClipboardTimeout > 120*time.Second) {
		return ErrInvalidClipboardTimeout
	}
	// Not spec-mandated: reveal_timeout has no stated range, only that it
	// must function at all, which requires a positive duration.
	if c.Security.RevealTimeout <= 0 {
		return ErrInvalidRevealTimeout
	}
	// Not spec-mandated: 128 is a generous ceiling far beyond any
	// practical password length, chosen to catch an obvious config typo
	// rather than to enforce a real product limit.
	if c.Generator.DefaultLength < 1 || c.Generator.DefaultLength > 128 {
		return ErrInvalidGeneratorLength
	}
	// Not spec-mandated: "keep the last N backups" is meaningless for
	// N < 1, regardless of whether Backups.Enabled is true.
	if c.Backups.Retention < 1 {
		return ErrInvalidBackupRetention
	}
	if strings.TrimSpace(c.UI.Leader) == "" {
		return ErrEmptyLeader
	}
	return nil
}
