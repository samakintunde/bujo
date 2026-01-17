package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Date navigation
	PrevDay key.Binding
	NextDay key.Binding
	Today   key.Binding
	GoTo    key.Binding

	// Actions
	Toggle    key.Binding
	Add       key.Binding
	Migrate   key.Binding
	Schedule  key.Binding
	Review    key.Binding
	ChainPrev key.Binding
	ChainNext key.Binding

	// General
	Confirm key.Binding
	Cancel  key.Binding
	Quit    key.Binding
	Help    key.Binding
}

// DefaultKeyMap returns the default key bindings
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "prev day"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "next day"),
	),
	PrevDay: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "prev day"),
	),
	NextDay: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "next day"),
	),
	Today: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "today"),
	),
	GoTo: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "go to date"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle status"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add entry"),
	),
	Migrate: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "migrate"),
	),
	Schedule: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "schedule"),
	),
	Review: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "review"),
	),
	ChainPrev: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "chain prev"),
	),
	ChainNext: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "chain next"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// ReviewKeyMap for review mode actions
type ReviewKeyMap struct {
	Migrate  key.Binding
	Schedule key.Binding
	Complete key.Binding
	Delete   key.Binding
	Keep     key.Binding
	Cancel   key.Binding
}

var DefaultReviewKeyMap = ReviewKeyMap{
	Migrate: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "migrate to today"),
	),
	Schedule: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "schedule"),
	),
	Complete: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "complete"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "cancel task"),
	),
	Keep: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "keep/skip"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "exit review"),
	),
}

// ScopeKeyMap for review scope selection
type ScopeKeyMap struct {
	Option1 key.Binding
	Option2 key.Binding
	Option3 key.Binding
	Option4 key.Binding
	Cancel  key.Binding
}

var DefaultScopeKeyMap = ScopeKeyMap{
	Option1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "last active day"),
	),
	Option2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "last 2 days"),
	),
	Option3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "last week"),
	),
	Option4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "all stale"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "cancel"),
	),
}
