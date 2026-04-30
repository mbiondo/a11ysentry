// Package generic provides a ProjectFramework implementation for generic
// frontend projects (React SPA, Vue CLI, Svelte, etc.) that do not follow a
// file-system routing convention.
//
// Root detection strategy: any file that is NOT imported by any other file in
// the project is considered a root entry point.
package generic

import (
	"io/fs"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

// Framework is the generic (fallback) project framework.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Generic" }

// Probe always returns false — generic is only used as a fallback.
func (f *Framework) Probe(_ string) bool { return false }

// CollectFiles walks dir and returns all UI component files and CSS files.
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
		base := strings.ToLower(d.Name())

		if scanner.CSSExtensions[ext] {
			cssFiles = append(cssFiles, path)
			return nil
		}
		if base == "tailwind.config.js" || base == "tailwind.config.ts" || base == "tailwind.config.mjs" {
			cssFiles = append(cssFiles, path)
			return nil
		}
		if scanner.UIExtensions[ext] {
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

// BuildPageTrees treats every file that is not imported by any other file as a
// root. Each root becomes its own PageTree with its full transitive import
// closure.
func (f *Framework) BuildPageTrees(allFiles []string, importGraph map[string][]string, projectRoot string) []scanner.PageTree {
	importedByAnyone := make(map[string]bool)
	for _, deps := range importGraph {
		for _, dep := range deps {
			importedByAnyone[dep] = true
		}
	}

	var trees []scanner.PageTree
	for _, file := range allFiles {
		if importedByAnyone[file] {
			continue
		}
		tree := scanner.CollectTree(file, importGraph, make(map[string]bool))
		trees = append(trees, scanner.PageTree{
			Label: shortPath(file, projectRoot),
			Files: tree,
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
