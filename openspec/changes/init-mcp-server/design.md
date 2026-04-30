# Design: Initialize MCP Server Module

## Technical Approach
We will implement a Model Context Protocol (MCP) server in the `mcp/` module using the `mark3labs/mcp-go` SDK. The server will provide a bridge between AI agents and the A11ySentry core engine. It will expose tools that trigger the ingestion and analysis pipeline.

## Architecture Decisions

### Decision: Stdio Transport
**Choice**: Use `stdio` for protocol communication.
**Alternatives considered**: SSE (Server-Sent Events).
**Rationale**: `stdio` is the standard transport for local development tools used by agents like Claude Desktop and Gemini. It is simpler to implement and deploy for the initial release.

### Decision: Tool Granularity
**Choice**: Expose a single `analyze_accessibility` tool initially.
**Alternatives considered**: Separate tools for parsing, normalization, and analysis.
**Rationale**: AI agents benefit from high-level, atomic actions. A single tool that takes a path and returns a full report is more efficient for agent usage than making multiple calls.

## Data Flow

    Agent (Client) ──[JSON-RPC/Stdio]──→ MCP Server
                                           │
    Violations ←──[JSON-RPC/Stdio]─── Handler Tool (analyze_accessibility)
                                           │
                                     ┌─────┴─────┐
                                     ▼           ▼
                                Web Adapter    Core Engine
                                (Normalization) (Analysis)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `mcp/main.go` | Create | MCP server entry point and tool registrations. |

## Interfaces / Contracts

### MCP Tool: `analyze_accessibility`
**Arguments**:
```json
{
  "path": {
    "type": "string",
    "description": "The absolute or relative path to the HTML file to analyze."
  }
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Tool handler logic | Mock the engine and adapter to verify handler calls. |
| Integration | Handshake and Discovery | Use an MCP client/inspector to verify server starts and lists tools. |

## Migration / Rollout
No migration required. This is a new module.

## Open Questions
- None.
