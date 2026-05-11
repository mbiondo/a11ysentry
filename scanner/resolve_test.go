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
		"root.tsx": {"local.tsx", "pkg://@mui/material/Button"},
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
