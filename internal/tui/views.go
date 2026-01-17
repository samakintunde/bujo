package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/samakintunde/bujo/internal/models"
)

func (a *App) renderDailyView() string {
	var b strings.Builder

	b.WriteString(a.renderHeader())
	b.WriteString("\n")
	b.WriteString(a.renderEntryList())
	b.WriteString("\n")
	b.WriteString(a.renderStatusBar())

	if a.message != "" {
		b.WriteString("\n")
		b.WriteString(SignifierCompletedStyle.Render(a.message))
	}

	if a.err != nil {
		b.WriteString("\n")
		b.WriteString(InputErrorStyle.Render(fmt.Sprintf("Error: %v", a.err)))
	}

	return AppStyle.Render(b.String())
}

func (a *App) renderHeader() string {
	dateStr := a.currentDate.Format("2006-01-02")

	var todayBadge string
	if a.isToday() {
		todayBadge = TodayBadgeStyle.Render(" (Today)")
	} else {
		daysAgo := int(time.Since(a.currentDate).Hours() / 24)
		if daysAgo > 0 {
			todayBadge = NavHintStyle.Render(fmt.Sprintf(" (%d days ago)", daysAgo))
		} else if daysAgo < 0 {
			todayBadge = NavHintStyle.Render(fmt.Sprintf(" (in %d days)", -daysAgo))
		}
	}

	dateDisplay := DateStyle.Render(dateStr) + todayBadge
	navHint := NavHintStyle.Render("[h←] [→l]")

	if len(a.migrationChain) > 0 {
		chainInfo := fmt.Sprintf("Chain: %d/%d", a.migrationChainIndex+1, len(a.migrationChain))
		navHint += "  " + ChainStyle.Render(chainInfo) + NavHintStyle.Render(" [[] []]")
	}

	header := fmt.Sprintf("%s  %s", dateDisplay, navHint)
	return HeaderStyle.Render(header)
}

func (a *App) renderEntryList() string {
	if len(a.entries) == 0 {
		return EmptyStateStyle.Render("No entries for this day. Press 'a' to add one.")
	}

	var b strings.Builder
	for i, entry := range a.entries {
		cursor := "  "
		if i == a.cursor {
			cursor = CursorStyle.Render("> ")
		}

		line := a.renderEntry(entry, i == a.cursor)
		b.WriteString(cursor + line + "\n")
	}
	return b.String()
}

func (a *App) renderEntry(entry models.Entry, selected bool) string {
	signifier := a.getSignifierStyled(entry)
	content := entry.Content

	if entry.Status == models.EntryStatusCancelled {
		content = SignifierCancelledStyle.Render(content)
	}

	line := fmt.Sprintf("%s %s", signifier, content)

	if entry.ParentID != "" || entry.MigrationCount > 0 || entry.RescheduleCount > 0 ||
		entry.Status == models.EntryStatusMigrated || entry.Status == models.EntryStatusScheduled {
		line += ChainStyle.Render(" 🔗")
	}

	if selected {
		return SelectedEntryStyle.Render(line)
	}
	return EntryStyle.Render(line)
}

func (a *App) getSignifierStyled(entry models.Entry) string {
	switch entry.Type {
	case models.EntryTypeTask:
		switch entry.Status {
		case models.EntryStatusOpen:
			return SignifierOpenStyle.Render("•")
		case models.EntryStatusCompleted:
			return SignifierCompletedStyle.Render("x")
		case models.EntryStatusMigrated:
			return SignifierMigratedStyle.Render(">")
		case models.EntryStatusScheduled:
			return SignifierScheduledStyle.Render("<")
		case models.EntryStatusCancelled:
			return SignifierCancelledStyle.Render("-")
		}
	case models.EntryTypeEvent:
		return SignifierEventStyle.Render("*")
	case models.EntryTypeNote:
		return SignifierNoteStyle.Render("-")
	}
	return " "
}

func (a *App) renderStatusBar() string {
	keys := []string{
		KeyStyle.Render("a") + DescStyle.Render("dd"),
		KeyStyle.Render("space") + DescStyle.Render(" toggle"),
		KeyStyle.Render("m") + DescStyle.Render("igrate"),
		KeyStyle.Render("s") + DescStyle.Render("chedule"),
		KeyStyle.Render("r") + DescStyle.Render("eview"),
		KeyStyle.Render("d") + DescStyle.Render("ate"),
		KeyStyle.Render("q") + DescStyle.Render("uit"),
	}

	if len(a.entries) > 0 && a.cursor < len(a.entries) {
		e := a.entries[a.cursor]
		if e.ParentID != "" || e.MigrationCount > 0 || e.RescheduleCount > 0 ||
			e.Status == models.EntryStatusMigrated || e.Status == models.EntryStatusScheduled {
			keys = append(keys, KeyStyle.Render("[")+DescStyle.Render("/")+KeyStyle.Render("]")+DescStyle.Render(" history"))
		}
	}

	return StatusBarStyle.Render(strings.Join(keys, "  "))
}

