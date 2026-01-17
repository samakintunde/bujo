package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samakintunde/bujo-cli/internal/storage"
)

func setupSyncer(t *testing.T) (string, *Syncer) {
	dir := t.TempDir()
	db, err := storage.NewDBStore(dir)
	if err != nil {
		t.Fatalf("New DB error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	syncer := NewSyncer(dir, db)
	return dir, syncer
}

func TestSync_SyncsMarkdownFiles(t *testing.T) {
	dir, syncer := setupSyncer(t)

	mdPath := filepath.Join(dir, "2024-01-15.md")
	content := `- [ ] First task
- * An event
- A note`
	if err := os.WriteFile(mdPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := syncer.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	entries, err := syncer.DB.GetEntriesByFile(mdPath)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	if entries[0].Content != "First task" {
		t.Errorf("entries[0].Content = %q, want %q", entries[0].Content, "First task")
	}
}

func TestSync_SkipsHiddenDirectories(t *testing.T) {
	dir, syncer := setupSyncer(t)

	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	hiddenFile := filepath.Join(gitDir, "config.md")
	if err := os.WriteFile(hiddenFile, []byte("- [ ] Should be ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	visiblePath := filepath.Join(dir, "visible.md")
	if err := os.WriteFile(visiblePath, []byte("- [ ] Visible task"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := syncer.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	hiddenEntries, _ := syncer.DB.GetEntriesByFile(hiddenFile)
	if len(hiddenEntries) != 0 {
		t.Errorf("hidden file synced %d entries, want 0", len(hiddenEntries))
	}

	visibleEntries, _ := syncer.DB.GetEntriesByFile(visiblePath)
	if len(visibleEntries) != 1 {
		t.Errorf("visible file synced %d entries, want 1", len(visibleEntries))
	}
}

func TestSync_SkipsNonMarkdownFiles(t *testing.T) {
	dir, syncer := setupSyncer(t)

	txtPath := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(txtPath, []byte("- [ ] In txt file"), 0644); err != nil {
		t.Fatal(err)
	}

	mdPath := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(mdPath, []byte("- [ ] In md file"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := syncer.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	txtEntries, _ := syncer.DB.GetEntriesByFile(txtPath)
	if len(txtEntries) != 0 {
		t.Errorf("txt file synced %d entries, want 0", len(txtEntries))
	}

	mdEntries, _ := syncer.DB.GetEntriesByFile(mdPath)
	if len(mdEntries) != 1 {
		t.Errorf("md file synced %d entries, want 1", len(mdEntries))
	}
}

func TestSync_AssignsIDsAndWritesBack(t *testing.T) {
	dir, syncer := setupSyncer(t)

	mdPath := filepath.Join(dir, "2024-01-15.md")
	content := "- [ ] Task without ID"
	if err := os.WriteFile(mdPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := syncer.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	entries, err := syncer.DB.GetEntriesByFile(mdPath)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}

	if entries[0].ID == "" {
		t.Error("entry ID should be assigned")
	}

	updatedContent, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updatedContent), `<!-- {"id":"`) {
		t.Errorf("file should contain ID metadata, got: %s", updatedContent)
	}
}

func TestSync_PreservesExistingIDs(t *testing.T) {
	dir, syncer := setupSyncer(t)

	mdPath := filepath.Join(dir, "2024-01-15.md")
	content := `- [ ] Task with ID <!-- {"id":"existing123"} -->`
	if err := os.WriteFile(mdPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := syncer.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	entries, err := syncer.DB.GetEntriesByFile(mdPath)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if entries[0].ID != "existing123" {
		t.Errorf("entry ID = %q, want %q", entries[0].ID, "existing123")
	}
}
