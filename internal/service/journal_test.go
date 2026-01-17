package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
)

func setupTestService(t *testing.T) (*JournalService, *storage.FSStore, *storage.DBStore, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "bujo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	fs, err := storage.NewFSStore(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	db, err := storage.NewDBStore(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	syncer := sync.NewSyncer(tmpDir, db)
	svc := NewJournalService(fs, db, syncer)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return svc, fs, db, cleanup
}

func TestAddEntry(t *testing.T) {
	svc, fs, db, cleanup := setupTestService(t)
	defer cleanup()

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

func TestUpdateEntryStatus(t *testing.T) {
	svc, _, db, cleanup := setupTestService(t)
	defer cleanup()

	entry, err := svc.AddEntry("Test task", models.EntryTypeTask, time.Now())
	if err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entries, _ := db.GetEntriesByFile(entry.FilePath)
	dbEntry := entries[0]

	err = svc.UpdateEntryStatus(dbEntry, models.EntryStatusCompleted)
	if err != nil {
		t.Fatalf("UpdateEntryStatus failed: %v", err)
	}

	bytes, _ := os.ReadFile(entry.FilePath)
	if !strings.Contains(string(bytes), "[x]") {
		t.Errorf("Expected file to contain [x], got: %s", string(bytes))
	}

	updatedEntries, _ := db.GetEntriesByFile(entry.FilePath)
	if updatedEntries[0].Status != models.EntryStatusCompleted {
		t.Errorf("Expected status completed, got %s", updatedEntries[0].Status)
	}
}

func TestMigrateTask(t *testing.T) {
	svc, _, db, cleanup := setupTestService(t)
	defer cleanup()

	yesterday := time.Now().AddDate(0, 0, -1)
	entry, err := svc.AddEntry("Old task", models.EntryTypeTask, yesterday)
	if err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entries, _ := db.GetEntriesByFile(entry.FilePath)
	dbEntry := entries[0]

	newEntry, err := svc.MigrateTask(dbEntry)
	if err != nil {
		t.Fatalf("MigrateTask failed: %v", err)
	}

	if newEntry.ParentID != dbEntry.ID {
		t.Errorf("Expected ParentID %s, got %s", dbEntry.ID, newEntry.ParentID)
	}
	if newEntry.MigrationCount != dbEntry.MigrationCount+1 {
		t.Errorf("Expected MigrationCount %d, got %d", dbEntry.MigrationCount+1, newEntry.MigrationCount)
	}

	oldEntries, _ := db.GetEntriesByFile(dbEntry.FilePath)
	if oldEntries[0].Status != models.EntryStatusMigrated {
		t.Errorf("Expected old entry status to be migrated, got %s", oldEntries[0].Status)
	}
}

func TestScheduleTask(t *testing.T) {
	svc, _, db, cleanup := setupTestService(t)
	defer cleanup()

	entry, err := svc.AddEntry("Task to schedule", models.EntryTypeTask, time.Now())
	if err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entries, _ := db.GetEntriesByFile(entry.FilePath)
	dbEntry := entries[0]

	targetDate := time.Now().AddDate(0, 0, 3)
	newEntry, err := svc.ScheduleTask(dbEntry, targetDate)
	if err != nil {
		t.Fatalf("ScheduleTask failed: %v", err)
	}

	if newEntry.ParentID != dbEntry.ID {
		t.Errorf("Expected ParentID %s, got %s", dbEntry.ID, newEntry.ParentID)
	}
	if newEntry.RescheduleCount != dbEntry.RescheduleCount+1 {
		t.Errorf("Expected RescheduleCount %d, got %d", dbEntry.RescheduleCount+1, newEntry.RescheduleCount)
	}

	oldEntries, _ := db.GetEntriesByFile(dbEntry.FilePath)
	if oldEntries[0].Status != models.EntryStatusScheduled {
		t.Errorf("Expected old entry status to be scheduled, got %s", oldEntries[0].Status)
	}
}

func TestGetEntriesByDate(t *testing.T) {
	svc, _, _, cleanup := setupTestService(t)
	defer cleanup()

	today := time.Now()
	_, _ = svc.AddEntry("Task 1", models.EntryTypeTask, today)
	_, _ = svc.AddEntry("Task 2", models.EntryTypeTask, today)
	_, _ = svc.AddEntry("Event", models.EntryTypeEvent, today)

	entries, err := svc.GetEntriesByDate(today)
	if err != nil {
		t.Fatalf("GetEntriesByDate failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestGetStaleTasks(t *testing.T) {
	svc, _, _, cleanup := setupTestService(t)
	defer cleanup()

	yesterday := time.Now().AddDate(0, 0, -1)
	_, _ = svc.AddEntry("Old task", models.EntryTypeTask, yesterday)

	tasks, err := svc.GetStaleTasks(0)
	if err != nil {
		t.Fatalf("GetStaleTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 stale task, got %d", len(tasks))
	}
}

func TestGetMigrationChain(t *testing.T) {
	svc, _, db, cleanup := setupTestService(t)
	defer cleanup()

	yesterday := time.Now().AddDate(0, 0, -1)
	original, _ := svc.AddEntry("Original task", models.EntryTypeTask, yesterday)

	entries, _ := db.GetEntriesByFile(original.FilePath)
	dbEntry := entries[0]

	migrated, _ := svc.MigrateTask(dbEntry)

	chain, err := svc.GetMigrationChain(migrated.ID)
	if err != nil {
		t.Fatalf("GetMigrationChain failed: %v", err)
	}

	if len(chain) != 2 {
		t.Errorf("Expected chain length 2, got %d", len(chain))
	}
}
