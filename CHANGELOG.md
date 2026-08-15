# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog,
and this project follows Semantic Versioning.

## [Unreleased]

## [0.4.0] - 2026-08-15
### Added
- Added additional cleanup command shorthands:
	- `-mb` and `-mbr` for `merged-branches`
	- `-rb` and `-rbr` for `remove-branches`
	- `-wk` and `-wkt` for `worktrees`
	- `-mwk` and `-mwkt` for `merged-worktrees`
	- `-rwk` and `-rwkt` for `remove-worktrees`
- Added test coverage for shorthand normalization across cleanup commands.
- Expanded README installation instructions for both source installs and prebuilt GitHub Release binaries.

### Changed
- Made shorthand parsing backward-compatible so previous alias forms continue to work.
- Improved release workflow behavior to skip cleanly when no tag is resolved instead of failing the pipeline.

## [0.3.0] - 2026-08-15
### Added
- Migrated CLI command routing to Cobra.
- Added cleanup aliases for worktree listing: worktree, wk, and -wk.
- Added automatic tagging workflow on pushes to main with UTC timestamp plus short commit hash.
- Added release workflow to build and publish Linux and Windows CLI artifacts to GitHub Releases.
- Added Linux arm64 release builds.
- Added packaged release archives per platform.
- Added release checksums and a Homebrew snippet artifact for Linux installs.

### Changed
- Updated CLI documentation to include shorthand cleanup command usage.
