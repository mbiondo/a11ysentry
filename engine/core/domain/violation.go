package domain

import (
	"context"
	"fmt"
	"sort"
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
	ID            int64
	RunID         string // Unique identifier for the analysis session
	ProjectName   string
	ProjectRoot   string
	FilePath      string
	Platform      Platform
	Timestamp     int64 // Unix timestamp
	Violations    []Violation
	Hierarchy     *FileNode // Root of the analysis tree for this report
	OpacityMetric float64   // Percentage of opaque nodes in the tree (0.0 to 100.0)
}

// Analyzer is the core logic that takes USN nodes and returns violations.
type Analyzer interface {
	Analyze(ctx context.Context, nodes []USN, cfg ProjectConfig) ([]Violation, error)
}

// UniqueViolation wraps a Violation with cross-tree deduplication metadata.
type UniqueViolation struct {
	Violation Violation
	PageCount int // How many PageTrees contained this exact violation
}

// violationKey returns a stable string key identifying a violation by rule + location.
// Two violations with the same key are considered the same issue regardless of which page tree surfaced them.
func violationKey(v Violation) string {
	return fmt.Sprintf("%s|%s|%d|%d", v.ErrorCode, v.SourceRef.FilePath, v.SourceRef.Line, v.SourceRef.Column)
}

// DeduplicateCrossTree collapses violations that appear in multiple ViolationReports
// (e.g. a shared component included in 19 pages) into a single UniqueViolation,
// tracking how many pages surfaced each one.
//
// The result is sorted by: severity (errors first), then source file path, then line, then column.
func DeduplicateCrossTree(reports []ViolationReport) []UniqueViolation {
	type entry struct {
		violation Violation
		count     int
	}
	seen := make(map[string]*entry)
	// Preserve insertion order for stable output
	var order []string

	for _, r := range reports {
		for _, v := range r.Violations {
			k := violationKey(v)
			if e, ok := seen[k]; ok {
				e.count++
			} else {
				seen[k] = &entry{violation: v, count: 1}
				order = append(order, k)
			}
		}
	}

	result := make([]UniqueViolation, 0, len(order))
	for _, k := range order {
		e := seen[k]
		result = append(result, UniqueViolation{
			Violation: e.violation,
			PageCount: e.count,
		})
	}

	// Sort: errors first, then by file path, then by line, then column
	sort.SliceStable(result, func(i, j int) bool {
		vi, vj := result[i].Violation, result[j].Violation
		if vi.Severity != vj.Severity {
			return vi.Severity == SeverityError
		}
		if vi.SourceRef.FilePath != vj.SourceRef.FilePath {
			return vi.SourceRef.FilePath < vj.SourceRef.FilePath
		}
		if vi.SourceRef.Line != vj.SourceRef.Line {
			return vi.SourceRef.Line < vj.SourceRef.Line
		}
		return vi.SourceRef.Column < vj.SourceRef.Column
	})

	return result
}
