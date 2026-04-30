## Exploration: init-core-engine

### Current State
The project is at an "Architectural Root" stage. Only the `AGENT.md` file exists, defining the theoretical architecture (Universal Semantic Node, Adapter-Hub Pattern), along with a static `landing/index.html` file. No Go module has been initialized, and there is no source code structure.

### Affected Areas
- `root/` — Go module initialization (`go mod init`).
- `internal/core/` — Definition of core models (`USN`, `SemanticRole`, etc.).
- `internal/adapter/` — Base structure for platform adapters.
- `cmd/A11ySentry/` — CLI/Server entry point.

### Approaches
1. **Clean Architecture / Hexagonal** — Structure the project with clear separation between the Domain (USN), Ports (Ingestion/Analysis Interfaces), and Adapters (Web, Android, iOS).
   - Pros: Scalability, testability (crucial for TDD), alignment with AGENT.md.
   - Cons: Higher number of initial files.
   - Effort: Medium

2. **Flat Structure** — Put everything in the `main` package or a single `core` package for rapid prototyping.
   - Pros: Extreme initial speed.
   - Cons: Hard to maintain as complex adapters (Tree-sitter, etc.) are added, breaks the defined "Adapter-Hub" pattern.
   - Effort: Low

### Recommendation
I recommend **Approach 1 (Clean Architecture)**. Since `AGENT.md` already defines a robust and deterministic design, starting with a professional structure from day 1 will facilitate the implementation of WCAG 2.2 validators without generating immediate technical debt.

### Risks
- Initial over-engineering for a project that has no code yet.
- Complexity in managing AST dependencies (Tree-sitter) on Windows.

### Ready for Proposal
Yes. We are ready to propose the base Go infrastructure and USN models.

