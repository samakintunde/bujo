package models

import (
	"strings"
	"testing"
)

func TestRawString(t *testing.T) {
	tests := []struct {
		name       string
		entry      Entry
		wantPrefix string
	}{
		// Tasks
		{
			name:       "open task",
			entry:      Entry{Type: EntryTypeTask, Status: EntryStatusOpen, Content: "Buy milk", ID: "abc"},
			wantPrefix: "- [ ] Buy milk",
		},
		{
			name:       "completed task",
			entry:      Entry{Type: EntryTypeTask, Status: EntryStatusCompleted, Content: "Buy milk", ID: "abc"},
			wantPrefix: "- [x] Buy milk",
		},
		{
			name:       "migrated task",
			entry:      Entry{Type: EntryTypeTask, Status: EntryStatusMigrated, Content: "Buy milk", ID: "abc"},
			wantPrefix: "- [>] Buy milk",
		},
		{
			name:       "scheduled task",
			entry:      Entry{Type: EntryTypeTask, Status: EntryStatusScheduled, Content: "Buy milk", ID: "abc"},
			wantPrefix: "- [<] Buy milk",
		},
		{
			name:       "cancelled task",
			entry:      Entry{Type: EntryTypeTask, Status: EntryStatusCancelled, Content: "Buy milk", ID: "abc"},
			wantPrefix: "- [-] Buy milk",
		},
		// Event
		{
			name:       "event",
			entry:      Entry{Type: EntryTypeEvent, Status: EntryStatusOpen, Content: "Meeting", ID: "abc"},
			wantPrefix: "- * Meeting",
		},
		// Note
		{
			name:       "note",
			entry:      Entry{Type: EntryTypeNote, Status: EntryStatusOpen, Content: "Remember this", ID: "abc"},
			wantPrefix: "- Remember this",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.RawString()

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("RawString() = %q, want prefix %q", got, tt.wantPrefix)
			}

			if !strings.Contains(got, `<!-- {"id":"abc"`) {
				t.Errorf("RawString() = %q, missing metadata comment", got)
			}
		})
	}
}

func TestMetadata(t *testing.T) {
	entry := Entry{
		ID:              "test-id",
		MigrationCount:  2,
		ParentID:        "parent-id",
		RescheduleCount: 1,
	}

	meta := entry.Metadata()

	if meta.ID != "test-id" {
		t.Errorf("Metadata().ID = %q, want %q", meta.ID, "test-id")
	}
	if meta.Mig != 2 {
		t.Errorf("Metadata().Mig = %d, want %d", meta.Mig, 2)
	}
	if meta.PID != "parent-id" {
		t.Errorf("Metadata().PID = %q, want %q", meta.PID, "parent-id")
	}
	if meta.Rsch != 1 {
		t.Errorf("Metadata().Rsch = %d, want %d", meta.Rsch, 1)
	}
}

func TestMetadataString_OmitsZeroValues(t *testing.T) {
	meta := Metadata{ID: "abc"}
	got := meta.String()

	if !strings.Contains(got, `"id":"abc"`) {
		t.Errorf("String() = %q, missing id", got)
	}
	if strings.Contains(got, `"mig"`) {
		t.Errorf("String() = %q, should omit zero mig", got)
	}
	if strings.Contains(got, `"pid"`) {
		t.Errorf("String() = %q, should omit empty pid", got)
	}
	if strings.Contains(got, `"rsch"`) {
		t.Errorf("String() = %q, should omit zero rsch", got)
	}
}

func TestMetadataString_IncludesAllFields(t *testing.T) {
	meta := Metadata{ID: "abc", Mig: 2, PID: "parent", Rsch: 1}
	got := meta.String()

	if !strings.Contains(got, `"id":"abc"`) {
		t.Errorf("String() = %q, missing id", got)
	}
	if !strings.Contains(got, `"mig":2`) {
		t.Errorf("String() = %q, missing mig", got)
	}
	if !strings.Contains(got, `"pid":"parent"`) {
		t.Errorf("String() = %q, missing pid", got)
	}
	if !strings.Contains(got, `"rsch":1`) {
		t.Errorf("String() = %q, missing rsch", got)
	}
	if !strings.HasPrefix(got, "<!-- ") || !strings.HasSuffix(got, " -->") {
		t.Errorf("String() = %q, missing HTML comment wrapper", got)
	}
}
