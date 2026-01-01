# thx

A command-line interface for [Things 3](https://culturedcode.com/things/) on macOS.

**Love Things 3 but wish you could manage tasks without leaving your terminal?** `thx` lets you view, add, complete, and organize your todos right from the command line. Perfect for developers, keyboard enthusiasts, and anyone who lives in the terminal.

## Why thx?

- **Stay in flow** — Add tasks without switching apps
- **Script your workflows** — Automate task creation with shell scripts, cron jobs, or CI pipelines
- **JSON output** — Pipe your tasks to `jq`, build dashboards, or integrate with other tools
- **Fast** — Native Go binary, instant response times

## Installation

### Homebrew (recommended)

```bash
brew tap peeomid/tap
brew install thx
```

### From source

```bash
go install github.com/peeomid/thx/cmd/thx@latest
```

### Manual build

```bash
git clone https://github.com/peeomid/thx.git
cd thx
go build -o thx ./cmd/thx
sudo mv thx /usr/local/bin/
```

## Quick Start

```bash
# See what's on your plate today
thx today

# Add a task
thx add "Buy groceries" --when today

# Mark it done
thx done <task-id>

# Quick entry with natural language (opens Things quick entry)
thx quick "Call mom tomorrow #personal"
```

## Usage

### Viewing Tasks

```bash
# View tasks by list
thx inbox              # Inbox items
thx today              # Today's tasks
thx upcoming           # Upcoming scheduled tasks
thx anytime            # Anytime tasks
thx someday            # Someday tasks
thx logbook            # Completed tasks

# View organization
thx projects           # All projects
thx areas              # All areas
thx tags               # All tags

# View details
thx show <id>                      # Show task details
thx show <id> --include-checklist  # Include checklist items
```

### Searching

```bash
# Basic search
thx search "meeting"

# Filter by tag, project, or area
thx search --tag work
thx search --project "Q1 Planning"
thx search --area "Personal"

# Filter by status
thx search --status open       # Open tasks (default)
thx search --status done       # Completed tasks
thx search --status canceled   # Canceled tasks

# Filter by deadline
thx search --deadline today
thx search --deadline 2025-03-15

# Combine filters
thx search "report" --tag work --status open
```

### Adding Tasks

```bash
# Simple task
thx add "Buy milk"

# Scheduled task
thx add "Submit report" --when today
thx add "Call dentist" --when tomorrow
thx add "Review goals" --when 2025-03-15

# With deadline
thx add "Tax return" --deadline 2025-04-15

# With tags
thx add "Fix bug" --tags "work,urgent"

# With notes and checklist
thx add "Pack for trip" --notes "Don't forget passport" --checklist "Clothes,Charger,Toiletries"

# Assign to project or area
thx add "Design mockup" --project "Website Redesign"
thx add "Workout" --area "Health"
```

**Scheduling options for `--when`:**
- `inbox` — Add to Inbox (default)
- `today` — Today
- `tomorrow` — Tomorrow
- `evening` — This Evening
- `anytime` — Anytime
- `someday` — Someday
- `YYYY-MM-DD` — Specific date

### Adding Projects

```bash
thx add-project "Launch Website" --area "Work" --todos "Design,Develop,Test,Deploy"
```

### Updating Tasks

```bash
# Reschedule
thx update <id> --when tomorrow

# Change title
thx update <id> --title "New title"

# Add deadline
thx update <id> --deadline 2025-03-20

# Move to project
thx update <id> --project "Q2 Goals"
```

### Completing & Canceling

```bash
thx done <id>      # Mark complete
thx cancel <id>    # Cancel task
```

### Moving Tasks

```bash
thx move <id> --to-project "New Project"
thx move <id> --to-area "Work"
```

### Opening in Things

```bash
thx open <id>       # Open specific task in Things
thx open today      # Open Today view
thx open inbox      # Open Inbox
```

### Quick Entry

Opens Things' quick entry window with pre-filled text:

```bash
thx quick "Buy milk tomorrow #errands"
```

## Output Formats

```bash
# Default human-readable output
thx today

# JSON output (great for scripting)
thx today --json

# Quiet mode (IDs only, one per line)
thx today --quiet
```

### Example: JSON with jq

```bash
# Get titles of today's tasks
thx today --json | jq -r '.[].title'

# Count open tasks
thx search --status open --json | jq length
```

## Configuration

Create `~/.config/thx/config.yaml`:

```yaml
# Default output format: default, json, or quiet
format: default

# Custom database path (usually not needed)
database: ""

# Defaults for new tasks
defaults:
  when: inbox          # Default scheduling: inbox, today, anytime, etc.
  tags: []             # Default tags for new tasks
```

### Environment Variables

You can also configure via environment variables:

```bash
export THX_FORMAT=json
export THX_DATABASE="/path/to/database"
export THX_DEFAULTS_WHEN=today
```

## Pagination

For commands that return lists, you can paginate results:

```bash
thx inbox --limit 10           # First 10 items
thx inbox --limit 10 --offset 10   # Next 10 items
```

## Troubleshooting

### "database not found" or permission errors

`thx` reads directly from the Things 3 SQLite database. On macOS, this requires **Full Disk Access** for your terminal app.

1. Open **System Settings** → **Privacy & Security** → **Full Disk Access**
2. Add your terminal app (Terminal, iTerm2, Warp, etc.)
3. Restart your terminal

### Custom database location

If Things stores data in a non-default location, you have three options:

**Option 1: Environment variable (recommended for persistent config)**
```bash
export THX_DATABASE="/path/to/Things Database.thingsdatabase/main.sqlite"
```

Add this to your `~/.zshrc` or `~/.bashrc` to make it permanent.

**Option 2: Command-line flag**
```bash
thx today --database "/path/to/Things Database.thingsdatabase/main.sqlite"
```

**Option 3: Config file**

Add to `~/.config/thx/config.yaml`:
```yaml
database: "/path/to/Things Database.thingsdatabase/main.sqlite"
```

## How It Works

- **Reading**: `thx` queries the Things 3 SQLite database directly (read-only)
- **Writing**: `thx` uses the [Things URL Scheme](https://culturedcode.com/things/support/articles/2803573/) to add/modify tasks safely

This means `thx` never writes directly to your database — all modifications go through Things 3's official API.

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR on [GitHub](https://github.com/peeomid/thx).
