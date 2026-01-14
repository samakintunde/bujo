package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDayPath(t *testing.T) {
	fs := &FSStore{Root: "/journal"}

	tests := []struct {
		name string
		date string
		want string
	}{
		{
			name: "valid date",
			date: "2024-01-15",
			want: "/journal/2024/01/2024-01-15.md",
		},
		{
			name: "different month",
			date: "2024-12-01",
			want: "/journal/2024/12/2024-12-01.md",
		},
		{
			name: "invalid date falls back to simple path",
			date: "not-a-date",
			want: "/journal/not-a-date.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.GetDayPath(tt.date)
			if got != tt.want {
				t.Errorf("GetDayPath(%q) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestEnsureDayPath(t *testing.T) {
	dir := t.TempDir()
	fs := &FSStore{Root: dir}

	path, err := fs.EnsureDayPath("2024-03-20")
	if err != nil {
		t.Fatalf("EnsureDayPath() error: %v", err)
	}

	expectedPath := filepath.Join(dir, "2024", "03", "2024-03-20.md")
	if path != expectedPath {
		t.Errorf("EnsureDayPath() = %q, want %q", path, expectedPath)
	}

	parentDir := filepath.Dir(path)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("parent directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("parent path is not a directory")
	}
}

func TestAppendLine(t *testing.T) {
	dir := t.TempDir()
	fs := &FSStore{Root: dir}
	path := filepath.Join(dir, "test.md")

	if err := fs.AppendLine(path, "first line"); err != nil {
		t.Fatalf("first AppendLine() error: %v", err)
	}

	if err := fs.AppendLine(path, "second line"); err != nil {
		t.Fatalf("second AppendLine() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	got := string(content)
	want := "first line\nsecond line\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestUpdateLine(t *testing.T) {
	dir := t.TempDir()
	fs := &FSStore{Root: dir}
	path := filepath.Join(dir, "test.md")

	initial := "line one\nline two\nline three"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	if err := fs.UpdateLine(path, 2, "REPLACED"); err != nil {
		t.Fatalf("UpdateLine() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if lines[1] != "REPLACED" {
		t.Errorf("line 2 = %q, want %q", lines[1], "REPLACED")
	}
	if lines[0] != "line one" {
		t.Errorf("line 1 = %q, want %q (should be unchanged)", lines[0], "line one")
	}
	if lines[2] != "line three" {
		t.Errorf("line 3 = %q, want %q (should be unchanged)", lines[2], "line three")
	}
}

func TestUpdateLine_OutOfRange(t *testing.T) {
	dir := t.TempDir()
	fs := &FSStore{Root: dir}
	path := filepath.Join(dir, "test.md")

	if err := os.WriteFile(path, []byte("one\ntwo"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	if err := fs.UpdateLine(path, 0, "bad"); err == nil {
		t.Error("UpdateLine(0) should error for line 0")
	}

	if err := fs.UpdateLine(path, 5, "bad"); err == nil {
		t.Error("UpdateLine(5) should error for out of range")
	}
}
