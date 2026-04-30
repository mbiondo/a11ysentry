# cli-dashboard Specification

## Purpose
Provide a rich, interactive terminal experience for managing accessibility audits and history.

## Requirements

### Requirement: Interactive Dashboard
The system MUST provide an interactive dashboard view that displays the history of analyzed projects and files.

#### Scenario: Dashboard Initialization
- GIVEN the user starts the CLI with the `dashboard` command
- WHEN the application initializes
- THEN the system MUST render a Bubbletea-based view showing the analysis history.

### Requirement: Violation Detail View
The system MUST provide a detailed view of current analysis violations with syntax-highlighted source snippets.

#### Scenario: Browsing Violations
- GIVEN a list of violations in the TUI
- WHEN the user selects a specific violation
- THEN the system MUST display the exact file path, line/column, and the code snippet causing the issue.

### Requirement: History Search
The system SHOULD allow the user to search through the analysis history.

#### Scenario: Filter by File
- GIVEN the dashboard view
- WHEN the user types a search query
- THEN the history list MUST be filtered to match the file name.
