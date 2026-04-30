# Proposal: Web CSS Analysis (Phase 1: Internal Styles)

## Intent
Enhance the Web Adapter to extract accessibility-relevant CSS properties (color, background-color) from internal `<style>` blocks and map them to HTML elements via their `class` attributes. This will significantly increase the detection rate for WCAG 1.4.3 (Color Contrast) violations.

## Scope

### In Scope
- Parsing text content from `<style>` tags within the ingested HTML document.
- Extracting basic class selectors (`.classname`) and their `color` / `background-color` properties.
- Applying these extracted properties to the `Traits` map of the generated `USN` if the node has matching classes.

### Out of Scope
- Parsing external `.css` files via `<link rel="stylesheet">`.
- Complex CSS specificity (e.g., `#id .class > div`), media queries, or pseudo-classes (`:hover`).
- CSS Variables (`var(--primary-color)`).

## Capabilities

### Modified Capabilities
- `web-adapter`: Will now maintain a transient state of parsed CSS classes during the `Ingest` phase and map them to nodes.

## Approach
1. In `adapters/web/adapter.go`, before traversing the body, traverse the `<head>` (or whole doc) to find all `<style>` nodes.
2. Extract the text and use a simple string tokenizer/regex to build a `map[string]map[string]string` (e.g., `{"my-class": {"color": "#FFF"}}`).
3. During the standard `traverse`, when an element has `class="my-class"`, merge the mapped properties into the `USN.Traits`.
4. The `Analyzer` in `engine/core/domain` will automatically pick these up and validate contrast!

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `adapters/web/adapter.go` | Modify | Add CSS extraction logic and trait merging. |
| `adapters/web/adapter_test.go` | Modify | Add tests for internal stylesheet parsing. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CSS parsing bugs | Medium | Use robust regex; fallback gracefully if a block is malformed. Only target `#hex` or `rgb` values to start. |

## Dependencies
- `regexp` and `strings` standard libraries.

## Success Criteria
- [ ] A document with `<style>.btn { color: #555; background-color: #FFF; }</style>` and `<button class="btn">` correctly triggers a WCAG_1_4_3 violation.
- [ ] Existing tests continue to pass.