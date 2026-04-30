# Tasks: Initialize Core Engine Infrastructure (Monorepo)

## Phase 1: Foundation & Infrastructure (Monorepo)

- [x] 1.1 Initialize `go.work` and modules: `engine`, `cli`, `mcp`, `adapters/*` (a11ysentry namespace)
- [x] 1.2 Create directory structure following Go Workspace layout
- [x] 1.3 Implement `USN` and core models in `engine/core/domain`

## Phase 2: Core Interfaces (Ports)

- [x] 2.1 Create `engine/core/ports/pipeline.go` defining `Adapter`, `Analyzer`, and `Emitter` interfaces
- [x] 2.2 Define `Violation` and `ViolationReport` types in `engine/core/domain/violation.go`

## Phase 3: CLI Entry Point (CI/CD Guard)

- [x] 3.1 Create `cli/main.go` with basic flag parsing
- [x] 3.2 Implement basic execution flow in `cli` calling a POC HTML adapter

## Phase 4: Testing & Verification

- [x] 4.1 Unit tests for `USN` struct in `engine/core/domain`
- [x] 4.2 Verify basic HTML analysis with the CLI (manual/smoke test)
- [x] 4.3 Verify workspace-wide build with `go build ./...`

## Phase 5: Documentation & Cleanup

- [ ] 5.1 Add doc comments in `engine/core/domain`
- [ ] 5.2 Run `go fmt` and `go vet` across the workspace

