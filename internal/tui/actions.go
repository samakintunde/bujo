package tui

import (
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samakintunde/bujo/internal/models"
)

func (a *App) cycleEntryStatus() tea.Cmd {
	if len(a.entries) == 0 || a.cursor >= len(a.entries) {
		return nil
	}

	entry := a.entries[a.cursor]

	if entry.Type != models.EntryTypeTask {
		return nil
	}

	var newStatus models.EntryStatus
	switch entry.Status {
	case models.EntryStatusOpen:
		newStatus = models.EntryStatusCompleted
	case models.EntryStatusCompleted:
		newStatus = models.EntryStatusCancelled
	case models.EntryStatusCancelled:
		newStatus = models.EntryStatusOpen
	default:
		return nil
	}

	return func() tea.Msg {
		err := a.service.UpdateEntryStatus(entry, newStatus)
		return entryUpdatedMsg{err: err}
	}
}

func (a *App) addEntry(text string) tea.Cmd {
	return func() tea.Msg {
		entryType := models.EntryTypeTask
		content := text

		if strings.HasPrefix(text, "!") {
			entryType = models.EntryTypeEvent
			content = strings.TrimPrefix(text, "!")
			content = strings.TrimSpace(content)
		} else if strings.HasPrefix(text, "-") {
			entryType = models.EntryTypeNote
			content = strings.TrimPrefix(text, "-")
			content = strings.TrimSpace(content)
		}

		_, err := a.service.AddEntry(content, entryType, a.currentDate)
		return entryAddedMsg{err: err}
	}
}

func (a *App) loadEntries(targetID ...string) tea.Cmd {
	return func() tea.Msg {
		tid := ""
		if len(targetID) > 0 {
			tid = targetID[0]
		}

		entries, err := a.service.GetEntriesByDate(a.currentDate)
		return entriesLoadedMsg{entries: entries, err: err, targetID: tid}
	}
}

func (a *App) loadReviewTasks(daysBack int) tea.Cmd {
	return func() tea.Msg {
		tasks, err := a.service.GetStaleTasks(daysBack)
		return reviewTasksLoadedMsg{tasks: tasks, err: err}
	}
}

func (a *App) migrateCurrentReviewTask() tea.Cmd {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return nil
	}

	task := a.reviewTasks[a.reviewCursor]
	return func() tea.Msg {
		_, err := a.service.MigrateTask(task)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}
		a.reviewSummary.Migrated++
		return reviewActionCompleteMsg{}
	}
}

func (a *App) completeCurrentReviewTask() tea.Cmd {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return nil
	}

	task := a.reviewTasks[a.reviewCursor]
	return func() tea.Msg {
		err := a.service.UpdateEntryStatus(task, models.EntryStatusCompleted)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}
		a.reviewSummary.Completed++
		return reviewActionCompleteMsg{}
	}
}

func (a *App) cancelCurrentReviewTask() tea.Cmd {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return nil
	}

	task := a.reviewTasks[a.reviewCursor]
	return func() tea.Msg {
		err := a.service.UpdateEntryStatus(task, models.EntryStatusCancelled)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}
		a.reviewSummary.Cancelled++
		return reviewActionCompleteMsg{}
	}
}

func (a *App) scheduleCurrentReviewTask(targetDate string) tea.Cmd {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return nil
	}

	task := a.reviewTasks[a.reviewCursor]
	return func() tea.Msg {
		parsed, err := time.Parse(time.DateOnly, targetDate)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		_, err = a.service.ScheduleTask(task, parsed)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}
		a.reviewSummary.Scheduled++
		return reviewActionCompleteMsg{}
	}
}

type reviewActionCompleteMsg struct{}

func (a *App) migrateEntryFromDailyView(entry models.Entry) tea.Cmd {
	return func() tea.Msg {
		_, err := a.service.MigrateTask(entry)
		return entryUpdatedMsg{err: err}
	}
}

func (a *App) scheduleEntryFromDailyView(entry models.Entry, targetDate string) tea.Cmd {
	return func() tea.Msg {
		parsed, err := time.Parse(time.DateOnly, targetDate)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		_, err = a.service.ScheduleTask(entry, parsed)
		return entryUpdatedMsg{err: err}
	}
}

func (a *App) loadMigrationChain(entryID string, direction int) tea.Cmd {
	return func() tea.Msg {
		chain, err := a.service.GetMigrationChain(entryID)
		if err != nil || len(chain) == 0 {
			return chainLoadedMsg{err: err}
		}

		currentIdx := -1
		for i, e := range chain {
			if e.ID == entryID {
				currentIdx = i
				break
			}
		}

		if currentIdx == -1 {
			return chainLoadedMsg{}
		}

		newIdx := currentIdx + direction
		if newIdx < 0 || newIdx >= len(chain) {
			return chainLoadedMsg{}
		}

		return chainLoadedMsg{chain: chain, index: newIdx}
	}
}

func (a *App) checkFirstOpenToday() tea.Cmd {
	return func() tea.Msg {
		lastOpened, _ := a.service.GetLastOpenedAt()
		today := time.Now()

		isFirstOpen := lastOpened.Year() != today.Year() ||
			lastOpened.Month() != today.Month() ||
			lastOpened.Day() != today.Day()

		staleCount, _ := a.service.CountStaleTasks(0)

		_ = a.service.SetLastOpenedAt(today)

		return initCheckMsg{isFirstOpenToday: isFirstOpen, staleTaskCount: staleCount}
	}
}

func extractDateFromPath(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
