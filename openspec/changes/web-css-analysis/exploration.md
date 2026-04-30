# Exploration: Web CSS Analysis

## Context
Currently, the Web Adapter (`adapters/web/adapter.go`) only captures `color` and `background-color` traits if they are defined inline via the `style` attribute. This is a severe limitation because modern web development relies heavily on CSS classes (e.g., `<style>` blocks, external stylesheets, or utility classes). As a result, the engine misses many potential Color Contrast (WCAG 1.4.3) violations.

## Options Considered

### Option 1: Full CSSOM Parser (e.g., `github.com/tdewolff/parse/v2`)
- **Pros**: Robust, handles complex selectors, media queries, and variables.
- **Cons**: Adds a heavy dependency. Complex to map specific computed styles to the DOM tree without a full layout engine.

### Option 2: Basic Regex/Tokenizer for `<style>` blocks (Chosen)
- **Pros**: No external dependencies. Can easily target simple `.class` selectors which cover 80% of use cases for simple HTML templates and components.
- **Cons**: Won't handle complex inheritance or specificity rules accurately.

## Technical Feasibility
We already use `golang.org/x/net/html`. We can find all `<style>` nodes, extract their text content, and parse blocks like `.my-class { color: #fff; background-color: #000; }`. 
Then, when traversing the DOM to create `USN` nodes, we can split the `class` attribute, look up the extracted CSS properties, and inject them into `USN.Traits`. 
This allows the existing `Analyzer` (which already looks at `Traits["color"]`) to seamlessly validate contrast without any changes to the core domain!

## Conclusion
We will implement Option 2 as a first iteration to drastically improve the detection rate of contrast violations on the web.