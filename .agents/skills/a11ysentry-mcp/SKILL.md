---
name: a11ysentry-mcp
description: >
  Expert guidance for auditing and fixing accessibility violations using the A11ySentry MCP server.
  Triggers: "check accessibility", "audit UI", "fix a11y", "WCAG compliance", or UI changes in PRs.
allowed-tools: [analyze_accessibility, read_file, glob, grep_search]
version: "1.2.0"
---

# A11ySentry Expert Skill

You are an accessibility expert. Your goal is to ensure the project complies with WCAG 2.2 standards using the `a11ysentry` MCP tool.

## 🧠 Strategic Instructions

### 1. The Recursive Context Rule (CRITICAL)
Accessibility context is often fragmented. An `<input>` in one file might be labeled by a `<label>` in a parent or sibling.
- **NEVER** analyze a single component file in isolation if it belongs to a tree.
- **ACTION**: Use `grep_search` or `read_file` to find where the component is used.
- **ACTION**: Pass ALL related files to `analyze_accessibility` as a comma-separated list.

### 2. Multi-Platform Awareness
A11ySentry uses specific adapters. When you see these files, apply the corresponding mindset:
- `.astro`, `.vue`, `.svelte`, `.tsx`: Web/HTML standards.
- `.kt`, `.xml`: Android (Compose/Views).
- `.swift`: iOS (SwiftUI).
- `.dart`: Flutter.

### 3. Fixing Protocol
When violations are found:
1. **Analyze**: Read the `Message` and `ErrorCode` (e.g., WCAG 1.1.1).
2. **Consult**: Check the `DocumentationURL` if the fix isn't obvious.
3. **Execute**: Use the `FixSnippet` as a base, but ensure it matches the project's styling/component patterns.
4. **Verify**: ALWAYS re-run `analyze_accessibility` after making a change to confirm the fix.

## 🛠 Available Tools (via MCP)

- `analyze_accessibility(path: string)`: Audits files or directories. `path` can be a single file path, a directory path, or a comma-separated list of files.
  - **Directory Mode**: If a directory is provided, A11ySentry automatically discovers project roots (e.g., Next.js, Astro, Android) and resolves full component trees for analysis.
  - **File Mode**: Audits specific files. Use comma-separated paths to provide multi-file context (Recursive Context Rule).
  **Note**: The output is in **TOON (Token-Oriented Object Notation)** for efficiency.
  **Format**: `violations[count]{code,sev,file,line,snippet,msg,fix}:`
  - `code`: WCAG Error Code.
  - `sev`: Severity (E = Error, W = Warning).
  - `snippet`: Truncated HTML/Code fragment.
  - `msg`: Description of the issue.
  - `fix`: Suggested fix.

## 📋 Examples

### Auditing a Full Project
If the user asks to check the entire project:
1. Identify the project root directory.
2. Run: `analyze_accessibility(".")` or `analyze_accessibility("src/")`.

### Auditing a Page (Legacy/Specific)
If the user asks to check `src/pages/index.astro` and you want to be specific:
1. Identify components used in `index.astro` (e.g., `Header.astro`, `Footer.astro`).
2. Run: `analyze_accessibility("src/pages/index.astro, src/components/Header.astro, src/components/Footer.astro")`.


### Fixing a contrast issue
1. Tool reports `WCAG_1_4_3` (Contrast) in `Button.tsx`.
2. Look for CSS variables or Tailwind classes defining the color.
3. Apply fix and re-run analysis.
