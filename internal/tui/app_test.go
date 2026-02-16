package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
)

func setupTestApp(t *testing.T) (*App, func()) {
	dir := t.TempDir()
	db, err := storage.NewDBStore(dir)
	if err != nil {
		t.Fatalf("NewDBStore() error: %v", err)
	}
	fs, err := storage.NewFSStore(dir)
	if err != nil {
		t.Fatalf("NewFSStore() error: %v", err)
	}
	syncer := sync.NewSyncer(dir, db)

	app := NewApp(db, fs, syncer)

	cleanup := func() {
		db.Close()
	}

	return app, cleanup
}

func TestNewApp(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	if app.state != StateDailyView {
		t.Errorf("initial state = %v, want StateDailyView", app.state)
	}
	if app.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", app.cursor)
	}
	if !isToday(app.currentDate) {
		t.Errorf("initial date = %v, want today", app.currentDate)
	}
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		initial   AppState
		key       string
		wantState AppState
	}{
		{"daily to add", StateDailyView, "a", StateAddEntry},
		{"daily to datepicker", StateDailyView, "d", StateDatePicker},
		{"daily to tag filter", StateDailyView, "f", StateTagFilter},
		{"daily to review", StateDailyView, "r", StateReviewScope},
		{"add cancel", StateAddEntry, "esc", StateDailyView},
		{"datepicker cancel", StateDatePicker, "esc", StateDailyView},
		{"tag filter cancel", StateTagFilter, "esc", StateDailyView},
		{"review scope cancel", StateReviewScope, "esc", StateDailyView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestApp(t)
			defer cleanup()

			app.state = tt.initial

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "esc" {
				keyMsg = tea.KeyMsg{Type: tea.KeyEscape}
			}

			newModel, _ := app.Update(keyMsg)
			newApp := newModel.(*App)

			if newApp.state != tt.wantState {
				t.Errorf("state after %q = %v, want %v", tt.key, newApp.state, tt.wantState)
			}
		})
	}
}

func TestCursorNavigation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "First"},
		{ID: "2", Content: "Second"},
		{ID: "3", Content: "Third"},
	}
	app.cursor = 0

	downKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ := app.Update(downKey)
	app = newModel.(*App)
	if app.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", app.cursor)
	}

	newModel, _ = app.Update(downKey)
	app = newModel.(*App)
	if app.cursor != 2 {
		t.Errorf("cursor after second j = %d, want 2", app.cursor)
	}

	newModel, _ = app.Update(downKey)
	app = newModel.(*App)
	if app.cursor != 2 {
		t.Errorf("cursor should not go past last entry, got %d", app.cursor)
	}

	upKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newModel, _ = app.Update(upKey)
	app = newModel.(*App)
	if app.cursor != 1 {
		t.Errorf("cursor after k = %d, want 1", app.cursor)
	}

	app.cursor = 0
	newModel, _ = app.Update(upKey)
	app = newModel.(*App)
	if app.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", app.cursor)
	}
}

func TestDateNavigation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	today := time.Now()
	app.currentDate = today

	prevKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	newModel, _ := app.Update(prevKey)
	app = newModel.(*App)

	expectedDate := today.AddDate(0, 0, -1)
	if !sameDay(app.currentDate, expectedDate) {
		t.Errorf("date after h = %v, want %v", app.currentDate.Format(time.DateOnly), expectedDate.Format(time.DateOnly))
	}

	nextKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	newModel, _ = app.Update(nextKey)
	app = newModel.(*App)

	if !sameDay(app.currentDate, today) {
		t.Errorf("date after l = %v, want %v", app.currentDate.Format(time.DateOnly), today.Format(time.DateOnly))
	}

	app.currentDate = today.AddDate(0, 0, -5)
	todayKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	newModel, _ = app.Update(todayKey)
	app = newModel.(*App)

	if !isToday(app.currentDate) {
		t.Errorf("date after t = %v, want today", app.currentDate.Format(time.DateOnly))
	}
}

func TestReviewPromptKeys(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.state = StateReviewPrompt
	app.staleTaskCount = 5

	yesKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, _ := app.Update(yesKey)
	app = newModel.(*App)

	if app.state != StateReviewScope {
		t.Errorf("state after y = %v, want StateReviewScope", app.state)
	}

	app.state = StateReviewPrompt
	noKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, _ = app.Update(noKey)
	app = newModel.(*App)

	if app.state != StateDailyView {
		t.Errorf("state after n = %v, want StateDailyView", app.state)
	}
}

func TestEntriesLoadedPreservesCursor(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "First"},
		{ID: "2", Content: "Second"},
		{ID: "3", Content: "Third"},
	}
	app.cursor = 2

	msg := entriesLoadedMsg{
		entries: []models.Entry{
			{ID: "1", Content: "First"},
			{ID: "2", Content: "Second"},
			{ID: "3", Content: "Third"},
		},
	}

	newModel, _ := app.Update(msg)
	app = newModel.(*App)

	if app.cursor != 2 {
		t.Errorf("cursor after reload = %d, want 2 (preserved)", app.cursor)
	}
}

func TestEntriesLoadedAdjustsCursorWhenOutOfBounds(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "First"},
		{ID: "2", Content: "Second"},
		{ID: "3", Content: "Third"},
	}
	app.cursor = 2

	msg := entriesLoadedMsg{
		entries: []models.Entry{
			{ID: "1", Content: "First"},
		},
	}

	newModel, _ := app.Update(msg)
	app = newModel.(*App)

	if app.cursor != 0 {
		t.Errorf("cursor after reload with fewer entries = %d, want 0", app.cursor)
	}
}

