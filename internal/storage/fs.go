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

func (fs *FSStore) GetDayPath(dateStr string) string {
	parsedDate, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return filepath.Join(fs.Root, dateStr+".md")
	}
	year := parsedDate.Format("2006")
	month := parsedDate.Format("01")
	fileName := dateStr + ".md"

	return filepath.Join(fs.Root, year, month, fileName)
}

func (fs *FSStore) EnsureDayPath(dateStr string) (string, error) {
	path := fs.GetDayPath(dateStr)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return path, nil
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
