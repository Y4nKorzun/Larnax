package clipboard

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
)

type fakeClipboard struct {
	content []byte

	reads, writes, clears       int
	readErr, writeErr, clearErr error
}

func (f *fakeClipboard) ReadText(ctx context.Context) ([]byte, error) {
	f.reads++
	if f.readErr != nil {
		return nil, f.readErr
	}
	out := make([]byte, len(f.content))
	copy(out, f.content)
	return out, nil
}

func (f *fakeClipboard) WriteText(ctx context.Context, value []byte) error {
	f.writes++
	if f.writeErr != nil {
		return f.writeErr
	}
	f.content = append([]byte(nil), value...)
	return nil
}

func (f *fakeClipboard) Clear(ctx context.Context) error {
	f.clears++
	if f.clearErr != nil {
		return f.clearErr
	}
	f.content = nil
	return nil
}

func TestSecureCopyWritesAndReturnsMatchingHash(t *testing.T) {
	cb := &fakeClipboard{}
	secret := []byte("hunter2")

	hash, err := SecureCopy(context.Background(), cb, secret)
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}
	if cb.writes != 1 {
		t.Errorf("writes = %d, want 1", cb.writes)
	}
	if string(cb.content) != "hunter2" {
		t.Errorf("clipboard content = %q, want %q", cb.content, "hunter2")
	}
	if want := sha256.Sum256(secret); hash != want {
		t.Errorf("hash = %x, want %x", hash, want)
	}
}

func TestSecureCopyPropagatesWriteError(t *testing.T) {
	sentinel := errors.New("clipboard unavailable")
	cb := &fakeClipboard{writeErr: sentinel}

	_, err := SecureCopy(context.Background(), cb, []byte("hunter2"))
	if !errors.Is(err, sentinel) {
		t.Errorf("SecureCopy() error = %v, want %v", err, sentinel)
	}
}

func TestClearIfOwnedClearsWhenClipboardUnchanged(t *testing.T) {
	cb := &fakeClipboard{}
	ctx := context.Background()

	hash, err := SecureCopy(ctx, cb, []byte("hunter2"))
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}

	if err := ClearIfOwned(ctx, cb, hash); err != nil {
		t.Fatalf("ClearIfOwned() error = %v", err)
	}
	if cb.clears != 1 {
		t.Errorf("clears = %d, want 1", cb.clears)
	}
	if cb.content != nil {
		t.Errorf("clipboard content = %q, want cleared", cb.content)
	}
}

// TestClearIfOwnedLeavesNewerCopyAlone is the exact scenario spec section
// 11.3 calls out by name: a naive "sleep then clear" would delete text the
// user copied after the secret. This proves the ownership-hash check
// prevents that.
func TestClearIfOwnedLeavesNewerCopyAlone(t *testing.T) {
	cb := &fakeClipboard{}
	ctx := context.Background()

	secretHash, err := SecureCopy(ctx, cb, []byte("hunter2"))
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}

	// The user copies something unrelated before the timeout fires.
	if err := cb.WriteText(ctx, []byte("totally unrelated text")); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}

	if err := ClearIfOwned(ctx, cb, secretHash); err != nil {
		t.Fatalf("ClearIfOwned() error = %v", err)
	}
	if cb.clears != 0 {
		t.Errorf("clears = %d, want 0 (clipboard was no longer ours)", cb.clears)
	}
	if string(cb.content) != "totally unrelated text" {
		t.Errorf("clipboard content = %q, want unchanged", cb.content)
	}
}

func TestClearIfOwnedTracksMostRecentCopy(t *testing.T) {
	cb := &fakeClipboard{}
	ctx := context.Background()

	hashA, err := SecureCopy(ctx, cb, []byte("secretA"))
	if err != nil {
		t.Fatalf("SecureCopy(A) error = %v", err)
	}
	hashB, err := SecureCopy(ctx, cb, []byte("secretB"))
	if err != nil {
		t.Fatalf("SecureCopy(B) error = %v", err)
	}

	// A's timeout firing after B was copied must not clear B's value.
	if err := ClearIfOwned(ctx, cb, hashA); err != nil {
		t.Fatalf("ClearIfOwned(A) error = %v", err)
	}
	if cb.clears != 0 {
		t.Fatalf("clears = %d after stale ClearIfOwned(A), want 0", cb.clears)
	}

	// B's own timeout firing must clear it.
	if err := ClearIfOwned(ctx, cb, hashB); err != nil {
		t.Fatalf("ClearIfOwned(B) error = %v", err)
	}
	if cb.clears != 1 {
		t.Errorf("clears = %d after ClearIfOwned(B), want 1", cb.clears)
	}
}

func TestClearIfOwnedNoOpOnEmptyClipboard(t *testing.T) {
	cb := &fakeClipboard{}
	ctx := context.Background()

	hash, err := SecureCopy(ctx, cb, []byte("hunter2"))
	if err != nil {
		t.Fatalf("SecureCopy() error = %v", err)
	}

	cb.content = nil // clipboard was cleared by something else entirely

	if err := ClearIfOwned(ctx, cb, hash); err != nil {
		t.Fatalf("ClearIfOwned() error = %v", err)
	}
	if cb.clears != 0 {
		t.Errorf("clears = %d, want 0", cb.clears)
	}
}

func TestClearIfOwnedPropagatesReadError(t *testing.T) {
	sentinel := errors.New("clipboard unavailable")
	cb := &fakeClipboard{readErr: sentinel}

	err := ClearIfOwned(context.Background(), cb, sha256.Sum256([]byte("hunter2")))
	if !errors.Is(err, sentinel) {
		t.Errorf("ClearIfOwned() error = %v, want %v", err, sentinel)
	}
	if cb.clears != 0 {
		t.Errorf("clears = %d, want 0", cb.clears)
	}
}
