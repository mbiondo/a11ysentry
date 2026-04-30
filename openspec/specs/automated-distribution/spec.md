# automated-distribution Specification

## Purpose
Automate the build, release, and installation of A11ySentry across multi-platform environments.

## Requirements

### Requirement: Cross-Platform Builds
The system MUST be capable of producing optimized binaries for Linux, macOS, and Windows (amd64/arm64) using a single command.

#### Scenario: Build Release
- GIVEN the release pipeline is triggered
- WHEN GoReleaser runs
- THEN it MUST generate compressed archives for all supported OS/Architecture combinations.

### Requirement: Automated Installer
The system MUST provide an installation script that simplifies binary download and environment setup.

#### Scenario: Shell Installation
- GIVEN a clean Linux/macOS environment
- WHEN the `install.sh` script is executed
- THEN it MUST download the latest release and symlink it to the user's PATH.

### Requirement: Zero-Config MCP Registration
The installer MUST automatically register the A11ySentry MCP server with installed AI agents.

#### Scenario: Register with Claude
- GIVEN the `install.sh` script detects a Claude Desktop installation
- WHEN the installation script runs
- THEN it MUST append the A11ySentry configuration to `claude_desktop_config.json` if not already present.
