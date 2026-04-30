# MCP Integration Guide

## Overview

A11ySentry integrates with AI agents via the **Model Context Protocol (MCP)**, allowing AI assistants to analyze accessibility violations directly within your development workflow. All analysis happens **locally** — no cloud calls, no telemetry.

---

## Supported AI Agents

| Agent | Platform | Config Location |
|-------|----------|-----------------|
| **Claude Desktop** | macOS, Windows | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)<br>`%APPDATA%\Claude\claude_desktop_config.json` (Windows) |
| **Cursor** | macOS, Windows, Linux | `~/.cursor/mcp.json` |
| **VS Code** | All | `.vscode/mcp.json` or user settings |
| **Gemini CLI** | All | `~/.gemini/mcp.json` |
| **OpenCode** | All | `~/.opencode/mcp.json` |

---

## Installation

### Step 1: Install A11ySentry

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.ps1 | iex
```

**Unix (Bash):**
```bash
curl -sSL https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.sh | bash
```

**Manual (Go):**
```bash
cd cli
go build -o a11ysentry.exe .   # Windows
go build -o a11ysentry .       # Unix
```

### Step 2: Initialize Your Project (Recommended)

Run `init` in your project root to auto-scaffold everything:

```bash
a11ysentry init
```

This creates:
- `.github/workflows/a11y.yml` — GitHub Actions with SARIF upload (push + PR)
- `.git/hooks/pre-commit` — local pre-commit accessibility check
- `a11ysentry.json` — project config (paths, format, exit codes)

Options:
```bash
a11ysentry init --force           # overwrite existing files
a11ysentry init --skip-hooks      # skip pre-commit hook
a11ysentry init --skip-actions    # skip GitHub Actions workflow
```

### Step 3: Register MCP

```bash
# Register in all detected agents automatically
a11ysentry mcp --register

# Verify registration
a11ysentry mcp --check-mcp
```

---

## Manual Configuration

### Claude Desktop

Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "a11ysentry": {
      "command": "a11ysentry",
      "args": ["mcp"]
    }
  }
}
```

### Cursor

Edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "a11ysentry": {
      "command": "a11ysentry",
      "args": ["mcp"]
    }
  }
}
```

### OpenCode

Edit `~/.opencode/mcp.json`:

```json
{
  "mcpServers": {
    "a11ysentry": {
      "command": "a11ysentry",
      "args": ["mcp"]
    }
  }
}
```

---

## Available Tools

The MCP server exposes **one tool**:

### `analyze_accessibility`

Audits one or more source files for WCAG 2.2 accessibility violations.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | ✅ | Path(s) to analyze. Comma-separate multiple files for component tree context. |

**Example — single file:**

```json
{
  "name": "analyze_accessibility",
  "arguments": {
    "path": "src/components/Button.tsx"
  }
}
```

**Example — component tree (recommended):**

```json
{
  "name": "analyze_accessibility",
  "arguments": {
    "path": "src/pages/Home.astro, src/components/Header.astro, src/components/Input.tsx"
  }
}
```

**Response (violations found):**

```json
[
  {
    "ErrorCode": "WCAG_1_1_1",
    "Severity": "error",
    "Message": "Image missing alternative text. Every image must have an 'alt', 'aria-label', or platform-specific description attribute.",
    "SourceRef": {
      "FilePath": "src/components/Header.astro",
      "Line": 12,
      "Column": 5,
      "RawHTML": "<img src=\"logo.svg\"/>"
    },
    "FixSnippet": "Add a descriptive label for users with screen readers.",
    "DocumentationURL": "https://www.w3.org/WAI/WCAG22/Techniques/general/G94"
  }
]
```

**Response (no violations):**

```
✅ No accessibility violations found in src/components/Button.tsx.
```

---

## CLI Reference

For use outside of AI agents, the binary offers a full CLI:

```bash
# Analyze one or more files
a11ysentry file.tsx Component.vue

# Analyze an entire project (resolves import graph automatically)
a11ysentry --dir ./src

