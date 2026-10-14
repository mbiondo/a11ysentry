package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
