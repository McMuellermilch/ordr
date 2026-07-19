# Changelog

All notable changes to ordr are documented here.

## [v0.2.0] – 2026-07-19

### Added
- **Folder move**: move entire directories as a unit using `match.type: dir` and `action: move`
- **Folder flatten**: extract all files from a directory into a target using `action: flatten`
- `remove_empty` option to delete the source directory after flattening if it ends up empty

### Fixed
- Rules were listed and applied twice when the only config file was `~/.ordrrc`

---

## [v0.1.0] – 2026-07-12

Initial release.

### Added
- `ordr clean` — organize files by applying rules from `.ordrrc`
- `ordr preview` — dry-run to see what would move without touching the filesystem
- `ordr undo` — reverse the last clean session
- `ordr rules` — list, add, remove, and test rules interactively
- `ordr init` — create a starter `.ordrrc` config
- `ordr status` — show current config and rule summary
- Rule matchers: `extensions`, `pattern`, `min_size`, `max_size`, `older_than`, `newer_than`
- Conflict strategies: `rename` (default), `skip`, `overwrite`
- Local + global config merge (`~/.ordrrc` as fallback)
- Scope restrictions via `scope.dirs`
- Homebrew distribution via `mcmuellermilch/tap`
