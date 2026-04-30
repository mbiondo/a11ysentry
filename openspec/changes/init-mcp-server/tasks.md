# Tasks: Initialize MCP Server Module

## Phase 1: Foundation & Infrastructure

- [ ] 1.1 Install MCP SDK `go get github.com/mark3labs/mcp-go` in the `mcp` directory
- [ ] 1.2 Run `go mod tidy` in all modules to ensure dependencies are resolved

## Phase 2: Core Server Implementation

- [ ] 2.1 Implement MCP server boilerplate in `mcp/main.go` using `mcp.NewServer`
- [ ] 2.2 Define tool metadata for `analyze_accessibility` (parameters schema)
- [ ] 2.3 Implement the JSON-RPC tool handler function

## Phase 3: Engine Integration

- [ ] 3.1 Wire `adapters/web` into the tool handler to ingest HTML files
- [ ] 3.2 Map Core Engine violations to MCP `TextContent` responses
- [ ] 3.3 Ensure errors (file not found, parse error) are returned gracefully

## Phase 4: Testing & Verification

- [ ] 4.1 Build the MCP server `go build -o a11ysentry-mcp.exe ./mcp/main.go`
- [ ] 4.2 Verify handshake and tool discovery via manual stdio test
- [ ] 4.3 Run analysis on `landing/index.html` via MCP tool and verify output
