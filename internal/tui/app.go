package tui

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/service"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
)

type AppState int

const (
	StateDailyView AppState = iota
	StateAddEntry
	StateDatePicker
	StateTagFilter
	StateReviewScope
	StateReviewTask
	StateReviewPrompt
)

type App struct {
	state       AppState
	currentDate time.Time
	entries     []models.Entry
	cursor      int

	keys       KeyMap
	reviewKeys ReviewKeyMap
	scopeKeys  ScopeKeyMap

	input     textinput.Model
	inputErr  string
	inputMode string

	reviewTasks   []models.Entry
	reviewCursor  int
	reviewScope   int
	reviewSummary ReviewSummary

	staleTaskCount int
	message        string

	activeTagFilter    []string
	activeTagMatchAll  bool
	tagOptions         []string
	tagCursor          int
	pendingTagFilter   map[string]bool
	pendingTagMatchAll bool

	migrationChain      []models.Entry
	migrationChainIndex int

	db      *storage.DBStore
	fs      *storage.FSStore
	syncer  *sync.Syncer
	service *service.JournalService

	width  int
	height int
	err    error
}

type ReviewSummary struct {
	Migrated  int
	Scheduled int
	Completed int
	Cancelled int
	Skipped   int
}

func NewApp(db *storage.DBStore, fs *storage.FSStore, syncer *sync.Syncer) *App {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 256
	ti.Width = 40

	svc := service.NewJournalService(fs, db, syncer)

	return &App{
		state:       StateDailyView,
		currentDate: time.Now(),
		entries:     []models.Entry{},
		cursor:      0,
		keys:        DefaultKeyMap,
		reviewKeys:  DefaultReviewKeyMap,
		scopeKeys:   DefaultScopeKeyMap,
		input:       ti,
		db:          db,
		fs:          fs,
		syncer:      syncer,
		service:     svc,
		width:       80,
		height:      24,
	}
}

type entriesLoadedMsg struct {
	entries  []models.Entry
	err      error
	targetID string
}

type reviewTasksLoadedMsg struct {
	tasks []models.Entry
	err   error
}

type entryUpdatedMsg struct {
	err error
}

type entryAddedMsg struct {
	err error
}

type initCheckMsg struct {
	isFirstOpenToday bool
	staleTaskCount   int
}

type chainLoadedMsg struct {
	chain []models.Entry
	index int
	err   error
}

