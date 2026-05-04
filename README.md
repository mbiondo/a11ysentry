# 🛡️ A11ySentry

[![Release](https://img.shields.io/github/v/release/mbiondo/a11ysentry?color=00ADD8&label=version)](https://github.com/mbiondo/a11ysentry/releases)
[![Build Status](https://github.com/mbiondo/a11ysentry/actions/workflows/release.yml/badge.svg)](https://github.com/mbiondo/a11ysentry/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mbiondo/a11ysentry)](https://goreportcard.com/report/github.com/mbiondo/a11ysentry)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mbiondo/a11ysentry)](https://golang.org/doc/devel/release.html)
[![Code Size](https://img.shields.io/github/languages/code-size/mbiondo/a11ysentry)](https://github.com/mbiondo/a11ysentry)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A11ySentry** is a professional-grade, multi-platform accessibility auditing engine. It doesn't just look at files; it understands **Component Architectures**.

Built for modern frontend engineering, it reconstructs the rendering hierarchy (PageTrees) of your application to provide precise, context-aware accessibility analysis with zero false positives.

---

## 🚀 Key Features

- **Architectural Awareness**: Reconstructs full component trees for Next.js, Nuxt, SvelteKit, Astro, and mobile platforms.
- **Formal CSS Parsing**: Integrated `tdewolff` parser for perfect color and variable resolution (including CSS Vars & Dark Mode).
- **W3C ACT Standardized**: Every rule is mapped to official W3C Accessibility Conformance Testing IDs (e.g., `23a2a8`).
- **A11y-Ignore System**: Suppress specific violations directly in code via `<!-- a11y-ignore: RULE_ID -->`.
- **Session-Based History**: Every audit is a "Snapshot," allowing you to track progress without mixing stale data.
- **Universal Semantic Node (USN)**: A unified data model that allows auditing Web, Mobile (Android/iOS/Flutter/RN), and Desktop (DotNet/Electron/PyQt) apps with the same rules.

---

## 🛠️ Usage

### Installation
```bash
go build -o a11ysentry ./cmd/a11ysentry
```

### Analyze a Project
```bash
./a11ysentry --dir ./my-project
```

### Interactive Dashboard (TUI)
Explore your component hierarchy and violations in a beautiful terminal interface.
```bash
./a11ysentry --tui
```
> **Pro Tip**: Press `t` inside a project to toggle the **Component Explorer**.

### Clear History
```bash
./a11ysentry clear
```

---

## 📱 Platform Support

| Category | Frameworks / Tech |
| :--- | :--- |
| **Web** | React, Vue, Svelte, Angular, Astro, Next.js, Nuxt, SvelteKit |
| **Mobile** | Android (Compose/XML), iOS (SwiftUI), Flutter, React Native |
| **Desktop** | Electron, .NET (XAML/MAUI), PyQt |
| **Backend** | Django Templates, Flask (Jinja2) |

---

## 🧩 IDE Integration (Alpha)
A11ySentry can run as a high-performance **Linter Engine** for VS Code, Cursor, and other IDEs.
```bash
./a11ysentry --linter
```
It accepts file paths on `stdin` and outputs real-time, PageTree-aware JSON diagnostics.

---

## 📜 Compliance
All rules follow the **WCAG 2.2** guidelines and are verified against **ACT Rules** to ensure professional compliance standards.

---
**Made with ❤️ for Accessible Engineering.**
