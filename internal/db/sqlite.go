package storage

import (
	"database/sql"
	"path/filepath"
	"time"

	"github.com/samakintunde/bujo-cli/internal/models"
	_ "modernc.org/sqlite"
)

type DBStore struct {
	db *sql.DB
}

const DB_FILEPATH = "db.sqlite"

func New(basePath string) (*DBStore, error) {
	dbFilePath := filepath.Join(basePath, DB_FILEPATH)
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	store := &DBStore{db: db}

	err = store.enableWALMode()
	if err != nil {
		return nil, err
	}
	err = store.createSchema()
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *DBStore) enableWALMode() error {
	_, err := s.db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;")
	return err
}

func (s *DBStore) Close() error {
	return s.db.Close()
}

func (s *DBStore) createSchema() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
CREATE TABLE IF NOT EXISTS entries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'task' CHECK(type IN ('task', 'event', 'note')),
    status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open', 'completed', 'migrated', 'scheduled', 'cancelled')),
    content TEXT NOT NULL,
    raw_content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    migration_count INTEGER DEFAULT 0,
    reschedule_count INTEGER DEFAULT 0,
    parent_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME,
    is_deleted BOOLEAN DEFAULT 0
);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
CREATE TABLE IF NOT EXISTS files (
    path TEXT PRIMARY KEY,
    last_synced_at DATETIME,
    hash TEXT
);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_status_date ON entries(status, file_path);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_parent ON entries(parent_id);")
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStore) GetFileLastSync(path string) (time.Time, error) {
	var lastSyncedAt sql.NullTime
	err := s.db.QueryRow("SELECT last_synced_at FROM files WHERE path = ?", path).Scan(&lastSyncedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return lastSyncedAt.Time, nil
}

func (s *DBStore) SyncEntries(path string, entries []models.Entry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM entries WHERE file_path = ?", path); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO entries (
        id, type, status, content, raw_content, file_path, line_number,
        migration_count, reschedule_count, parent_id, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		if e.Type == models.EntryTypeIgnore {
			continue
		}
		_, err = stmt.Exec(
			e.ID, e.Type, e.Status, e.Content, e.RawContent, e.FilePath, e.LineNumber,
			e.MigrationCount, e.RescheduleCount, e.ParentID, e.CreatedAt, e.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`INSERT OR REPLACE INTO files (path, last_synced_at) VALUES (?, ?)`, path, time.Now()); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *DBStore) GetEntriesByFile(path string) ([]models.Entry, error) {
	query := `
        SELECT id, type, status, content, raw_content, file_path, line_number,
               migration_count, reschedule_count, parent_id, created_at, updated_at
        FROM entries
        WHERE file_path = ?
        ORDER BY line_number ASC`

	rows, err := s.db.Query(query, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		var e models.Entry
		err := rows.Scan(
			&e.ID, &e.Type, &e.Status, &e.Content, &e.RawContent, &e.FilePath, &e.LineNumber,
			&e.MigrationCount, &e.RescheduleCount, &e.ParentID, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
