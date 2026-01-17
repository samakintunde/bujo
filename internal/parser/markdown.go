package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/samakintunde/bujo/internal/models"
)

// Matches: - [x] Buy milk
var taskRegex = regexp.MustCompile(`^-\s\[(.)\]\s+(.*)`)

// Matches: - * Meeting
var eventRegex = regexp.MustCompile(`^-\s\*\s+(.*)`)

// Matches: - Just a note
var noteRegex = regexp.MustCompile(`^-\s+(.*)`)

// Matches hidden comment: <!-- {...} -->
var metaRegex = regexp.MustCompile(`<!--\s*(\{.*\})\s*-->`)

// Parse reads a markdown file and returns valid entries.
// It filters out EntryTypeIgnore lines.
// Use this for reading/indexing data (e.g. DB import).
func Parse(path string) ([]models.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Entry{}, fmt.Errorf("No entries found")
		}
		return []models.Entry{}, err
	}
	defer f.Close()

	entries := make([]models.Entry, 0)

	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		entry := parseLine(scanner.Text())
		if entry.Type == models.EntryTypeIgnore {
			i++
			continue
		}
		entry.LineNumber = i + 1
		entry.FilePath = path
		entries = append(entries, entry)
		i++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// ParseRaw reads a markdown file and returns ALL lines as entries,
// including EntryTypeIgnore lines.
// Use this for file rewriting/syncing to preserve structure.
func ParseRaw(path string) ([]models.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Entry{}, fmt.Errorf("No entries found")
		}
		return []models.Entry{}, err
	}
	defer f.Close()

	entries := make([]models.Entry, 0)

	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		entry := parseLine(scanner.Text())
		entry.LineNumber = i + 1
		entry.FilePath = path
		entries = append(entries, entry)
		i++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func parseLine(line string) models.Entry {
	entry := models.Entry{RawContent: line}

	if match := metaRegex.FindStringSubmatch(line); len(match) > 1 {
		var meta models.Metadata
		if err := json.Unmarshal([]byte(match[1]), &meta); err == nil {
			entry.ID = meta.ID
			entry.MigrationCount = meta.Mig
			entry.ParentID = meta.PID
			entry.RescheduleCount = meta.Rsch
		}

		line = strings.Replace(line, match[0], "", 1)
	}

	line = strings.TrimSpace(line)

	if match := taskRegex.FindStringSubmatch(line); len(match) > 1 {
		entry.Type = models.EntryTypeTask
		entry.Content = strings.TrimSpace(match[2])

		switch match[1] {
		case " ":
			entry.Status = models.EntryStatusOpen
		case "x":
			entry.Status = models.EntryStatusCompleted
		case ">":
			entry.Status = models.EntryStatusMigrated
		case "<":
			entry.Status = models.EntryStatusScheduled
		case "-":
			entry.Status = models.EntryStatusCancelled
		}
	} else if match := eventRegex.FindStringSubmatch(line); len(match) > 1 {
		entry.Type = models.EntryTypeEvent
		entry.Content = strings.TrimSpace(match[1])
		entry.Status = models.EntryStatusOpen
	} else if match := noteRegex.FindStringSubmatch(line); len(match) > 1 {
		entry.Type = models.EntryTypeNote
		entry.Content = strings.TrimSpace(match[1])
		entry.Status = models.EntryStatusOpen
	} else {
		entry.Type = models.EntryTypeIgnore
	}
	return entry
}
