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
- `internal/config`: Configuration loading (file, env, defaults).

## Testing

Run tests with:

```bash
go test ./...
```

## Building

```bash
# Build binary
go build -o thx ./cmd/thx

# Install to $GOPATH/bin
go install ./cmd/thx
```

## Releasing

This project uses [Semantic Versioning](https://semver.org/):
- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** (v0.1.0 → v0.2.0): New features, backward compatible
- **PATCH** (v0.1.0 → v0.1.1): Bug fixes, backward compatible

### Release Checklist

When preparing a release:

1. **Update CHANGELOG.md** — Add a new section for the version with:
   - `### Added` — New features
   - `### Changed` — Changes to existing functionality
   - `### Fixed` — Bug fixes
   - `### Removed` — Removed features

2. **Commit the changelog**:
   ```bash
   git add CHANGELOG.md
   git commit -m "Prepare release vX.Y.Z"
   ```

3. **Create and push the tag**:
   ```bash
   git tag vX.Y.Z
   git push origin main
   git push origin vX.Y.Z
   ```

4. **Run GoReleaser** (or let CI do it):
   ```bash
   GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
   ```

### Example: Minor Version Release

For adding new features (e.g., v0.1.0 → v0.2.0):

```bash
# 1. Update CHANGELOG.md with new features
# 2. Commit
git add CHANGELOG.md
git commit -m "Prepare release v0.2.0"
git push origin main

# 3. Tag and release
git tag v0.2.0
git push origin v0.2.0
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```

## Notes

- Default database paths are defined in `internal/db/path.go`.
- Be careful with `index` columns in SQL; quote as `"index"` when used.
- GoReleaser config is in `.goreleaser.yaml`.
- Homebrew formula is auto-updated in `peeomid/homebrew-tap`.