type tagsLoadedMsg struct {
	tags []string
	err  error
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.loadEntries(), a.checkFirstOpenToday())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case entriesLoadedMsg:
		if msg.err != nil {
			a.entries = []models.Entry{}
		} else {
			a.entries = msg.entries
		}

		if msg.targetID != "" {
			found := false
			for i, e := range a.entries {
				if e.ID == msg.targetID {
					a.cursor = i
					found = true
					break
				}
			}
			if !found && a.cursor >= len(a.entries) {
				a.cursor = max(0, len(a.entries)-1)
			}
		} else if a.cursor >= len(a.entries) {
			a.cursor = max(0, len(a.entries)-1)
		}
		return a, nil

	case initCheckMsg:
		if msg.isFirstOpenToday && msg.staleTaskCount > 0 {
			a.staleTaskCount = msg.staleTaskCount
			a.state = StateReviewPrompt
		}
		return a, nil

	case reviewTasksLoadedMsg:
		if msg.err != nil {
			a.err = msg.err
			a.state = StateDailyView
		} else {
			a.reviewTasks = msg.tasks
			a.reviewCursor = 0
			if len(msg.tasks) == 0 {
				a.state = StateDailyView
			} else {
				a.state = StateReviewTask
			}
		}
		return a, nil

	case entryUpdatedMsg:
		if msg.err != nil {
			a.err = msg.err
		}
		return a, a.loadEntries()

	case entryAddedMsg:
		if msg.err != nil {
			a.err = msg.err
		}
		a.state = StateDailyView
		a.input.Reset()
		return a, a.loadEntries()

	case reviewActionCompleteMsg:
		return a.advanceReview()

	case chainLoadedMsg:
		if msg.err != nil {
			a.err = msg.err
		} else if len(msg.chain) > 0 && msg.index >= 0 && msg.index < len(msg.chain) {
			a.migrationChain = msg.chain
			a.migrationChainIndex = msg.index
			entry := msg.chain[msg.index]
			parsed, err := time.Parse(time.DateOnly, extractDateFromPath(entry.FilePath))
			if err == nil {
				a.currentDate = parsed
				return a, a.loadEntries(entry.ID)
			}
		}
		return a, nil

	case tagsLoadedMsg:
		if msg.err != nil {
			a.err = msg.err
			a.tagOptions = []string{}
			return a, nil
		}

		a.tagOptions = msg.tags
		valid := make(map[string]struct{}, len(msg.tags))
		for _, tag := range msg.tags {
			valid[tag] = struct{}{}
		}
		for tag := range a.pendingTagFilter {
			if _, exists := valid[tag]; !exists {
				delete(a.pendingTagFilter, tag)
			}
		}
		if a.tagCursor >= len(a.tagOptions) {
			a.tagCursor = max(0, len(a.tagOptions)-1)
		}
		return a, nil

	case tea.KeyMsg:
		return a.handleKeyMsg(msg)
	}

	if a.state == StateAddEntry || a.state == StateDatePicker {
		var cmd tea.Cmd
		a.input, cmd = a.input.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.state {
	case StateDailyView:
		return a.handleDailyViewKeys(msg)
	case StateAddEntry:
		return a.handleAddEntryKeys(msg)
	case StateDatePicker:
		return a.handleDatePickerKeys(msg)
	case StateTagFilter:
		return a.handleTagFilterKeys(msg)
	case StateReviewScope:
		return a.handleReviewScopeKeys(msg)
	case StateReviewTask:
		return a.handleReviewTaskKeys(msg)
	case StateReviewPrompt:
		return a.handleReviewPromptKeys(msg)
	}
	return a, nil
}

func (a *App) handleReviewPromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		a.state = StateReviewScope
		a.reviewSummary = ReviewSummary{}
		return a, nil
	case "n", "N", "esc", "q":
		a.state = StateDailyView
		return a, nil
	}
	return a, nil
}

func (a *App) handleDailyViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	a.message = ""
	a.err = nil

	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.Up):
		if a.cursor > 0 {
			a.cursor--
		}
		a.clearChainState()

	case key.Matches(msg, a.keys.Down):
		if a.cursor < len(a.entries)-1 {
			a.cursor++
		}
		a.clearChainState()

	case key.Matches(msg, a.keys.PrevDay):
		a.currentDate = a.currentDate.AddDate(0, 0, -1)
		a.clearChainState()
		return a, a.loadEntries()

	case key.Matches(msg, a.keys.NextDay):
		a.currentDate = a.currentDate.AddDate(0, 0, 1)
		a.clearChainState()
		return a, a.loadEntries()

	case key.Matches(msg, a.keys.Today):
		a.currentDate = time.Now()
		a.clearChainState()
		return a, a.loadEntries()

	case key.Matches(msg, a.keys.GoTo):
		a.state = StateDatePicker
		a.input.Reset()
		a.input.Placeholder = "YYYY-MM-DD"
		a.input.Focus()
		a.inputErr = ""
		return a, textinput.Blink

	case key.Matches(msg, a.keys.Toggle):
		return a, a.cycleEntryStatus()

	case key.Matches(msg, a.keys.Add):
		a.state = StateAddEntry
		a.input.Reset()
		a.input.Placeholder = "New entry (prefix: ! for event, - for note)"
		a.input.Focus()
		a.inputErr = ""
		return a, textinput.Blink

	case key.Matches(msg, a.keys.Review):
		a.state = StateReviewScope
		a.reviewSummary = ReviewSummary{}

	case key.Matches(msg, a.keys.Filter):
		return a, a.openTagFilter()

	case key.Matches(msg, a.keys.ClearFilter):
		if len(a.activeTagFilter) > 0 {
			a.activeTagFilter = nil
			a.activeTagMatchAll = false
			a.message = "Tag filter cleared"
			return a, a.loadEntries()
		}

	case key.Matches(msg, a.keys.Migrate):
		if !a.isToday() && len(a.entries) > 0 && a.cursor < len(a.entries) {
			entry := a.entries[a.cursor]
			if entry.Type == models.EntryTypeTask && entry.Status == models.EntryStatusOpen {
				return a, a.migrateEntryFromDailyView(entry)
			}
		}

	case key.Matches(msg, a.keys.Schedule):
		if len(a.entries) > 0 && a.cursor < len(a.entries) {
			entry := a.entries[a.cursor]
			if entry.Type == models.EntryTypeTask && entry.Status == models.EntryStatusOpen {
				a.state = StateDatePicker
				a.inputMode = "schedule_daily"
				a.input.Reset()
				a.input.Placeholder = "Schedule to YYYY-MM-DD"
				a.input.Focus()
				return a, textinput.Blink
			}
		}

	case key.Matches(msg, a.keys.ChainPrev):
		if len(a.entries) > 0 && a.cursor < len(a.entries) {
			entry := a.entries[a.cursor]
			if entry.ParentID != "" || entry.MigrationCount > 0 || entry.RescheduleCount > 0 {
				return a, a.loadMigrationChain(entry.ID, -1)
			}
		}

	case key.Matches(msg, a.keys.ChainNext):
		if len(a.entries) > 0 && a.cursor < len(a.entries) {
			entry := a.entries[a.cursor]
			return a, a.loadMigrationChain(entry.ID, 1)
		}
	}

	return a, nil
}

