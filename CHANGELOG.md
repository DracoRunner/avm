# Changelog

All notable changes to `avm` are recorded here.

The format follows Keep a Changelog style, and releases use semantic versioning.

## [Unreleased]

### Added
- Rust CLI workspace with `avm-cli`, `avm-core`, `avm-shims`, `avm-runtime`, `avm-plugin-api`, and `avm-plugin-node`.
- Docker-based test harness and Rust integration test coverage.
- User-facing LLM and agent onboarding docs.

### Changed
- Package scope moved to `@prajanova/avm`.
- Project name updated to Any Version Manager.

## [0.2.6] - 2026-05-13

### Added
- Baseline Rust rewrite structure.
- Node provider direction for package script discovery and Node version resolution.
- Shim model for plain command interception.

### Changed
- Replaced the legacy project layout with Rust workspace boundaries.
- Updated npm package ownership and repository links to Prajanova.
