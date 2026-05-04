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

#### `analyze_accessibility`

Audits source files or full project directories for accessibility violations.

**Parameters:**
```json
{
  "path": "string (required) - Absolute or relative path to the source file(s) or directory to analyze. Supports comma-separated paths for multi-file context."
}
```

**Response:**
Returns a list of violations in **TOON** format for token efficiency.

#### `get_component_context`

Returns the architectural hierarchy (parents and children) of a component to provide better context for analysis.

**Parameters:**
```json
{
  "path": "string (required) - Absolute path to the component file to investigate."
}
```

#### `get_audit_history`

Retrieves the history of past accessibility audits from the local database.

**Parameters:**
```json
{
  "limit": "number (default: 10) - Maximum number of reports to retrieve."
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
| `WCAG_1_3_1` | 1.3.1 Info and Relationships | Heading hierarchy skipped or jump |
| `WCAG_1_3_1_LEGEND` | 1.3.1 Info and Relationships | Fieldset missing a legend |
| `WCAG_1_3_5` | 1.3.5 Identify Input Purpose | Input missing autocomplete attribute |
| `WCAG_1_4_1` | 1.4.1 Use of Color | Information conveyed solely by color (e.g. no-underline links) |
| `WCAG_1_4_3` | 1.4.3 Contrast (Minimum) | Low color contrast (< 4.5:1) |
| `WCAG_1_4_3_DARK` | 1.4.3 Contrast (Minimum) | Low contrast in dark mode |
| `WCAG_1_4_3_UNRESOLVED` | 1.4.3 Contrast (Minimum) | Contrast could not be statically resolved |
| `WCAG_1_4_11` | 1.4.11 Non-text Contrast | Low contrast for UI boundaries (< 3:1) |
| `WCAG_2_1_1` | 2.1.1 Keyboard | Element with click handler lacks keyboard support |
| `WCAG_2_4_1` | 2.4.1 Bypass Blocks | Multiple <main> landmarks found |
| `WCAG_2_4_3` | 2.4.3 Focus Order | Positive tabindex or problematic focus sequence |
| `WCAG_2_4_3_HIDDEN` | 2.4.3 Focus Order | Focusable element inside aria-hidden container |
| `WCAG_2_4_4` | 2.4.4 Link Purpose | Ambiguous links (same label, different destination) |
| `WCAG_2_4_6` | 2.4.6 Headings and Labels | Empty or non-descriptive heading |
| `WCAG_2_4_7` | 2.4.7 Focus Visible | Hidden focus indicator (outline: none) |
| `WCAG_2_5_5` | 2.5.5 Target Size | Touch target < 44px (mobile) |
| `WCAG_2_5_8` | 2.5.8 Target Size (Minimum) | Touch target < 24px (web) |
| `WCAG_3_1_1` | 3.1.1 Language of Page | Missing `lang` attribute |
| `WCAG_3_3_2` | 3.3.2 Labels or Instructions | Input missing label |
| `WCAG_4_1_1` | 4.1.1 Parsing | Duplicate ID found |
| `WCAG_4_1_2` | 4.1.2 Name, Role, Value | Missing ARIA state (pressed/expanded/checked) or role |
| `WCAG_4_1_3` | 4.1.3 Status Messages | Live region missing aria-live attribute |
| `G141` | Heading Structure | Missing H1 heading |

### Advanced Rules

| Code | Description |
|------|-------------|
| `ADV_FOCUS_TRAP` | Modal/Dialog missing 'aria-modal="true"' |
| `ARIA_1_1` | Multiple landmarks of same type without unique labels |
| `REDUNDANT_TITLE` | Title attribute is identical to the element's text/label |
