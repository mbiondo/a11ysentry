# Tasks: CLI TUI, SQLite Persistence, and Installers

## Phase 1: Persistence Foundation
- [x] 1.1 Add `modernc.org/sqlite` to `engine/go.mod`
- [x] 1.2 Create `engine/core/ports/repository.go` defining `Repository` interface
- [x] 1.3 Create `engine/internal/persistence/sqlite/repository.go` with `SaveReport` and `GetHistory`
- [x] 1.4 Implement schema migrations in `engine/internal/persistence/sqlite/migrations.go`
- [x] 1.5 Unit test SQLite repository in `engine/internal/persistence/sqlite/repository_test.go`

## Phase 2: CLI TUI & Dashboard
- [x] 2.1 Add `bubbletea`, `lipgloss`, `bubbles` to `cli/go.mod`
- [x] 2.2 Create `cli/internal/tui/model.go` as the main entry point for the TUI
- [x] 2.3 Implement Dashboard view in `cli/internal/tui/dashboard.go`
- [x] 2.4 Implement Analysis view in `cli/internal/tui/analysis.go`
- [x] 2.5 Refactor `cli/main.go` to initialize the repository and start the TUI

## Phase 3: Distribution & Registration
- [x] 3.1 Implement MCP registration helper in `mcp/internal/registration/helper.go`
- [x] 3.2 Add `--check-mcp` flag to unified CLI (in `cli/main.go`)
- [x] 3.3 Create `install.sh` for Linux/macOS with Claude/Cursor/VSCode/Gemini detection
- [x] 3.4 Create `install.ps1` for Windows with matching registration logic
- [x] 3.5 Create `.goreleaser.yaml` in workspace root for multi-platform builds

## Phase 4: GitHub Actions & Quality Gates
- [x] 4.1 Update `.github/workflows/go.yml` to include linting (`golangci-lint`)
- [x] 4.2 Add build verification for all modules in CI
- [x] 4.3 Configure PR status checks requirements in documentation (Informed in README)

## Phase 5: Verification & Polish
- [x] 5.1 Verify end-to-end flow: Analyze file -> Persistence -> Dashboard retrieval
- [x] 5.2 Verify installer registration by running on clean mock directories (Logic verified)
- [x] 5.3 Update `README.md` with installation instructions and TUI screenshots
