---
name: a11ysentry-mcp
description: >
  Expert guidance for auditing and fixing accessibility violations with A11ySentry, a
  multi-platform WCAG 2.2 engine available both as an MCP server and as a CLI. Use it when the
  task involves checking accessibility, auditing a UI, fixing an a11y or WCAG violation,
  reviewing UI changes in a PR, or reasoning about contrast, labels, focus, or ARIA. Covers Web,
  Mobile (Android, iOS, Flutter, React Native), Desktop (.NET, Electron, PyQt), and backend
  templates (Django, Flask, Angular, Vue).
version: "2.0.0"
---

# A11ySentry Accessibility Expert

You are an accessibility expert. Your goal is to make the project comply with WCAG 2.2 using
A11ySentry, a deterministic, offline engine with no telemetry. It has two surfaces that share the
same engine; pick whichever the task calls for:

- **MCP tools** — when the agent runtime has the A11ySentry MCP server registered.
- **CLI** — when the binary is on `PATH`, for scripts, CI, or bulk runs.

---

## 1. Strategic rules

### The Recursive Context Rule (CRITICAL)
Accessibility context is fragmented. An `<input>` in one file is often labeled by a `<label>` in a
parent or a sibling component, and a heading's level only makes sense inside the page tree.

- **Never** audit a component in isolation when it belongs to a tree.
- Find where the component is used (filesystem read/search), then pass **all** related files to
  `analyze_accessibility` as a comma-separated list.
- When you need the surrounding hierarchy first, call `get_component_context` on the file.

### Prefer directory mode over file mode
Passing a **directory** makes A11ySentry discover project roots, resolve the full import graph,
and build complete PageTrees. That yields context-aware results with far fewer false positives
than auditing files one by one. Use directory mode for anything larger than a single component.

### Multi-platform awareness
The adapter is selected by extension, so apply the matching mental model:

| Files | Platform |
| :--- | :--- |
| `.html`, `.htm`, `.astro`, `.vue`, `.svelte`, `.tsx`, `.jsx`, `.ts`, `.js` | Web / HTML |
| `.kt`, `.xml` (Android), `.java` | Android (Compose / Views) |
| `.swift` | iOS / SwiftUI |
| `.dart` | Flutter |
| `.xaml`, `.cs` | .NET XAML |
| `.razor` | Blazor |
| `.fxml` | JavaFX |
| `.prefab`, `.unity`, `.tscn` | Unity / Godot |

`.ts`/`.js` containing `android` or `ios` in the path are treated as React Native.

### Fixing protocol
1. **Analyze** — read the `code` (for example `WCAG_1_4_3`) and `msg`.
2. **Consult** — check `DocumentationURL` when the fix is not obvious.
3. **Execute** — use `fix` as a base, adapted to the project's component patterns and styling.
4. **Verify** — always re-run the analysis after a change. An unverified fix is not a fix.

---

## 2. MCP tools

### `analyze_accessibility(path: string)`
Audits source files or a full project directory.

- **Directory mode** — pass one directory. Project roots are discovered automatically, component
  trees are resolved, and every violation across the tree is reported once.
- **File mode** — pass one or more paths, comma-separated, to supply multi-file context
  (see the Recursive Context Rule).

Returns **TOON** (Token-Oriented Object Notation) to stay cheap on context:

```text
violations[2]{code,sev,file,line,snippet,msg,fix}:
  WCAG_1_1_1,E,src/components/Header.astro,12,<img src="logo.svg"/>,Image missing alternative text; add alt, aria-label or a platform description,Add a descriptive label for users with screen readers.
  WCAG_1_4_3,E,src/components/Button.tsx,8,<button class="text-gray-400">,Contrast ratio 2.8:1 is below the 4.5:1 minimum,Darken the text color or use a CSS variable that resolves to a compliant value.
```

- `code` — the rule identifier (`WCAG_x_y_z`, `ARIA_1_1`, `G141`, `REDUNDANT_TITLE`, `ADV_FOCUS_TRAP`).
- `sev` — `E` error or `W` warning.
- `file`, `line` — exact source location.
- `snippet` — the offending fragment, truncated to 60 characters.
- `msg` — what is wrong.
- `fix` — the suggested fix.

Commas inside `snippet`, `msg`, and `fix` are escaped to `;` so the columns do not shift.

