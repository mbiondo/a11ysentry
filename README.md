# 🛡️ A11ySentry

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
