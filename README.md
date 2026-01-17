# bujo

> **Terminal-based Bullet Journal for developers**
>
> ![Status](https://img.shields.io/badge/status-v0.1.0--(Beta)-blue) ![License](https://img.shields.io/badge/license-GPLv3-green)

`bujo` is a minimalist, terminal-based Bullet Journaling tool designed for developers who value speed, intentionality, and the "File over App" philosophy. It allows you to manage your daily tasks, events, and notes without ever leaving your terminal.

## Philosophy

- **Speed**: Capture thoughts and tasks instantly via the CLI.
- **Local-first (Markdown files)**: Your data lives in plain Markdown files on your machine. No proprietary formats, no cloud lock-in.
- **Intentionality**: Built-in workflows for migrating tasks and reviewing stale items to keep you focused on what matters.

## Installation

### Homebrew (macOS/Linux)
```bash
brew install samakintunde/tap/bujo
```

### Go
```bash
go install github.com/samakintunde/bujo@latest
```

### Manual
Download the pre-compiled binaries for your platform from the [GitHub Releases](https://github.com/samakintunde/bujo/releases) page.

## Quick Start / Workflow

### 1. Capture (CLI)
Capture entries instantly from your shell. By default, entries are added to today's log as tasks.

```bash
# Add a task for today
bujo add "Finish the project documentation"

# Add an event
bujo add -t event "Sprint Planning at 2 PM"

# Add a note
bujo add -t note "Meeting ID: 123-456-789"
```

### 2. Plan (TUI)
Launch the interactive Terminal User Interface (TUI) to manage your day, migrate tasks, and review your progress.

```bash
bujo
```

## TUI Keybindings

| Key | Action | Description |
| :--- | :--- | :--- |
| **Navigation** | | |
| `j` / `k` | **Move Cursor** | Select next/previous entry |
| `h` / `l` | **Change Date** | Navigate to previous/next day |
| `t` | **Jump Today** | Go straight to today's log |
| **Actions** | | |
| `Space` | **Toggle State** | Cycle: Open → Done → Cancelled → Open |
| `a` | **Add** | Add a new entry to the current day |
| `m` | **Migrate** | Move open task to today |
| `s` | **Schedule** | Move open task to a specific future date |
| **Advanced** | | |
| `r` | **Review** | Enter **Review Mode** to process stale tasks |
| `[` / `]` | **History** | Trace a task's migration history backward/forward |
| `q` | **Quit** | Exit the application |

## Data Storage

`bujo` follows the "File over App" philosophy. Your data lives in `~/.bujo` (or your configured path) as plain Markdown files.

The directory structure follows a logical date-based hierarchy:
```
~/.bujo/
└── 2026/
    └── 01/
        └── 2026-01-17.md
```

You can open and edit these files directly with any text editor. `bujo` will automatically sync changes when you launch the TUI or use CLI commands.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
