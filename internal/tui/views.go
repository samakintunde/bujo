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

	return a.renderApp(b.String())
}

func (a *App) renderHeader() string {
	dateStr := a.currentDate.Format("02, January, 2006")

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

	if len(a.activeTagFilter) > 0 {
		filterInfo := fmt.Sprintf("Filter %s: %s", a.filterModeLabel(a.activeTagMatchAll), a.tagFilterSummary(a.activeTagFilter))
		navHint += "  " + ChainStyle.Render(filterInfo)
	}

	header := fmt.Sprintf("%s  %s", dateDisplay, navHint)
	return HeaderStyle.Copy().Width(a.contentWidth()).Render(header)
}

func (a *App) renderEntryList() string {
	if len(a.entries) == 0 {
		if len(a.activeTagFilter) > 0 {
			return EmptyStateStyle.Render("No entries match current tags. Press 'f' to adjust or 'c' to clear.")
		}
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
		KeyStyle.Render("f") + DescStyle.Render("ilter"),
		KeyStyle.Render("c") + DescStyle.Render("lear"),
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

	return StatusBarStyle.Copy().Width(a.contentWidth()).Render(strings.Join(keys, "  "))
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

	return a.renderApp(b.String())
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

	return a.renderApp(ModalStyle.Copy().Width(a.contentWidth()).Render(b.String()))
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

	return a.renderApp(ModalStyle.Copy().Width(a.contentWidth()).Render(b.String()))
}

func (a *App) renderReviewTask() string {
	if len(a.reviewTasks) == 0 || a.reviewCursor >= len(a.reviewTasks) {
		return a.renderApp("No tasks to review.")
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

	return a.renderApp(ReviewCardStyle.Copy().Width(a.contentWidth()).Render(b.String()))
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

	return a.renderApp(ModalStyle.Copy().Width(a.contentWidth()).Render(b.String()))
}

func (a *App) renderTagFilter() string {
	var b strings.Builder

	title := ModalTitleStyle.Render("Filter entries by tags")
	b.WriteString(title + "\n\n")

	modeLine := fmt.Sprintf("Mode: %s (press 'o' to toggle)", a.filterModeLabel(a.pendingTagMatchAll))
	b.WriteString(ModalOptionStyle.Render(modeLine) + "\n\n")

	if len(a.tagOptions) == 0 {
		b.WriteString(EmptyStateStyle.Render("No tags found yet. Add entries with #tags first."))
		b.WriteString("\n\n")
		b.WriteString(ModalHintStyle.Render("[Enter] Apply  [Esc] Cancel"))
		return a.renderApp(ModalStyle.Copy().Width(a.contentWidth()).Render(b.String()))
	}

	for i, tag := range a.tagOptions {
		cursor := "  "
		if i == a.tagCursor {
			cursor = "> "
		}
		checkbox := "[ ]"
		if a.pendingTagFilter[tag] {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s%s #%s", cursor, checkbox, tag)
		if i == a.tagCursor {
			b.WriteString(ModalOptionSelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(ModalOptionStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n")
	selected := len(a.pendingSelectedTags())
	b.WriteString(ModalOptionStyle.Render(fmt.Sprintf("Selected: %d", selected)) + "\n")
	b.WriteString(ModalHintStyle.Render("[j/k] Move  [Space] Toggle  [o] OR/AND  [Enter] Apply  [Esc] Cancel"))

	return a.renderApp(ModalStyle.Copy().Width(a.contentWidth()).Render(b.String()))
}

func (a *App) tagFilterSummary(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	tagged := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagged = append(tagged, "#"+tag)
	}
	return strings.Join(tagged, ", ")
}

func (a *App) renderApp(content string) string {
	return AppStyle.Copy().Width(a.appWidth()).Render(content)
}

func (a *App) appWidth() int {
	if a.width <= 0 {
		return 80
	}
	return a.width
}

func (a *App) contentWidth() int {
	width := a.appWidth() - AppStyle.GetHorizontalFrameSize()
	if width < 20 {
		return 20
	}
	return width
}
