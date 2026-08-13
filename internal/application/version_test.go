package application

import "testing"

func TestVersionIsNonEmpty(t *testing.T) {
	if Version == "" {
		t.Error("Version is empty")
	}
}
