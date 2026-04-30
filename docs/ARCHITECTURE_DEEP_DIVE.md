# Architecture Deep Dive

## Overview

A11ySentry implements the **Adapter-Hub Pattern**, a specialized architectural pattern for multi-platform accessibility validation.

```
┌─────────────────────────────────────────────────────────────────┐
│                         A11ySentry                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   Web    │  │ Android  │  │   iOS    │  │  Unity   │       │
│  │ Adapter  │  │ Adapter  │  │ Adapter  │  │ Adapter  │  ...  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
│       │             │             │             │               │
│       └─────────────┴─────────────┴─────────────┘               │
│                          │                                      │
│                          ▼                                      │
│         ┌──────────────────────────────────┐                   │
│         │   Universal Semantic Node (USN)  │                   │
│         │   Platform-Agnostic Abstraction  │                   │
│         └──────────────────────────────────┘                   │
│                          │                                      │
│                          ▼                                      │
│         ┌──────────────────────────────────┐                   │
│         │      Core Analysis Engine        │                   │
│         │   (WCAG 2.2 Rules - Stateless)   │                   │
│         └──────────────────────────────────┘                   │
│                          │                                      │
│                          ▼                                      │
│         ┌──────────────────────────────────┐                   │
│         │        Emission Layer            │                   │
│         │  CLI │ TUI │ MCP │ JSON │ SQLite │                   │
│         └──────────────────────────────────┘                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Pipeline Stages

### Stage 1: Ingestion (Adapters)

**Purpose:** Parse platform-specific source files and extract semantic information.

**Characteristics:**
- Stateless and concurrent (processes files in parallel)
- Platform-specific parsing logic
- Extracts raw attributes, styles, and hierarchy

**Example Flow (Web Adapter):**
```
HTML File → html.Parse() → AST Traversal → USN Extraction
```

**Code Structure:**
```go
func (a *htmlAdapter) Ingest(ctx context.Context, files []string) ([]USN, error) {
    // 1. Read file content
    // 2. Parse into AST (golang.org/x/net/html)
    // 3. Extract CSS from <style> tags
    // 4. Traverse AST and create USN nodes
    // 5. Return normalized nodes
}
```

### Stage 2: Normalization (USN Mapping)

**Purpose:** Transform platform-specific attributes into the Universal Semantic Node schema.

**Mapping Examples:**

| Platform | Source | USN Field |
|----------|--------|-----------|
| Web React | `aria-label="Submit"` | `Label: "Submit"` |
| Android Compose | `Modifier.semantics { label = "Submit" }` | `Label: "Submit"` |
| iOS SwiftUI | `.accessibilityLabel("Submit")` | `Label: "Submit"` |
| Flutter | `Semantics(label: "Submit")` | `Label: "Submit"` |

**Tailwind CSS Resolution:**
```go
// Input: class="w-12 h-12 text-red-500"
// Output:
USN.Traits = {
    "width": 48.0,           // 12 * 4px
    "height": 48.0,          // 12 * 4px
    "color": "#ef4444"       // red-500 hex
}
```

### Stage 3: Analysis (Core Engine)

**Purpose:** Apply WCAG 2.2 rules to USN trees.

**Key Design Decisions:**

1. **Stateless Analysis:** The analyzer does not maintain state between runs. This ensures:
   - Deterministic results (same input = same output)
   - Thread-safe concurrent execution
   - Easy testing and debugging

2. **Two-Pass Validation:**
   - **Pass 1:** Collect metadata (labels, document info)
   - **Pass 2:** Apply rules using collected context

3. **Platform-Agnostic Rules:** Rules operate on USN, not source code.

**Example Rule (Image Alt Text):**
```go
if node.Role == RoleImage && node.Label == "" {
    violations = append(violations, Violation{
        ErrorCode: "WCAG_1_1_1",
        Message: "Image missing alternative text.",
        SourceRef: node.Source,
        FixSnippet: "Add alt=\"Description\" to the image tag.",
    })
}
```

### Stage 4: Emission

**Purpose:** Output results in multiple formats.

**Outputs:**
- **CLI:** Text output with colored violations
- **TUI:** Interactive dashboard with history
- **MCP:** JSON-RPC tools for AI agents
- **JSON/JSON-LD:** Machine-readable reports
- **SQLite:** Persistent audit history

---

## Data Flow Diagram

```
┌─────────────┐
│ Source File │ (e.g., Button.tsx)
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│ 1. Ingestion (React Adapter)    │
│    - Parse JSX                  │
│    - Extract props              │
│    - Map Tailwind classes       │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│ 2. USN Normalization            │
│    {                            │
│      UID: "submit-btn",         │
│      Role: "button",            │
│      Label: "Submit",           │
│      Traits: {                  │
│        "width": 120,            │
│        "height": 44,            │
│        "aria-pressed": "false"  │
│      },                         │
│      Source: {                  │
│        Platform: "WEB_REACT",   │
│        FilePath: "src/Button.tsx",
│        Line: 42                 │
│      }                          │
│    }                            │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│ 3. Analysis (WCAG Rules)        │
│    - Check label presence ✓     │
│    - Check touch target size ✓  │
│    - Check ARIA states ✗        │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│ 4. Emission                     │
│    Violation: {                 │
│      ErrorCode: "WCAG_4_1_2",   │
│      Message: "Interactive      │
│                button missing   │
│                state attribute",│
│      Line: 42,                  │
│      FixSnippet: "Add           │
│        aria-pressed={pressed}"  │
│    }                            │
└─────────────────────────────────┘
```

---

## Persistence Layer

### SQLite Schema

```sql
CREATE TABLE audit_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    platform TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE violations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    error_code TEXT NOT NULL,
    message TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line INTEGER,
    column INTEGER,
    raw_html TEXT,
    fix_snippet TEXT,
    documentation_url TEXT,
    FOREIGN KEY (report_id) REFERENCES audit_reports(id)
);

