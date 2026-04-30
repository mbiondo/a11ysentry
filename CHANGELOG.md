# Changelog

All notable changes to A11ySentry will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
