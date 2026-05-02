// Package astro provides a ProjectFramework for Astro projects.
//
// Astro uses a file-system routing convention: every .astro file under
// src/pages/ (or pages/) is a route entry point. Layouts are imported
// explicitly via import statements — they are NOT inferred from directory
// structure. The generic root-detection strategy (files not imported by anyone
// = roots) therefore works correctly for Astro out of the box.
//
// This package customises CollectFiles to also include .astro files outside
// src/pages/ (components, layouts) and correctly probes for the Astro project
// marker (astro.config.*).
package astro

import (
	"io/fs"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

// Framework implements scanner.ProjectFramework for Astro projects.
type Framework struct{}

// New returns a new Astro framework scanner.
func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Astro" }

// Probe returns true when dir contains an Astro project.
func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "astro.config.mjs")) ||
		scanner.FileExists(filepath.Join(dir, "astro.config.ts")) ||
		scanner.FileExists(filepath.Join(dir, "astro.config.js"))
}

// CollectFiles walks dir and returns all UI and CSS files.
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
		if ext == ".xml" {
			return nil
		}
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

// BuildPageTrees uses the generic strategy: files not imported by anyone are
// roots. In Astro, src/pages/*.astro files are entry points that import
// layouts and components explicitly — so this naturally gives correct trees.
func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	importedByAnyone := make(map[string]bool)
	for _, deps := range importGraph {
		for _, dep := range deps {
			importedByAnyone[dep] = true
		}
	}

	// Identify files in pages directory
	isPage := make(map[string]bool)
	hasPagesDir := false
	for _, file := range allFiles {
		rel, _ := filepath.Rel(projectRoot, file)
		slashRel := filepath.ToSlash(rel)
		if strings.HasPrefix(slashRel, "src/pages") || strings.HasPrefix(slashRel, "pages") {
			isPage[file] = true
			hasPagesDir = true
		}
	}

	var trees []scanner.PageTree
	for _, file := range allFiles {
		// If project has a pages dir, only consider files there as potential roots.
		// If not, fall back to the "not imported" strategy.
		if hasPagesDir && !isPage[file] {
			continue
		}
		
		if importedByAnyone[file] {
			continue
		}
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
