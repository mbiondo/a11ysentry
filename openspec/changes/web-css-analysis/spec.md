# Specification: Web CSS Analysis

## Purpose
Enable the Web Adapter to detect accessibility-relevant traits (color, background-color) from internal CSS blocks.

## Requirements

### Requirement: Internal Style Ingestion
The system MUST parse text content from `<style>` tags during the ingestion of HTML files.

#### Scenario: Extracting Color from Class
- GIVEN an HTML file with `<style>.warning { color: #FF0000; }</style>`
- WHEN the Web Adapter ingests the file
- THEN it MUST identify that the class `.warning` has the trait `color: #FF0000`.

### Requirement: Trait Merging via Class Attribute
The system MUST map properties from identified CSS classes to the `Traits` of a `USN` node if that node's `class` attribute contains the class name.

#### Scenario: Applying Class Style to Node
- GIVEN a parsed CSS map containing `.btn { background-color: #000; }`
- AND an HTML element `<button class="primary btn">`
- WHEN the element is converted to a `USN`
- THEN the `USN.Traits` MUST contain `"background-color": "#000"`.

### Requirement: Conflict Resolution (Specificity)
The system SHALL prioritize inline styles over CSS classes when a property is defined in both places.

#### Scenario: Inline Style Overrides Class
- GIVEN a class `.text { color: red; }`
- AND an element `<p class="text" style="color: blue;">`
- WHEN the `USN` is generated
- THEN the `Traits["color"]` MUST be `blue`.
