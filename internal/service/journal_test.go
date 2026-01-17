package service

import (
	"os"
	"testing"
	"time"

	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
)

func TestAddEntry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bujo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs, err := storage.NewFSStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	db, err := storage.NewDBStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	syncer := sync.NewSyncer(tmpDir, db)
	svc := NewJournalService(fs, db, syncer)

	content := "Test entry"
	entryType := models.EntryTypeTask
	now := time.Now()

	entry, err := svc.AddEntry(content, entryType, now)
	if err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	dateStr := now.Format(time.DateOnly)
	expectedPath := fs.GetDayPath(dateStr)
	if entry.FilePath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, entry.FilePath)
	}

	bytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(bytes) != entry.RawString()+"\n" {
		t.Errorf("File content mismatch. Got %q, want %q", string(bytes), entry.RawString()+"\n")
	}

	entries, err := db.GetEntriesByFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to get entries from DB: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry in DB, got %d", len(entries))
	}
	if entries[0].ID != entry.ID {
		t.Errorf("DB entry ID mismatch. Got %s, want %s", entries[0].ID, entry.ID)
	}
}