func (a *App) handleTagFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Cancel):
		a.state = StateDailyView
		return a, nil

	case key.Matches(msg, a.keys.Up):
		if a.tagCursor > 0 {
			a.tagCursor--
		}

	case key.Matches(msg, a.keys.Down):
		if a.tagCursor < len(a.tagOptions)-1 {
			a.tagCursor++
		}

	case key.Matches(msg, a.keys.Toggle):
		if len(a.tagOptions) > 0 && a.tagCursor < len(a.tagOptions) {
			tag := a.tagOptions[a.tagCursor]
			if a.pendingTagFilter[tag] {
				delete(a.pendingTagFilter, tag)
			} else {
				a.pendingTagFilter[tag] = true
			}
		}

	case key.Matches(msg, a.keys.ToggleFilterMode):
		a.pendingTagMatchAll = !a.pendingTagMatchAll

	case key.Matches(msg, a.keys.Confirm):
		a.activeTagFilter = a.pendingSelectedTags()
		a.activeTagMatchAll = a.pendingTagMatchAll
		a.state = StateDailyView
		if len(a.activeTagFilter) == 0 {
			a.message = "Tag filter cleared"
		} else {
			a.message = fmt.Sprintf("Applied %s tag filter (%d selected)", a.filterModeLabel(a.activeTagMatchAll), len(a.activeTagFilter))
		}
		return a, a.loadEntries()
	}

	return a, nil
}

func (a *App) handleAddEntryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Cancel):
		a.state = StateDailyView
		a.input.Reset()
		return a, nil

	case key.Matches(msg, a.keys.Confirm):
		text := a.input.Value()
		if text == "" {
			a.state = StateDailyView
			return a, nil
		}
		return a, a.addEntry(text)
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

func (a *App) handleDatePickerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Cancel):
		if a.inputMode == "schedule" {
			a.state = StateReviewTask
		} else {
			a.state = StateDailyView
		}
		a.input.Reset()
		a.inputErr = ""
		a.inputMode = ""
		return a, nil

	case key.Matches(msg, a.keys.Confirm):
		dateStr := a.input.Value()
		parsed, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			a.inputErr = "Invalid date format. Use YYYY-MM-DD"
			return a, nil
		}

		if a.inputMode == "schedule" {
			a.input.Reset()
			a.inputErr = ""
			a.inputMode = ""
			return a, a.scheduleCurrentReviewTask(dateStr)
		}

		if a.inputMode == "schedule_daily" {
			if len(a.entries) > 0 && a.cursor < len(a.entries) {
				entry := a.entries[a.cursor]
				a.input.Reset()
				a.inputErr = ""
				a.inputMode = ""
				return a, a.scheduleEntryFromDailyView(entry, dateStr)
			}
			a.state = StateDailyView
			a.input.Reset()
			a.inputErr = ""
			a.inputMode = ""
			return a, nil
		}

		a.currentDate = parsed
		a.state = StateDailyView
		a.input.Reset()
		a.inputErr = ""
		return a, a.loadEntries()
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

