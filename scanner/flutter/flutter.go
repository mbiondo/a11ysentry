package flutter

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/scanner"
	"io/fs"
	"path/filepath"
	"strings"
)

// Framework implements scanner.ProjectFramework for Flutter projects.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Flutter (Dart)" }

// Probe returns true when dir contains a Flutter project.
func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "pubspec.yaml"))
}

// CollectFiles walks dir and returns UI component files.
func (f *Framework) CollectFiles(dir string) ([]string, []string, error) {
	var uiFiles, cssFiles []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if scanner.SkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".dart" {
			uiFiles = append(uiFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

// ResolveImports delegates to the shared resolver.
func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	return scanner.ResolveImports(filePath, projectRoot, fileSet)
}

// BuildPageTrees groups all files into a single analysis unit for simplicity.
func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	var trees []scanner.PageTree
	for _, file := range allFiles {
		trees = append(trees, scanner.PageTree{
			Label: shortPath(file, projectRoot),
			Root:  &domain.FileNode{FilePath: file},
		})
	}
	return trees
}

func shortPath(abs, base string) string {
	rel, err := filepath.Rel(base, abs)
	if err != nil {
		return abs
	}
	return rel
}
