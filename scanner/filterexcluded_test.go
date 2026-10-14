package scanner

import (
	"path/filepath"
	"testing"
)

// An exclude entry must be matched against the path inside the project, not the absolute path.
//
// Matching the absolute path lets a directory name anywhere above the project act as a global
// exclude: the repository's own a11ysentry.json excludes "landing", so every file under
// <checkout>/landing matched it and pointing at that project with --dir analyzed nothing at all.
func TestFilterExcludedMatchesInsideTheProjectOnly(t *testing.T) {
	root := filepath.Join(string(filepath.Separator)+"work", "landing")
	files := []string{
		filepath.Join(root, "src", "pages", "index.astro"),
		filepath.Join(root, "node_modules", "pkg", "index.js"),
		filepath.Join(root, "src", "legacy", "old.tsx"),
	}

	kept := FilterExcluded(files, []string{"node_modules", "landing", "legacy"}, root)

	// "landing" names the project itself: it must not exclude the project's own files.
	if len(kept) != 1 {
		t.Fatalf("kept %d files, want 1 (index.astro): %v", len(kept), kept)
	}
	if filepath.Base(kept[0]) != "index.astro" {
		t.Errorf("kept %q, want index.astro", kept[0])
	}
}

// An exclude still has to work for directories inside the project, at any depth.
func TestFilterExcludedStillMatchesInsideProject(t *testing.T) {
	root := filepath.Join(string(filepath.Separator)+"work", "app")
	files := []string{
		filepath.Join(root, "src", "a.tsx"),
		filepath.Join(root, "src", "generated", "b.tsx"),
	}

	kept := FilterExcluded(files, []string{"generated"}, root)

	if len(kept) != 1 || filepath.Base(kept[0]) != "a.tsx" {
		t.Errorf("kept %v, want only a.tsx", kept)
	}
}
