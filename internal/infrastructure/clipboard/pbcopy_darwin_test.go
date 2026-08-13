//go:build darwin

// These tests exercise the real macOS system pasteboard (spec section
// 24.5's per-platform clipboard integration tests) rather than a fake —
// running them overwrites whatever is currently on the developer's
// clipboard, the same tradeoff any clipboard-manipulating test suite on a
// real OS makes.
package clipboard

import (
	"context"
	"os/exec"
	"testing"
)

func requirePasteboardTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pbcopy"); err != nil {
		t.Skip("pbcopy not available in PATH")
	}
	if _, err := exec.LookPath("pbpaste"); err != nil {
		t.Skip("pbpaste not available in PATH")
	}
}

func TestDarwinClipboardImplementsInterface(t *testing.T) {
	var _ Clipboard = DarwinClipboard{}
}

func TestDarwinClipboardWriteThenReadBack(t *testing.T) {
	requirePasteboardTools(t)
	cb := DarwinClipboard{}
	ctx := context.Background()

	if err := cb.WriteText(ctx, []byte("hello from Larnax test")); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	got, err := cb.ReadText(ctx)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if string(got) != "hello from Larnax test" {
		t.Errorf("ReadText() = %q, want %q", got, "hello from Larnax test")
	}
}

func TestDarwinClipboardWriteThenReadBackUnicode(t *testing.T) {
	requirePasteboardTools(t)
	cb := DarwinClipboard{}
	ctx := context.Background()
	const value = "пароль-🔒-密码"

	if err := cb.WriteText(ctx, []byte(value)); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	got, err := cb.ReadText(ctx)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if string(got) != value {
		t.Errorf("ReadText() = %q, want %q", got, value)
	}
}

func TestDarwinClipboardClearEmptiesPasteboard(t *testing.T) {
	requirePasteboardTools(t)
	cb := DarwinClipboard{}
	ctx := context.Background()

	if err := cb.WriteText(ctx, []byte("something")); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	if err := cb.Clear(ctx); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	got, err := cb.ReadText(ctx)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ReadText() after Clear() = %q, want empty", got)
	}
}

// TestDarwinClipboardSecureCopyThenClearIfOwned exercises the real
// pasteboard through the same SecureCopy/ClearIfOwned pair copy_field.go
// uses in production — spec section 11.3/11.4's "copy, then clear after
// timeout" flow end to end, not just against the in-memory fake
// application's own tests use.
func TestDarwinClipboardSecureCopyThenClearIfOwned(t *testing.T) {
	requirePasteboardTools(t)
	cb := DarwinClipboard{}
	ctx := context.Background()

	hash, err := SecureCopy(ctx, cb, []byte("s3cr3t-password"))
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}
	if err := ClearIfOwned(ctx, cb, hash); err != nil {
		t.Fatalf("ClearIfOwned() error = %v", err)
	}
	got, err := cb.ReadText(ctx)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("clipboard not cleared: %q", got)
	}
}

// TestDarwinClipboardClearIfOwnedDoesNotClearReplacedValue is spec
// section 24.5 item 4 ("does not clear a clipboard the user changed")
// against the real pasteboard.
func TestDarwinClipboardClearIfOwnedDoesNotClearReplacedValue(t *testing.T) {
	requirePasteboardTools(t)
	cb := DarwinClipboard{}
	ctx := context.Background()

	hash, err := SecureCopy(ctx, cb, []byte("original-secret"))
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}

	// The user copies something else before the timeout fires.
	if err := cb.WriteText(ctx, []byte("user typed this after copying the secret")); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}

	if err := ClearIfOwned(ctx, cb, hash); err != nil {
		t.Fatalf("ClearIfOwned() error = %v", err)
	}
	got, err := cb.ReadText(ctx)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if string(got) != "user typed this after copying the secret" {
		t.Errorf("ClearIfOwned() cleared a replaced value: got %q", got)
	}
}