func TestQuitKey(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	quitKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := app.Update(quitKey)

	if cmd == nil {
		t.Error("quit key should return a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("quit key command returned %T, want tea.QuitMsg", msg)
	}
}

func TestCursorMovementClearsChainState(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "First", ParentID: "0"},
		{ID: "2", Content: "Second"},
		{ID: "3", Content: "Third"},
	}
	app.cursor = 0
	app.migrationChain = []models.Entry{
		{ID: "0", Content: "Original"},
		{ID: "1", Content: "First"},
	}
	app.migrationChainIndex = 1

	downKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ := app.Update(downKey)
	app = newModel.(*App)

	if len(app.migrationChain) != 0 {
		t.Errorf("migrationChain should be cleared after cursor move, got %d entries", len(app.migrationChain))
	}
	if app.migrationChainIndex != 0 {
		t.Errorf("migrationChainIndex should be 0 after cursor move, got %d", app.migrationChainIndex)
	}

	app.cursor = 1
	app.migrationChain = []models.Entry{{ID: "0"}, {ID: "1"}}
	app.migrationChainIndex = 1

	upKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newModel, _ = app.Update(upKey)
	app = newModel.(*App)

	if len(app.migrationChain) != 0 {
		t.Errorf("migrationChain should be cleared after cursor up, got %d entries", len(app.migrationChain))
	}
}

func TestDateNavigationClearsChainState(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{{ID: "1", Content: "Task"}}
	app.cursor = 0
	app.migrationChain = []models.Entry{{ID: "0"}, {ID: "1"}}
	app.migrationChainIndex = 1

	prevKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	newModel, _ := app.Update(prevKey)
	app = newModel.(*App)

	if len(app.migrationChain) != 0 {
		t.Errorf("migrationChain should be cleared after h (prev day), got %d entries", len(app.migrationChain))
	}

	app.migrationChain = []models.Entry{{ID: "0"}, {ID: "1"}}
	app.migrationChainIndex = 1
	nextKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	newModel, _ = app.Update(nextKey)
	app = newModel.(*App)

	if len(app.migrationChain) != 0 {
		t.Errorf("migrationChain should be cleared after l (next day), got %d entries", len(app.migrationChain))
	}

	app.currentDate = time.Now().AddDate(0, 0, -5)
	app.migrationChain = []models.Entry{{ID: "0"}, {ID: "1"}}
	app.migrationChainIndex = 1

	todayKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	newModel, _ = app.Update(todayKey)
	app = newModel.(*App)

	if len(app.migrationChain) != 0 {
		t.Errorf("migrationChain should be cleared after t (today), got %d entries", len(app.migrationChain))
	}
}

func TestStatusBarChainHintFormat(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "Migrated task", ParentID: "0"},
	}
	app.cursor = 0

	statusBar := app.renderStatusBar()

	if !contains(statusBar, "history") {
		t.Error("status bar should show 'history' hint for entries with ParentID")
	}
}

func TestStatusBarNoChainHintForRegularEntries(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.entries = []models.Entry{
		{ID: "1", Content: "Regular task"},
	}
	app.cursor = 0

	statusBar := app.renderStatusBar()

	if contains(statusBar, "history") {
		t.Error("status bar should NOT show history hint for regular entries")
	}
}

func TestTagFilterFlow_CommitAndClear(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	today := time.Now()
	_, _ = app.service.AddEntry("First #work", models.EntryTypeTask, today)
	_, _ = app.service.AddEntry("Second #deep_focus #work", models.EntryTypeTask, today)

	filterKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	newModel, cmd := app.Update(filterKey)
	app = newModel.(*App)

	if app.state != StateTagFilter {
		t.Fatalf("state after f = %v, want StateTagFilter", app.state)
	}
	if cmd == nil {
		t.Fatal("filter key should return load tags command")
	}

	msg := cmd()
	newModel, _ = app.Update(msg)
	app = newModel.(*App)

	if len(app.tagOptions) != 2 {
		t.Fatalf("tag options = %d, want 2", len(app.tagOptions))
	}

	spaceKey := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	newModel, _ = app.Update(spaceKey) // select first tag
	app = newModel.(*App)

	downKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ = app.Update(downKey)
	app = newModel.(*App)

	newModel, _ = app.Update(spaceKey) // select second tag
	app = newModel.(*App)

	toggleModeKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	newModel, _ = app.Update(toggleModeKey)
	app = newModel.(*App)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = app.Update(enterKey)
	app = newModel.(*App)

	if app.state != StateDailyView {
		t.Fatalf("state after enter = %v, want StateDailyView", app.state)
	}
	if !app.activeTagMatchAll {
		t.Fatal("activeTagMatchAll should be true after toggling to AND")
	}
	if len(app.activeTagFilter) != 2 {
		t.Fatalf("activeTagFilter len = %d, want 2", len(app.activeTagFilter))
	}
	if cmd == nil {
		t.Fatal("enter should trigger filtered load command")
	}

	clearKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	newModel, cmd = app.Update(clearKey)
	app = newModel.(*App)

	if len(app.activeTagFilter) != 0 {
		t.Fatalf("activeTagFilter len after clear = %d, want 0", len(app.activeTagFilter))
	}
	if app.activeTagMatchAll {
		t.Fatal("activeTagMatchAll should reset to false after clear")
	}
	if cmd == nil {
		t.Fatal("clear should trigger entry reload")
	}
}

func TestRenderHeader_ShowsActiveTagFilter(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.activeTagFilter = []string{"deep_focus", "work"}
	app.activeTagMatchAll = false

	header := app.renderHeader()

	if !contains(header, "Filter OR:") || !contains(header, "#deep_focus") || !contains(header, "#work") {
		t.Errorf("header should show active filter summary, got: %s", header)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
