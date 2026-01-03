![TAGtastic banner](assets/what-is-tagtastic.png)

# TAGtastic 🏷️

[![Release](https://img.shields.io/github/actions/workflow/status/infravillage/tagtastic/release.yml?label=release&logo=githubactions&logoColor=white)](https://github.com/infravillage/tagtastic/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/go-1.25.5-00ADD8?logo=go&logoColor=white)](https://go.dev/doc/devel/release)
[![Go Report Card](https://goreportcard.com/badge/github.com/infravillage/tagtastic)](https://goreportcard.com/report/github.com/infravillage/tagtastic)
[![License](https://img.shields.io/github/license/infravillage/tagtastic?color=2C3E50)](LICENSE)

TAGtastic is a lightweight CLI for generating human-readable release codenames.

It is designed for release automation, CI/CD pipelines, and teams that want
consistent naming with a clear audit trail. The project follows Semantic
Versioning, Keep a Changelog, and GoReleaser.

TAGtastic focuses on *naming*, not release orchestration.

---

## What TAGtastic Does

TAGtastic generates deterministic, human-readable codenames that can be
associated with SemVer tags.

Example use cases:
- assigning a memorable codename to a release
- improving release communication
- providing a stable, human-friendly identifier alongside version numbers

---

## Example Usage

```bash
$ tagtastic generate --theme crayola_colors --seed 1
Almond

$ tagtastic generate --theme birds --exclude albatross
Blue Heron

$ tagtastic list --theme birds
Albatross
Blue Heron
Crane
Dove
Eagle

$ tagtastic themes
birds
cities
crayola_colors
landmarks

$ tagtastic validate "Almond" --theme crayola_colors
Found in theme 'crayola_colors'

$ tagtastic generate --theme birds --seed 2 --format shell
RELEASE_CODENAME=blue-heron
```

Note: shell/CI output uses the first alias for a codename (slug style), so
multi-word names become dash-separated when using the `shell` format.

CI/automation options:
- `--quiet` suppresses non-essential output.
- `--json-errors` emits machine-readable error output.
- `config init/reset --dry-run` previews changes without writing.

---

## Similar Tools

TAGtastic complements existing release tooling such as GoReleaser.

It does not replace release automation or versioning systems. Instead, it
fits into the same workflow by providing a stable, human-readable codename
for each SemVer tag.

---

## Project Status

* Current focus: stabilizing the v1 specification
* Release phases: alpha → beta → stable

Scope expansion is intentional and conservative.

---

## Repository Layout

* `cmd/tagtastic/` — CLI entrypoint
* `internal/` — application logic
* `data/` — local datasets used by the CLI

---

## Code Quality (Go Report Card)
Go Report Card runs on the hosted service, but you can validate locally using the CLI:

```bash
go install github.com/gojp/goreportcard/cmd/goreportcard-cli@latest
goreportcard-cli
```

The badge above links to the hosted report for this repo.

Local quality checks (modern tooling):
```bash
make quality
```

---

## Release Naming (Crayola Colors)
Each release uses a codename from the Crayola color list in the Corpora repo:
https://github.com/dariusk/corpora/blob/master/data/colors/crayola.json

Rules:
- Names are assigned in alphabetical order for each release.
- The codename is recorded in `CHANGELOG.md` and the GitHub Release title.
- SemVer tags remain the source of truth (e.g., `v1.0.0-beta.1`).

Data and tooling:
- `data/crayola.json` is the raw source snapshot.
- `data/themes.yaml` contains the structured theme data used by the CLI.
- `internal/data/themes.yaml` is the embedded copy (sync with `go run ./cmd/tools/sync-themes`).
- `go run ./cmd/tools/next-codename` prints the next available codename based on `CHANGELOG.md`.

---

## GoReleaser + Codename Flow
GoReleaser does not generate codenames by itself. TAGtastic provides the codename and passes it into GoReleaser via `RELEASE_CODENAME`.

Recommended flow:
1) Quick path (release helper, optional and CI-friendly):
   - `go run ./cmd/tools/release 0.1.0-alpha.2 --commit`
   - This auto-selects the next codename, updates `CHANGELOG.md`/`VERSION`, and creates the annotated tag.
2) Manual path (TAGtastic + git):
   - `CODENAME=$(make codename -s)`
   - Update `CHANGELOG.md` and `VERSION`
   - `git tag -a vX.Y.Z[-alpha.N] -m "vX.Y.Z – ${CODENAME}"`
3) Push the tag:
   - `git push origin vX.Y.Z[-alpha.N]`
4) GitHub Actions runs GoReleaser and publishes the release.

## Automating CHANGELOG.md and VERSION
GoReleaser does not auto-edit `CHANGELOG.md` or `VERSION`. You can automate this in CI or locally by adding a small script (Go or shell) that:
- Reads the next codename (`make codename`)
- Updates the latest `CHANGELOG.md` section
- Writes the new `VERSION`

TAGtastic includes a dedicated release helper at `cmd/tools/release` that performs these steps and creates an annotated git tag.

Example (local, auto-picks codename):
```bash
go run ./cmd/tools/release 0.1.0-alpha.2
```

Example (with commit):
```bash
go run ./cmd/tools/release 0.1.0-alpha.2 --commit
```

Example (dry run):
```bash
go run ./cmd/tools/release 0.1.0-alpha.2 --dry-run
```

CI/automation options:
- `--quiet` suppresses non-essential output.
- `--json-errors` emits machine-readable error output.

Makefile shortcut:
```bash
make release-prep VERSION=0.1.0-alpha.2
```

---

## Release Checklist (Appendix)
- Confirm tests pass: `go test ./...`
- Run `make codename` to select the next Crayola color via TAGtastic.
- Run the release helper to update `CHANGELOG.md`, `VERSION`, and create the tag.
- Tag the release with SemVer (`vX.Y.Z[-alpha.N]`) and the TAGtastic codename.
- Ensure GitHub Actions runs GoReleaser and publishes the release.
- Use GoReleaser for all release binaries (no manual packaging).

---

## Scope Note

TAGtastic is a standalone utility.

It is not part of any core product, engine, or proprietary system and is
developed independently as a general-purpose tool.

---

## Changelog
`CHANGELOG.md` follows Keep a Changelog and Semantic Versioning.

---

## License
MIT. See `LICENSE`.

---

## Credits
- CLI parsing: Kong (https://github.com/alecthomas/kong)
- Release automation: GoReleaser (https://goreleaser.com/)
- Codename themes data: Corpora by Darius Kazemi (https://github.com/dariusk/corpora)