func (a *App) handleReviewScopeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.scopeKeys.Cancel):
		a.state = StateDailyView
		return a, nil

	case key.Matches(msg, a.scopeKeys.Option1):
		a.reviewScope = 1
		return a, a.loadReviewTasks(1)

	case key.Matches(msg, a.scopeKeys.Option2):
		a.reviewScope = 2
		return a, a.loadReviewTasks(2)

	case key.Matches(msg, a.scopeKeys.Option3):
		a.reviewScope = 3
		return a, a.loadReviewTasks(7)

	case key.Matches(msg, a.scopeKeys.Option4):
		a.reviewScope = 4
		return a, a.loadReviewTasks(0)
	}

	return a, nil
}

func (a *App) handleReviewTaskKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.reviewKeys.Cancel):
		a.state = StateDailyView
		a.currentDate = time.Now()
		return a, a.loadEntries()

	case key.Matches(msg, a.reviewKeys.Migrate):
		return a, a.migrateCurrentReviewTask()

	case key.Matches(msg, a.reviewKeys.Complete):
		return a, a.completeCurrentReviewTask()

	case key.Matches(msg, a.reviewKeys.Delete):
		return a, a.cancelCurrentReviewTask()

	case key.Matches(msg, a.reviewKeys.Keep):
		a.reviewSummary.Skipped++
		return a.advanceReview()

	case key.Matches(msg, a.reviewKeys.Schedule):
		a.state = StateDatePicker
		a.inputMode = "schedule"
		a.input.Reset()
		a.input.Placeholder = "Schedule to YYYY-MM-DD"
		a.input.Focus()
		return a, textinput.Blink
	}

	return a, nil
}

func (a *App) advanceReview() (tea.Model, tea.Cmd) {
	if a.reviewCursor >= len(a.reviewTasks)-1 {
		a.state = StateDailyView
		a.currentDate = time.Now()
		a.message = fmt.Sprintf("Review complete. Migrated: %d, Completed: %d, Cancelled: %d, Skipped: %d",
			a.reviewSummary.Migrated, a.reviewSummary.Completed, a.reviewSummary.Cancelled, a.reviewSummary.Skipped)
		return a, a.loadEntries()
	}
	a.reviewCursor++
	return a, nil
}

func (a *App) View() string {
	switch a.state {
	case StateDailyView:
		return a.renderDailyView()
	case StateAddEntry:
		return a.renderAddEntry()
	case StateDatePicker:
		return a.renderDatePicker()
	case StateTagFilter:
		return a.renderTagFilter()
	case StateReviewScope:
		return a.renderReviewScope()
	case StateReviewTask:
		return a.renderReviewTask()
	case StateReviewPrompt:
		return a.renderReviewPrompt()
	}
	return ""
}

func (a *App) clearChainState() {
	a.migrationChain = nil
	a.migrationChainIndex = 0
}

func (a *App) openTagFilter() tea.Cmd {
	a.state = StateTagFilter
	a.tagCursor = 0
	a.pendingTagMatchAll = a.activeTagMatchAll
	a.pendingTagFilter = make(map[string]bool, len(a.activeTagFilter))
	for _, tag := range a.activeTagFilter {
		a.pendingTagFilter[tag] = true
	}

	return a.loadTags()
}

func (a *App) pendingSelectedTags() []string {
	if len(a.pendingTagFilter) == 0 {
		return nil
	}

	selected := make([]string, 0, len(a.pendingTagFilter))
	for _, tag := range a.tagOptions {
		if a.pendingTagFilter[tag] {
			selected = append(selected, tag)
		}
	}
	for tag := range a.pendingTagFilter {
		if !slices.Contains(a.tagOptions, tag) {
			selected = append(selected, tag)
		}
	}
	return selected
}

func (a *App) filterModeLabel(matchAll bool) string {
	if matchAll {
		return "AND"
	}
	return "OR"
}
