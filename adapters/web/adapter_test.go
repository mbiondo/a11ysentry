package web

import (
	"context"
	"os"
	"strings"
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
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if _, err := tmpFile.WriteString(htmlContent); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	adapter := NewHTMLAdapter()
	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: tmpFile.Name()})
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

func TestHTMLAdapter_ComplexCSS(t *testing.T) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				:root { --main-bg: #ffffff; --main-fg: #000000; }
				.btn, .link { color: var(--main-fg); background-color: var(--main-bg); }
				@media (prefers-color-scheme: dark) {
					:root { --main-bg: #000000; --main-fg: #ffffff; }
					.btn { border-color: #ff00ff; }
				}
			</style>
		</head>
		<body>
			<button class="btn" id="btn1">Button</button>
			<a class="link" id="link1">Link</a>
		</body>
		</html>
	`
	tmpFile, _ := os.CreateTemp("", "test_complex_css_*.html")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = os.WriteFile(tmpFile.Name(), []byte(htmlContent), 0644)

	adapter := NewHTMLAdapter().(*htmlAdapter)
	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	var btnNode *domain.USN
	for i := range nodes {
		if nodes[i].UID == "btn1" {
			btnNode = &nodes[i]
		}
	}

	if btnNode == nil {
		t.Fatal("button node not found")
	}

	// Basic color resolution via variables
	if btnNode.Traits["color"] != "#000000" {
		t.Errorf("expected btn color #000000, got %v", btnNode.Traits["color"])
	}

	// Dark mode override check
	if adapter.darkCSSMap["btn"]["border-color"] != "#ff00ff" {
		t.Errorf("expected dark mode border-color #ff00ff, got %v", adapter.darkCSSMap["btn"]["border-color"])
	}
}

func TestHTMLAdapter_OpaqueNode(t *testing.T) {
	// GIVEN a file using an external component
	htmlContent := `<MuiButton aria-label="Submit" disabled />`
	tmpFile, _ := os.CreateTemp("", "test_opaque_*.tsx")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	err := os.WriteFile(tmpFile.Name(), []byte(htmlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// AND a FileNode tree where MuiButton is marked as opaque
	root := &domain.FileNode{
		FilePath: tmpFile.Name(),
		Children: []*domain.FileNode{
			{
				FilePath:     "pkg://@mui/material/MuiButton",
				IsOpaque:     true,
				OpaqueSource: "@mui/material/MuiButton",
			},
		},
	}

	adapter := NewHTMLAdapter()
	
	// WHEN ingesting the tree
	nodes, err := adapter.Ingest(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}

	// THEN we MUST find a USN for the opaque component with its traits
	var opaqueNode *domain.USN
	for i := range nodes {
		if strings.EqualFold(nodes[i].UID, "MuiButton") {
			opaqueNode = &nodes[i]
		}
	}

	if opaqueNode == nil {
		t.Fatal("opaque node MuiButton not found")
	}

	if opaqueNode.Label != "Submit" {
		t.Errorf("expected label 'Submit', got %s", opaqueNode.Label)
	}

	if !opaqueNode.IsOpaque {
		t.Error("expected node to be marked as IsOpaque")
	}

	if opaqueNode.Source.OpaqueSource != "@mui/material/MuiButton" {
		t.Errorf("expected OpaqueSource '@mui/material/MuiButton', got %s", opaqueNode.Source.OpaqueSource)
	}
}
