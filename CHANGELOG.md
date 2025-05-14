# Changelog

## [v0.3.0] - 2025-05-14

### Added
- Comprehensive Makefile for build, test, install, etc.
- Scripts for config creation (`scripts/create_config.sh`) and live development (`scripts/dev.sh`).
- GoReleaser configuration (`.config/goreleaser.yaml`) for automated releases.
- TUI: Realtime detail panel showing Name, Key, Tags, Usage, and Last Seen.
- TUI: List display format updated to be denser (`Name` then `# Tags`).
- TUI: Form Add/Edit/Delete now correctly displayed full screen (not centered over list).
- TUI: Automatic Mnemonic name generation if Name field is empty when adding a new connection.
- CLI: New `gsm import` command with `--secret KEY[#tags]` and `--file keys.txt` (with `KEY[#tags]` format).
- CLI: Colorful output for `import` command statuses.
- Config: Added `LastConnected` timestamp and `Usage` count to connections.

### Changed
- `README.md` significantly updated with new features, installation instructions, and more details.
- Internal wordlist for mnemonic generation is now embedded into the binary using `go:embed`.
- Improved TUI status messages.

### Fixed
- Various TUI layout and rendering issues.
- Linter errors and build issues related to `go:embed` and package structure.
