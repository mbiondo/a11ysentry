package domain

import (
	"context"
)

// Severity represents whether a violation is definitive or context-dependent.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Violation represents an accessibility issue found by the engine.
type Violation struct {
	ErrorCode        string
	Severity         Severity
	Message          string
	SourceRef        Source
	FixSnippet       string
	DocumentationURL string
}

// ViolationReport aggregates all violations for a single analysis session.
type ViolationReport struct {
	ID          int64
	RunID       string    // Unique identifier for the analysis session
	ProjectName string
	ProjectRoot string
	FilePath    string
	Platform    Platform
	Timestamp   int64 // Unix timestamp
	Violations  []Violation
	Hierarchy   *FileNode // Root of the analysis tree for this report
}

// Analyzer is the core logic that takes USN nodes and returns violations.
type Analyzer interface {
	Analyze(ctx context.Context, nodes []USN, cfg ProjectConfig) ([]Violation, error)
}
