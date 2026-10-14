# Pi agent support and agent-facing skill documentation

Objective: make A11ySentry installable and understandable by the `pi` agent runtime, and make the
agent-facing skill a complete, single-source description of the tool's real features.
Branch: `feat/pi-agent-support` · Delivery strategy: `single-pr` · Test runner: `go -C <module> test ./...`
(there is no root module; run per module, e.g. `go -C engine test ./...`)

## Specs

- **S1 — Pi MCP registration.** `a11ysentry mcp --register` must be able to register the MCP server in
  the `pi` agent runtime. Pi reads MCP servers from `~/.pi/agent/mcp.json` on every platform, using the
  standard `{"mcpServers": {"<name>": {"command": ..., "args": [...]}}}` shape. The user asked to
  "darle soporte a PI agents".
- **S2 — Pi skill installation.** The agent skill must be installed into the directories `pi` actually
  discovers: `.pi/skills`, `.agent/skills`, `.agents/skills` (project) and `~/.pi/agent/skills`,
  `~/.config/agents/skills`, `~/.agents/skills` (global). Today only `~/.gemini/skills` is written.
- **S3 — Single source of truth for the skill.** The skill body must exist in exactly one file, consumed
  by both the repository and the compiled binary, so the two copies can never diverge again. The file
  must be embedded with `//go:embed` from inside the `a11ysentry/mcp` module, because Go cannot embed
  across module boundaries.
- **S4 — The skill must describe the tool's real features.** Driven by the user's words: "mejorar la
  documentacion para los agentes, el skill digamos, asi sabe bien las features que tiene". The skill
  must document every tool the MCP server exposes, the CLI surface, the output formats, the exit codes,
  the `a11y-ignore` suppression system, the USN model, and `a11ysentry.json`, and its `allowed-tools`
  frontmatter must list only tools that actually exist.
- **S5 — Skill installation must not depend on a foreign path.** `install.sh` must stop reading the
  skill from a directory belonging to an unrelated project, and must install to the same agent
  directories on every platform.
- **S6 — Build instructions must be runnable.** A fresh clone must be able to build and test. The
  documentation must explain the workspace requirement and the correct per-module test command.

## Tasks

| ID | Specs | Route | Status | Evidence |
|----|-------|-------|--------|----------|
| T1 | — | inline | done | branch `feat/pi-agent-support`; this document |
| T2 | S3 | inline | done | skill at `mcp/registration/skills/a11ysentry-mcp/SKILL.md`, `//go:embed`; `const` deleted |
| T3 | S4 | inline | done | rewritten `SKILL.md` v2.0.0; false `allowed-tools` removed |
| T4 | S1, S2 | inline | done | `RegisterPi`, `GlobalSkillTargets`, `ProjectSkillTargets`, `--install-skills` |
| T5 | S5 | inline | done | `install.ps1`, `install.sh` delegate to the binary |
| T6 | S6 | inline | done | standards restored in `docs/standards.md`; build docs corrected |
| T7 | S1-S6 | delegated | pending | independent verifier over the whole feature |

## Log

- **L1** (user, 2026-10-14, verbatim): "Hola, como estas? podemos revisar el proyecto para darle
  soporte a PI agents y ademas para mejorar la documentacion para los agentes, el skill digamos, asi
  sabe bien las features que tiene... revisemos bien pro favor"
- **L2** (user decision, 2026-10-14): scope = "Pi + skill unificado" — Pi registration plus a complete,
  single-source skill; full documentation audit explicitly deferred.
- **L3** (user decision, 2026-10-14): skill deduplication strategy = "Mover + go:embed" — relocate the
  skill into the `a11ysentry/mcp` module and embed it.
- **L4** (explorer findings, 2026-10-14): `pi` MCP config is `~/.pi/agent/mcp.json` in the standard
  `mcpServers` shape, so the existing `patchConfig` already emits the right JSON. Pi skill discovery
  order: `.pi/skills`, `.agent/skills`, `.agents/skills`, then `~/.pi/agent/skills`,
  `~/.config/agents/skills`, `~/.agents/skills`.
- **L5** (explorer findings, 2026-10-14): `install.sh:103` hardcodes
  `$HOME/repositories/semantix/skills/a11ysentry-mcp`, a path belonging to an unrelated project, so the
  skill is never installed on Unix.
- **L6** (explorer findings, 2026-10-14): the repository has no root Go module. Local imports resolve
  only through a `go.work` that `.gitignore` excludes, so a fresh clone cannot build or test. Verified:
  `go build ./cmd/a11ysentry` exits 0 with a workspace and fails without one.
- **L7** (explorer findings, 2026-10-14): the uncommitted `.atl/skill-registry.md` deleted the project's
  "Project Standards" section (USN mapping, WCAG compliance, Hierarchy/Contrast/Context rules).
  Recoverable from `git show HEAD:.atl/skill-registry.md`.
- **L8** (explorer findings, 2026-10-14): `docs/MCP_INTEGRATION.md` states the server exposes "one
  tool"; it exposes three. The skill's `allowed-tools` lists `read_file`, `glob`, `grep_search`, none of
  which the server exposes.
- **L9** (explorer findings, 2026-10-14): the skill body existed in **four** places, drifted:
  `skills/a11ysentry-mcp/SKILL.md` (v1.3.0), `.agents/skills/a11ysentry-mcp/SKILL.md` (v1.2.0,
  tracked, and the path the registry pointed at), the `a11ysentrySkillContent` Go const (v1.3.0), and
  stale installed copies under the user profile. The two in-repo duplicates were deleted; the
  embedded v2.0.0 source supersedes all of them.
- **L10** (own checks, 2026-10-14): `go build ./cmd/a11ysentry` exit 0; `go vet` clean on
  `cmd/a11ysentry`, `engine`, `mcp`, `scanner`; `go -C mcp test ./registration/...` 8/8 pass;
  `go -C engine test ./...` and `go -C scanner test ./...` pass. New tests cover the Pi config path,
  the written JSON shape, preservation of existing MCP servers, idempotency, skill target discovery,
  the embedded skill's presence and documented tools, and installation into a directory.
- **L11** (own findings, 2026-10-14): `docs/DEVELOPER_GUIDE.md` told contributors to run
  `go test ./...` from the root, which fails; corrected to per-module invocations.
- **Next step**: independent verification of the whole feature against the branch point `5fd3fe3`.
