# bujo-cli

> **Status**: 🚧 Under Construction

A terminal-based Bullet Journal (BuJo) tool for developers who want to organize their day without leaving the command line.

It focuses on **speed** (capture fast) and **intentionality** (review often).

## Philosophy

- **Capture Speed**: `bujo add "meeting"` is faster than switching windows.
- **Focus**: A TUI interface forces you to process yesterday's unfinished work before starting today (The "Migration").
- **Local First**: Data lives in a local SQLite database (`~/.bujo.db`).

## Installation

```bash
go install github.com/samakintunde/bujo-cli@latest
```

_(Note: Not yet released, build from source for now)_

## Usage

### 1. Rapid Logging (CLI)

Capture thoughts immediately from your shell.

```bash
# Add a task for today (default)
bujo add "Review PR #42"

# Add an event
bujo add -t event "Team Standup at 10am"

# Add a note
bujo add -t note "Server IP: 192.168.1.5"

# Add a task for a specific date (YYYY-MM-DD)
bujo add -d 2024-02-01 "Doctor appointment"
```

### 2. Daily View (TUI)

Launch the interactive view to manage your day.

```bash
bujo
```

- **Space**: Toggle status (Open -> Done -> Irrelevant)
- **a**: Add new item
- **m**: Migrate task (Workflow triggered on startup if pending tasks exist)
- **q**: Quit

### 3. List (CLI)

Quickly see what's on your plate without entering the full TUI.

```bash
bujo list
```

## Data Location

All data is stored in `~/.bujo.db`. You can back this file up or sync it across machines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