### `get_component_context(path: string)`
Returns the component's ancestors (layouts and parents) and its children, so you can see how it
fits into the rendering hierarchy before judging it. Accepts absolute or repository-relative
paths. Use it when a violation's cause probably lives in a parent.

### `get_audit_history(limit: number)`
Returns the last N saved audit reports as plain text lines with a timestamp, file, platform, and
a pass marker or violation count. Default limit is 10. Use it to tell whether a file was already
audited, or to compare against a previous state.

---

## 3. CLI surface

The binary is `a11ysentry`. Useful when scripting, in CI, or for bulk work.

```bash
# Scaffold CI/CD into the current project
a11ysentry init                     # GitHub Actions + pre-commit hook + a11ysentry.json
a11ysentry init --force             # overwrite existing files
a11ysentry init --skip-hooks        # skip the git pre-commit hook
a11ysentry init --skip-actions      # skip the GitHub Actions workflow

a11ysentry Button.tsx Modal.astro   # analyze one or more files
a11ysentry --dir ./src              # analyze a project, resolving component trees
a11ysentry --exclude node_modules   # comma-separated directories to skip
a11ysentry --css global.css, tokens.css   # pre-load external CSS for color resolution
a11ysentry --format json            # text (default) | json | json-ld | sarif
a11ysentry --output results.txt      # write results to a file
a11ysentry --watch                  # re-analyze on file changes
a11ysentry --show-tree              # print the resolved component tree
a11ysentry --platform vue           # force a platform
a11ysentry --linter                 # JSON stream on stdin, project-aware diagnostics
a11ysentry --tui                    # interactive dashboard (press 't' for the component explorer)
a11ysentry clear                    # delete the audit history database
a11ysentry mcp --register           # register the MCP server in detected agents
a11ysentry --version
```

`--format sarif` emits SARIF 2.1.0 for GitHub Code Scanning.

### Exit codes
| Code | Meaning |
| :--- | :--- |
| `0` | No issues found |
| `1` | One or more errors found |
| `2` | Warnings only |

---

## 4. Suppression and configuration

### Inline suppression
```html
<!-- a11y-ignore: WCAG_1_1_1 -->
<img src="decorative.svg" />

<!-- a11y-ignore: all -->
<canvas id="game" />
```
Ignores are inherited by descendant elements, and `all` suppresses every rule on the subtree.
Suppress deliberately and narrowly; a blanket `all` on a large subtree hides real defects.

### `a11ysentry.json`
```json
{
  "version": "1.0",
  "paths": ["src", "app", "pages", "components"],
  "exclude": ["node_modules", ".git", "examples", "landing"],
  "format": "text",
  "exitOnWarning": false,
  "platform": "",
  "showTree": false,
  "rules": {
    "WCAG_1_4_3": { "enabled": true, "severity": "error" }
  }
}
```
It is read from the analyzed project root and merged over defaults, so only list what you change.
`rules` accepts `enabled` and `severity` (`error` or `warning`) per rule code.

---

## 5. What the engine does

- **PageTree reconstruction** — builds the real component hierarchy from the import graph, so a
  violation is judged in the tree it actually renders in.
- **Universal Semantic Node (USN)** — every platform element is normalized to one node with a
  semantic `Role`, accessible `Label`, `State`, platform `Traits`, and accurate `Source`
  (file, line, column). One rule set then covers 16 platform families.
- **Formal CSS parsing** — colors and variables are resolved through a real CSS parser, including
  CSS custom properties and dark mode, so contrast verdicts are computed, not guessed.
- **Deduplication** — identical violations are collapsed, so a component reused across many pages
  is reported once.
- **Offline and local** — no network calls, no telemetry. Audit history lives in
  `~/.a11ysentry/history.db`.

---

## 6. Worked examples

**Audit a whole project.** Call `analyze_accessibility(".")` or a source directory. Read the TOON
rows, group them by `file`, and fix the highest-severity ones first.

**Audit one component properly.** Search for usages of the component, collect the entry files and
their parents, then call `analyze_accessibility` with all of them comma-separated.

**Diagnose a contrast failure.** Note `code` and `line`. Look for CSS variables or utility classes
that set the color, resolve the actual value (remember dark mode), apply the fix, and re-run the
analysis on the same paths to confirm the row is gone.

**Check whether anything was audited before.** Call `get_audit_history(5)`.

**Run it in CI.** Use the CLI, not the MCP server, and prefer `--format sarif` with Code Scanning
or rely on the exit codes.
