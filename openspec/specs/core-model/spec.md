# Core Model Specification

## Purpose
Define the Universal Semantic Node (USN) schema and supporting data types as the foundational data model for A11ySentry.

## Requirements

### Requirement: USN Schema Definition
The system MUST provide a `USN` (Universal Semantic Node) data structure that captures the essential semantic properties of a UI element in a platform-agnostic way.

#### Scenario: USN Structure Completeness
- GIVEN a need to represent a UI element
- WHEN a `USN` instance is created
- THEN it MUST include fields for: `UID`, `Role`, `Label`, `State`, `Traits`, `Geometry`, `Hierarchy`, and `Source`.

### Requirement: Semantic Role Enumeration
The system SHALL define a restricted set of `SemanticRole` values to ensure consistent mapping across different UI frameworks.

#### Scenario: Validating Semantic Roles
- GIVEN a `USN` instance
- WHEN the `Role` field is assigned
- THEN it MUST be one of the pre-defined roles such as `button`, `heading`, `link`, `input`, `image`, `live-region`, or `modal`.

### Requirement: Platform Source Tracking
Every `USN` instance MUST track its origin platform to allow for platform-specific validation logic.

#### Scenario: Identifying Source Platform
- GIVEN a `USN` node generated from an AST
- WHEN the `Source.Platform` field is accessed
- THEN it MUST report one of the supported platforms (e.g., `WEB_REACT`, `ANDROID_COMPOSE`, `IOS_SWIFTUI`).

