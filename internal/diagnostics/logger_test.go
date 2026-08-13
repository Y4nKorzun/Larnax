package diagnostics

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func newTestLogger(t *testing.T) (*Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	return New(slog.New(handler)), &buf
}

func lastRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var record map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &record); err != nil {
		t.Fatalf("unmarshalling log line %q: %v", lines[len(lines)-1], err)
	}
	return record
}

func TestHashPathIsDeterministic(t *testing.T) {
	if HashPath("/home/user/personal.kdbx") != HashPath("/home/user/personal.kdbx") {
		t.Error("HashPath() is not deterministic for the same input")
	}
}

func TestHashPathDiffersForDifferentPaths(t *testing.T) {
	if HashPath("/home/user/personal.kdbx") == HashPath("/home/user/work.kdbx") {
		t.Error("HashPath() produced the same hash for two different paths")
	}
}

func TestHashPathLength(t *testing.T) {
	if got := len(HashPath("/home/user/personal.kdbx")); got != 16 {
		t.Errorf("len(HashPath()) = %d, want 16", got)
	}
}

func TestOperationLogsSuccessWithoutErrorFields(t *testing.T) {
	logger, buf := newTestLogger(t)
	hash := HashPath("/home/user/personal.kdbx")

	logger.Operation("open", hash, "")

	record := lastRecord(t, buf)
	if record["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", record["level"])
	}
	if record["op"] != "open" {
		t.Errorf("op = %v, want open", record["op"])
	}
	if record["path_hash"] != hash {
		t.Errorf("path_hash = %v, want %v", record["path_hash"], hash)
	}
	if _, present := record["category"]; present {
		t.Error("category key present on a successful operation")
	}
}

func TestOperationLogsFailureCategoryNotRawError(t *testing.T) {
	logger, buf := newTestLogger(t)
	hash := HashPath("/home/user/personal.kdbx")

	logger.Operation("save", hash, "disk-full")

	record := lastRecord(t, buf)
	if record["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", record["level"])
	}
	if record["category"] != "disk-full" {
		t.Errorf("category = %v, want disk-full", record["category"])
	}
}

func TestVersionLogsFields(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.Version("0.1.0", "go1.26.5", "darwin", "arm64")

	record := lastRecord(t, buf)
	if record["app"] != "0.1.0" || record["go"] != "go1.26.5" || record["os"] != "darwin" || record["arch"] != "arm64" {
		t.Errorf("record = %v, want app/go/os/arch fields set", record)
	}
}

func TestKDBXVersionLogsField(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.KDBXVersion("4.1")

	record := lastRecord(t, buf)
	if record["version"] != "4.1" {
		t.Errorf("version = %v, want 4.1", record["version"])
	}
}

func TestEntryCountsLogsFields(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.EntryCounts(42, 7)

	record := lastRecord(t, buf)
	if record["entries"] != float64(42) || record["groups"] != float64(7) {
		t.Errorf("record = %v, want entries=42 groups=7", record)
	}
}

func TestTimingLogsMilliseconds(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.Timing("save", 250*time.Millisecond)

	record := lastRecord(t, buf)
	if record["operation"] != "save" {
		t.Errorf("operation = %v, want save", record["operation"])
	}
	if record["duration_ms"] != float64(250) {
		t.Errorf("duration_ms = %v, want 250", record["duration_ms"])
	}
}

func TestErrorCategoryLogsField(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.ErrorCategory("hmac-mismatch")

	record := lastRecord(t, buf)
	if record["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", record["level"])
	}
	if record["category"] != "hmac-mismatch" {
		t.Errorf("category = %v, want hmac-mismatch", record["category"])
	}
}

func TestFeatureFlagsLogsEachFlag(t *testing.T) {
	logger, buf := newTestLogger(t)
	logger.FeatureFlags(map[string]bool{"totp": true})

	record := lastRecord(t, buf)
	if record["totp"] != true {
		t.Errorf("totp = %v, want true", record["totp"])
	}
}

func TestNewWithNilUsesDefaultWithoutPanicking(t *testing.T) {
	logger := New(nil)
	logger.EntryCounts(0, 0) // must not panic
}
