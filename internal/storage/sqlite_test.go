package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samakintunde/bujo/internal/models"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	store, err := NewDBStore(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Error("New() returned store with nil db")
	}
}

func TestNewDBStore_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "subdir")
	if _, err := os.Stat(nonExistentDir); !os.IsNotExist(err) {
		t.Fatalf("setup: expected %s to not exist", nonExistentDir)
	}

	store, err := NewDBStore(nonExistentDir)
	if err != nil {
		t.Fatalf("NewDBStore() with non-existent dir failed: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("NewDBStore() did not create directory %s", nonExistentDir)
	}
}

func TestSyncAndGetEntries(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDBStore(dir)
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
	store, err := NewDBStore(dir)
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
	store, err := NewDBStore(dir)
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
	store, err := NewDBStore(dir)
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

func TestCountStaleTasks(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDBStore(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	yesterday := time.Now().AddDate(0, 0, -1)
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	tenDaysAgo := time.Now().AddDate(0, 0, -10)

	entries := []models.Entry{
		{ID: "t1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task 1", RawContent: "- [ ] Task 1", FilePath: "/test/old.md", LineNumber: 1, CreatedAt: yesterday},
		{ID: "t2", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task 2", RawContent: "- [ ] Task 2", FilePath: "/test/old.md", LineNumber: 2, CreatedAt: threeDaysAgo},
		{ID: "t3", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task 3", RawContent: "- [ ] Task 3", FilePath: "/test/old.md", LineNumber: 3, CreatedAt: tenDaysAgo},
		{ID: "t4", Type: models.EntryTypeTask, Status: models.EntryStatusCompleted, Content: "Done", RawContent: "- [x] Done", FilePath: "/test/old.md", LineNumber: 4, CreatedAt: yesterday},
		{ID: "t5", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Today", RawContent: "- [ ] Today", FilePath: "/test/today.md", LineNumber: 1, CreatedAt: time.Now()},
	}

	if err := store.SyncEntries("/test/old.md", entries[:4]); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}
	if err := store.SyncEntries("/test/today.md", entries[4:]); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	count, err := store.CountStaleTasks(0)
	if err != nil {
		t.Fatalf("CountStaleTasks(0) error: %v", err)
	}
	if count != 3 {
		t.Errorf("CountStaleTasks(0) = %d, want 3 (all stale open tasks)", count)
	}

	count, err = store.CountStaleTasks(2)
	if err != nil {
		t.Fatalf("CountStaleTasks(2) error: %v", err)
	}
	if count != 1 {
		t.Errorf("CountStaleTasks(2) = %d, want 1 (only yesterday's task)", count)
	}

	count, err = store.CountStaleTasks(7)
	if err != nil {
		t.Fatalf("CountStaleTasks(7) error: %v", err)
	}
	if count != 2 {
		t.Errorf("CountStaleTasks(7) = %d, want 2 (yesterday + 3 days ago)", count)
	}
}

func TestGetStaleTasks(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDBStore(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	yesterday := time.Now().AddDate(0, 0, -1)
	threeDaysAgo := time.Now().AddDate(0, 0, -3)

	entries := []models.Entry{
		{ID: "t1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Yesterday task", RawContent: "- [ ] Yesterday task", FilePath: "/test/old.md", LineNumber: 1, CreatedAt: yesterday},
		{ID: "t2", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Old task", RawContent: "- [ ] Old task", FilePath: "/test/old.md", LineNumber: 2, CreatedAt: threeDaysAgo},
	}

	if err := store.SyncEntries("/test/old.md", entries); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	tasks, err := store.GetStaleTasks(0)
	if err != nil {
		t.Fatalf("GetStaleTasks(0) error: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("GetStaleTasks(0) returned %d tasks, want 2", len(tasks))
	}
	if len(tasks) >= 2 && tasks[0].ID != "t2" {
		t.Errorf("GetStaleTasks(0) first task = %s, want t2 (oldest first)", tasks[0].ID)
	}

	tasks, err = store.GetStaleTasks(2)
	if err != nil {
		t.Fatalf("GetStaleTasks(2) error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("GetStaleTasks(2) returned %d tasks, want 1", len(tasks))
	}
}

func TestLastOpenedAt(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDBStore(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	lastOpened, err := store.GetLastOpenedAt()
	if err != nil {
		t.Fatalf("GetLastOpenedAt() error: %v", err)
	}
	if !lastOpened.IsZero() {
		t.Errorf("GetLastOpenedAt() for fresh db = %v, want zero time", lastOpened)
	}

	today := time.Now()
	if err := store.SetLastOpenedAt(today); err != nil {
		t.Fatalf("SetLastOpenedAt() error: %v", err)
	}

	lastOpened, err = store.GetLastOpenedAt()
	if err != nil {
		t.Fatalf("GetLastOpenedAt() after set error: %v", err)
	}

	if lastOpened.Format(time.DateOnly) != today.Format(time.DateOnly) {
		t.Errorf("GetLastOpenedAt() = %v, want %v", lastOpened.Format(time.DateOnly), today.Format(time.DateOnly))
	}
}

func TestUpdateEntryStatus(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDBStore(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()

	path := "/test/file.md"
	entries := []models.Entry{
		{ID: "task1", Type: models.EntryTypeTask, Status: models.EntryStatusOpen, Content: "Task", RawContent: "- [ ] Task", FilePath: path, LineNumber: 1},
	}

	if err := store.SyncEntries(path, entries); err != nil {
		t.Fatalf("SyncEntries() error: %v", err)
	}

	if err := store.UpdateEntryStatus("task1", models.EntryStatusCompleted); err != nil {
		t.Fatalf("UpdateEntryStatus() error: %v", err)
	}

	got, err := store.GetEntriesByFile(path)
	if err != nil {
		t.Fatalf("GetEntriesByFile() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("GetEntriesByFile() returned %d entries, want 1", len(got))
	}
	if got[0].Status != models.EntryStatusCompleted {
		t.Errorf("Entry status = %s, want %s", got[0].Status, models.EntryStatusCompleted)
	}
}
