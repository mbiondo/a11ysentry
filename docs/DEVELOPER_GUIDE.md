# Developer Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Creating a New Adapter](#creating-a-new-adapter)
3. [Adding a New WCAG Rule](#adding-a-new-wcag-rule)
4. [Testing Guidelines](#testing-guidelines)
5. [USN Mapping Examples](#usn-mapping-examples)
6. [Debugging Tips](#debugging-tips)

---

## Getting Started

### Prerequisites

- Go 1.24+
- Git
- PowerShell (Windows) or Bash (Unix)

### Project Structure

```
semantix/
├── engine/              # Core validation engine
│   ├── core/
│   │   ├── domain/     # USN, Violation, Analyzer
│   │   └── ports/      # Interfaces (Adapter, Repository)
│   └── persistence/
│       └── sqlite/     # SQLite implementation
├── adapters/            # Platform-specific adapters
│   ├── web/            # HTML/React/Vue/Angelo
│   ├── android/        # Kotlin/XML (Compose/View)
│   ├── ios/            # Swift (SwiftUI)
│   ├── flutter/        # Dart
│   └── ...             # Other platforms
├── cmd/
│   └── a11ysentry/     # CLI & TUI entry point
├── scanner/             # Framework detection (Next.js, Astro, generic)
├── mcp/                 # MCP server & registration
├── examples/            # Sample files for testing
└── docs/                # Documentation
```

### Build & Run

```bash
# Build CLI (from workspace root)
go build -o cmd/a11ysentry/a11ysentry ./cmd/a11ysentry

# Run tests
go test ./...

# Analyze a file
./cmd/a11ysentry/a11ysentry ../examples/example-astro/src/pages/index.astro
```

---

## Creating a New Adapter

### Step 1: Create Adapter Structure

```go
// adapters/myplatform/adapter.go
package myplatform

import (
    "context"
    "a11ysentry/engine/core/domain"
    "a11ysentry/engine/core/ports"
)

type myPlatformAdapter struct {
    platform domain.Platform
}

func NewMyPlatformAdapter() ports.Adapter {
    return &myPlatformAdapter{
        platform: domain.Platform("MY_PLATFORM"),
    }
}
```

### Step 2: Implement Ingest Method

```go
func (a *myPlatformAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
    var allNodes []domain.USN
    
    for _, file := range files {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            content, err := os.ReadFile(file)
            if err != nil {
                return nil, err
            }
            
            // Parse platform-specific syntax
            nodes := a.parseFile(string(content), file)
            allNodes = append(allNodes, nodes...)
        }
    }
    
    return allNodes, nil
}
```

### Step 3: Map to USN

```go
func (a *myPlatformAdapter) parseFile(content, filePath string) []domain.USN {
    var nodes []domain.USN
    
    // Example: Parse a button component
    // <MyButton label="Submit" disabled={true} />
    
    usn := domain.USN{
        UID:   "submit-btn",
        Role:  domain.RoleButton,
        Label: "Submit",
        State: domain.USNState{
            Disabled: true,
        },
        Traits: map[string]any{
            "variant": "primary",
        },
        Source: domain.Source{
            Platform: a.platform,
            FilePath: filePath,
            Line:     42,
            Column:   10,
        },
    }
    
    nodes = append(nodes, usn)
    return nodes
}
```

### Step 4: Register in CLI

Edit `cmd/a11ysentry/main.go`:

```go
import "a11ysentry/adapters/myplatform"

// In the switch statement:
case ".myext":
    adapter = myplatform.NewMyPlatformAdapter()
    platform = domain.Platform("MY_PLATFORM")
```

### Step 5: Add Tests

```go
// adapters/myplatform/adapter_test.go
package myplatform

import (
    "context"
    "testing"
)

func TestIngest_ButtonWithLabel(t *testing.T) {
    adapter := NewMyPlatformAdapter()
    nodes, err := adapter.Ingest(context.Background(), []string{"testdata/button.myext"})
    
    if err != nil {
        t.Fatalf("Ingest failed: %v", err)
    }
    
    if len(nodes) != 1 {
        t.Errorf("Expected 1 node, got %d", len(nodes))
    }
    
    if nodes[0].Label != "Submit" {
        t.Errorf("Expected label 'Submit', got '%s'", nodes[0].Label)
    }
}
```

---

## Adding a New WCAG Rule

### Step 1: Identify the Rule

Example: WCAG 2.5.7 - Dragging Movements

**Requirement:** Functions that use dragging movements should also be operable without dragging.

### Step 2: Add Validation Logic

Edit `engine/core/domain/violation.go`:

```go
// In the Analyze method, add:

// Rule: Dragging Movements (WCAG 2.5.7)
if node.Role == RoleImage || node.Role == "slider" {
    if draggable, ok := node.Traits["draggable"].(bool); ok && draggable {
        // Check if there's an alternative non-drag interaction
        hasAlternative := false
        if _, ok := node.Traits["aria-keyshortcuts"]; ok {
            hasAlternative = true
        }
        if _, ok := node.Traits["onDoubleClick"]; ok {
            hasAlternative = true
        }
        
        if !hasAlternative {
            violations = append(violations, Violation{
                ErrorCode:        "WCAG_2_5_7",
                Message:          "Draggable element lacks alternative non-drag interaction (keyboard or double-click).",
                SourceRef:        node.Source,
                FixSnippet:       "Add aria-keyshortcuts or double-click handler as alternative.",
                DocumentationURL: "https://www.w3.org/WAI/WCAG22/Understanding/dragging-movements.html",
            })
        }
    }
}
```

### Step 3: Add Test Cases

```go
// engine/core/domain/analyzer_test.go

func TestAnalyze_DraggingMovements(t *testing.T) {
    tests := []struct {
        name          string
        nodes         []USN
        expectViolation bool
    }{
        {
            name: "Draggable image without alternative",
            nodes: []USN{
                {
                    Role:  RoleImage,
                    Traits: map[string]any{
                        "draggable": true,
                    },
                },
            },
            expectViolation: true,
        },
        {
            name: "Draggable image with keyboard alternative",
            nodes: []USN{
                {
                    Role:  RoleImage,
                    Traits: map[string]any{
                        "draggable":       true,
                        "aria-keyshortcuts": "Enter",
                    },
                },
            },
            expectViolation: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            analyzer := NewAnalyzer()
            violations, _ := analyzer.Analyze(context.Background(), tt.nodes)
            
            if tt.expectViolation && len(violations) == 0 {
                t.Error("Expected violation, got none")
            }
            if !tt.expectViolation && len(violations) > 0 {
                t.Errorf("Expected no violation, got %d", len(violations))
            }
        })
    }
}
```

### Step 4: Update Documentation

Add to `docs/API_REFERENCE.md`:

| Code | WCAG Reference | Description |
|------|----------------|-------------|
| `WCAG_2_5_7` | 2.5.7 Dragging Movements | Missing alternative to drag interaction |

---

## Testing Guidelines

### Unit Tests

**Location:** Same package as the code being tested.

**Pattern:**
```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange
    input := ...
    
    // Act
    result, err := FunctionUnderTest(input)
    
    // Assert
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Tests

**Location:** `tests/` directory.

**Example:**
```go
// tests/integration_test.go
func TestFullPipeline_WebButton(t *testing.T) {
    // 1. Ingest
    adapter := web.NewHTMLAdapter()
    nodes, _ := adapter.Ingest(context.Background(), []string{"../examples/example-astro/src/pages/index.astro"})
    
    // 2. Analyze
    analyzer := domain.NewAnalyzer()
    violations, _ := analyzer.Analyze(context.Background(), nodes)
    
    // 3. Assert
    if len(violations) != 0 {
        t.Errorf("Expected no violations, got %d", len(violations))
    }
}
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./engine/core/domain

# Verbose output
go test -v ./...
```

---

## USN Mapping Examples

### Web (React)

**Source:**
```jsx
<button 
  id="submit-btn"
  aria-label="Submit form"
  className="w-48 h-12 bg-blue-500 text-white"
  disabled={false}
>
  Submit
</button>
```

**USN:**
```go
USN{
    UID:   "submit-btn",
    Role:  RoleButton,
    Label: "Submit form",
    State: USNState{Disabled: false},
    Traits: map[string]any{
        "width":  192.0,  // 48 * 4px
        "height": 48.0,   // 12 * 4px
        "background-color": "#3b82f6",
        "color": "#ffffff",
    },
    Source: Source{
        Platform: PlatformWebReact,
        FilePath: "src/Button.tsx",
        Line:     10,
    },
}
```

### Android (Compose)

**Source:**
```kotlin
Button(
    onClick = { submit() },
    modifier = Modifier
        .width(192.dp)
        .height(48.dp)
        .semantics {
            contentDescription = "Submit form"
        },
    enabled = true
) {
    Text("Submit")
}
```

**USN:**
```go
USN{
    UID:   "button-1",
    Role:  RoleButton,
    Label: "Submit form",
    State: USNState{Disabled: false},
    Traits: map[string]any{
        "width":  192.0,
        "height": 48.0,
    },
    Source: Source{
        Platform: PlatformAndroidCompose,
        FilePath: "MainActivity.kt",
        Line:     42,
    },
}
```

### iOS (SwiftUI)

**Source:**
```swift
Button(action: submit) {
    Text("Submit")
        .frame(width: 192, height: 48)
        .background(Color.blue)
}
.accessibilityLabel("Submit form")
.disabled(false)
```

**USN:**
```go
USN{
    UID:   "button-1",
    Role:  RoleButton,
    Label: "Submit form",
    State: USNState{Disabled: false},
    Traits: map[string]any{
        "width":  192.0,
        "height": 48.0,
        "background-color": "#007AFF",  // iOS blue
    },
    Source: Source{
        Platform: PlatformIOSSwiftUI,
        FilePath: "ContentView.swift",
        Line:     15,
    },
}
```

### Flutter

**Source:**
```dart
Semantics(
  label: 'Submit form',
  child: ElevatedButton(
    onPressed: submit,
    style: ElevatedButton.styleFrom(
      fixedSize: Size(192, 48),
      backgroundColor: Colors.blue,
    ),
    child: Text('Submit'),
  ),
)
```

**USN:**
```go
USN{
    UID:   "button-1",
    Role:  RoleButton,
    Label: "Submit form",
    State: USNState{Disabled: false},
    Traits: map[string]any{
        "width":  192.0,
        "height": 48.0,
        "background-color": "#1976D2",  // Material blue
    },
    Source: Source{
        Platform: PlatformFlutterDart,
        FilePath: "lib/main.dart",
        Line:     25,
    },
}
```

---

## Debugging Tips

### Enable Verbose Logging

```bash
# Set debug environment variable
export A11YSENTRY_DEBUG=1

# Run with verbose flag (if implemented)
a11ysentry --verbose file.html
```

### Inspect USN Output

Add debug print in adapter:

```go
func (a *htmlAdapter) traverse(...) []USN {
    // ...
    usn := domain.USN{...}
    
    // Debug: Print USN
    fmt.Printf("DEBUG USN: %+v\n", usn)
    
    nodes = append(nodes, usn)
    // ...
}
```

### Test with Example Files

```bash
# Analyze example file
a11ysentry ../examples/example-astro/src/pages/index.astro

# Check generated USN (add --debug flag if implemented)
a11ysentry --debug examples/example-astro/src/pages/index.astro
```

### Use TUI for History Inspection

```bash
# Open TUI dashboard
a11ysentry --tui

# Navigate to recent audits
# Press 'Enter' to view detailed violations
```

### Common Issues

**Issue:** Adapter not detecting file type

**Solution:** Check file extension mapping in `cmd/a11ysentry/main.go`:
```go
switch ext {
case ".html", ".htm":
    adapter = web.NewHTMLAdapter()
// Add your extension here
```

**Issue:** False positive violations

**Solution:** Check if USN mapping is extracting the label correctly:
```go
fmt.Printf("Label extracted: '%s'\n", usn.Label)
```

**Issue:** Touch target false positives

**Solution:** Verify Tailwind class mapping:
```go
// Check if w-12 → 48px is working
fmt.Printf("Width trait: %v\n", usn.Traits["width"])
```

---

## Code Style

### Formatting

```bash
# Format all Go files
go fmt ./...

# Lint
go vet ./...
```

### Naming Conventions

- **Functions:** `CamelCase` (exported), `camelCase` (unexported)
- **Interfaces:** `-er` suffix (`Analyzer`, `Adapter`, `Repository`)
- **Constants:** `PascalCase` with prefix (`RoleButton`, `PlatformWeb`)
- **Test Functions:** `TestFunctionName_Scenario`

### Comments

```go
// Good: Explains WHY, not WHAT
// Two-pass validation ensures we collect all labels before checking inputs
for _, node := range nodes {
    // ...
}

// Bad: Redundant
// Loop through nodes
for _, node := range nodes {
    // ...
}
```

---

## Performance Optimization

### Concurrent File Processing

```go
// Good: Process files in parallel
nodeChan := make(chan []USN, len(files))
for _, file := range files {
    go func(f string) {
        nodes, _ := adapter.Ingest(ctx, []string{f})
        nodeChan <- nodes
    }(file)
}

// Bad: Sequential processing
for _, file := range files {
    nodes, _ := adapter.Ingest(ctx, []string{file})
    allNodes = append(allNodes, nodes...)
}
```

### Avoid Unnecessary Allocations

```go
// Good: Pre-allocate slice with known size
violations := make([]Violation, 0, len(nodes))

// Bad: Append without capacity hint
var violations []Violation
```

### Memory-Efficient USN Creation

```go
// Good: Only set non-zero values
usn := domain.USN{
    UID:   id,
    Role:  role,
    Label: label,
}

// Bad: Set all fields explicitly even when zero
usn := domain.USN{
    UID:   id,
    Role:  role,
    Label: label,
    State: domain.USNState{},  // Unnecessary if all fields are zero
}
```

---

## Contributing Workflow

1. **Fork the repository**
2. **Create a branch:**
   ```bash
   git checkout -b feat/add-gaming-platform-adapter
   ```
3. **Make changes and write tests**
4. **Run tests:**
   ```bash
   go test ./...
   ```
5. **Format code:**
   ```bash
   go fmt ./...
   go vet ./...
   ```
6. **Commit with conventional commits:**
   ```bash
   git commit -m "feat(godot): add Godot engine adapter"
   ```
7. **Push and open PR**

---

## Resources

- [WCAG 2.2 Guidelines](https://www.w3.org/WAI/WCAG22/quickref/)
- [Universal Semantic Node Spec](./API_REFERENCE.md#universal-semantic-node-usn)
- [Architecture Deep Dive](./ARCHITECTURE_DEEP_DIVE.md)
- [MCP Integration Guide](./MCP_INTEGRATION.md)

## Examples actuales
- example-astro: proyecto de ejemplo con Astro (archivo de configuración, layout y páginas de muestra en src/…)
- example-nextjs: Next.js App Router de ejemplo con app/layout.tsx, app/page.tsx, y componentes de ejemplo
- example-nuxt: Nuxt 3 de ejemplo con nuxt.config.ts, layouts/default.vue y pages
- example-sveltekit: SvelteKit de ejemplo con +layout.svelte, +page.svelte y rutas de ejemplo

- Se eliminó el directorio examples/web y toda su contenido asociado. Las referencias a ese conjunto ya no forman parte de la guía ni de la ejecución de desarrollo.

### Estructura y descubrimiento de proyectos
- El sistema de descubrimiento de proyectos para monorepos utiliza FindProjectRoots para recorrer directorios y detectar roots de proyectos en los que exista un package.json o indicios de frameworks soportados.
- Nuxt/SvelteKit quedan integrados como scanners del módulo a11ysentry/scanner, y se detectan durante la ejecución de Detect().

### Compilación para desarrollo
- Compilar el binario de desarrollo desde la raíz del repositorio: 
  go build -o cmd/a11ysentry/a11ysentry.exe ./cmd/a11ysentry
- El workspace Go se resuelve mediante go.work; los módulos locales se resuelven automáticamente y no se necesita un cli separado.
