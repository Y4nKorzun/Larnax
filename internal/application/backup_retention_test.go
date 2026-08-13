package application

import (
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/config"
)

func TestBackupRetentionDisabledReturnsZero(t *testing.T) {
	cfg := config.BackupsConfig{Enabled: false, Retention: 10}
	if got := BackupRetention(cfg); got != 0 {
		t.Errorf("BackupRetention() = %d, want 0", got)
	}
}

func TestBackupRetentionEnabledReturnsConfiguredValue(t *testing.T) {
	cfg := config.BackupsConfig{Enabled: true, Retention: 5}
	if got := BackupRetention(cfg); got != 5 {
		t.Errorf("BackupRetention() = %d, want 5", got)
	}
}

func TestBackupRetentionWithDefaultConfig(t *testing.T) {
	got := BackupRetention(config.Default().Backups)
	if got != 10 {
		t.Errorf("BackupRetention(Default()) = %d, want 10", got)
	}
}
