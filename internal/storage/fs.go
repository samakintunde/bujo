package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samakintunde/bujo-cli/internal/config"
)

type FSStore struct {
	BasePath string
}

func NewFSStore(cfg config.JournalConfig) (*FSStore, error) {
	return &FSStore{BasePath: cfg.Path}, nil
}

func (fs *FSStore) GetTodayPath() (string, error) {
	date := time.Now().Format(time.DateOnly)
	dateParts := strings.Split(date, "-")
	year := dateParts[0]
	month := dateParts[1]
	fileName := fmt.Sprintf("%s.md", date)

	dirPath := filepath.Join(fs.BasePath, year, month)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dirPath, fileName), nil
}

func (fs *FSStore) EnsureDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FSStore) AppendLine(path, content string) error {
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
