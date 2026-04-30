# Delta for MCP Interface

## MODIFIED Requirements

### Requirement: Protocol Compliance
The system MUST implement the Model Context Protocol (MCP) version 1.0.0 or higher over `stdio` transport. The binary MUST be capable of identifying itself and its registration status when queried by installers.
(Previously: The system MUST implement the Model Context Protocol (MCP) version 1.0.0 or higher over `stdio` transport.)

#### Scenario: Server Handshake
- GIVEN an MCP client starts the server
- WHEN the client sends an `initialize` request
- THEN the server MUST respond with its capabilities and server information.

#### Scenario: Registration Status Check
- GIVEN the A11ySentry binary
- WHEN executed with a `--check-mcp` flag
- THEN it MUST output the registration status for supported platforms.
