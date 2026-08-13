package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveThenLoadRoundTripsDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	want := Default()

	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestSaveProducesSpecShapedTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Save(path, Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	content := string(data)

	for _, want := range []string{
		"[ui]", `leader = "space"`, `color_mode = "auto"`,
		"[security]", `auto_lock = "15m"`, `clipboard_timeout = "15s"`, `reveal_timeout = "5s"`,
		"[generator]", "default_length = 24",
		"[backups]", "retention = 10",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("saved TOML missing %q; full content:\n%s", want, content)
		}
	}
}

func TestSaveDisabledDurationsRoundTripToZero(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := Default()
	cfg.Security.AutoLock = 0
	cfg.Security.ClipboardTimeout = 0

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Security.AutoLock != 0 || got.Security.ClipboardTimeout != 0 {
		t.Errorf("disabled durations = %v / %v, want 0 / 0", got.Security.AutoLock, got.Security.ClipboardTimeout)
	}
}

func TestSaveCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "config.toml")
	if err := Save(path, Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("saved file not found: %v", err)
	}
}

func TestLoadRejectsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.toml")
	if _, err := Load(path); err == nil {
		t.Error("Load() of a missing file succeeded, want error")
	}
}

func TestLoadRejectsMalformedTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("this is not [valid toml"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Error("Load() of malformed TOML succeeded, want error")
	}
}

func TestLoadRejectsConfigFailingValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := Default()
	cfg.Security.ClipboardTimeout = 2 * time.Second // below the 5s spec minimum
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := Load(path); !errors.Is(err, ErrInvalidClipboardTimeout) {
		t.Errorf("Load() error = %v, want %v", err, ErrInvalidClipboardTimeout)
	}
}

func TestFormatDurationMatchesSpecExamples(t *testing.T) {
	cases := map[time.Duration]string{
		0:                "0s",
		15 * time.Minute: "15m",
		15 * time.Second: "15s",
		5 * time.Second:  "5s",
		90 * time.Minute: "90m",
	}
	for d, want := range cases {
		if got := formatDuration(d); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", d, got, want)
		}
	}
}
