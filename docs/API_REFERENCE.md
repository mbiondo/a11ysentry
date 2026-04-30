# API Reference

## Core Domain Types

### Universal Semantic Node (USN)

The USN is the core abstraction that allows A11ySentry to validate accessibility across 15+ platforms with a single set of rules.

```go
type USN struct {
    UID       string       // Unique identifier (from id attribute or tag name)
    Role      SemanticRole // Semantic role (button, heading, link, etc.)
    Label     string       // Accessible label or description
    State     USNState     // Interactive state (disabled, hidden, etc.)
    Traits    map[string]any // Platform-specific attributes and styles
    Geometry  Geometry     // Physical bounds (x, y, width, height)
    Hierarchy Hierarchy    // Parent-child relationships
    Source    Source       // Origin platform and file location
}
```

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `UID` | `string` | Unique identifier from `id` attribute or fallback to tag name |
| `Role` | `SemanticRole` | Normalized semantic role (see below) |
| `Label` | `string` | Accessible name from `aria-label`, `alt`, or text content |
| `State` | `USNState` | Interactive state flags |
| `Traits` | `map[string]any` | Dynamic attributes (styles, ARIA states, platform props) |
| `Geometry` | `Geometry` | Spatial dimensions for touch target validation |
| `Hierarchy` | `Hierarchy` | Tree structure with parent/children references |
| `Source` | `Source` | Platform, file path, line/column for error reporting |

### Semantic Roles

```go
const (
    RoleButton     SemanticRole = "button"
    RoleHeading    SemanticRole = "heading"
    RoleLink       SemanticRole = "link"
    RoleInput      SemanticRole = "input"
    RoleImage      SemanticRole = "image"
    RoleLiveRegion SemanticRole = "live-region"
    RoleModal      SemanticRole = "modal"
)
```

### Platforms

```go
const (
    // Web
    PlatformWebReact    Platform = "WEB_REACT"
    PlatformBlazor      Platform = "BLAZOR"
    PlatformElectron    Platform = "ELECTRON"
    PlatformTauri       Platform = "TAURI"
    
    // Mobile
    PlatformAndroidCompose Platform = "ANDROID_COMPOSE"
    PlatformAndroidView    Platform = "ANDROID_VIEW"
    PlatformIOSSwiftUI     Platform = "IOS_SWIFTUI"
    PlatformFlutterDart    Platform = "FLUTTER_DART"
    PlatformReactNative    Platform = "REACT_NATIVE"
    
    // Desktop
    PlatformDotNetXAML  Platform = "DOTNET_XAML"
    PlatformDotNetCSharp Platform = "DOTNET_CSHARP"
    PlatformJavaFX      Platform = "JAVA_FX"
    PlatformJavaSwing   Platform = "JAVA_SWING"
    
    // Gaming
    PlatformUnity  Platform = "UNITY"
    PlatformGodot  Platform = "GODOT"
)
```

### Violation

```go
type Violation struct {
    ErrorCode        string  // WCAG reference (e.g., "WCAG_1_1_1")
    Message          string  // Human-readable description
    SourceRef        Source  // Exact location in source code
    FixSnippet       string  // Suggested fix
    DocumentationURL string  // Link to WCAG documentation
}
```

### ViolationReport

```go
type ViolationReport struct {
    ID         int64        // Database ID
    FilePath   string       // Analyzed file path
    Platform   Platform     // Source platform
    Timestamp  int64        // Unix timestamp
    Violations []Violation  // List of violations found
}
```

---

## Interfaces

### Analyzer

Core validation engine interface.

```go
type Analyzer interface {
    Analyze(ctx context.Context, nodes []USN) ([]Violation, error)
}
```

**Usage:**
```go
analyzer := domain.NewAnalyzer()
violations, err := analyzer.Analyze(ctx, usnNodes)
```

### Adapter

Platform-specific ingestion interface.

```go
type Adapter interface {
    Ingest(ctx context.Context, files []string) ([]USN, error)
}
```

**Available Adapters:**

| Adapter | Platform | Extensions |
|---------|----------|------------|
| `web.NewHTMLAdapter()` | Web (React, Vue, Angular, Astro) | `.html`, `.htm` |
| `android.NewAndroidAdapter()` | Android | `.kt`, `.xml`, `.java` |
| `ios.NewIOSAdapter()` | iOS | `.swift` |
| `flutter.NewFlutterAdapter()` | Flutter | `.dart` |
| `dotnet.NewDotNetAdapter()` | .NET | `.xaml`, `.cs` |
| `javadesktop.NewJavaDesktopAdapter()` | Java Desktop | `.fxml` |
| `reactnative.NewReactNativeAdapter()` | React Native | `.js`, `.jsx`, `.ts`, `.tsx` |
| `blazor.NewBlazorAdapter()` | Blazor | `.razor` |
| `unity.NewUnityAdapter()` | Unity | `.prefab`, `.unity` |
| `godot.NewGodotAdapter()` | Godot | `.tscn` |