CREATE INDEX idx_reports_timestamp ON audit_reports(timestamp DESC);
CREATE INDEX idx_violations_report ON violations(report_id);
```

### Repository Pattern

```go
type Repository interface {
    SaveReport(ctx context.Context, report ViolationReport) error
    GetReports(ctx context.Context, limit int) ([]ViolationReport, error)
    GetReportByID(ctx context.Context, id int64) (*ViolationReport, error)
    DeleteReport(ctx context.Context, id int64) error
}
```

**Implementation:** `engine/persistence/sqlite/repository.go`

---

## MCP Integration Architecture

### Server Components

```
┌────────────────────────────────────────────┐
│            MCP Server (stdio)              │
├────────────────────────────────────────────┤
│                                            │
│  ┌──────────────────────────────────┐     │
│  │  JSON-RPC Request Handler        │     │
│  │  - Parse incoming tool calls     │     │
│  │  - Validate parameters           │     │
│  └──────────────┬───────────────────┘     │
│                 │                          │
│  ┌──────────────▼───────────────────┐     │
│  │  Tool Executor                   │     │
│  │  - analyze_file                  │     │
│  │  - get_audit_history            │     │
│  │  - check_wcag_compliance        │     │
│  └──────────────┬───────────────────┘     │
│                 │                          │
│  ┌──────────────▼───────────────────┐     │
│  │  Core Engine & Repository        │     │
│  │  - Run analysis pipeline         │     │
│  │  - Query SQLite history          │     │
│  └──────────────────────────────────┘     │
│                                            │
└────────────────────────────────────────────┘
```

### Registration Process

When running `a11ysentry mcp --register`:

1. **Detect Installed Agents:**
   - Claude Desktop: `%APPDATA%\Claude\claude_desktop_config.json` (Windows) or `~/Library/Application Support/Claude/` (macOS)
   - Cursor: `~/.cursor/mcp.json`
   - VS Code: Extension-specific config
   - Gemini CLI: `~/.gemini/mcp.json`

2. **Update Configuration:**
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

3. **Verify Registration:** Test that the server starts correctly.

---

## TUI Dashboard Architecture

### Component Stack

```
┌──────────────────────────────────────┐
│         Bubbletea Program            │
│  (github.com/charmbracelet/bubbletea)│
└──────────────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│ History │ │ Report  │ │ Stats   │
│  List   │ │ Viewer  │ │ Panel   │
└─────────┘ └─────────┘ └─────────┘
```

### State Management

```go
type MainModel struct {
    repo          Repository
    reports       []ViolationReport
    cursor        int
    selectedReport *ViolationReport
    viewMode      ViewMode  // list | detail | stats
}
```

**Views:**
- **List:** Paginated list of past audits
- **Detail:** Full violation report with navigation
- **Stats:** Aggregate metrics (total violations by type, platform distribution)

---

## Performance Characteristics

### Benchmarks

| Stage | Time (per file) | Memory |
|-------|-----------------|--------|
| Ingestion | 10-50ms | 1-5MB |
| Normalization | 5-10ms | <1MB |
| Analysis | 1-5ms | <1MB |
| **Total** | **16-65ms** | **2-7MB** |

### Concurrency Model

- **File-level parallelism:** Multiple files processed concurrently
- **Channel-based synchronization:** Results collected via channels
- **Context cancellation:** Graceful shutdown support

```go
nodeChan := make(chan []USN, len(files))
errChan := make(chan error, len(files))

