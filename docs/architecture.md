# Architecture: The Adapter-Hub Pattern

A11ySentry follows the **Adapter-Hub pattern**. This architectural decision ensures that the core validation engine remains pure and platform-agnostic.

## 1. Ingestion (Adapters)
Adapters are responsible for:
- Reading source files (`.html`, `.kt`, `.swift`).
- Identifying semantic elements.
- Generating **Universal Semantic Nodes (USN)**.

## 2. Normalization (USN)
The USN is the "Lingua Franca" of A11ySentry. It translates framework-specific attributes (like `className` in React or `contentDescription` in Compose) into a single standard schema.

## 3. Analysis (Core Engine)
The engine only speaks USN. It iterates through the tree and applies WCAG 2.2 Success Criteria.

## 4. Emission
Violations are reported via the **MCP Server** or **CLI**, providing exact source pointers for the developer to fix.
