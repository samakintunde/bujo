package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRepo(t *testing.T) {
	dir := t.TempDir()

	if IsRepo(dir) {
		t.Error("IsRepo() = true for empty dir, want false")
	}

	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !IsRepo(dir) {
		t.Error("IsRepo() = false after .git created, want true")
	}
}
