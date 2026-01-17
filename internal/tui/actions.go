package tui

import (
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samakintunde/bujo-cli/internal/models"
)

func (a *App) cycleEntryStatus() tea.Cmd {
	if len(a.entries) == 0 || a.cursor >= len(a.entries) {
		return nil
	}

	entry := &a.entries[a.cursor]

	if entry.Type != models.EntryTypeTask {
		return nil
	}

	switch entry.Status {
	case models.EntryStatusOpen:
		entry.Status = models.EntryStatusCompleted
	case models.EntryStatusCompleted:
		entry.Status = models.EntryStatusCancelled
	case models.EntryStatusCancelled:
		entry.Status = models.EntryStatusOpen
	default:
		return nil
	}

	return a.updateEntryInFile(*entry)
}

func (a *App) updateEntryInFile(entry models.Entry) tea.Cmd {
	return func() tea.Msg {
		newLine := entry.RawString()
		err := a.fs.UpdateLine(entry.FilePath, entry.LineNumber, newLine)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(entry.FilePath); err != nil {
			return entryUpdatedMsg{err: err}
		}

		return entryUpdatedMsg{err: nil}
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

		entry := models.NewEntry(entryType, content)

		dateStr := a.currentDate.Format(time.DateOnly)
		path, err := a.fs.EnsureDayPath(dateStr)
		if err != nil {
			return entryAddedMsg{err: err}
		}

		line := entry.RawString()
		if err := a.fs.AppendLine(path, line); err != nil {
			return entryAddedMsg{err: err}
		}

		if err := a.syncer.SyncFile(path); err != nil {
			return entryAddedMsg{err: err}
		}

		return entryAddedMsg{err: nil}
	}
}

func (a *App) loadReviewTasks(daysBack int) tea.Cmd {
	return func() tea.Msg {
		tasks, err := a.db.GetStaleTasks(daysBack)
		return reviewTasksLoadedMsg{tasks: tasks, err: err}
	}
}

func (a *App) migrateCurrentReviewTask() tea.Cmd {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return nil
	}

	task := a.reviewTasks[a.reviewCursor]
	return func() tea.Msg {
		task.Status = models.EntryStatusMigrated
		newLine := task.RawString()
		if err := a.fs.UpdateLine(task.FilePath, task.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		migratedTask := models.NewEntry(models.EntryTypeTask, task.Content)
		migratedTask.MigrationCount = task.MigrationCount + 1
		migratedTask.ParentID = task.ID

		todayPath, err := a.fs.EnsureDayPath(time.Now().Format(time.DateOnly))
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.fs.AppendLine(todayPath, migratedTask.RawString()); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(task.FilePath); err != nil {
			return entryUpdatedMsg{err: err}
		}
		if err := a.syncer.SyncFile(todayPath); err != nil {
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
		task.Status = models.EntryStatusCompleted
		newLine := task.RawString()
		if err := a.fs.UpdateLine(task.FilePath, task.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(task.FilePath); err != nil {
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
		task.Status = models.EntryStatusCancelled
		newLine := task.RawString()
		if err := a.fs.UpdateLine(task.FilePath, task.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(task.FilePath); err != nil {
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
		task.Status = models.EntryStatusScheduled
		newLine := task.RawString()
		if err := a.fs.UpdateLine(task.FilePath, task.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		scheduledTask := models.NewEntry(models.EntryTypeTask, task.Content)
		scheduledTask.RescheduleCount = task.RescheduleCount + 1
		scheduledTask.ParentID = task.ID

		targetPath, err := a.fs.EnsureDayPath(targetDate)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.fs.AppendLine(targetPath, scheduledTask.RawString()); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(task.FilePath); err != nil {
			return entryUpdatedMsg{err: err}
		}
		if err := a.syncer.SyncFile(targetPath); err != nil {
			return entryUpdatedMsg{err: err}
		}

		a.reviewSummary.Scheduled++
		return reviewActionCompleteMsg{}
	}
}

type reviewActionCompleteMsg struct{}

func (a *App) migrateEntryFromDailyView(entry models.Entry) tea.Cmd {
	return func() tea.Msg {
		entry.Status = models.EntryStatusMigrated
		newLine := entry.RawString()
		if err := a.fs.UpdateLine(entry.FilePath, entry.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		migratedTask := models.NewEntry(models.EntryTypeTask, entry.Content)
		migratedTask.MigrationCount = entry.MigrationCount + 1
		migratedTask.ParentID = entry.ID

		todayPath, err := a.fs.EnsureDayPath(time.Now().Format(time.DateOnly))
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.fs.AppendLine(todayPath, migratedTask.RawString()); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(entry.FilePath); err != nil {
			return entryUpdatedMsg{err: err}
		}
		if err := a.syncer.SyncFile(todayPath); err != nil {
			return entryUpdatedMsg{err: err}
		}

		return entryUpdatedMsg{err: nil}
	}
}

func (a *App) scheduleEntryFromDailyView(entry models.Entry, targetDate string) tea.Cmd {
	return func() tea.Msg {
		entry.Status = models.EntryStatusScheduled
		newLine := entry.RawString()
		if err := a.fs.UpdateLine(entry.FilePath, entry.LineNumber, newLine); err != nil {
			return entryUpdatedMsg{err: err}
		}

		scheduledTask := models.NewEntry(models.EntryTypeTask, entry.Content)
		scheduledTask.RescheduleCount = entry.RescheduleCount + 1
		scheduledTask.ParentID = entry.ID

		targetPath, err := a.fs.EnsureDayPath(targetDate)
		if err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.fs.AppendLine(targetPath, scheduledTask.RawString()); err != nil {
			return entryUpdatedMsg{err: err}
		}

		if err := a.syncer.SyncFile(entry.FilePath); err != nil {
			return entryUpdatedMsg{err: err}
		}
		if err := a.syncer.SyncFile(targetPath); err != nil {
			return entryUpdatedMsg{err: err}
		}

		return entryUpdatedMsg{err: nil}
	}
}

func (a *App) loadMigrationChain(entryID string, direction int) tea.Cmd {
	return func() tea.Msg {
		chain, err := a.db.GetMigrationChain(entryID)
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

func extractDateFromPath(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
