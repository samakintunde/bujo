package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSecondary = lipgloss.Color("#06B6D4") // Cyan
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorDanger    = lipgloss.Color("#EF4444") // Red
	colorText      = lipgloss.Color("#F9FAFB") // White
	colorSubtle    = lipgloss.Color("#9CA3AF") // Light gray
)

// Layout styles
var (
	// Container for the entire app
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header with date
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorMuted).
			MarginBottom(1).
			Width(50)

	// Date text in header
	DateStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	// "Today" badge
	TodayBadgeStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Navigation hints in header
	NavHintStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Entry list styles
var (
	// Selected entry (cursor on it)
	SelectedEntryStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Background(lipgloss.Color("#374151"))

	// Normal entry
	EntryStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// Cursor indicator
	CursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Chain indicator
	ChainStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			MarginLeft(1)
)

// Entry signifier styles (by status)
var (
	SignifierOpenStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	SignifierCompletedStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	SignifierMigratedStyle = lipgloss.NewStyle().
				Foreground(colorWarning)

	SignifierScheduledStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	SignifierCancelledStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Strikethrough(true)

	SignifierEventStyle = lipgloss.NewStyle().
				Foreground(colorWarning)

	SignifierNoteStyle = lipgloss.NewStyle().
				Foreground(colorSubtle)
)

// Status bar (bottom)
var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorMuted).
			MarginTop(1).
			Width(50)

	KeyStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Modal styles
var (
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Width(40)

	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	ModalOptionStyle = lipgloss.NewStyle().
				Foreground(colorText)

	ModalOptionSelectedStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Background(lipgloss.Color("#374151"))

	ModalHintStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)
)

// Input styles
var (
	InputPromptStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	InputTextStyle = lipgloss.NewStyle().
			Foreground(colorText)

	InputErrorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)
)

// Review mode styles
var (
	ReviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWarning).
				MarginBottom(1)

	ReviewCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1, 2).
			Width(50)

	ReviewTaskStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	ReviewMetaStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	ReviewActionStyle = lipgloss.NewStyle().
				Foreground(colorText)

	ReviewProgressStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Align(lipgloss.Right)

	TurtleStyle = lipgloss.NewStyle().
			Foreground(colorWarning)
)

// Empty state
var (
	EmptyStateStyle = lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		Align(lipgloss.Center)
)
