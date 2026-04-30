package sqlite

import (
	"a11ysentry/engine/core/domain"
	"context"
	"os"
	"testing"
)

func TestSQLiteRepository(t *testing.T) {
	dbPath := "test_history.db"
	// Always clean up before and after to ensure isolation.
	os.Remove(dbPath)
	t.Cleanup(func() { os.Remove(dbPath) })

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Test Save
	report := domain.ViolationReport{
		FilePath: "test.html",
		Platform: domain.PlatformWebReact,
		Violations: []domain.Violation{
			{ErrorCode: "WCAG_1_1_1", Severity: domain.SeverityError, Message: "Missing alt"},
		},
	}

	if err := repo.SaveReport(ctx, report); err != nil {
		t.Fatalf("Failed to save report: %v", err)
	}

	// Test History
	history, err := repo.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 report in history, got %d", len(history))
	}

	if history[0].FilePath != "test.html" {
		t.Errorf("Expected file path 'test.html', got '%s'", history[0].FilePath)
	}

	if len(history[0].Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(history[0].Violations))
	}
}
