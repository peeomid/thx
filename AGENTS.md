# AGENTS

This repo contains `thx`, a Go CLI for Things 3 on macOS.

## Working Agreement

- Read data from the Things SQLite database; never write to it directly.
- Write operations must use the Things URL scheme via the `open` command.
- Prefer `modernc.org/sqlite` (pure Go) unless CGO is explicitly requested.
- Keep commands scriptable: JSON and quiet outputs should remain stable.
- Favor explicit error messages with actionable hints (e.g., Full Disk Access).

## Structure

- `cmd/thx`: CLI entrypoint.
- `internal/cli`: Cobra commands and flag wiring.
- `internal/db`: Database path resolution and query logic.
- `internal/things`: Data models, date encoding, URL scheme helpers.
- `internal/output`: Output formatting.

## Testing

Run tests with:

```bash
go test ./...
```

## Notes

- Default database paths are defined in `internal/db/path.go`.
- Be careful with `index` columns in SQL; quote as `"index"` when used.
