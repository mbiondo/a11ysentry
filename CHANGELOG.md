# Changelog

All notable changes to A11ySentry will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.4] - 2026-05-02

### Fixed
- **CI**: Fixed Homebrew Tap publishing by using correct GitHub tokens.

## [0.0.3] - 2026-05-02

### Added
- **Landing**: Added Windows installation command and improved terminal mockup realism.
- **Landing**: Improved mobile layout and accessibility landmarks.

### Changed
- **Config**: Updated default configuration for better out-of-the-box experience.
- **Cleanup**: Removed redundant analysis reports from the repository.

### Fixed
- **Landing**: Solved structural accessibility issues identified by MCP analysis.

## [0.0.2] - 2026-05-02

### Added
- **Rules**: Advanced WCAG rules for landmarks, modals, and fieldsets.
- **Rules**: Implementation of WCAG 2.1.1 keyboard navigation rules.
- **TUI**: Professional responsive TUI refactor with project-aware navigation.
- **TUI**: New CLI progress bar with file logging for better feedback during analysis.
- **Engine**: Hierarchical FileNode analysis for better context inheritance in adapters.
- **Distribution**: Global install scripts (`install.ps1`, `install.sh`) for easier setup.
- **MCP**: Auto-resolving MCP registration for seamless integration with AI agents.

### Changed
- **CLI**: Simplified usage by removing the redundant `--dir` flag (project root is now inferred).

### Fixed
- **CI**: Robust project detection and improved SARIF generation.
- **CI**: Resolved race conditions in the a11y workflow.
- **Linter**: Addressed various lint issues across the codebase.

## [0.0.1] - 2026-04-30

### Added
- **Core Engine**: Deterministic validation pipeline based on the Universal Semantic Node (USN) abstraction.
- **Platform Support**: Ingestion adapters for over 15 platforms including Web (React, Vue, Astro), Mobile (Android, iOS, Flutter), Desktop, and Gaming (Unity, Godot).
- **Scanner**: Automatic project discovery for monorepos and multi-framework setups.
- **Tailwind CSS 4**: Heuristic color resolution and dark mode support for Tailwind-based projects.
- **MCP Server**: Full Model Context Protocol support for integration with AI agents (Claude, Cursor, Gemini, etc.).
- **TUI Dashboard**: Interactive terminal UI for auditing history and managing violations.
- **SARIF Output**: Native support for Static Analysis Results Interchange Format for GitHub Code Scanning.
- **Project Scaffolding**: `a11ysentry init` command to inject GitHub Actions, pre-commit hooks, and configuration files.
- **Watch Mode**: Real-time analysis of file changes.
- **Documentation**: Comprehensive guides for developers, architects, and end-users.
