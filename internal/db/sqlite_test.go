package storage

import (
	"testing"
	"time"

	"github.com/samakintunde/bujo-cli/internal/models"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	store, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Error("New() returned store with nil db")
	}
}

func TestSyncAndGetEntries(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	entries := []models.Entry{
		{
			ID:         "id1",
			Type:       models.EntryTypeTask,
			Status:     models.EntryStatusOpen,
			Content:    "First task",
			RawContent: "- [ ] First task",
			FilePath:   "/test/2024-01-01.md",
			LineNumber: 1,
		},
		{
			ID:         "id2",
			Type:       models.EntryTypeEvent,
			Status:     models.EntryStatusOpen,
			Content:    "Meeting",
			RawContent: "- * Meeting",
			FilePath:   "/test/2024-01-01.md",
			LineNumber: 2,
		},
	}

	err = store.SyncEntries("/test/2024-01-01.md", entries)
	if err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	got, err := store.GetEntriesByFile("/test/2024-01-01.md")
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("GetEntriesByFile() returned %d entries, want 2", len(got))
	}

	if got[0].ID != "id1" || got[0].Content != "First task" {
		t.Errorf("got[0] = %+v, want id1/First task", got[0])
	}
	if got[1].ID != "id2" || got[1].Content != "Meeting" {
		t.Errorf("got[1] = %+v, want id2/Meeting", got[1])
	}
}

func TestSyncReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	path := "/test/file.md"

	first := []models.Entry{
		{ID: "old1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Old", RawContent: "- [ ] Old", FilePath: path, LineNumber: 1},
		{ID: "old2", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Also old", RawContent: "- [ ] Also old", FilePath: path, LineNumber: 2},
	}
	if err := store.SyncEntries(path, first); err != nil {
		t.Fatalf("first SyncEntries() error: %v", err)
	}

	second := []models.Entry{
		{ID: "new1", Type: models.EntryTypeNote, Status: models.EntryStatusOpen, Content: "New", RawContent: "- New", FilePath: path, LineNumber: 1},
	}
	if err := store.SyncEntries(path, second); err != nil {
		t.Fatalf("second SyncEntries() error: %v", err)
	}

	got, err := store.GetEntriesByFile(path)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("GetEntriesByFile() returned %d entries, want 1 (old entries should be deleted)", len(got))
	}
	if got[0].ID != "new1" {
		t.Errorf("got[0].ID = %q, want %q", got[0].ID, "new1")
	}
}

func TestSyncSkipsIgnoredEntries(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	path := "/test/file.md"
	entries := []models.Entry{
		{ID: "task1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task", RawContent: "- [ ] Task", FilePath: path, LineNumber: 1},
		{ID: "", Type: models.EntryTypeIgnore, Content: "", RawContent: "# Heading", FilePath: path, LineNumber: 2},
		{ID: "note1", Type: models.EntryTypeNote, Status: models.EntryStatusOpen, Content: "Note", RawContent: "- Note", FilePath: path, LineNumber: 3},
	}

	if err := store.SyncEntries(path, entries); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	got, err := store.GetEntriesByFile(path)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("GetEntriesByFile() returned %d entries, want 2 (ignored entry should be skipped)", len(got))
	}
}

func TestGetFileLastSync(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	path := "/test/file.md"

	lastSync, err := store.GetFileLastSync(path)
	if err != nil {
		t.Fatalf("GetFileLastSync() error: %v", err)
	}
	if !lastSync.IsZero() {
		t.Errorf("GetFileLastSync() for unknown file = %v, want zero time", lastSync)
	}

	before := time.Now().Add(-time.Second)
	entries := []models.Entry{
		{ID: "id1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task", RawContent: "- [ ] Task", FilePath: path, LineNumber: 1},
	}
	if err := store.SyncEntries(path, entries); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	lastSync, err = store.GetFileLastSync(path)
	if err != nil {
		t.Fatalf("GetFileLastSync() after sync error: %v", err)
	}
	if lastSync.Before(before) {
		t.Errorf("GetFileLastSync() = %v, want after %v", lastSync, before)
	}
}
