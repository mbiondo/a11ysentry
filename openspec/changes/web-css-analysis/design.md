# Design: Web CSS Analysis

## Technical Approach
We will implement a two-pass ingestion strategy in the Web Adapter.
1. **Pass 1 (CSS Extraction)**: Pre-scan the document for `<style>` nodes and build a class-to-property map.
2. **Pass 2 (DOM Traversal)**: Standard traversal where nodes look up their classes in the map created in Pass 1.

## Architecture Decisions

### Decision: Regex-based CSS Tokenizer
**Choice**: Use `regexp` to find `.classname { prop: val; }` patterns.
**Rationale**: Keeps the adapter zero-dependency and lightweight. Since we only care about `color` and `background-color` for accessibility, a full CSS parser is overkill for Phase 1.

### Decision: Class Matching
**Choice**: Split the `class` attribute by whitespace and iterate through classes to find matches in the CSS map.
**Rationale**: Handles multiple classes (`class="btn primary"`) correctly.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `adapters/web/adapter.go` | Modify | Add `extractCSS` method and update `traverse` to merge traits. |
| `adapters/web/adapter_test.go` | Modify | Add table-driven tests for CSS class mapping. |

## Testing Strategy
We will add a new test case to `adapter_test.go` that provides an HTML string with a `<style>` block and verifies the resulting `USN` nodes have the expected traits.
