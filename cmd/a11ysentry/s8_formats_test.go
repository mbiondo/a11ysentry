package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"a11ysentry/cmd/a11ysentry/internal/sarif"
	"a11ysentry/engine/core/domain"
)

func s8Violation(code, file string, line, col int, sev domain.Severity) domain.Violation {
	return domain.Violation{
		ErrorCode: code,
		Severity:  sev,
		Message:   code + " message",
		SourceRef: domain.Source{FilePath: file, Line: line, Column: col},
	}
}

// s8Reports models the shape S8 is about: one shared component ingested into three page trees
// (the same violation repeated per tree), two distinct violations inside a single tree, and one
// tree with no violations.
func s8Reports() []domain.ViolationReport {
	shared := s8Violation("WCAG_1_1_1", "components/Shared.tsx", 2, 10, domain.SeverityError)
	own1 := s8Violation("WCAG_1_1_1", "app/page.tsx", 6, 7, domain.SeverityError)
	own2 := s8Violation("G141", "app/page.tsx", 1, 1, domain.SeverityWarning)
	return []domain.ViolationReport{
		{FilePath: "app/page.tsx", Violations: []domain.Violation{own1, own2, shared}},
		{FilePath: "app/about/page.tsx", Violations: []domain.Violation{shared}},
		{FilePath: "app/contact/page.tsx", Violations: []domain.Violation{shared}},
		{FilePath: "app/clean/page.tsx", Violations: nil},
	}
}

func renderS8(t *testing.T, reports []domain.ViolationReport, format string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out."+format)
	printReports(reports, format, ".", out, false)
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("printReports(%s) wrote no output: %v", format, err)
	}
	return string(b)
}

// S8: every format reports a shared component once. text, json and sarif must agree on the number
// of distinct problems, and the json pageCount values must sum to the un-deduplicated total the
// old per-tree shape carried.
func TestS8AllFormatsAgreeOnDistinctCount(t *testing.T) {
	reports := s8Reports()

	var unique []domain.UniqueViolation
	if err := json.Unmarshal([]byte(renderS8(t, reports, "json")), &unique); err != nil {
		t.Fatalf("json output is not a list of unique violations: %v", err)
	}
	if len(unique) != 3 {
		t.Errorf("json: got %d entries, want 3 distinct problems", len(unique))
	}
	raw := 0
	for _, r := range reports {
		raw += len(r.Violations)
	}
	sum := 0
	for _, u := range unique {
		sum += u.PageCount
	}
	if sum != raw {
		t.Errorf("json: sum(pageCount) = %d, want the raw per-tree total %d", sum, raw)
	}

	var log sarif.Log
	if err := json.Unmarshal([]byte(renderS8(t, reports, "sarif")), &log); err != nil {
		t.Fatalf("sarif output is not valid SARIF: %v", err)
	}
	if got := len(log.Runs[0].Results); got != len(unique) {
		t.Errorf("sarif: got %d results, want %d (the json distinct count)", got, len(unique))
	}
	if got := len(log.Runs[0].Tool.Driver.Rules); got != 2 {
		t.Errorf("sarif: got %d rules, want 2", got)
	}

	text := renderS8(t, reports, "text")
	if !strings.Contains(text, "Summary: 2 error(s), 1 warning(s)") {
		t.Errorf("text summary does not match the 3-entry unique list:\n%s", text)
	}
	if n := strings.Count(text, "Shared.tsx:2:10"); n != 1 {
		t.Errorf("text repeated the shared component %d times, want 1:\n%s", n, text)
	}
}

// S8 edge case: an empty report list must still produce valid, empty output for every format.
// A scan with no violations must emit `results` as an empty array: GitHub Code Scanning rejects
// `"results": null` as invalid SARIF ("is not of a type(s) array").
func TestS8EmptyReportList(t *testing.T) {
	if got := renderS8(t, nil, "json"); got != "[]" {
		t.Errorf("empty json = %q, want []", got)
	}

	rawSARIF := renderS8(t, nil, "sarif")
	if strings.Contains(rawSARIF, `"results": null`) {
		t.Errorf("empty sarif emitted results as null; Code Scanning rejects that:\n%s", rawSARIF)
	}
	var log sarif.Log
	if err := json.Unmarshal([]byte(rawSARIF), &log); err != nil {
		t.Fatalf("empty sarif is not valid SARIF: %v", err)
	}
	if log.Runs[0].Results == nil {
		t.Errorf("empty sarif results must be an empty array, not null:\n%s", rawSARIF)
	}
	if len(log.Runs[0].Results) != 0 || len(log.Runs[0].Tool.Driver.Rules) != 0 {
		t.Errorf("empty sarif should carry no results and no rules")
	}
}

// S8 edge case: a violation with no source file must not panic or be dropped.
func TestS8ViolationWithEmptySourceFilePath(t *testing.T) {
	reports := []domain.ViolationReport{{
		FilePath:   "p",
		Violations: []domain.Violation{s8Violation("R", "", 0, 0, domain.SeverityError)},
	}}
	var unique []domain.UniqueViolation
	if err := json.Unmarshal([]byte(renderS8(t, reports, "json")), &unique); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	if len(unique) != 1 {
		t.Fatalf("got %d entries, want 1", len(unique))
	}
}

// S8 edge case: a single tree with many violations keeps every one of them.
func TestS8SingleTreeManyViolations(t *testing.T) {
	vs := make([]domain.Violation, 0, 50)
	for i := 1; i <= 50; i++ {
		vs = append(vs, s8Violation("R", "a.tsx", i, 1, domain.SeverityError))
	}
	var unique []domain.UniqueViolation
	if err := json.Unmarshal([]byte(renderS8(t, []domain.ViolationReport{{FilePath: "p", Violations: vs}}, "json")), &unique); err != nil {
		t.Fatal(err)
	}
	if len(unique) != 50 {
		t.Errorf("got %d entries, want 50", len(unique))
	}
}
