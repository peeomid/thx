# thx - Things CLI

A fast, scriptable command-line interface for Things 3.

## Overview

**thx** (pronounced "thanks") is a Go-based CLI for interacting with Things 3 on macOS. It provides read access via direct SQLite database queries and write access via Things URL Scheme.

## Goals

1. **Fast** - Native Go binary, instant startup
2. **Scriptable** - JSON output, composable with Unix tools
3. **Simple** - Intuitive commands that mirror Things concepts
4. **Offline** - Works entirely locally, no network required
5. **Predictable** - Clear list semantics, stable identifiers, consistent filters

## Non-Goals

- GUI or TUI (keep it pure CLI)
- Sync with Things Cloud (out of scope)
- Cross-platform (Things 3 is macOS only)

---

## Research: Existing Implementations

| Project | Language | Approach | Useful For |
|---------|----------|----------|------------|
| [tli](https://github.com/changkun/tli) | Go | Email to Things inbox | Write only (email-based), not DB |
| [things.py](https://github.com/thingsapi/things.py) | Python | SQLite read | **Best schema reference** |
| [things.sh](https://github.com/AlexanderWillner/things.sh) | Bash | SQLite read | **SQL query reference** |
| [things-cli](https://github.com/thingsapi/things-cli) | Python | Uses things.py | CLI design reference |
| [things-mcp](https://github.com/hald/things-mcp) | Python | SQLite + URL Scheme | MCP integration reference |

**Key finding:** No existing Go library for reading Things 3 database - need to implement from scratch.

---

## Architecture

### Reading Data

Things 3 stores data in a SQLite database. Location varies by version:

```
# Version 3.15.16+
~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite

# Earlier versions
~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite
```

Direct read access is safe and used by other tools (things-py, things-cli, things.sh).

**Permissions note:** reading the Things database may require Full Disk Access on macOS. The CLI should surface a clear error and remediation hint if the file is not readable.

### Writing Data

Things 3 doesn't support direct database writes. Use [Things URL Scheme](https://culturedcode.com/things/help/url-scheme/):
```
things:///add?title=Buy%20milk&when=today
```

Execute via `open` command on macOS.

---

## Database Schema

Reference: [things.py database docs](https://thingsapi.github.io/things.py/things/database.html)

### Tables

| Table | Purpose |
|-------|---------|
| `TMTask` | Tasks, projects, headings (main table) |
| `TMArea` | Areas |
| `TMTag` | Tags |
| `TMTaskTag` | Task-tag junction table |
| `TMAreaTag` | Area-tag junction table |
| `TMChecklistItem` | Checklist items within tasks |
| `TMSettings` | Application settings |
| `Meta` | Metadata including database version |

### TMTask Key Columns

| Column | Type | Values/Description |
|--------|------|-------------------|
| `uuid` | TEXT | Primary key (UUID) |
| `title` | TEXT | Task title |
| `status` | INT | 0=open, 2=canceled, 3=completed |
| `type` | INT | 0=todo, 1=project, 2=heading |
| `start` | INT | 0=inbox, 1=anytime, 2=someday |
| `trashed` | INT | 0=no, 1=yes |
| `area` | TEXT | Area UUID (FK) |
| `project` | TEXT | Project UUID (FK) |
| `heading` | TEXT | Heading UUID (FK) |
| `notes` | TEXT | Markdown notes |
| `index` | INT | Sort order |
| `todayIndex` | INT | Today view sort order |
| `creationDate` | REAL | Unix timestamp |
| `userModificationDate` | REAL | Unix timestamp |
| `stopDate` | REAL | Completion timestamp |
| `startDate` | INT | Things date format (see below) |
| `deadline` | INT | Things date format (see below) |

### Date Encoding

Things uses custom binary encoding for dates:

```go
// Things date to ISO date
func thingsDateToISO(date int64) (year, month, day int) {
    year = int(date >> 16)
    month = int((date >> 12) & 0xF)
    day = int((date >> 7) & 0x1F)
    return
}

// ISO date to Things date
func isoToThingsDate(year, month, day int) int64 {
    return int64((year << 16) | (month << 12) | (day << 7))
}
```

### Common Query Patterns

From [things.sh](https://github.com/AlexanderWillner/things.sh):

```sql
-- Today's tasks
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND start = 1 AND startDate IS NOT NULL
ORDER BY todayIndex;

-- Inbox
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND start = 0 AND type = 0;

-- Projects
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND type = 1;

-- Tasks with tags (join)
SELECT t.*, GROUP_CONCAT(tag.title) as tags
FROM TMTask t
LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
LEFT JOIN TMTag tag ON tt.tags = tag.uuid
WHERE t.status = 0 AND t.trashed = 0
GROUP BY t.uuid;
```

---

## Commands

### List Commands (Read)

```bash
# View lists
thx inbox                    # Show inbox items
thx today                    # Show today's tasks
thx upcoming                 # Show upcoming tasks
thx anytime                  # Show anytime tasks
thx someday                  # Show someday tasks
thx logbook                  # Show completed tasks

# View by entity
thx projects                 # List all projects
thx areas                    # List all areas
thx tags                     # List all tags

# View specific items
thx show <id>                # Show item details
thx show <project-name>      # Show project and its tasks
```

### Search Commands (Read)

```bash
thx search "query"           # Search by title/notes
thx search --tag "work"      # Filter by tag
thx search --project "proj"  # Filter by project
thx search --area "area"     # Filter by area
thx search --status done     # Filter by status (open/done/canceled)
thx search --deadline today  # Filter by deadline
```

### Create Commands (Write)

```bash
# Add todo
thx add "Buy milk"
thx add "Buy milk" --when today
thx add "Buy milk" --when tomorrow --deadline 2024-01-15
thx add "Buy milk" --tags "errands,shopping"
thx add "Buy milk" --project "Groceries"
thx add "Buy milk" --checklist "Whole milk,Eggs,Bread"
thx add "Buy milk" --notes "From Costco"

# Add project
thx add-project "Q1 Planning"
thx add-project "Q1 Planning" --area "Work" --deadline 2024-03-31
thx add-project "Q1 Planning" --todos "Task 1,Task 2,Task 3"

# Quick add (natural language, opens Things quick entry)
thx quick "Buy milk tomorrow #errands"
```

### Update Commands (Write)

```bash
# Complete/cancel
thx done <id>                # Mark as complete
thx cancel <id>              # Mark as canceled

# Update fields
thx update <id> --title "New title"
thx update <id> --when today
thx update <id> --deadline 2024-01-15
thx update <id> --tags "work,urgent"
thx update <id> --project "Project Name"
thx update <id> --notes "Additional notes"

# Move
thx move <id> --to-project "Project"
thx move <id> --to-area "Area"
```

### Utility Commands

```bash
thx open <id>                # Open item in Things app
thx open today               # Open Today view in Things
thx open inbox               # Open Inbox in Things

thx version                  # Show version
thx help                     # Show help
```

---

## Output Formats

### Default (Human-readable)

```
$ thx today
□ Buy milk                           [errands]     due: today
□ Review PR #123                     [work]
■ Write documentation                [work]        3/5 items
```

### JSON (--json flag)

```bash
$ thx today --json
```

```json
[
  {
    "id": "ABC123",
    "title": "Buy milk",
    "status": "open",
    "when": "today",
    "deadline": "2024-01-10",
    "tags": ["errands"],
    "project": null,
    "area": "Personal",
    "notes": "",
    "checklist": [],
    "created": "2024-01-09T10:00:00Z",
    "modified": "2024-01-09T10:00:00Z"
  }
]
```

### Quiet (--quiet flag)

```bash
$ thx today --quiet
ABC123
DEF456
GHI789
```

Outputs only IDs, useful for scripting:
```bash
thx today --quiet | xargs -I {} thx done {}
```

---

## Configuration

Config file at `~/.config/thx/config.yaml` (optional):

```yaml
# Default output format
format: default  # default | json | quiet

# Custom database path (if non-standard)
database: ~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite

# Default values for new todos
defaults:
  when: inbox  # inbox | today | tomorrow | evening | someday | anytime
  tags: []
```

---

## Semantics and Filtering

### Config Precedence

1. CLI flags
2. Environment variables (prefix `THX_`)
3. Config file
4. Built-in defaults

### ID Resolution

- `id` is always the Things UUID in JSON/quiet output.
- Commands that accept `<id>` also accept an exact title match; if multiple matches exist, return an error with suggestions.

### List Semantics

- All list commands default to `status = open` and `trashed = 0` unless explicitly querying logbook or status filters.
- `logbook` includes completed and canceled items (not trashed), ordered by completion date.
- `upcoming` includes items with a future `startDate` or `deadline` within the next 14 days (configurable via `--range-days`).

### When Mapping

| CLI "when" | DB fields |
|------------|-----------|
| inbox | `start = 0` |
| today | `start = 1` and `startDate = today` |
| tomorrow | `start = 1` and `startDate = tomorrow` |
| evening | `start = 1` and `startDate = today` + `todayIndex` after "Evening" separator (best-effort) |
| anytime | `start = 1` and `startDate IS NULL` |
| someday | `start = 2` |
| date | `start = 1` and `startDate = <encoded date>` |

### Checklist Support

- Read: include checklist items for `show` and `--include-checklist`.
- Write: `--checklist` accepts comma-separated values and maps to Things URL scheme.

### Pagination

- Large list commands should support `--limit` and `--offset` to aid scripting.

---

## Data Model

### Todo

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier (UUID) |
| title | string | Task title |
| status | string | open, completed, canceled |
| type | string | to-do, project, heading |
| when | string | today, tomorrow, evening, anytime, someday, or date |
| deadline | string | ISO date (YYYY-MM-DD) |
| tags | []string | Tag names |
| project | string | Project ID (nullable) |
| project_title | string | Project name |
| area | string | Area ID (nullable) |
| area_title | string | Area name |
| notes | string | Markdown notes |
| checklist | []ChecklistItem | Checklist items |
| created | datetime | Creation timestamp |
| modified | datetime | Last modified timestamp |

### ChecklistItem

| Field | Type | Description |
|-------|------|-------------|
| title | string | Item text |
| completed | bool | Completion status |

### Project

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |
| title | string | Project name |
| status | string | open, completed, canceled |
| when | string | Schedule |
| deadline | string | ISO date |
| tags | []string | Tag names |
| area | string | Area ID |
| area_title | string | Area name |
| notes | string | Project notes |
| todos | []Todo | Tasks in project |
| headings | []Heading | Section headings |

### Area

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |
| title | string | Area name |
| projects | []Project | Projects in area |
| todos | []Todo | Loose tasks in area |

---

## Implementation Notes

### Go Libraries

- **SQLite**: `modernc.org/sqlite` (pure Go, no CGO)
- **SQLite (optional)**: `github.com/mattn/go-sqlite3` (CGO, faster; enable via build tag)
- **CLI**: `github.com/spf13/cobra`
- **Config**: `github.com/spf13/viper`
- **JSON**: `encoding/json` (stdlib)

### URL Scheme Execution

```go
func executeURL(url string) error {
    return exec.Command("open", url).Run()
}
```

### Database Location

```go
func getDatabasePath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home,
        "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac",
        "Things Database.thingsdatabase/main.sqlite")
}
```

### Database Path Resolution

1. CLI flag `--database`
2. Config value `database`
3. Default path
4. If default path not found, search `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`
5. If multiple matches, choose the most recently modified

### Error Messages

Errors should include a short actionable hint, for example:
- Missing DB path: "Database not found. Set --database or grant Full Disk Access."
- Permission denied: "Permission denied reading Things DB. Grant Full Disk Access in System Settings."

---

## Example Workflows

### Morning Review

```bash
# See what's due today
thx today

# Check inbox and process
thx inbox
thx update ABC123 --when today
thx move DEF456 --to-project "Work"
```

### Quick Capture

```bash
# Add task from terminal
thx add "Call dentist" --when tomorrow --tags "health"

# Pipe from another command
echo "Review: $(git log -1 --pretty=%s)" | xargs thx add
```

### Automation

```bash
# Complete all tasks in a project
thx search --project "Sprint 1" --quiet | xargs -I {} thx done {}

# Export today's tasks to JSON
thx today --json > today-backup.json

# Count tasks by tag
thx search --tag "work" --quiet | wc -l
```

### Integration with other tools

```bash
# Create task from clipboard
pbpaste | xargs -I {} thx add "{}"

# fzf integration
thx today --quiet | fzf --preview 'thx show {}' | xargs thx open
```

---

## Release Plan

### v0.1.0 - MVP

- [ ] Read: inbox, today, upcoming, anytime, someday, logbook
- [ ] Read: projects, areas, tags
- [ ] Read: show, search (basic)
- [ ] Write: add (basic - title, when, tags)
- [ ] Output: default, json, quiet
- [ ] Command: open, version, help

### v0.2.0 - Full CRUD

- [ ] Write: add (all options - notes, checklist, deadline, project)
- [ ] Write: add-project
- [ ] Write: done, cancel
- [ ] Write: update, move
- [ ] Search: all filters

### v0.3.0 - Polish

- [ ] Config file support
- [ ] Shell completions (bash, zsh, fish)
- [ ] Man pages
- [ ] Homebrew formula

---

## Testing Plan

- Unit tests for Things date encoding/decoding and date range logic.
- Unit tests for query builders (filters, list semantics, pagination).
- Unit tests for URL scheme generation for add/update/move/done/cancel.
- Config precedence tests (flags/env/config/defaults).
- Integration tests with a temporary SQLite fixture DB (seeded with minimal TMTask/TMTag/TMArea tables).

---

## References

### Official
- [Things URL Scheme](https://culturedcode.com/things/help/url-scheme/) - Official URL scheme docs
- [Things Data Export](https://culturedcode.com/things/support/articles/2982272/) - Official export support

### Database Schema
- [things.py database docs](https://thingsapi.github.io/things.py/things/database.html) - **Best schema reference**
- [things.py GitHub](https://github.com/thingsapi/things.py) - Python library source

### CLI References
- [things.sh](https://github.com/AlexanderWillner/things.sh) - Bash CLI with SQL queries
- [things-cli](https://github.com/thingsapi/things-cli) - Python CLI
- [tli](https://github.com/changkun/tli) - Go CLI (email-based only)

### Other Implementations
- [things-mcp](https://github.com/hald/things-mcp) - MCP server (Python)
- [pythings](https://github.com/mdbraber/pythings) - Python with Peewee ORM
- [things-api GraphQL](https://github.com/evelion-apps/things-api) - Unofficial GraphQL API

### Go Libraries to Use
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Pure Go SQLite (no CGO)
- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - CGO SQLite (faster)
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/spf13/viper](https://github.com/spf13/viper) - Config management
