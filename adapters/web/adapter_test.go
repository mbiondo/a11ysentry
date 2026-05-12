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

func TestHTMLAdapter_ThemeBlock(t *testing.T) {
	// GIVEN a CSS file using @theme (Tailwind CSS v4 / Astro convention)
	cssContent := `@theme {
  --color-neon-cyan: #22d3ee;
  --color-brand-primary: #6366f1;
}
`
	cssFile, _ := os.CreateTemp("", "test_theme_*.css")
	defer func() { _ = os.Remove(cssFile.Name()) }()
	_ = os.WriteFile(cssFile.Name(), []byte(cssContent), 0644)

	htmlContent := `<!DOCTYPE html>
<html><body>
  <a id="link1" class="text-neon-cyan bg-brand-primary">Hello</a>
  <a id="link2" class="text-neon-cyan/50">Faded</a>
</body></html>`
	htmlFile, _ := os.CreateTemp("", "test_theme_*.html")
	defer func() { _ = os.Remove(htmlFile.Name()) }()
	_ = os.WriteFile(htmlFile.Name(), []byte(htmlContent), 0644)

	adapter := NewHTMLAdapter()
	LoadProjectCSS(adapter, []string{cssFile.Name()})

	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: htmlFile.Name()})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	var link1, link2 *domain.USN
	for i := range nodes {
		switch nodes[i].UID {
		case "link1":
			link1 = &nodes[i]
		case "link2":
			link2 = &nodes[i]
		}
	}

	if link1 == nil {
		t.Fatal("link1 not found")
	}
	if link1.Traits["color"] != "#22d3ee" {
		t.Errorf("expected text-neon-cyan → #22d3ee, got %v", link1.Traits["color"])
	}
	if link1.Traits["background-color"] != "#6366f1" {
		t.Errorf("expected bg-brand-primary → #6366f1, got %v", link1.Traits["background-color"])
	}

	if link2 == nil {
		t.Fatal("link2 not found")
	}
	// text-neon-cyan/50 → 50% opacity blended over white (#ffffff default)
	if link2.Traits["color"] == "" || link2.Traits["color"] == "#22d3ee" {
		t.Errorf("expected opacity-blended color for text-neon-cyan/50, got %v", link2.Traits["color"])
	}
}

func TestHTMLAdapter_ArbitraryValues(t *testing.T) {
	// GIVEN a page using Tailwind JIT arbitrary color values
	htmlContent := `<!DOCTYPE html>
<html><body>
  <a id="link1" class="bg-[#1a1b26] text-[#a9b1d6]">Dark BG</a>
  <a id="link2" class="bg-[rgb(26,27,38)] text-[#ffffff]">RGB BG</a>
</body></html>`
	htmlFile, _ := os.CreateTemp("", "test_arbitrary_*.html")
	defer func() { _ = os.Remove(htmlFile.Name()) }()
	_ = os.WriteFile(htmlFile.Name(), []byte(htmlContent), 0644)

	adapter := NewHTMLAdapter()
	nodes, err := adapter.Ingest(context.Background(), &domain.FileNode{FilePath: htmlFile.Name()})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	var link1, link2 *domain.USN
	for i := range nodes {
		switch nodes[i].UID {
		case "link1":
			link1 = &nodes[i]
		case "link2":
			link2 = &nodes[i]
		}
	}

	if link1 == nil {
		t.Fatal("link1 not found")
	}
	if link1.Traits["background-color"] != "#1a1b26" {
		t.Errorf("bg-[#1a1b26] → expected #1a1b26, got %v", link1.Traits["background-color"])
	}
	if link1.Traits["color"] != "#a9b1d6" {
		t.Errorf("text-[#a9b1d6] → expected #a9b1d6, got %v", link1.Traits["color"])
	}

	if link2 == nil {
		t.Fatal("link2 not found")
	}
	if link2.Traits["background-color"] != "#1a1b26" {
		t.Errorf("bg-[rgb(26,27,38)] → expected #1a1b26, got %v", link2.Traits["background-color"])
	}
}

func TestHTMLAdapter_NormalizeHSL(t *testing.T) {
	a := &htmlAdapter{}
	cases := []struct {
		input string
		want  string
	}{
		{"hsl(0, 100%, 50%)", "#ff0000"},
		{"hsl(120, 100%, 50%)", "#00ff00"},
		{"hsl(240, 100%, 50%)", "#0000ff"},
		{"hsla(0, 0%, 100%, 1)", "#ffffff"},
		{"hsla(0, 0%, 0%, 1)", "#000000"},
	}
	for _, c := range cases {
		got := a.normalizeColor(c.input)
		if got != c.want {
			t.Errorf("normalizeColor(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
