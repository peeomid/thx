# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.0] - 2025-01-01

### Added
- `thx doctor` command for diagnosing setup issues
- Long descriptions and examples for all commands
- Custom help template with better formatting
- Comprehensive README with installation, usage, and troubleshooting guides

### Changed
- Improved help output with examples and config hints

## [0.1.0] - 2025-01-01

### Added
- Initial release
- Read commands: `inbox`, `today`, `upcoming`, `anytime`, `someday`, `logbook`
- Organization commands: `projects`, `areas`, `tags`
- Search with filters: `search` (by tag, project, area, status, deadline)
- Detail view: `show` with checklist support
- Write commands: `add`, `add-project`, `done`, `cancel`, `update`, `move`
- Quick entry: `quick`, `open`
- Output formats: default, JSON, quiet
- Configuration via file (`~/.config/thx/config.yaml`) and environment variables
- Homebrew installation via `peeomid/tap`
