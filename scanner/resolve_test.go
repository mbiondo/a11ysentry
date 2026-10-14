package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveImports_Opaque(t *testing.T) {
	// GIVEN a temporary project with an external import
	tmpDir, _ := os.MkdirTemp("", "sentry-test-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	filePath := filepath.Join(tmpDir, "index.tsx")
	content := `import { Button } from "@mui/material";
import Local from "./Local";

function App() {
	return <Button>Click</Button>;
}`
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "Local.tsx"), []byte("export default {}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	fileSet := map[string]bool{
		filepath.Join(tmpDir, "index.tsx"): true,
		filepath.Join(tmpDir, "Local.tsx"): true,
	}

	// WHEN resolving imports
	resolved := ResolveImports(filePath, tmpDir, fileSet, LoadTSConfigPaths(tmpDir))

	// THEN it MUST include the local file normally
	// AND it MUST include the external package with pkg:// prefix
	foundLocal := false
	foundExternal := false
	for _, r := range resolved {
		if strings.Contains(r, "Local.tsx") {
			foundLocal = true
		}
		if r == "pkg://@mui/material/Button" {
			foundExternal = true
		}
	}

	if !foundLocal {
		t.Error("expected local import to be resolved")
	}
	if !foundExternal {
		t.Error("expected external import to be resolved as pkg://@mui/material/Button")
	}
}

func TestCollectTree_Opaque(t *testing.T) {
	// GIVEN an import graph with an opaque node
	graph := map[string][]string{
		"root.tsx":  {"local.tsx", "pkg://@mui/material/Button"},
		"local.tsx": {},
	}

	// WHEN building the tree
	tree := CollectTree("root.tsx", graph, make(map[string]bool))

	// THEN the tree MUST contain the opaque node correctly flagged
	if len(tree.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(tree.Children))
	}

	foundOpaque := false
	for _, child := range tree.Children {
		if child.IsOpaque {
			if child.OpaqueSource != "@mui/material/Button" {
				t.Errorf("expected OpaqueSource '@mui/material/Button', got %s", child.OpaqueSource)
			}
			foundOpaque = true
		}
	}

	if !foundOpaque {
		t.Error("expected to find an opaque child node")
	}
}

func writeResolveFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A component import used to be resolved twice: .tsx/.jsx appear both in UIExtensions and in the
// follow-up extension list, and add() appended without checking. Every component edge in the graph
// was therefore duplicated, which put each shared component twice in a page tree.
func TestResolveImportsDoesNotDuplicateAnEdge(t *testing.T) {
	dir := t.TempDir()

	layout := writeResolveFixture(t, dir, "layout.tsx",
		"import { Widget } from \"./Widget\";\nimport { helper } from \"./helper\";\n")
	widget := writeResolveFixture(t, dir, "Widget.tsx", "export function Widget() { return null; }")
	helper := writeResolveFixture(t, dir, "helper.ts", "export const helper = 1;")

	fileSet := map[string]bool{layout: true, widget: true, helper: true}

	imports := ResolveImports(layout, dir, fileSet, nil)

	counts := map[string]int{}
	for _, imp := range imports {
		counts[filepath.Base(imp)]++
	}

	for name, count := range counts {
		if count != 1 {
			t.Errorf("%s resolved %d times, want exactly 1 (full list: %v)", name, count, imports)
		}
	}
	if counts["Widget.tsx"] != 1 {
		t.Errorf("Widget.tsx missing from the resolved imports: %v", imports)
	}
	if counts["helper.ts"] != 1 {
		t.Errorf("helper.ts missing from the resolved imports: %v", imports)
	}
}

// The extension probing must still find a component imported without an extension.
func TestResolveImportsStillFindsExtensionlessComponent(t *testing.T) {
	dir := t.TempDir()

	page := writeResolveFixture(t, dir, "page.tsx", "import { Card } from \"./Card\";\n")
	writeResolveFixture(t, dir, "Card.tsx", "export function Card() { return null; }")

	fileSet := map[string]bool{page: true, filepath.Join(dir, "Card.tsx"): true}

	imports := ResolveImports(page, dir, fileSet, nil)

	var found bool
	for _, imp := range imports {
		if strings.HasSuffix(imp, "Card.tsx") {
			found = true
		}
	}
	if !found {
		t.Errorf("an extensionless component import was not resolved: %v", imports)
	}
}