for _, file := range files {
    go func(f string) {
        // Process file
        nodes, err := adapter.Ingest(ctx, []string{f})
        if err != nil {
            errChan <- err
            return
        }
        nodeChan <- nodes
    }(file)
}
```

---

## Security Considerations

1. **No LLM Inference:** All validation is deterministic code-based logic
2. **File System Access:** Only reads user-specified files
3. **No Network Calls:** Except for documentation URLs in violation reports
4. **SQLite Isolation:** Database stored in user home directory with 0755 permissions

---

## Extensibility

### Adding a New Platform

1. **Create Adapter:** Implement `ports.Adapter` interface
2. **Map to USN:** Define platform-specific → USN mappings
3. **Register in CLI:** Add case in `main.go` switch statement
4. **Test:** Add adapter tests with platform-specific examples

### Adding a New Rule

1. **Define in Analyzer:** Add validation logic in `Analyze()` method
2. **Assign WCAG Code:** Use standard WCAG reference
3. **Provide Fix Snippet:** Give actionable remediation
4. **Test:** Add test case with passing/failing examples

---

## Design Decisions

### Why USN Instead of Direct Analysis?

**Problem:** Building platform-specific rules for 15+ platforms would require 15x the effort.

**Solution:** USN provides a single abstraction layer.

**Trade-offs:**
- ✅ Single rule implementation serves all platforms
- ✅ Easier to maintain and test
- ⚠️ Some platform-specific nuances may be lost (mitigated via `Traits` map)

### Why Stateless Analysis?

**Problem:** Stateful analyzers are hard to test, debug, and scale.

**Solution:** Pure function approach (input → output, no side effects).

**Benefits:**
- ✅ Deterministic results
- ✅ Thread-safe by default
- ✅ Easy to unit test
- ✅ Can run multiple analyses concurrently

### Why SQLite for Persistence?

**Problem:** Need to store audit history without external dependencies.

**Solution:** Embedded SQLite database.

**Alternatives Considered:**
- PostgreSQL: Overkill, requires server
- JSON files: Poor query performance
- In-memory: Loses data on exit

---

## Future Architecture Considerations

1. **gRPC Service:** Expose engine as a network service for CI/CD integration
2. **Plugin System:** Allow third-party adapters via WASM
3. **Real-time Watching:** File watcher for continuous analysis
4. **Cloud Sync:** Optional cloud backup of audit history
