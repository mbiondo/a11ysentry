## Exploration: init-mcp-server

### Current State
A11ySentry is structured as a Go Workspace monorepo. We have the `engine` (domain/ports), `adapters/web` (HTML parsing), and `cli` (standalone tool). The `mcp` module is currently an empty Go module (`go mod init a11ysentry/mcp`).

### Affected Areas
- `mcp/` — Implementation of the MCP server.
- `engine/core/ports` — Potential need for a unified "Service" or "Use Case" to simplify calling the engine from multiple interfaces (CLI and MCP).

### Approaches
1. **MCP Golang SDK (mark3labs/mcp-go)** — Use a community-standard SDK to implement the Model Context Protocol.
   - Pros: Handles JSON-RPC, stdio/SSE transport, and protocol handshake automatically. Consistent with the Go ecosystem.
   - Cons: External dependency.
   - Effort: Medium

2. **Custom JSON-RPC Implementation** — Build the protocol handling from scratch over stdio.
   - Pros: Zero external dependencies.
   - Cons: High effort to maintain compliance with the MCP spec, prone to bugs in handshake/lifecycle management.
   - Effort: High

### Recommendation
I recommend **Approach 1 (MCP Golang SDK)**. Using `mark3labs/mcp-go` will allow us to focus on the accessibility logic rather than the plumbing of the protocol. It is well-documented and widely used for Go-based MCP servers.

### Risks
- SDK breaking changes (MCP is a relatively new protocol).
- Complexity in bridging the Go monorepo imports within the MCP server.

### Ready for Proposal
Yes. We are ready to define the `analyze-accessibility` tool in an MCP server.
