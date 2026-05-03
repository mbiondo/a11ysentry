package pyqt

import (
	"io/fs"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "PyQt" }

func (f *Framework) Probe(dir string) bool {
	hasUIFiles := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && scanner.SkipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if strings.HasSuffix(d.Name(), ".ui") {
			hasUIFiles = true
			return filepath.SkipDir
		}
		return nil
	})
	return hasUIFiles
}

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

		if strings.HasSuffix(d.Name(), ".ui") {
			uiFiles = append(uiFiles, path)
		}
		if strings.HasSuffix(d.Name(), ".qss") || strings.HasSuffix(d.Name(), ".css") {
			cssFiles = append(cssFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	return []string{} // UI files don't typically import each other in PyQt
}

func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	var trees []scanner.PageTree
	for _, file := range allFiles {
		root := scanner.CollectTree(file, importGraph, make(map[string]bool))
		trees = append(trees, scanner.PageTree{
			Label: shortPath(file, projectRoot),
			Root:  root,
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
