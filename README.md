# thx

A fast, scriptable command-line interface for Things 3 on macOS.

## Build

```bash
go build -o thx ./cmd/thx
```

## Usage

```bash
./thx today
./thx inbox --json
./thx upcoming --range-days 7
./thx projects
./thx show <id-or-title> --include-checklist
./thx search "query" --tag work --status open
```

## Writing to Things

```bash
./thx add "Buy milk" --when today --tags "errands"
./thx add-project "Q1 Planning" --area "Work" --todos "Task 1,Task 2"
./thx update <id> --title "New title" --when tomorrow
./thx done <id>
./thx cancel <id>
./thx move <id> --to-project "Project" --to-area "Area"
./thx quick "Buy milk tomorrow #errands"
```

## Database Path

If your Things database is not in the default location, pass `--database`:

```bash
./thx today --database "/Users/you/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite"
```

Note: reading the Things database may require granting Full Disk Access to your terminal.
