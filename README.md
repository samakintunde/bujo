# bujo-cli

> **Status**: v0.1.0 (Beta)

A terminal-based Bullet Journal (BuJo) tool for developers who want to organize their day without leaving the command line.

It focuses on **speed** (capture fast) and **intentionality** (review often).

## Philosophy

- **[File over app](https://stephango.com/file-over-app)**
- **Speed**
- **Focus**
- **Local First**
- **Git-friendly**

## Installation

### Homebrew (macOS/Linux)

```bash
brew install samakintunde/tap/bujo
```

### Go Install

```bash
go install github.com/samakintunde/bujo@latest
```

### Manual

Download the latest binary from the [Releases page](https://github.com/samakintunde/bujo/releases).

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

All data is stored in `~/.bujo`

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
