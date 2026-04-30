# Engine Pipeline Specification

## Purpose
Define the high-level architecture and interfaces for the stateless validation pipeline.

## Requirements

### Requirement: Deterministic Pipeline Flow
The core engine MUST implement a deterministic pipeline consisting of four distinct stages: Ingestion, Normalization, Analysis, and Emission.

#### Scenario: Pipeline Execution Sequence
- GIVEN a set of source files
- WHEN the pipeline is executed
- THEN it MUST first parse the source (Ingestion), then map to USN (Normalization), then apply rules (Analysis), and finally output results (Emission).

### Requirement: Stateless Analysis
The Analysis stage MUST be stateless, meaning the validation of a USN tree MUST NOT depend on the state of previous analyses.

#### Scenario: Reproducible Results
- GIVEN a specific `USN` tree
- WHEN the Analysis stage is run multiple times on the same tree
- THEN it MUST always return the exact same set of violations.

### Requirement: Platform-Agnostic Core
The core validation logic (Analysis) MUST operate exclusively on the `USN` structure, remaining agnostic to the original source framework.

#### Scenario: Universal Rule Application
- GIVEN a validation rule for "Alt text in images"
- WHEN applied to a `USN` tree
- THEN it MUST produce consistent violations regardless of whether the node originated from React, Compose, or SwiftUI.

