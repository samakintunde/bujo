package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FSStore struct {
	Root string
}

func NewFSStore(basePath string) (*FSStore, error) {
	return &FSStore{Root: basePath}, nil
}

func (fs *FSStore) GetDayPath(dateStr string) (string, error) {
	parsedDate, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return "", err
	}
	date := parsedDate.Format(time.DateOnly)
	dateParts := strings.Split(date, "-")
	year := dateParts[0]
	month := dateParts[1]
	fileName := fmt.Sprintf("%s.md", date)

	dirPath := filepath.Join(fs.Root, year, month)

	if err := fs.EnsureDirectory(dirPath); err != nil {
		return "", err
	}

	return filepath.Join(dirPath, fileName), nil
}

func (fs *FSStore) GetTodayPath() (string, error) {
	return fs.GetDayPath(time.Now().Format(time.DateOnly))
}

func (fs *FSStore) EnsureDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FSStore) AppendLine(path, content string) error {
	// Adding O_CREATE creates the file if missing. Neat!
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(content + "\n"); err != nil {
		return err
	}

	return nil
}

func (fs *FSStore) UpdateLine(path string, lineNum int, content string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(f), "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return fmt.Errorf("line number out of range")
	}
	lines[lineNum-1] = content
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(newContent), 0644)
}
