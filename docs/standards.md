# A11ySentry Standards

## Testing
- Every rule must have a corresponding test case in `analyzer_test.go`.
- Every adapter must demonstrate ingestion in its respective `adapter_test.go`.

## Multi-Platform
- Do not add HTML-specific rules to the global analyzer unless scoped to `PlatformWeb`.

## AI Integration
- AI Agents are consumers of the MCP Server.
- Use the `FixSnippet` field to provide agents with ready-to-paste fixes.
