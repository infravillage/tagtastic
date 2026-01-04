# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [Unreleased]

### Added
- N/A

### Changed
- N/A

### Fixed
- N/A

## [0.2.0-beta.1] – "Asparagus" – 2026-01-04

### Added

- GitHub Actions CI workflow for automated quality checks on PRs and commits
- Security scanning with gosec (runs on every PR and commit)
- `make security` target for local security scanning
- CI workflow badge in README.md
- Security considerations documentation in README.md

### Changed

- Updated Go version requirement to 1.24.0 (required by dependencies)
- Updated README.md with CI/CD integration examples and security scanning details

## [0.1.1-beta.1] – "Aquamarine" – 2026-01-04

### Added
- Placeholder version entry for this release.

## [0.1.0-beta.2] – "Apricot" – 2026-01-04

### Added
- N/A

### Changed
- N/A

### Fixed
- Release helper now handles flags after the version argument correctly in dry-run mode.

### Deprecated
- N/A

### Removed
- N/A

### Security
- N/A

## [0.1.0-beta.1] – "Almond" – 2026-01-03

### Added
- Initial CLI scaffold and project documentation.
- Release helper tool to prepare releases, update files, and tag versions.
- Repo-local config support with precedence (`--config-path`, `TAGTASTIC_CONFIG`, `./.tagtastic.yaml`).
- `generate --record` to write selected codenames into repo config.
- CI-friendly flags: `--quiet`, `--json-errors`, and `--dry-run` (where applicable).
- Banner asset and repo badges (Go Report Card, release status, license, Go version).
- Quality checks via `make quality` and Go Report Card guidance.
- Expanded documentation with real-world release workflows and CI usage examples.

### Changed
- Shell output now uses aliases (slug style) with a safe fallback.
- JSON output is pretty-printed for readability.
- Banner and help behavior tuned for interactive vs. CI usage.
- Codename lookup prefers git tags, then config, then changelog.

### Fixed
- Changelog reference links corrected for beta and unreleased entries.

## [0.1.0-alpha.1] – "Antique Brass" – 2026-01-01

### Added
- Placeholder version entry for the first alpha release.





[Unreleased]: https://github.com/infravillage/tagtastic/compare/v0.2.0-beta.1...HEAD
[0.2.0-beta.1]: https://github.com/infravillage/tagtastic/compare/vUnreleased...v0.2.0-beta.1
[Unreleased]: https://github.com/infravillage/tagtastic/compare/v0.1.1-beta.1...HEAD
[0.1.1-beta.1]: https://github.com/infravillage/tagtastic/compare/v0.1.0-beta.2...v0.1.1-beta.1
[0.1.0-beta.2]: https://github.com/infravillage/tagtastic/compare/v0.1.0-beta.1...v0.1.0-beta.2
[0.1.0-beta.1]: https://github.com/infravillage/tagtastic/compare/v0.1.0-alpha.1...v0.1.0-beta.1
[0.1.0-alpha.1]: https://github.com/infravillage/tagtastic/releases/tag/v0.1.0-alpha.1