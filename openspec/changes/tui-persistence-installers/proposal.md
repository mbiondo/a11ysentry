# Proposal: CLI TUI, SQLite Persistence, and Installers

## Intent
Transform A11ySentry from a POC CLI into a production-ready tool with historical analysis tracking, an interactive TUI dashboard, and automated multi-platform distribution with zero-config MCP registration.

## Scope

### In Scope
- SQLite persistence (Pure Go) for analysis history and violation reports.
- Bubbletea-based TUI for interactive dashboard and analysis feedback.
- GoReleaser configuration for cross-platform binaries.
- Smart installers (`install.sh`, `install.ps1`) with automated MCP registration for major AI agents.

### Out of Scope
- Cloud sync for analysis history (SQLite is local only).
- GUI (Desktop app) — focus remains on CLI/TUI.

## Capabilities

### New Capabilities
- `cli-dashboard`: Interactive TUI for browsing history and active analysis.
- `persistence-layer`: SQLite schema and repository for analysis state.
- `automated-distribution`: Installers and release pipelines.

### Modified Capabilities
- `mcp-interface`: Add requirement for automated registration during installation.

## Approach
Implement a `Persistence` port in the engine and a SQLite adapter using `modernc.org/sqlite`. Refactor the `cli` module to use Bubbletea as its main execution mode. Create platform-specific scripts to automate the download and configuration of MCP settings.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `engine/` | Modified | Add repository port and SQLite adapter. |
| `cli/` | Modified | Migrate to Bubbletea TUI and add dashboard. |
| `mcp/` | Modified | Add registration helper logic. |
| `/` | New | Add `.goreleaser.yaml`, `install.sh`, `install.ps1`. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CGO issues with SQLite | Low | Using `modernc.org/sqlite` (Pure Go). |
| MCP Config Corruption | Med | Use atomic writes and robust JSON parsing/backups. |

## Rollback Plan
Binary removal and restoration of MCP config backups (`.bak` files).

## Success Criteria
- [ ] Binary builds for Win/Linux/Mac with `go work sync && go build`.
- [ ] Analysis history persists across CLI restarts.
- [ ] `install.sh` successfully registers MCP in Claude Desktop.
