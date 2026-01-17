package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/samakintunde/bujo/internal/id"
)

type EntryType string

const (
	EntryTypeTask   EntryType = "task"
	EntryTypeNote   EntryType = "note"
	EntryTypeEvent  EntryType = "event"
	EntryTypeIgnore EntryType = ""
)

type EntryStatus string

const (
	EntryStatusOpen      EntryStatus = "open"
	EntryStatusCompleted EntryStatus = "completed"
	EntryStatusMigrated  EntryStatus = "migrated"
	EntryStatusCancelled EntryStatus = "cancelled"
	EntryStatusScheduled EntryStatus = "scheduled"
)

type Entry struct {
	ID              string
	Type            EntryType
	Status          EntryStatus
	Content         string
	RawContent      string
	FilePath        string
	LineNumber      int
	MigrationCount  int
	RescheduleCount int
	ParentID        string
	IsDeleted       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewEntry(entryType EntryType, content string) *Entry {
	return &Entry{
		ID:      id.New(),
		Type:    entryType,
		Status:  EntryStatusOpen,
		Content: content,
	}
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s %s", e.getMarkdownSignifier(), e.Content)
}

func (e *Entry) Metadata() Metadata {
	return Metadata{
		ID:   e.ID,
		Mig:  e.MigrationCount,
		PID:  e.ParentID,
		Rsch: e.RescheduleCount,
	}
}

func (e *Entry) RawString() string {
	return fmt.Sprintf("%s %s %s", e.getMarkdownSignifier(), e.Content, e.Metadata().String())
}

func (e *Entry) DisplayString() string {
	return fmt.Sprintf("%s %s", e.getDisplaySignifier(), e.Content)
}

func (e *Entry) getDisplaySignifier() string {
	switch e.Type {
	case EntryTypeTask:
		switch e.Status {
		case EntryStatusOpen:
			return "•"
		case EntryStatusCompleted:
			return "x"
		case EntryStatusMigrated:
			return ">"
		case EntryStatusCancelled:
			return "-"
		case EntryStatusScheduled:
			return "<"
		default:
			return ""
		}
	case EntryTypeEvent:
		return "•"
	case EntryTypeNote:
		return "-"
	default:
		return ""
	}
}

func (e *Entry) getMarkdownSignifier() string {
	switch e.Type {
	case EntryTypeTask:
		switch e.Status {
		case EntryStatusOpen:
			return "- [ ]"
		case EntryStatusCompleted:
			return "- [x]"
		case EntryStatusMigrated:
			return "- [>]"
		case EntryStatusCancelled:
			return "- [-]"
		case EntryStatusScheduled:
			return "- [<]"
		default:
			return "- [ ]"
		}
	case EntryTypeEvent:
		return "- *"
	case EntryTypeNote:
		return "-"
	default:
		return "- [ ]"
	}
}

type Metadata struct {
	ID   string `json:"id"`
	Mig  int    `json:"mig,omitempty"`
	PID  string `json:"pid,omitempty"`
	Rsch int    `json:"rsch,omitempty"`
}

func (m Metadata) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("<!-- %s -->", b)
}