### Repository

Persistence layer interface.

```go
type Repository interface {
    SaveReport(ctx context.Context, report ViolationReport) error
    GetReports(ctx context.Context, limit int) ([]ViolationReport, error)
    GetReportByID(ctx context.Context, id int64) (*ViolationReport, error)
    DeleteReport(ctx context.Context, id int64) error
}
```

**Implementation:** `sqlite.NewSQLiteRepository(dbPath string)`

---

## MCP Server

### Tools

The MCP server exposes the following tools to AI agents:

#### `analyze_file`

Analyzes a file for accessibility violations.

**Parameters:**
```json
{
  "file_path": "string (required)",
  "platform": "string (optional, auto-detected)"
}
```

**Response:**
```json
{
  "violations": [
    {
      "error_code": "WCAG_1_1_1",
      "message": "Image missing alternative text.",
      "line": 42,
      "column": 10,
      "fix_snippet": "Add alt=\"Description\" to the image tag."
    }
  ],
  "total": 1
}
```

#### `get_audit_history`

Retrieves past audit reports from the SQLite database.

**Parameters:**
```json
{
  "limit": "number (default: 10)"
}
```

#### `check_wcag_compliance`

Checks if a code snippet meets specific WCAG criteria.

**Parameters:**
```json
{
  "code": "string (required)",
  "wcag_level": "string (A, AA, AAA)",
  "platform": "string"
}
```

---

## CLI Commands

### Direct Analysis

```bash
a11ysentry path/to/file.html [path/to/file2.kt ...]
```

### Output Formats

```bash
# Text (default)
a11ysentry file.html

# JSON
a11ysentry --format json file.html

# JSON-LD
a11ysentry --format json-ld file.html
```

### TUI Dashboard

```bash
a11ysentry --tui
```

### MCP Server

```bash
# Start MCP server (stdio)
a11ysentry mcp

# Register in all detected AI agents
a11ysentry mcp --register

# Check registration status
a11ysentry mcp --check-mcp

# Register with custom binary path
a11ysentry mcp --register --binary /usr/local/bin/a11ysentry
```

---

## Configuration

### SQLite Database Location

Default: `~/.a11ysentry/history.db`

Override via environment variable:
```bash
export A11YSENTRY_DB_PATH=/custom/path/history.db
```

### Tailwind CSS Mapping

The web adapter includes built-in Tailwind CSS 4 utility mapping:

| Pattern | Maps To |
|---------|---------|
| `w-{n}` | `width: n * 4px` |
| `h-{n}` | `height: n * 4px` |
| `text-{color}` | `color: #{hex}` |
| `bg-{color}` | `background-color: #{hex}` |

Example: `w-12 h-12 text-red-500` → `width: 48px, height: 48px, color: #ef4444`

---

## Error Codes

### WCAG Violations

| Code | WCAG Reference | Description |
|------|----------------|-------------|
| `WCAG_1_1_1` | 1.1.1 Non-text Content | Missing alt text or label |
| `WCAG_1_3_1` | 1.3.1 Info and Relationships | Heading hierarchy skipped |
| `WCAG_1_4_3` | 1.4.3 Contrast (Minimum) | Low color contrast |
| `WCAG_2_4_7` | 2.4.7 Focus Visible | Hidden focus indicator |
| `WCAG_2_5_5` | 2.5.5 Target Size | Touch target < 44px (mobile) |
| `WCAG_2_5_8` | 2.5.8 Target Size (Minimum) | Touch target < 24px (web) |
| `WCAG_3_1_1` | 3.1.1 Language of Page | Missing `lang` attribute |
| `WCAG_3_3_2` | 3.3.2 Labels or Instructions | Input missing label |
| `WCAG_4_1_1` | 4.1.1 Parsing | Duplicate ID |
| `WCAG_4_1_2` | 4.1.2 Name, Role, Value | Missing ARIA state or role |
| `G141` | Heading Structure | Missing H1 heading |

### Advanced Rules

| Code | Description |
|------|-------------|
| `ADV_FOCUS_TRAP` | Modal missing focus trap |
