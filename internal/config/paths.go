package config

import (
	"os"
	"path/filepath"
)

// appDirName names this application's subdirectory within the OS's
// per-user config directory. It matches the binary name from spec section
// 18.3 (cmd/kdbx-tui).
const appDirName = "kdbx-tui"

// configFileName is the TOML file spec section 21.1 documents the shape
// of. Loading it is separate, later work — this package only computes
// where it lives.
const configFileName = "config.toml"

// ConfigDir returns this application's per-user config directory:
// os.UserConfigDir's platform default (XDG_CONFIG_HOME or ~/.config on
// Linux, ~/Library/Application Support on macOS, %AppData% on Windows)
// joined with appDirName. It does not create the directory.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName), nil
}

// ConfigFilePath returns the path of the config.toml file described by
// spec section 21.1. It does not check that the file exists — spec
// section 21 has no default vault location, so callers must not infer one
// from this path either.
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}