# Pre-load external CSS for accurate color/contrast resolution
a11ysentry --css global.css,tokens.css Component.vue

# Output formats
a11ysentry --format text    # default — human-readable
a11ysentry --format json    # machine-readable JSON
a11ysentry --format json-ld # JSON-LD with schema.org context
a11ysentry --format sarif   # SARIF 2.1.0 (GitHub Code Scanning)

# Watch mode — re-analyze on file changes
a11ysentry --watch Component.vue

# Force platform
a11ysentry --platform vue Component.vue
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No issues found |
| `1` | One or more errors found |
| `2` | Warnings only (no errors) |

---

## Supported Platforms

| Extension | Adapter |
|-----------|---------|
| `.html`, `.htm`, `.astro`, `.vue`, `.svelte` | Web (HTML parser) |
| `.tsx`, `.jsx` | React / Web |
| `.ts`, `.js` | React Native (auto-detected) or Web |
| `.kt` | Android Compose (Kotlin) |
| `.xml` | Android View |
| `.java` | Android (Java) |
| `.swift` | iOS / SwiftUI |
| `.dart` | Flutter |
| `.xaml`, `.cs` | .NET XAML |
| `.razor` | Blazor |
| `.fxml` | JavaFX |
| `.prefab`, `.unity` | Unity |
| `.tscn` | Godot |

---

## CI/CD Integration

### GitHub Actions with SARIF (Code Scanning)

```yaml
# .github/workflows/a11y.yml
name: Accessibility Check

on: [push, pull_request]

jobs:
  a11y:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4

      - name: Install A11ySentry
        run: curl -sSL https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.sh | bash

      - name: Run Accessibility Audit
        run: a11ysentry --format sarif --dir ./src > results.sarif
        continue-on-error: true

      - name: Upload SARIF to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Exit-code-based Gate

```yaml
      - name: Run Accessibility Audit
        run: |
          a11ysentry --dir ./src
          # exits 0=ok, 1=errors, 2=warnings
```

---

## Workflow Examples

### Cursor IDE

1. Open a component, ask in chat:
   ```
   @A11ySentry check src/components/Button.tsx for accessibility issues
   ```

2. The agent calls `analyze_accessibility` with the file path.

3. Fix violations using the returned `FixSnippet` and `DocumentationURL`.

### Claude Desktop

```
Paste your component and ask:
"Audit this component for WCAG 2.2 violations"
```

Claude will call `analyze_accessibility` with the temporary file and return structured violations.

---

## Testing the MCP Server Manually

```bash
# Start the server in stdio mode (no output = running correctly)
a11ysentry mcp

# Send a JSON-RPC request
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"analyze_accessibility","arguments":{"path":"examples/example-astro/src/pages/index.astro"}}}' | a11ysentry mcp
```

---

## Troubleshooting

### MCP server not starting

```bash
# Verify binary is in PATH
where a11ysentry    # Windows
which a11ysentry    # Unix

# Test manually
a11ysentry mcp      # Should produce no output (stdio mode)
```

### Tools not appearing in agent

1. Restart the AI agent (Claude Desktop / Cursor / VS Code).
2. Validate your config JSON syntax:
   ```bash
   cat ~/.cursor/mcp.json | jq .
   ```

### Analysis returns no violations

- Ensure the file path is correct (absolute paths are safest via MCP).
- Test with a known-violation file:
  ```bash
  a11ysentry examples/blazor/Settings.razor
  ```

---

## Security

- **Local processing**: All analysis runs on-device; no data leaves the machine.
- **No telemetry**: Zero analytics or usage tracking.
- **File access**: Only reads the files you explicitly provide.
- **SQLite**: Audit history stored in `~/.a11ysentry/history.db`.

---

## Resources

- [API Reference](./API_REFERENCE.md)
- [Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md)
- [Developer Guide](./DEVELOPER_GUIDE.md)
- [MCP Specification](https://modelcontextprotocol.io/)
- [WCAG 2.2 Quick Reference](https://www.w3.org/WAI/WCAG22/quickref/)
