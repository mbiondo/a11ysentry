package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"a11ysentry/cmd/a11ysentry/internal/sarif"
	"a11ysentry/engine/core/domain"
)

func sharedImageViolation() domain.Violation {
	return domain.Violation{
		ErrorCode: "WCAG_1_1_1",
		Severity:  domain.SeverityError,
		Message:   "Image missing alt",
		SourceRef: domain.Source{FilePath: "src/Shared.tsx", Line: 3, Column: 1},
	}
}

// sharedComponentReports mimics a shared component that was ingested into three page trees:
// the exact same problem is present once per tree.
func sharedComponentReports() []domain.ViolationReport {
	reports := make([]domain.ViolationReport, 0, 3)
	for _, page := range []string{"page-a", "page-b", "page-c"} {
		reports = append(reports, domain.ViolationReport{
			FilePath:   page,
			Violations: []domain.Violation{sharedImageViolation()},
		})
	}
	return reports
}

// A shared component is ingested into every page tree that imports it, so a single problem is
// reported once per page. Every output format must collapse that back to one entry.
func TestPrintReportsDeduplicatesSharedComponentAcrossTrees(t *testing.T) {
	reports := sharedComponentReports()
	dir := t.TempDir()

	jsonOut := filepath.Join(dir, "out.json")
	printReports(reports, "json", ".", jsonOut, false)

	rawJSON, err := os.ReadFile(jsonOut)
	if err != nil {
		t.Fatal(err)
	}

	var unique []domain.UniqueViolation
	if err := json.Unmarshal(rawJSON, &unique); err != nil {
		t.Fatalf("json output is not a list of unique violations: %v\n%s", err, rawJSON)
	}
	if len(unique) != 1 {
		t.Errorf("json: got %d entries for 1 distinct problem:\n%s", len(unique), rawJSON)
	} else if unique[0].PageCount != 3 {
		t.Errorf("json: PageCount = %d, want 3", unique[0].PageCount)
	}

	sarifOut := filepath.Join(dir, "out.sarif")
	printReports(reports, "sarif", ".", sarifOut, false)

	rawSARIF, err := os.ReadFile(sarifOut)
	if err != nil {
		t.Fatal(err)
	}

	var sarifLog sarif.Log
	if err := json.Unmarshal(rawSARIF, &sarifLog); err != nil {
		t.Fatalf("sarif output is not valid SARIF: %v\n%s", err, rawSARIF)
	}
	if got := len(sarifLog.Runs[0].Results); got != 1 {
		t.Errorf("sarif: got %d results for 1 distinct problem:\n%s", got, rawSARIF)
	}
}
