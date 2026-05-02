# 🛡️ A11ySentry
### Universal Accessibility Engine for Multi-Platform UI

A11ySentry is a deterministic, multi-platform accessibility validator. It provides a single semantic source of truth for all major UI frameworks.

[![CI/CD Pipeline](https://github.com/mbiondo/a11ysentry/actions/workflows/go.yml/badge.svg)](https://github.com/mbiondo/a11ysentry/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🚀 The Vision
Most accessibility tools are platform-locked. A11ySentry introduces the **Universal Semantic Node (USN)**: an abstraction layer that allows the same WCAG 2.2 rules to validate over **15 platforms** with 100% deterministic certainty.

## 🛠️ Key Features
- **Massive Platform Support**:
  - **Web**: React, Vue, Svelte, Angular, Astro.
  - **Mobile**: Android (Compose/View), iOS (SwiftUI), React Native, Flutter.
  - **Desktop**: .NET (MAUI/WPF), Java (FX/Swing), Electron, Tauri.
  - **Gaming**: Unity, Godot.
- **Recursive Component Context**: Analyzes entire component trees (Page + Components) to avoid false positives by cross-referencing labels and states across files.
- **Tailwind 4 Support**: Built-in heuristic mapping for Tailwind CSS 4 utility classes (colors, spacing, etc.) for zero-runtime static analysis.
- **WCAG 2.2 Ready**: Implements the latest accessibility standards, including the new Web Touch Target (24px) and Aria State requirements.
- **Unified CLI**: Analyze files, view history in TUI, or start MCP from a single binary.
- **Interactive TUI Dashboard**: Visual dashboard for audit history and violation tracking.
- **Universal MCP Support**: Register A11ySentry in **Claude, Cursor, VS Code, Gemini CLI, Qwen-Coder, and OpenCode** with one command.
- **Deterministic Engine**: Pure engineering-based validation (no LLM hallucinations).
- **High Precision**: Reports exact file path, line, and column for every violation.

---

## 📦 Installation

### Quick Install (Automatic)
Run the smart installer to download the binary, add it to your PATH, and register MCP in all your AI agents:

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.ps1 | iex
```

**Unix (Bash):**
```bash
curl -sSL https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.sh | bash
```

---

## 🖥️ Usage

### CLI & TUI
```bash
# Initialize a project (scaffolds CI/CD, hooks, config)
a11ysentry init

# Analyze a file directly
a11ysentry path/to/file.html

# Analyze full project (resolves component trees)
a11ysentry .

# Output formats
a11ysentry --format sarif . > results.sarif   # GitHub Code Scanning
a11ysentry --format json .                    # Machine-readable JSON
a11ysentry --format text .                    # Human-readable (default)

# Pre-load CSS for accurate color contrast
a11ysentry --css global.css,branding.css Component.vue

# Watch mode — re-analyze on file changes
a11ysentry --watch Component.tsx

# Open the interactive TUI dashboard
a11ysentry --tui
```

### AI Agents & MCP
A11ySentry integrates with your favorite AI agents via the **Model Context Protocol**.

```bash
# Register in all detected agents (Claude, Cursor, Gemini, etc.)
a11ysentry mcp --register

# Verify registration status
a11ysentry mcp --check-mcp

# Start the MCP server manually (Stdio)
a11ysentry mcp
```

---

## 📂 Project Structure

```
semantix/
├── engine/              # Core WCAG rules and SQLite persistence
├── adapters/            # Platform-specific ingestion (15+ platforms)
├── cli/                 # Unified entry point and TUI (Bubbletea)
├── mcp/                 # MCP registration and server logic
├── examples/            # Multi-platform components for testing
├── docs/                # Comprehensive documentation
└── openspec/            # Spec-driven development specs
```

## 📚 Documentation

- **[API Reference](./docs/API_REFERENCE.md)** - Core types, interfaces, and CLI commands
- **[Architecture Deep Dive](./docs/ARCHITECTURE_DEEP_DIVE.md)** - Pipeline stages, data flow, design decisions
- **[Developer Guide](./docs/DEVELOPER_GUIDE.md)** - How to create adapters, add rules, and test
- **[MCP Integration Guide](./docs/MCP_INTEGRATION.md)** - AI agent setup and tool usage
- **[Examples Documentation](./examples/README.md)** - Platform-specific examples and patterns
- **[Release Guide](./docs/RELEASE.md)** - Build, deploy, and maintenance procedures

## 🤝 Contributing

We love contributors! See our contribution resources:

- **[Contributing Guidelines](./CONTRIBUTING.md)** - Commits, branches, and PR process
- **[Code of Conduct](./CODE_OF_CONDUCT.md)** - Community standards
- **[Developer Guide](./docs/DEVELOPER_GUIDE.md)** - Technical implementation guide
- **[Examples](./examples/README.md)** - Add platform examples

### Quick Start for Contributors

```bash
# Clone the repository
git clone https://github.com/mbiondo/a11ysentry.git
cd semantix

# Build CLI
cd cli && go build -o a11ysentry

# Run tests
go test ./...

# Analyze an example file
./a11ysentry ../examples/example-astro/src/pages/index.astro
```

## 📄 License & Security

- **License:** [MIT License](./LICENSE)
- **Security Policy:** [SECURITY.md](./SECURITY.md)
- **Changelog:** [CHANGELOG.md](./CHANGELOG.md)

## 🚀 Quick Links

| I want to... | Go to... |
|--------------|----------|
| Install A11ySentry | [Installation Section](#-installation) |
| Understand the architecture | [Architecture Deep Dive](./docs/ARCHITECTURE_DEEP_DIVE.md) |
| Add a new platform adapter | [Developer Guide - Creating Adapters](./docs/DEVELOPER_GUIDE.md#creating-a-new-adapter) |
| Integrate with AI agents | [MCP Integration Guide](./docs/MCP_INTEGRATION.md) |
| See examples | [Examples Documentation](./examples/README.md) |
| Report a bug | [GitHub Issues](https://github.com/mbiondo/a11ysentry/issues) |
| Release a new version | [Release Guide](./docs/RELEASE.md) |
