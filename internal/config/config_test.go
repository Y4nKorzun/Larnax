package config

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultConfigIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Errorf("Default().Validate() = %v, want nil", err)
	}
}

func TestValidateAutoLockAllowsDisabled(t *testing.T) {
	c := Default()
	c.Security.AutoLock = 0
	if err := c.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil (0 means disabled)", err)
	}
}

func TestValidateAutoLockRejectsBelowMinimum(t *testing.T) {
	c := Default()
	c.Security.AutoLock = 30 * time.Second
	if err := c.Validate(); !errors.Is(err, ErrInvalidAutoLock) {
		t.Errorf("Validate() = %v, want %v", err, ErrInvalidAutoLock)
	}
}

func TestValidateAutoLockRejectsAboveMaximum(t *testing.T) {
	c := Default()
	c.Security.AutoLock = 61 * time.Minute
	if err := c.Validate(); !errors.Is(err, ErrInvalidAutoLock) {
		t.Errorf("Validate() = %v, want %v", err, ErrInvalidAutoLock)
	}
}

func TestValidateAutoLockAllowsBoundaries(t *testing.T) {
	for _, d := range []time.Duration{time.Minute, 60 * time.Minute} {
		c := Default()
		c.Security.AutoLock = d
		if err := c.Validate(); err != nil {
			t.Errorf("Validate() with AutoLock=%v = %v, want nil", d, err)
		}
	}
}

func TestValidateClipboardTimeoutAllowsDisabled(t *testing.T) {
	c := Default()
	c.Security.ClipboardTimeout = 0
	if err := c.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil (0 means disabled)", err)
	}
}

func TestValidateClipboardTimeoutRejectsBelowMinimum(t *testing.T) {
	c := Default()
	c.Security.ClipboardTimeout = 4 * time.Second
	if err := c.Validate(); !errors.Is(err, ErrInvalidClipboardTimeout) {
		t.Errorf("Validate() = %v, want %v", err, ErrInvalidClipboardTimeout)
	}
}

func TestValidateClipboardTimeoutRejectsAboveMaximum(t *testing.T) {
	c := Default()
	c.Security.ClipboardTimeout = 121 * time.Second
	if err := c.Validate(); !errors.Is(err, ErrInvalidClipboardTimeout) {
		t.Errorf("Validate() = %v, want %v", err, ErrInvalidClipboardTimeout)
	}
}

func TestValidateClipboardTimeoutAllowsBoundaries(t *testing.T) {
	for _, d := range []time.Duration{5 * time.Second, 120 * time.Second} {
		c := Default()
		c.Security.ClipboardTimeout = d
		if err := c.Validate(); err != nil {
			t.Errorf("Validate() with ClipboardTimeout=%v = %v, want nil", d, err)
		}
	}
}

func TestValidateRevealTimeoutRejectsNonPositive(t *testing.T) {
	for _, d := range []time.Duration{0, -1 * time.Second} {
		c := Default()
		c.Security.RevealTimeout = d
		if err := c.Validate(); !errors.Is(err, ErrInvalidRevealTimeout) {
			t.Errorf("Validate() with RevealTimeout=%v = %v, want %v", d, err, ErrInvalidRevealTimeout)
		}
	}
}

func TestValidateGeneratorLengthBounds(t *testing.T) {
	cases := []struct {
		length  int
		wantErr error
	}{
		{0, ErrInvalidGeneratorLength},
		{1, nil},
		{128, nil},
		{129, ErrInvalidGeneratorLength},
	}
	for _, c := range cases {
		cfg := Default()
		cfg.Generator.DefaultLength = c.length
		err := cfg.Validate()
		if c.wantErr == nil && err != nil {
			t.Errorf("Validate() with length=%d = %v, want nil", c.length, err)
		}
		if c.wantErr != nil && !errors.Is(err, c.wantErr) {
			t.Errorf("Validate() with length=%d = %v, want %v", c.length, err, c.wantErr)
		}
	}
}

func TestValidateBackupRetentionRejectsNonPositive(t *testing.T) {
	for _, r := range []int{0, -1} {
		c := Default()
		c.Backups.Retention = r
		if err := c.Validate(); !errors.Is(err, ErrInvalidBackupRetention) {
			t.Errorf("Validate() with Retention=%d = %v, want %v", r, err, ErrInvalidBackupRetention)
		}
	}
}

func TestValidateRejectsEmptyLeader(t *testing.T) {
	for _, leader := range []string{"", "   "} {
		c := Default()
		c.UI.Leader = leader
		if err := c.Validate(); !errors.Is(err, ErrEmptyLeader) {
			t.Errorf("Validate() with Leader=%q = %v, want %v", leader, err, ErrEmptyLeader)
		}
	}
}
