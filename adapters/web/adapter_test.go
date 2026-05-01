package web

import (
	"context"
	"os"
	"testing"
	"a11ysentry/engine/core/domain"
)

func TestHTMLAdapter_CSSAnalysis(t *testing.T) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				.danger { color: #FF0000; background-color: #FFFFFF; }
				.safe { color: #000000; }
			</style>
		</head>
		<body>
			<button class="danger" id="btn1">Action</button>
			<p class="safe" style="color: #123456;" id="p1">Text</p>
		</body>
		</html>
	`
	tmpFile, _ := os.CreateTemp("", "test_css_*.html")
	defer os.Remove(tmpFile.Name()) //nolint:errcheck
	if _, err := tmpFile.WriteString(htmlContent); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	adapter := NewHTMLAdapter()
	nodes, err := adapter.Ingest(context.Background(), []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	var btnNode, pNode *domain.USN
	for i := range nodes {
		if nodes[i].UID == "btn1" {
			btnNode = &nodes[i]
		}
		if nodes[i].UID == "p1" {
			pNode = &nodes[i]
		}
	}

	if btnNode == nil {
		t.Fatal("button node not found")
	}
	if btnNode.Traits["color"] != "#FF0000" {
		t.Errorf("expected btn color #FF0000, got %v", btnNode.Traits["color"])
	}
	if btnNode.Traits["background-color"] != "#FFFFFF" {
		t.Errorf("expected btn bg #FFFFFF, got %v", btnNode.Traits["background-color"])
	}

	if pNode == nil {
		t.Fatal("p node not found")
	}
	// Inline style should override class style
	if pNode.Traits["color"] != "#123456" {
		t.Errorf("expected p color #123456 (inline override), got %v", pNode.Traits["color"])
	}
}
