package id

import (
	"testing"
)

func TestNew(t *testing.T) {
	id1 := New()
	id2 := New()

	if id1 == "" {
		t.Error("New() returned empty string")
	}

	if len(id1) != 26 {
		t.Errorf("New() length = %d, want 26", len(id1))
	}

	if id1 == id2 {
		t.Errorf("New() returned duplicate: %s", id1)
	}
}
