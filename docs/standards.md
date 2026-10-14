# A11ySentry Standards

Project-wide rules an agent or contributor must honor when changing this codebase.

## Universal Semantic Node (USN)

All platform-specific UI elements MUST be mapped to the USN schema, so the engine can apply one
rule set across Web, Mobile, Desktop, and Gaming. A USN node must carry:

- `Role` — normalized semantic role.
- `Label` — accessible name.
- `Traits` — platform attributes (styles, ARIA, props).
- `Source` — accurate file, line, and column.

Adding a platform means adding an adapter that produces USN nodes. It never means adding a rule
that only understands one platform's markup.

## Compliance

Validators MUST follow WCAG 2.2 / Section 508, and the logic MUST be deterministic and binary
(pass/fail) from engineering data — never from LLM inference.

- **Hierarchy Rule**: heading levels must not skip.
- **Contrast Rule**: minimum 4.5:1 for text, 3:1 for UI components.
- **Context Rule**: accessibility states (hidden, disabled) must propagate recursively down the
  component tree.

## Testing

- Every rule must have a corresponding test case in `analyzer_test.go`.
- Every adapter must demonstrate ingestion in its respective `adapter_test.go`.

## Multi-Platform

- Do not add HTML-specific rules to the global analyzer unless scoped to `PlatformWeb`.

## AI Integration

- AI agents are consumers of the MCP server.
- Use the `FixSnippet` field to provide agents with ready-to-paste fixes.

## Agent-facing assets

- The agent skill has exactly one source of truth:
  `mcp/registration/skills/a11ysentry-mcp/SKILL.md`. It is embedded with `go:embed` and installed
  by `a11ysentry mcp --install-skills`. Never copy its body elsewhere.
