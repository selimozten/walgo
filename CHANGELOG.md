# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New `deploy-http` command to publish the built site via Walrus HTTP APIs (publisher/aggregator) on Testnet. No wallet/funds required; returns quiltId and patchIds.
- `walgo setup --force` flag to overwrite an existing `sites-config.yaml` with absolute wallet/config paths.
- Optional `--network devnet` to scaffold config for HTTP-only workflows. (On-chain deploy on devnet still requires proper Walrus client objects.)

### Changed
- Removed deprecated `--context` flag usage in site-builder calls. Network selection now comes from `~/.config/walrus/sites-config.yaml`.
- Improved `init` guidance to show HTTP vs on-chain deployment options.

### Fixed
- `convert` command now parses Base36 cleanly even when `site-builder` prints log lines; outputs proper Base36 IDs and URLs.

### Known limitations
- On-chain `deploy`/`update` still require a correctly configured Walrus client and a funded Sui wallet on the selected network. Testnet HTTP publishing can be used as a no-funds alternative.

## [0.1.0] - 2024-12-19

### Added
- Initial public release
- Complete Hugo integration with `init`, `build`, `serve`, and `new` commands
- Full Walrus Sites deployment support with `deploy` and `update` commands
- Obsidian vault import functionality with wikilink conversion
- SuiNS domain configuration guidance via `domain` command
- Site status monitoring and object ID conversion utilities
- Asset optimization engine for HTML, CSS, and JavaScript
- Comprehensive CLI with help system and configuration management
- YAML-based configuration with sensible defaults
- Multi-format frontmatter support (YAML, TOML, JSON)
- Cross-platform support (macOS, Linux, Windows)

### Features
- 🚀 Hugo site initialization with Walrus-specific configuration
- 📦 One-command deployment to decentralized storage
- 🔄 Efficient site updates without full redeployment
- 📝 Content creation with `walgo new` command
- 🌐 SuiNS domain management guidance
- 📊 Site resource monitoring and status checking
- 🔧 Obsidian to Hugo migration toolkit
- ⚡ Built-in asset optimization and minification
- 🎯 Flexible configuration system
- 📋 Comprehensive documentation and examples

[Unreleased]: https://github.com/selimozten/walgo/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/selimozten/walgo/releases/tag/v0.1.0 