# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog,
and this project follows Semantic Versioning.

## [Unreleased]

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
