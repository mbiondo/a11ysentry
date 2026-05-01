// Package nextjs provides a ProjectFramework for Next.js App Router projects.
//
// # App Router layout hierarchy
//
// In the Next.js App Router, layouts wrap pages. The file system defines the
// nesting: a layout.tsx in a parent directory wraps all pages and nested
// layouts beneath it.
//
//	app/
//	  layout.tsx            ← root layout (html, lang, global providers)
//	  page.tsx              ← home route
//	  (app)/
//	    layout.tsx          ← nested layout
//	    activities/
//	      page.tsx          ← route: wrapped by (app)/layout + root layout
//	    admin/
//	      layout.tsx        ← nested layout
//	      page.tsx          ← route: wrapped by admin/layout + (app)/layout + root layout
//
// # Analysis units
//
// Each page.tsx is analyzed together with its full layout chain (from the
// nearest ancestor layout down to the root layout), plus its imported
// components. This ensures the engine sees html[lang], providers, and shared
// navigation that live in layouts — eliminating false positives for WCAG_3_1_1
// and giving accurate context for contrast and heading rules.
package nextjs

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"a11ysentry/scanner"
)

// Framework implements scanner.ProjectFramework for Next.js App Router.
type Framework struct{}

// New returns a new Next.js framework scanner.
func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Next.js App Router" }

// Probe returns true when dir contains a Next.js App Router project.
// Heuristic: presence of app/ directory AND next.config.* at the root.
func (f *Framework) Probe(dir string) bool {
	hasAppDir := scanner.DirExists(filepath.Join(dir, "app"))
	hasNextConfig := scanner.FileExists(filepath.Join(dir, "next.config.js")) ||
		scanner.FileExists(filepath.Join(dir, "next.config.ts")) ||
		scanner.FileExists(filepath.Join(dir, "next.config.mjs"))
	return hasAppDir && hasNextConfig
}

// CollectFiles walks dir and returns UI component files and CSS/config files.
// Only .tsx, .jsx, .html, .vue, .svelte, .astro are considered UI files;
// pure .ts/.js files are skipped since they are utilities, not components.
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

// BuildPageTrees builds one PageTree per page.tsx/page.jsx found under app/.
//
// Each tree contains, in order:
//  1. The layout chain: all layout.tsx files from the project root down to the
//     page's nearest ancestor directory, outermost first.
//  2. The page file itself.
//  3. The transitive closure of all components imported by the above files.
//
// This mirrors exactly what Next.js renders for a given route.
func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	// Index: absolute dir → layout file in that dir (if any).
	layoutByDir := make(map[string]string)
	var pages []string

	for _, file := range allFiles {
		base := strings.ToLower(filepath.Base(file))
		switch base {
		case "layout.tsx", "layout.jsx":
			layoutByDir[filepath.Dir(file)] = file
		case "page.tsx", "page.jsx":
			pages = append(pages, file)
		}
	}

	// Sort pages for deterministic output.
	sort.Strings(pages)

	var trees []scanner.PageTree
	for _, page := range pages {
		chain := layoutChain(page, layoutByDir, projectRoot)

		// Collect components: start from layouts and page, deduplicate.
		visited := make(map[string]bool)
		var treeFiles []string

		for _, layout := range chain {
			treeFiles = append(treeFiles, scanner.CollectTree(layout, importGraph, visited)...)
		}
		treeFiles = append(treeFiles, scanner.CollectTree(page, importGraph, visited)...)

		trees = append(trees, scanner.PageTree{
			Label: shortPath(page, projectRoot),
			Files: treeFiles,
		})
	}
	return trees
}

// layoutChain returns the ordered list of layout files that wrap page,
// from outermost (root) to innermost (nearest ancestor).
//
// It walks up the directory tree from page's directory toward projectRoot,
// collecting layout files along the way, then reverses so root comes first.
func layoutChain(page string, layoutByDir map[string]string, projectRoot string) []string {
	var chain []string
	dir := filepath.Dir(page)

	for {
		if layout, ok := layoutByDir[dir]; ok {
			chain = append(chain, layout)
		}
		if dir == projectRoot {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root — shouldn't happen in practice.
			break
		}
		dir = parent
	}

	// Reverse: root layout first, nearest layout last.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

func shortPath(abs, base string) string {
	rel, err := filepath.Rel(base, abs)
	if err != nil {
		return abs
	}
	return rel
}
