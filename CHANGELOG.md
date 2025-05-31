# Changelog

## [v0.3.2] - 2025-01-22

### Fixed
- Fixed critical delete bug in TUI where deleted connections were not saved to config file, causing items to reappear after restart.
- Added proper `config.Save()` call after successful deletion and improved error handling for save operations.

## [v0.3.1] - 2025-05-14 

### Changed
- Enhanced CLI file import (`gsm import --file`):
  - Now correctly parses optional tags appended to keys (e.g., `SECRETKEY#tag1,tag2`).
  - Skips lines not starting with an alphanumeric (base62-like) character (e.g., comments starting with `#` or `!`).
  - For lines starting with alphanumeric characters, it now correctly takes only the characters up to the first space or tab as the `KEY[#tags]` part, ignoring subsequent text on the same line as comments.
  - Improved feedback messages for skipped lines or errors during file parsing.
- Updated `README.md` with more detailed installation options, including `Makefile` usage and improved `go install` command.
- Refined `scripts/get.sh` for more robust installation, particularly in determining writable installation paths and verifying the installed binary using `gsm version` subcommand.

### Fixed
- Corrected `go install` command in `README.md` to point to the correct package path (`cmd/gsm`).
- Addressed linter warnings and minor bugs in build scripts and TUI rendering logic that arose during feature development.

## [v0.3.0] - 2025-05-14

### Added
- Comprehensive Makefile for build, test, install, etc.
- Scripts for config creation (`scripts/create_config.sh`) and live development (`scripts/dev.sh`).
- GoReleaser configuration (`.config/goreleaser.yaml`) for automated releases.
- TUI: Realtime detail panel showing Name, Key, Tags, Usage, and Last Seen.
- TUI: List display format updated to be denser (`Name` then `# Tags`).
- TUI: Form Add/Edit/Delete now correctly displayed (not centered over list when active).
- TUI: Automatic Mnemonic name generation if Name field is empty when adding a new connection.
- CLI: New `gsm import` command with `--secret KEY[#tags]` and `--file keys.txt` (with basic `KEY[#tags]` format).
- CLI: Colorful output for `import` command statuses.
- Config: Added `LastConnected` timestamp and `Usage` count to connections.

### Changed
- `README.md` significantly updated with new features, installation instructions, and more details.
- Internal wordlist for mnemonic generation is now embedded into the binary using `go:embed`.
- Improved TUI status messages and form placeholder text.

### Fixed
- Various TUI layout and rendering issues.
- Linter errors and build issues related to `go:embed` and package structure.
- Corrected `gsm version` verification in `scripts/get.sh`.
