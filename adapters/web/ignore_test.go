package web

import (
	"a11ysentry/engine/core/domain"
	"context"
	"os"
	"testing"
)

func TestHTMLAdapter_IgnoreInline(t *testing.T) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<body>
			<!-- a11y-ignore: WCAG_1_1_1 -->
			<img id="img1">
			
			<img id="img2"> <!-- This one should still fail -->
			
			<!-- a11y-ignore: all -->
			<button id="btn1"></button>
		</body>
		</html>
	`
	tmpFile, _ := os.CreateTemp("", "test_ignore_*.html")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = os.WriteFile(tmpFile.Name(), []byte(htmlContent), 0644)

	// We need to use the analyzer to see if filtering works
	adapter := NewHTMLAdapter()
	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	analyzer := domain.NewAnalyzer()
	violations, err := analyzer.Analyze(context.Background(), nodes, domain.DefaultConfig())
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Expecting:
	// 1. img2 missing alt (WCAG_1_1_1)
	// 2. Page missing H1 (G141)
	// 3. Page missing Lang (WCAG_3_1_1)
	
	foundImg1 := false
	foundImg2 := false
	foundBtn1 := false

	for _, v := range violations {
		if v.ErrorCode == "WCAG_1_1_1" {
			if v.SourceRef.RawHTML == "<img id=\"img1\">" {
				foundImg1 = true
			}
			if v.SourceRef.RawHTML == "<img id=\"img2\">" {
				foundImg2 = true
			}
		}
		if v.ErrorCode == "WCAG_4_1_2" && v.SourceRef.RawHTML == "<button id=\"btn1\">" {
			foundBtn1 = true
		}
	}

	if foundImg1 {
		t.Errorf("img1 should have been ignored for WCAG_1_1_1")
	}
	if !foundImg2 {
		t.Errorf("img2 should NOT have been ignored")
	}
	if foundBtn1 {
		t.Errorf("btn1 should have been ignored for 'all' rules")
	}
}
