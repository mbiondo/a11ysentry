# Design: Initialize Core Engine Infrastructure

## Technical Approach
We are implementing a Clean Architecture structure in Go 1.24+. The core domain will be isolated from external concerns. The pipeline will be interface-based to allow for stateless, deterministic validation as defined in `AGENT.md`.

## Architecture Decisions

### Decision: Project Structure (Clean Architecture)
**Choice**: Standard Go project layout (`cmd/`, `internal/`, `pkg/`).
**Alternatives considered**: Flat structure.
**Rationale**: Enables clear separation of concerns, crucial for the "Adapter-Hub" pattern. Domain logic lives in `internal/core/domain`.

### Decision: USN Representation
**Choice**: Go `struct` with explicit fields for core metadata and `map[string]any` for flexible `Traits`.
**Alternatives considered**: Using a dynamic JSON-like map for everything.
**Rationale**: Provides type safety for required fields (UID, Role, Label) while allowing platform-specific traits to be extensible.

## Data Flow
The data flow follows the deterministic pipeline stages:
1. **Ingestion**: `PlatformAdapter` reads source code.
2. **Normalization**: Adapter maps source nodes to `USN`.
3. **Analysis**: `RuleEngine` traverses the `USN` tree and applies WCAG rules.
4. **Emission**: `Reporter` outputs violations.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Create | Go module definition. |
| `cmd/A11ySentry/main.go` | Create | Application entry point. |
| `internal/core/domain/usn.go` | Create | `USN` struct and core types (`SemanticRole`, `Platform`). |
| `internal/core/ports/pipeline.go` | Create | Interfaces for `Adapter`, `Analyzer`, and `Emitter`. |

## Interfaces / Contracts

```go
package domain

type SemanticRole string

const (
    RoleButton      SemanticRole = "button"
    RoleLink        SemanticRole = "link"
    RoleHeading     SemanticRole = "heading"
    // ... rest of roles
)

type USN struct {
    UID       string
    Role      SemanticRole
    Label     string
    State     USNState
    Traits    map[string]any
    Geometry  Geometry
    Hierarchy Hierarchy
    Source    Source
}

type USNState struct {
    Disabled bool
    Hidden   bool
    Selected bool
    Expanded bool
    Invalid  bool
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | USN creation and role validation | Standard `go test` with table-driven tests. |
| Integration | Module initialization | Verify `go.mod` and imports are valid. |

## Migration / Rollout
No migration required. This is a fresh project initialization.

## Open Questions
- None.

