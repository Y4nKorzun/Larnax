package application

import "github.com/Y4nKorzun/Larnax/internal/config"

// BackupRetention converts config's backups section into the retention
// int SaveVault expects. SaveVault treats retention <= 0 as "disabled"
// (see its own doc comment), while config.BackupsConfig separates that
// into an explicit Enabled flag plus a Retention count that's always
// meaningful on its own terms — config.Validate rejects Retention < 1
// regardless of Enabled, so a disabled config can still carry a real
// number a re-enable would want back. This is the one place that gap
// between the two gets closed, rather than every SaveVault call site
// reimplementing "if not enabled, pass 0."
func BackupRetention(cfg config.BackupsConfig) int {
	if !cfg.Enabled {
		return 0
	}
	return cfg.Retention
}
