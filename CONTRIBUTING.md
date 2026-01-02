# Contributing to TAGtastic

Thanks for helping build TAGtastic in public. Please keep changes small, documented, and tested. This is a CLI-first project, so prioritize correctness, clear UX, and reproducible releases.

## Development Standards
- Go code is formatted with `gofmt -s`.
- Add doc comments for exported types and functions.
- Prefer clear, descriptive names; keep helpers unexported unless needed.

## Tests
- Use Go's `testing` package.
- Name tests `TestXxx` and place in `*_test.go` files.
- Run `go test -v -race ./...` before opening a PR.

## Documentation
- Update `README.md` when behavior changes or new commands are added.
- Keep `CHANGELOG.md` accurate and append release entries in Keep a Changelog format.
- When adding themes, update `data/themes.yaml` and sync the embedded copy.

## Changelog + Release Codenames
- Follow Keep a Changelog: https://keepachangelog.com/en/1.0.0/
- Use SemVer for tags and versions.
- Each release requires a Crayola color codename from the Corpora list:
  https://github.com/dariusk/corpora/blob/master/data/colors/crayola.json
- Codenames are assigned in alphabetical order, recorded in `CHANGELOG.md`, and used in the GitHub Release title.
- Use `go run ./cmd/tools/next-codename` to select the next available codename before tagging.
- If you edit `data/themes.yaml`, run `go run ./cmd/tools/sync-themes` to update the embedded copy.

## Commit Messages
Use Conventional Commits. Examples:
- `feat(cli): add list command`
- `fix(data): handle empty themes`
- `docs: update README`

## Pull Requests
- Summarize the change and link any related issues.
- Include tests or explain why tests are not needed.
- Update docs and changelog entries when behavior changes.
- For user-facing behavior, include example CLI output in the PR description.
