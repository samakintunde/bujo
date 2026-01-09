package domain

import "time"

type EntryType string

const (
	EntryTypeTask  EntryType = "task"
	EntryTypeNote  EntryType = "note"
	EntryTypeEvent EntryType = "event"
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

func (e *Entry) Signifier() string {
	m := map[EntryType]string{
		EntryTypeTask:  "[ ]",
		EntryTypeEvent: " ○ ",
		EntryTypeNote:  " – ",
	}

	switch e.Type {
	case EntryTypeTask:
		return m[EntryTypeTask]
	case EntryTypeEvent:
		return m[EntryTypeEvent]
	case EntryTypeNote:
		return m[EntryTypeNote]
	default:
		return ""
	}
}

type Metadata struct {
	ID   string
	Mig  int
	PID  string
	Rsch int
}
