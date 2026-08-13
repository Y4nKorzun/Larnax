package clipboard

import "testing"

func TestAvailableMatchesNewSecondReturn(t *testing.T) {
	_, ok := New()
	if Available() != ok {
		t.Errorf("Available() = %v, want it to match New()'s second return value %v", Available(), ok)
	}
}

func TestNewReturnsNonNilClipboardWhenAvailable(t *testing.T) {
	cb, ok := New()
	if ok && cb == nil {
		t.Error("New() returned ok=true but a nil Clipboard")
	}
	if !ok && cb != nil {
		t.Error("New() returned ok=false but a non-nil Clipboard")
	}
}
