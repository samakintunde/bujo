package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samakintunde/bujo/internal/models"
)

func TestParseRaw(t *testing.T) {
	tests := []struct {
		name string
		line string
		want models.Entry
	}{
		// Tasks
		{
			name: "open task",
			line: "- [ ] Buy milk",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusOpen,
				Content: "Buy milk",
			},
		},
		{
			name: "completed task",
			line: "- [x] Buy milk",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusCompleted,
				Content: "Buy milk",
			},
		},
		{
			name: "migrated task",
			line: "- [>] Buy milk",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusMigrated,
				Content: "Buy milk",
			},
		},
		{
			name: "scheduled task",
			line: "- [<] Buy milk",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusScheduled,
				Content: "Buy milk",
			},
		},
		{
			name: "cancelled task",
			line: "- [-] Buy milk",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusCancelled,
				Content: "Buy milk",
			},
		},
		{
			name: "unknown checkbox char becomes task with empty status",
			line: "- [?] Mystery",
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  "",
				Content: "Mystery",
			},
		},
		// Events
		{
			name: "event",
			line: "- * Meeting at 3pm",
			want: models.Entry{
				Type:    models.EntryTypeEvent,
				Status:  models.EntryStatusOpen,
				Content: "Meeting at 3pm",
			},
		},
		// Notes
		{
			name: "note",
			line: "- Just a note",
			want: models.Entry{
				Type:    models.EntryTypeNote,
				Status:  models.EntryStatusOpen,
				Content: "Just a note",
			},
		},
		// Ignored lines
		{
			name: "heading is ignored",
			line: "# Daily Log",
			want: models.Entry{
				Type: models.EntryTypeIgnore,
			},
		},
		{
			name: "whitespace line is ignored",
			line: "   ",
			want: models.Entry{
				Type: models.EntryTypeIgnore,
			},
		},
		{
			name: "random text is ignored",
			line: "Just some random text",
			want: models.Entry{
				Type: models.EntryTypeIgnore,
			},
		},
		// Metadata
		{
			name: "extracts id from metadata",
			line: `- [ ] Task <!-- {"id":"abc123"} -->`,
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusOpen,
				Content: "Task",
				ID:      "abc123",
			},
		},
		{
			name: "extracts all metadata fields",
			line: `- [ ] Task <!-- {"id":"abc","mig":2,"pid":"parent","rsch":1} -->`,
			want: models.Entry{
				Type:            models.EntryTypeTask,
				Status:          models.EntryStatusOpen,
				Content:         "Task",
				ID:              "abc",
				MigrationCount:  2,
				ParentID:        "parent",
				RescheduleCount: 1,
			},
		},
		{
			name: "malformed json metadata is ignored but entry still parsed",
			line: `- [ ] Task <!-- {invalid json} -->`,
			want: models.Entry{
				Type:    models.EntryTypeTask,
				Status:  models.EntryStatusOpen,
				Content: "Task",
				ID:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write single line to temp file
			dir := t.TempDir()
			path := filepath.Join(dir, "test.md")
			if err := os.WriteFile(path, []byte(tt.line), 0644); err != nil {
				t.Fatal(err)
			}

			entries, err := ParseRaw(path)
			if err != nil {
				t.Fatalf("ParseRaw error: %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}

			got := entries[0]

			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Status != tt.want.Status {
				t.Errorf("Status = %q, want %q", got.Status, tt.want.Status)
			}
			if got.Content != tt.want.Content {
				t.Errorf("Content = %q, want %q", got.Content, tt.want.Content)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.MigrationCount != tt.want.MigrationCount {
				t.Errorf("MigrationCount = %d, want %d", got.MigrationCount, tt.want.MigrationCount)
			}
			if got.ParentID != tt.want.ParentID {
				t.Errorf("ParentID = %q, want %q", got.ParentID, tt.want.ParentID)
			}
			if got.RescheduleCount != tt.want.RescheduleCount {
				t.Errorf("RescheduleCount = %d, want %d", got.RescheduleCount, tt.want.RescheduleCount)
			}
			if got.LineNumber != 1 {
				t.Errorf("LineNumber = %d, want 1", got.LineNumber)
			}
			if got.FilePath != path {
				t.Errorf("FilePath = %q, want %q", got.FilePath, path)
			}
			if got.RawContent != tt.line {
				t.Errorf("RawContent = %q, want %q", got.RawContent, tt.line)
			}
		})
	}
}

func TestParseRaw_LineNumbers(t *testing.T) {
	content := `# Heading
- [ ] First task
- Second is a note
- * Third is event`

	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ParseRaw(path)
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	for i, e := range entries {
		expectedLine := i + 1
		if e.LineNumber != expectedLine {
			t.Errorf("entry %d: LineNumber = %d, want %d", i, e.LineNumber, expectedLine)
		}
	}
}

func TestParse_FiltersIgnoredLines(t *testing.T) {
	content := `# Heading
- [ ] A task

- * An event
Random text
- A note`

	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Should only have 3 entries: task, event, note
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Verify none are Ignore type
	for i, e := range entries {
		if e.Type == models.EntryTypeIgnore {
			t.Errorf("entry %d has Type=Ignore, should be filtered out", i)
		}
	}

	// Verify types in order
	if entries[0].Type != models.EntryTypeTask {
		t.Errorf("entries[0].Type = %q, want task", entries[0].Type)
	}
	if entries[1].Type != models.EntryTypeEvent {
		t.Errorf("entries[1].Type = %q, want event", entries[1].Type)
	}
	if entries[2].Type != models.EntryTypeNote {
		t.Errorf("entries[2].Type = %q, want note", entries[2].Type)
	}

	// Verify line numbers are preserved (original file positions)
	if entries[0].LineNumber != 2 {
		t.Errorf("task LineNumber = %d, want 2", entries[0].LineNumber)
	}
	if entries[1].LineNumber != 4 {
		t.Errorf("event LineNumber = %d, want 4", entries[1].LineNumber)
	}
	if entries[2].LineNumber != 6 {
		t.Errorf("note LineNumber = %d, want 6", entries[2].LineNumber)
	}
}
