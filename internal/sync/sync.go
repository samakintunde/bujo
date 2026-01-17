package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samakintunde/bujo/internal/id"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/parser"
	"github.com/samakintunde/bujo/internal/storage"
)

type Syncer struct {
	Root string
	DB   *storage.DBStore
}

func NewSyncer(root string, db *storage.DBStore) *Syncer {
	return &Syncer{
		Root: root,
		DB:   db,
	}
}

func (s *Syncer) Sync() error {
	return filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		return s.syncFileWithInfo(path, info)
	})
}

func (s *Syncer) SyncFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return s.syncFileWithInfo(path, info)
}

func (s *Syncer) syncFileWithInfo(path string, info fs.FileInfo) error {
	lastSynced, err := s.DB.GetFileLastSync(path)
	if err != nil {
		return fmt.Errorf("failed to get sync status for %s: %w", path, err)
	}

	if info.ModTime().After(lastSynced) {
		entries, err := parser.ParseRaw(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		filename := filepath.Base(path)
		dateStr := strings.TrimSuffix(filename, filepath.Ext(filename))
		fileDate, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			fileDate = info.ModTime()
		}

		dirty := false
		for i := range entries {
			if entries[i].Type == models.EntryTypeIgnore {
				continue
			}
			entries[i].CreatedAt = fileDate
			entries[i].UpdatedAt = time.Now()

			if entries[i].ID == "" {
				entries[i].ID = id.New()
				dirty = true
			}
		}

		if dirty {
			var sb strings.Builder
			for _, e := range entries {
				if e.Type == models.EntryTypeIgnore {
					sb.WriteString(e.RawContent)
				} else {
					sb.WriteString(e.RawString())
				}
				sb.WriteString("\n")
			}
			if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
				return fmt.Errorf("failed to write back IDs to %s: %w", path, err)
			}
			fmt.Printf("Auto-repaired IDs in: %s\n", path)
		}

		if err := s.DB.SyncEntries(path, entries); err != nil {
			return fmt.Errorf("failed to sync entries for %s: %w", path, err)
		}
	}

	return nil
}
