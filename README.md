# ordr

Rule-based file organizer CLI. Define where files should go — ordr does the rest.

**[Documentation](https://mcmuellermilch.github.io/ordr/)**

---

## Overview

ordr reads rules from an `.ordrrc` config file and moves files into the right directories based on extension, name pattern, size, or age. Every operation is previewed before execution and can be reversed with `ordr undo`.

```sh
$ ordr preview ~/Desktop

  Preview — 3 file(s) would be moved

  ~/Desktop/invoice.pdf       →  ~/Desktop/documents/invoice.pdf [PDFs]
  ~/Desktop/Screenshot.png    →  ~/Pictures/Screenshots/Screenshot.png [Screenshots]
  ~/Desktop/backup.zip        →  ~/Desktop/archives/backup.zip [Archives]

$ ordr clean ~/Desktop --yes
  ✓ Moved 3 file(s). Run 'ordr undo' to reverse.
```

---

## Install

```sh
brew tap McMuellermilch/tap
brew install ordr
```

---

## Project Structure

```
ordr/
  cmd/ordr/           # Entrypoint (main.go)
  internal/
    cli/              # Cobra commands — thin glue between core and infra
    core/
      config/         # Domain types (Rule, MatchConfig, ...)
      rule/           # Rule matching engine (pure logic, no I/O)
      organizer/      # ExecutionPlan builder (pure logic, no I/O)
    infra/
      config/         # Config loader (walks up dirs) + writer
      fs/             # File system operations + undo log
  pkg/
    display/          # Terminal output formatting (lipgloss)
  docs/               # Astro + Starlight documentation website
  scripts/            # Local test scripts
```

The architecture follows a strict dependency rule:

```
cli → core, infra
infra → core
core → nothing (stdlib only)
```

This means the domain logic in `internal/core/` has zero external dependencies and is trivially testable.

---

## Development

**Requirements:** Go 1.22+

```sh
# Clone
git clone https://github.com/McMuellermilch/ordr
cd ordr

# Build
go build -o ordr ./cmd/ordr

# Run locally
./ordr --help

# Tests
go test ./...
go vet ./...

# End-to-end test scripts
./scripts/test-setup.sh        # create test dir + config
./ordr preview /tmp/ordr-test  # inspect
./ordr clean /tmp/ordr-test --yes
./ordr undo --yes
./scripts/test-teardown.sh     # clean up

# Full automated test suite
./scripts/test.sh
```

---

## Commands

| Command | Description |
|---|---|
| `ordr clean [dir]` | Move files by rules (with confirmation) |
| `ordr preview [dir]` | Dry-run — show what would be moved |
| `ordr undo` | Reverse the last clean |
| `ordr rules list` | Show all configured rules |
| `ordr rules add` | Add a rule interactively |
| `ordr rules remove <name>` | Remove a rule by name |
| `ordr rules test <file>` | Show which rules match a file |
| `ordr init` | Create a new `.ordrrc` config |
| `ordr status [dir]` | Summary of files vs. rules |

---

## Config

ordr reads `.ordrrc` (YAML), walking up from the current directory to `~/.ordrrc`.

```yaml
version: 1

rules:
  - name: "PDFs"
    target: "documents"
    match:
      extensions: [".pdf"]

  - name: "Screenshots"
    target: "~/Pictures/Screenshots"
    match:
      pattern: "^Screenshot.*"
      extensions: [".png"]
    scope:
      dirs: ["~/Desktop"]

  - name: "Old large videos"
    target: "~/Archive/Videos"
    match:
      extensions: [".mp4", ".mov"]
      min_size: "500MB"
      older_than: "90d"
```

Full config reference: [mcmuellermilch.github.io/ordr/config/overview](https://mcmuellermilch.github.io/ordr/config/overview)

---

## Release

Releases are handled by [GoReleaser](https://goreleaser.com/) via GitHub Actions. Pushing a `v*` tag builds cross-compiled binaries and updates the Homebrew formula automatically.

```sh
git tag v1.0.0
git push origin v1.0.0
```

Targets: `darwin/arm64`, `darwin/amd64`, `linux/arm64`, `linux/amd64`

---

## Docs

The documentation site lives in `docs/` and is built with [Astro Starlight](https://starlight.astro.build/). It deploys to GitHub Pages on every push to `main` that changes files under `docs/`.

```sh
cd docs
npm install
npm run dev   # http://localhost:4321/ordr/
```

---

## License

MIT