func (a *App) renderAddEntry() string {
	var b strings.Builder

	b.WriteString(a.renderHeader())
	b.WriteString("\n")
	b.WriteString(a.renderEntryList())
	b.WriteString("\n")

	prompt := InputPromptStyle.Render("> ")
	b.WriteString(prompt + a.input.View())

	if a.inputErr != "" {
		b.WriteString("\n" + InputErrorStyle.Render(a.inputErr))
	}

	b.WriteString("\n")
	b.WriteString(ModalHintStyle.Render("[Enter] Add  [Esc] Cancel"))

	return AppStyle.Render(b.String())
}

func (a *App) renderDatePicker() string {
	var b strings.Builder

	title := ModalTitleStyle.Render("Go to date:")
	b.WriteString(title + "\n\n")

	prompt := InputPromptStyle.Render("> ")
	b.WriteString(prompt + a.input.View())

	if a.inputErr != "" {
		b.WriteString("\n" + InputErrorStyle.Render(a.inputErr))
	}

	b.WriteString("\n\n")
	b.WriteString(ModalHintStyle.Render("[Enter] Confirm  [Esc] Cancel"))

	return AppStyle.Render(ModalStyle.Render(b.String()))
}

func (a *App) renderReviewScope() string {
	var b strings.Builder

	title := ModalTitleStyle.Render("Review tasks from:")
	b.WriteString(title + "\n\n")

	options := []string{
		fmt.Sprintf("[1] Last active day  (%d tasks)", a.countStaleTasks(1)),
		fmt.Sprintf("[2] Last 2 days      (%d tasks)", a.countStaleTasks(2)),
		fmt.Sprintf("[3] Last week        (%d tasks)", a.countStaleTasks(7)),
		fmt.Sprintf("[4] All stale        (%d tasks)", a.countStaleTasks(0)),
	}

	for _, opt := range options {
		b.WriteString(ModalOptionStyle.Render(opt) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(ModalHintStyle.Render("[1-4] Select  [Esc] Cancel"))

	return AppStyle.Render(ModalStyle.Render(b.String()))
}

func (a *App) renderReviewTask() string {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return AppStyle.Render("No tasks to review.")
	}

	task := a.reviewTasks[a.reviewCursor]
	var b strings.Builder

	header := ReviewHeaderStyle.Render(fmt.Sprintf("REVIEW: %d stale tasks", len(a.reviewTasks)))
	b.WriteString(header + "\n\n")

	taskContent := ReviewTaskStyle.Render(task.Content)

	daysAgo := int(time.Since(task.CreatedAt).Hours() / 24)
	fromDate := ReviewMetaStyle.Render(fmt.Sprintf("From: %s (%d days ago)", task.CreatedAt.Format("2006-01-02"), daysAgo))

	migrated := ReviewMetaStyle.Render(fmt.Sprintf("Migrated: %d times", task.MigrationCount))
	if task.MigrationCount > 3 {
		migrated = migrated + " " + TurtleStyle.Render("🐢")
	}

	b.WriteString(taskContent + "\n")
	b.WriteString(fromDate + "\n")
	b.WriteString(migrated + "\n\n")

	actions := []string{
		KeyStyle.Render("[m]") + " Migrate to today",
		KeyStyle.Render("[s]") + " Schedule for later",
		KeyStyle.Render("[x]") + " Mark complete",
		KeyStyle.Render("[d]") + " Cancel task",
		KeyStyle.Render("[k]") + " Keep (skip)",
	}
	for _, action := range actions {
		b.WriteString(ReviewActionStyle.Render(action) + "\n")
	}

	b.WriteString("\n")
	progress := ReviewProgressStyle.Render(fmt.Sprintf("[%d/%d]", a.reviewCursor+1, len(a.reviewTasks)))
	b.WriteString(progress)

	return AppStyle.Render(ReviewCardStyle.Render(b.String()))
}

func (a *App) isToday() bool {
	now := time.Now()
	return a.currentDate.Year() == now.Year() &&
		a.currentDate.Month() == now.Month() &&
		a.currentDate.Day() == now.Day()
}

func (a *App) countStaleTasks(daysBack int) int {
	count, _ := a.db.CountStaleTasks(daysBack)
	return count
}

func (a *App) renderReviewPrompt() string {
	var b strings.Builder

	title := ModalTitleStyle.Render("Review stale tasks?")
	b.WriteString(title + "\n\n")

	msg := fmt.Sprintf("You have %d stale task(s) from previous days.", a.staleTaskCount)
	b.WriteString(ModalOptionStyle.Render(msg) + "\n\n")

	b.WriteString(ModalOptionStyle.Render("Review them now?") + "\n\n")

	b.WriteString(ModalHintStyle.Render("[y] Yes  [n] No"))

	return AppStyle.Render(ModalStyle.Render(b.String()))
}
