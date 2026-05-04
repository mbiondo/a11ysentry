# A11ySentry Architecture Deep Dive

A11ySentry is designed as a **Universal Accessibility Engine**. Unlike traditional linters that analyze isolated files, A11ySentry reconstructs the **Rendering Hierarchy** to understand how components interact.

---

## 1. The Hub-and-Spoke Model
A11ySentry uses a central hub connected to multiple platform-specific adapters:
- **Scanners**: Detect project types (Next.js, Flutter, etc.) and build the **PageTree** (import graph).
- **Adapters**: Ingest raw source code and normalize it into **Universal Semantic Nodes (USN)**.
- **Engine**: Applies modular rules to the USN tree.

## 2. Universal Semantic Node (USN)
The USN is our core abstraction. Whether it's a `<button>` in React or an `ElevatedButton` in Flutter, they are both mapped to a `RoleButton` USN. This allows a single set of rules to audit any platform.

## 3. PageTree Awareness
The most critical feature of A11ySentry is its awareness of the component tree.
- **Context Inheritance**: If a parent Layout defines a background color, child Pages inherit that color for contrast calculations.
- **Linter Engine**: When running in `--linter` mode, the engine uses the full import graph to verify if a component (like a generic Input) receives its required accessibility attributes from a parent, preventing false positives.

## 4. Formal CSS Engine
Using the `tdewolff` parser, A11ySentry performs deep style resolution:
- **Recursive Variable Resolution**: Resolves `var(--foo)` chains across files.
- **At-Rule Detection**: Specifically handles `@media (prefers-color-scheme: dark)` to audit accessibility in dark mode.
- **Tailwind Integration**: Automatically resolves utility-first classes (e.g., `text-red-500`) using built-in and custom config palettes.

## 5. Compliance & Standardization
Every rule in the engine is mapped to:
- **WCAG 2.2**: Success criteria for international accessibility.
- **W3C ACT Rules**: Accessibility Conformance Testing IDs for industry-standard reporting.

---
*A11ySentry is built for engineers who care about structural integrity and inclusive design.*
