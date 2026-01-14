package config

import (
	"path/filepath"
	"testing"
)

func TestGetDBPath(t *testing.T) {
	cfg := &Config{Path: "/home/user/.bujo"}

	got := cfg.GetDBPath()
	want := filepath.Join("/home/user/.bujo", "db")

	if got != want {
		t.Errorf("GetDBPath() = %q, want %q", got, want)
	}
}

func TestGetJournalPath(t *testing.T) {
	cfg := &Config{Path: "/home/user/.bujo"}

	got := cfg.GetJournalPath()
	want := filepath.Join("/home/user/.bujo", "journal")

	if got != want {
		t.Errorf("GetJournalPath() = %q, want %q", got, want)
	}
}
