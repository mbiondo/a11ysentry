# 🛡️ A11ySentry
### Universal Accessibility Engine for Multi-Platform UI

A11ySentry is a deterministic, multi-platform accessibility validator. It provides a single semantic source of truth for all major UI frameworks using the **Universal Semantic Node (USN)** abstraction.

[![CI/CD Pipeline](https://github.com/mbiondo/a11ysentry/actions/workflows/go.yml/badge.svg)](https://github.com/mbiondo/a11ysentry/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🚀 The Vision
Most accessibility tools are platform-locked and rely on brittle heuristics or slow LLMs. A11ySentry introduces a **pure engineering approach**: a high-performance engine that translates any UI (Web, Mobile, Desktop, Gaming) into a common semantic language (USN) and applies a unified set of **deterministic WCAG 2.2 rules**.

## 🛠️ Key Features

### 🌍 Massive Platform Support
- **Web**: React, Vue, Svelte, Angular, Astro (with automatic layout-chain resolution).
- **Python**: **Django** and **Flask/Jinja2** (with full template inheritance support).
- **Mobile**: Android (Compose/XML), iOS (SwiftUI/UIKit), **React Native**, **Flutter**, and **MAUI**.
- **Desktop**: .NET (MAUI/WPF), **PyQt/PySide** (native .ui parsing), Java (FX/Swing), Electron, Tauri.
- **Gaming**: Unity, Godot.

### 🧠 The Deterministic Edge
Unlike LLM-based tools that guess, A11ySentry uses a **pure engineering approach**. It transforms every UI into a **Universal Semantic Node (USN)** tree, allowing for 100% deterministic validation with zero hallucinations.

### 🔍 Intelligent Analysis
- **Recursive Context Rule (CRITICAL)**: Accessibility states like `aria-hidden`, `disabled`, or platform-specific "exclude from semantics" traits are automatically propagated down the tree. This allows us to catch focus traps and "ghost" interactive elements that other scanners miss.
- **Structural Link Purpose (WCAG 2.4.4)**: A language-agnostic rule that detects ambiguous links (same label, different destination) by analyzing the entire USN tree for naming collisions.
- **Tailwind 4 & CSS Support**: Built-in resolution for Tailwind CSS 4 utility classes and external CSS variables. Resolves color contrast statically without a browser.

### ⚡ Developer Experience
- **Interactive TUI 2.0**: A beautiful terminal dashboard with platform badges, real-time project statistics, and detailed violation summaries.
- **Universal MCP Support**: One-command integration with AI agents (**Claude, Cursor, VS Code, Gemini, Qwen, OpenCode**) via the Model Context Protocol.
- **CI/CD Ready**: Built-in `init` command to scaffold GitHub Actions, pre-commit hooks, and configurations.
- **Zero Hallucinations**: 100% deterministic logic. If A11ySentry flags it, it's a real violation.

---

## 📦 Installation

### Quick Install (Automatic)
Run the smart installer to download the latest binary and add it to your PATH:

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

### CLI Commands
```bash
# Initialize CI/CD and config
a11ysentry init

# Analyze a single file
a11ysentry MyComponent.tsx

# Analyze a full project directory (Framework-aware)
a11ysentry --dir ./src

# Watch mode for real-time feedback
a11ysentry --watch .

# Open the TUI Dashboard
a11ysentry --tui
```

### AI Agent Integration (MCP)
```bash
# Register A11ySentry in all your AI tools at once
a11ysentry mcp --register
```

---

## 📂 Project Structure

- `engine/`: Core WCAG rules, USN schema, and SQLite history persistence.
- `adapters/`: Platform-specific parsers (Web, Android, iOS, Flutter, RN, etc.).
- `scanner/`: Framework detection and dependency graph resolution.
- `cmd/a11ysentry/`: Main entry point and TUI (Charmbracelet Bubbletea).
- `mcp/`: Model Context Protocol server and registration logic.
- `examples/`: Complex multi-platform test cases.

---

## 🤝 Contributing
We are building the future of inclusive software. Check our **[Developer Guide](./docs/DEVELOPER_GUIDE.md)** to learn how to add new adapters or rules.

- **[Architecture Deep Dive](./docs/ARCHITECTURE_DEEP_DIVE.md)**
- **[API Reference](./docs/API_REFERENCE.md)**
- **[MCP Integration Guide](./docs/MCP_INTEGRATION.md)**

## 📄 License
Licensed under the [MIT License](./LICENSE).
