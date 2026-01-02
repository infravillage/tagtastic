# TAGtastic

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
```

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

## Release Checklist (Appendix)
- Confirm tests pass: `go test ./...`
- Run `make codename` and assign the next Crayola color.
- Update `CHANGELOG.md` with the release entry and codename.
- Update `VERSION` if used for local builds.
- Tag the release with SemVer (`vX.Y.Z[-alpha.N]`) and push the tag.
- Ensure GitHub Actions runs GoReleaser and publishes the release.

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
