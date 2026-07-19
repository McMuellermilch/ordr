# Contributing to ordr

Thank you for your interest in contributing!

## Getting started

```bash
git clone https://github.com/McMuellermilch/ordr.git
cd ordr
go build ./...
go test ./...
```

## Making changes

- Keep the clean architecture in mind: `core` has zero external dependencies, `infra` implements I/O, `cli` is thin glue.
- Run `go vet ./...` before submitting.
- Use conventional commit prefixes: `feat:`, `fix:`, `docs:`, `chore:`.

## Submitting a pull request

1. Fork the repo and create a branch from `main`.
2. Make your changes and ensure `go build ./...` and `go test ./...` pass.
3. Open a pull request with a clear description of what and why.

## Reporting bugs

Use the [bug report issue template](https://github.com/McMuellermilch/ordr/issues/new?template=bug_report.md).

## Feature requests

Use the [feature request issue template](https://github.com/McMuellermilch/ordr/issues/new?template=feature_request.md).
