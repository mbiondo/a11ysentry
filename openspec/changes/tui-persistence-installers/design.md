# Design: CLI TUI, SQLite Persistence, and Installers

## Technical Approach
This design introduces a persistence layer to the Core Engine and replaces the existing POC CLI with a Bubbletea-based TUI. It also establishes a distribution pipeline using GoReleaser and platform-native installer scripts that handle MCP registration.

## Architecture Decisions

### Decision: SQLite Driver Selection
**Choice**: `modernc.org/sqlite`
**Alternatives considered**: `github.com/mattn/go-sqlite3` (CGO).
**Rationale**: `modernc.org/sqlite` is a pure Go implementation. This allows us to keep `CGO_ENABLED=0` for all GoReleaser builds, ensuring stable cross-compilation for Windows, Linux, and macOS without needing complex cross-compilers.

### Decision: Repository Pattern
**Choice**: Port-Adapter pattern in `engine`.
**Alternatives considered**: Direct DB calls from the CLI.
**Rationale**: By defining a `Repository` interface in `engine/core/ports`, we keep the core engine testable and decoupled from the SQLite implementation.

### Decision: TUI Framework
**Choice**: `charmbracelet/bubbletea`
**Alternatives considered**: `tview`, `termui`.
**Rationale**: Bubbletea is the modern standard for interactive CLI tools in Go, offering great composability and a robust ecosystem (Lipgloss, Bubbles).

## Data Flow

    [CLI TUI] <───> [Core Engine] ───> [Adapters]
       │               │
       └───────────────┴───> [SQLite Persistence]

1. **User Action**: User triggers audit or opens dashboard.
2. **Analysis**: Engine processes USNs through WCAG rules.
3. **Storage**: `cli` calls `engine.Repository.SaveReport()` to persist results.
4. **Display**: TUI queries the repository to render historical charts/lists.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `engine/core/ports/repository.go` | Create | Interface for analysis history storage. |
| `engine/internal/persistence/sqlite/` | Create | SQLite implementation of the Repository. |
| `cli/main.go` | Modify | Refactor to Bubbletea `tea.Program`. |
| `cli/internal/tui/` | Create | Bubbletea models for Dashboard and Analysis views. |
| `mcp/internal/registration/` | Create | Helper logic for detecting and updating MCP configs. |
| `install.sh` | Create | POSIX script for Linux/macOS installation. |
| `install.ps1` | Create | PowerShell script for Windows installation. |
| `.goreleaser.yaml` | Create | Release automation config. |

## Interfaces / Contracts

```go
// engine/core/ports/repository.go
type Repository interface {
    SaveReport(ctx context.Context, report domain.ViolationReport) error
    GetHistory(ctx context.Context, limit int) ([]domain.ViolationReport, error)
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | SQLite Repository | Use `go test` with a temporary file DB. |
| Unit | MCP Registration Logic | Mock filesystem to verify JSON injection. |
| Integration | TUI Initialization | Smoke test the Bubbletea model. |

## Migration / Rollout
No migration required for the initial release. Future updates will use an `internal/persistence/migrations` package.
