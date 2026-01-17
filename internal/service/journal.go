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

// AddEntry creates a new entry, saves it to disk, syncs to DB, and optionally commits to git.
func (s *JournalService) AddEntry(content string, entryType models.EntryType, date time.Time) (*models.Entry, error) {
	entry := models.NewEntry(entryType, content)
	dateStr := date.Format(time.DateOnly)

	path, err := s.fs.EnsureDayPath(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure day path: %w", err)
	}

	entry.FilePath = path

	line := entry.RawString()
	if err := s.fs.AppendLine(path, line); err != nil {
		return nil, fmt.Errorf("failed to write entry to file: %w", err)
	}

	if err := s.syncer.SyncFile(path); err != nil {
		return nil, fmt.Errorf("failed to sync file to db: %w", err)
	}

	if git.IsPresent() && git.IsRepo(s.fs.Root) {
		dir := filepath.Dir(path)
		message := fmt.Sprintf("feat(bujo): add %s #%s", entryType, entry.ID)
		if err := git.Commit(dir, message); err != nil {
			return nil, fmt.Errorf("git commit failed: %w", err)
		}
	}

	return entry, nil
}
