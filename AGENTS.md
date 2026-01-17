# PROJECT KNOWLEDGE BASE

**Generated:** Sat Jan 17 2026
**Commit:** v0.1.0
**Branch:** main

## OVERVIEW
**bujo** is a terminal-based Bullet Journal tool (Go, Cobra, Bubble Tea, SQLite).
Philosophy: **File over App**. Markdown files are source of truth; SQLite is an ephemeral index.

## STRUCTURE
```
bujo/
├── cmd/           # CLI entry points (add, list, root, version)
├── internal/
│   ├── tui/       # [COMPLEX] Bubble Tea UI state machine & views
│   ├── storage/   # [CRITICAL] SQLite schema & raw queries (No ORM)
│   ├── parser/    # Markdown <-> Entry parsing
│   ├── sync/      # File system <-> DB synchronization
│   ├── models/    # Domain types (Entry, Status)
│   └── git/       # Git integration
├── Makefile       # Build automation
└── main.go        # Entry point
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| **CLI Commands** | `cmd/*.go` | Cobra definitions |
| **UI Logic** | `internal/tui/` | Bubble Tea models & update loop |
| **DB Schema** | `internal/storage/sqlite.go` | Tables: `entries`, `files`, `app_state` |
| **Parsing** | `internal/parser/markdown.go` | Regex-based parsing of tasks/notes |
| **Config** | `internal/config/` | Viper setup (env vars, flags) |

## CONVENTIONS
- **Architecture**: CLI (Cobra) -> Domain Logic -> Storage (SQLite) / FS.
- **Persistence**: Hybrid.
  - **Read**: Fast queries from SQLite.
  - **Write**: Append to Markdown files -> Sync to SQLite.
- **Dates**: `YYYY-MM-DD` strings. No complex date objects in storage.
- **State Machine**: Open (`[ ]`) → Completed (`[x]`) → Migrated (`[>]`) → Scheduled (`[<]`) → Cancelled (`[-]`).
- **Testing**: Standard `go test`, no external libs. Table-driven tests preferred.

## ANTI-PATTERNS (STRICT)
- **NO ORM**: Use raw SQL via `database/sql` only.
- **NO Repository Interface**: Use `*sql.DB` directly. (YAGNI for Phase 1).
- **NO Complex Config**: Zero-config design (`~/.bujo`).
- **NO Monthly View**: Strictly Daily Log for MVP.
- **NO External Test Libs**: Use `testing` package only (no testify).

## COMMANDS
```bash
# Dev
make run        # Run via go run
make build      # Build binary
make test       # Run tests
make lint       # Run golangci-lint

# Release
git tag v0.1.0
git push origin v0.1.0 # Triggers GoReleaser via GitHub Actions
```

## NOTES
- **WAL Mode**: SQLite WAL is always enabled for concurrency.
- **Git Integration**: Optional. If `.git` exists in data dir, commits changes.
- **Versioning**: Build-time injection via `ldflags` (Version, Commit, Date).
