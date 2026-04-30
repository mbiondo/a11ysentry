# persistence-layer Specification

## Purpose
Ensure long-term durability of analysis reports and configuration through a lightweight local database.

## Requirements

### Requirement: Analysis Persistence
The system MUST persist every completed analysis report to a local SQLite database.

#### Scenario: Save Report
- GIVEN a completed accessibility audit
- WHEN the emission phase finishes
- THEN the system MUST store the file metadata, violation count, and timestamp in the database.

### Requirement: Cross-Session Recovery
The system MUST be able to retrieve historical reports across different CLI sessions.

#### Scenario: History Retrieval
- GIVEN the CLI is restarted
- WHEN the dashboard is requested
- THEN the system MUST load and display all reports previously saved to the database.

### Requirement: Database Migration
The system SHOULD handle automatic schema migrations when the data model evolves.

#### Scenario: Schema Update
- GIVEN an existing database from an older version
- WHEN a newer version of A11ySentry starts
- THEN the system SHOULD apply pending migrations without data loss.
