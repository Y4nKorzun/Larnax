package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDirJoinsAppName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.UserConfigDir reads %AppData%, not $HOME, on windows")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	if filepath.Base(dir) != appDirName {
		t.Errorf("ConfigDir() = %q, want base %q", dir, appDirName)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() = %q, want absolute path", dir)
	}
}

func TestConfigDirHonorsXDGConfigHome(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG_CONFIG_HOME only takes effect on linux")
	}

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	want := filepath.Join(xdg, appDirName)
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigFilePathIsInsideConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("AppData", home)
	}

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v", err)
	}

	if filepath.Dir(path) != dir {
		t.Errorf("ConfigFilePath() dir = %q, want %q", filepath.Dir(path), dir)
	}
	if filepath.Base(path) != configFileName {
		t.Errorf("ConfigFilePath() base = %q, want %q", filepath.Base(path), configFileName)
	}
}
