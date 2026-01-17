package service

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/samakintunde/bujo/internal/git"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
)

type JournalService struct {
	fs     *storage.FSStore
	db     *storage.DBStore
	syncer *sync.Syncer
}

func NewJournalService(fs *storage.FSStore, db *storage.DBStore, syncer *sync.Syncer) *JournalService {
	return &JournalService{
		fs:     fs,
		db:     db,
		syncer: syncer,
	}
}

func (s *JournalService) AddEntry(content string, entryType models.EntryType, date time.Time) (*models.Entry, error) {
	entry := models.NewEntry(entryType, content)
	dateStr := date.Format(time.DateOnly)

	path, err := s.fs.EnsureDayPath(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure day path: %w", err)
	}

	entry.FilePath = path

	if err := s.fs.AppendLine(path, entry.RawString()); err != nil {
		return nil, fmt.Errorf("failed to write entry to file: %w", err)
	}

	if err := s.syncer.SyncFile(path); err != nil {
		return nil, fmt.Errorf("failed to sync file to db: %w", err)
	}

	s.gitCommit(path, fmt.Sprintf("feat(bujo): add %s #%s", entryType, entry.ID))

	return entry, nil
}

func (s *JournalService) UpdateEntryStatus(entry models.Entry, newStatus models.EntryStatus) error {
	entry.Status = newStatus

	if err := s.fs.UpdateLine(entry.FilePath, entry.LineNumber, entry.RawString()); err != nil {
		return fmt.Errorf("failed to update entry in file: %w", err)
	}

	if err := s.syncer.SyncFile(entry.FilePath); err != nil {
		return fmt.Errorf("failed to sync file to db: %w", err)
	}

	s.gitCommit(filepath.Dir(entry.FilePath), fmt.Sprintf("feat(bujo): update %s #%s to %s", entry.Type, entry.ID, newStatus))

	return nil
}

func (s *JournalService) MigrateTask(entry models.Entry) (*models.Entry, error) {
	entry.Status = models.EntryStatusMigrated
	if err := s.fs.UpdateLine(entry.FilePath, entry.LineNumber, entry.RawString()); err != nil {
		return nil, fmt.Errorf("failed to update original entry: %w", err)
	}

	newEntry := models.NewEntry(models.EntryTypeTask, entry.Content)
	newEntry.MigrationCount = entry.MigrationCount + 1
	newEntry.ParentID = entry.ID

	todayPath, err := s.fs.EnsureDayPath(time.Now().Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("failed to ensure today path: %w", err)
	}
	newEntry.FilePath = todayPath

	if err := s.fs.AppendLine(todayPath, newEntry.RawString()); err != nil {
		return nil, fmt.Errorf("failed to write migrated entry: %w", err)
	}

	if err := s.syncer.SyncFile(entry.FilePath); err != nil {
		return nil, fmt.Errorf("failed to sync original file: %w", err)
	}
	if err := s.syncer.SyncFile(todayPath); err != nil {
		return nil, fmt.Errorf("failed to sync today file: %w", err)
	}

	s.gitCommit(filepath.Dir(todayPath), fmt.Sprintf("feat(bujo): migrate task #%s to #%s", entry.ID, newEntry.ID))

	return newEntry, nil
}

func (s *JournalService) ScheduleTask(entry models.Entry, targetDate time.Time) (*models.Entry, error) {
	entry.Status = models.EntryStatusScheduled
	if err := s.fs.UpdateLine(entry.FilePath, entry.LineNumber, entry.RawString()); err != nil {
		return nil, fmt.Errorf("failed to update original entry: %w", err)
	}

	newEntry := models.NewEntry(models.EntryTypeTask, entry.Content)
	newEntry.RescheduleCount = entry.RescheduleCount + 1
	newEntry.ParentID = entry.ID

	targetDateStr := targetDate.Format(time.DateOnly)
	targetPath, err := s.fs.EnsureDayPath(targetDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure target path: %w", err)
	}
	newEntry.FilePath = targetPath

	if err := s.fs.AppendLine(targetPath, newEntry.RawString()); err != nil {
		return nil, fmt.Errorf("failed to write scheduled entry: %w", err)
	}

	if err := s.syncer.SyncFile(entry.FilePath); err != nil {
		return nil, fmt.Errorf("failed to sync original file: %w", err)
	}
	if err := s.syncer.SyncFile(targetPath); err != nil {
		return nil, fmt.Errorf("failed to sync target file: %w", err)
	}

	s.gitCommit(filepath.Dir(targetPath), fmt.Sprintf("feat(bujo): schedule task #%s to %s as #%s", entry.ID, targetDateStr, newEntry.ID))

	return newEntry, nil
}

func (s *JournalService) GetEntriesByDate(date time.Time) ([]models.Entry, error) {
	dateStr := date.Format(time.DateOnly)
	path := s.fs.GetDayPath(dateStr)

	if err := s.syncer.SyncFile(path); err != nil {
		return nil, fmt.Errorf("failed to sync file: %w", err)
	}

	entries, err := s.db.GetEntriesByFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	return entries, nil
}

func (s *JournalService) GetStaleTasks(daysBack int) ([]models.Entry, error) {
	tasks, err := s.db.GetStaleTasks(daysBack)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale tasks: %w", err)
	}
	return tasks, nil
}

func (s *JournalService) CountStaleTasks(daysBack int) (int, error) {
	count, err := s.db.CountStaleTasks(daysBack)
	if err != nil {
		return 0, fmt.Errorf("failed to count stale tasks: %w", err)
	}
	return count, nil
}

func (s *JournalService) GetMigrationChain(entryID string) ([]models.Entry, error) {
	chain, err := s.db.GetMigrationChain(entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration chain: %w", err)
	}
	return chain, nil
}

func (s *JournalService) GetLastOpenedAt() (time.Time, error) {
	return s.db.GetLastOpenedAt()
}

func (s *JournalService) SetLastOpenedAt(t time.Time) error {
	return s.db.SetLastOpenedAt(t)
}

func (s *JournalService) gitCommit(dir, message string) {
	if git.IsPresent() && git.IsRepo(s.fs.Root) {
		_ = git.Commit(dir, message)
	}
}
