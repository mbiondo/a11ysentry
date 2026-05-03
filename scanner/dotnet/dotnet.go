package dotnet

import (
	"io/fs"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "DotNet" }

func (f *Framework) Probe(dir string) bool {
	hasDotNetFiles := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && scanner.SkipDirs[d.Name()] {
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".csproj" || ext == ".sln" {
			hasDotNetFiles = true
			return filepath.SkipDir // Found it, stop scanning
		}
		return nil
	})
	return hasDotNetFiles
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

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".xaml" || ext == ".cs" || ext == ".razor" {
			uiFiles = append(uiFiles, path)
		}
		if scanner.CSSExtensions[ext] {
			cssFiles = append(cssFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	return []string{} // .NET typically uses global namespaces, not file imports
}

func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	var trees []scanner.PageTree
	for _, file := range allFiles {
		// Treat each file as a standalone tree for now
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
