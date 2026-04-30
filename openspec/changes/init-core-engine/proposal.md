# Proposal: Initialize Core Engine Infrastructure

## Intent
Establish the technical foundation of A11ySentry following `AGENT.md`. We need to transform the theoretical architecture into a real Go code structure to begin implementing the Universal Semantic Node (USN) and WCAG validators.


## Scope

### In Scope
- Go module initialization (`go mod init A11ySentry`).
- Clean Architecture directory structure (`cmd/`, `internal/core`, `internal/adapter`).
- Definition of the `USN` struct and related types (`SemanticRole`, `Platform`) in the domain.
- Boilerplate for the application entry point.

### Out of Scope
- Implementation of real adapters (Tree-sitter, etc.).
- WCAG validation logic (to be done in subsequent changes).
- Complete MCP/gRPC interface.

## Capabilities

### New Capabilities
- `core-model`: Definition of the Universal Semantic Node (USN) and base data schema.
- `engine-pipeline`: Basic structure of the Ingestion -> Normalization -> Analysis -> Emission pipeline.

### Modified Capabilities
- None

## Approach
We will use **Clean Architecture**. The domain (`internal/core/domain`) will contain the `USN` without external dependencies. The pipeline will be defined through interfaces in `internal/core/ports`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `root/` | New | `go.mod` and base structure. |
| `internal/core/domain` | New | Definition of `USN` and types. |
| `internal/core/ports` | New | Pipeline interfaces. |
| `cmd/A11ySentry` | New | CLI entry point. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Inconsistency with AGENT.md | Low | Continuous validation against the original spec. |
| Type complexity in Go | Low | Keep USN simple and extensible using `map[string]any` for Traits. |

## Rollback Plan
Delete the created files and the `go.mod` file. As this is an initialization, the rollback impact is minimal.

## Dependencies
- Go 1.24+

## Success Criteria
- [x] `go mod init` successfully completed.
- [x] `USN` struct defined and compiling.
- [x] Folder structure created according to the proposed design.

