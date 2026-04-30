// Package nuxt provides a ProjectFramework implementation for Nuxt 3 projects.
//
// Detection: looks for nuxt.config.ts / nuxt.config.js / nuxt.config.mjs
//
// Page tree strategy:
//   - Pages live in pages/ (.vue files, file-system routing)
//   - Default layout is layouts/default.vue — wraps every page
//   - Each tree: [default layout (if exists), page.vue, imported components]
package nuxt

import (
	"io/fs"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

// Framework is the Nuxt 3 project framework.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Nuxt 3" }

// Probe returns true when a nuxt.config.* file is present at the project root.
func (f *Framework) Probe(dir string) bool {
	for _, name := range []string{"nuxt.config.ts", "nuxt.config.js", "nuxt.config.mjs"} {
		if scanner.FileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

// CollectFiles walks dir and returns all .vue/.astro/.tsx/.jsx/.svelte UI files
// and CSS files, skipping build/vendor directories.
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
		if scanner.CSSExtensions[ext] {
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

// BuildPageTrees builds one PageTree per page in pages/:
//
//	[layouts/default.vue (if exists)] + [page.vue] + [imported components]
func (f *Framework) BuildPageTrees(allFiles []string, importGraph map[string][]string, projectRoot string) []scanner.PageTree {
	pagesDir := filepath.Join(projectRoot, "pages")
	defaultLayout := filepath.Join(projectRoot, "layouts", "default.vue")
	hasDefaultLayout := scanner.FileExists(defaultLayout)

	var trees []scanner.PageTree

	for _, file := range allFiles {
		// Only files directly inside pages/ (at any depth) are page roots.
		rel, err := filepath.Rel(pagesDir, file)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}

		// Build component closure for this page.
		closure := scanner.CollectTree(file, importGraph, make(map[string]bool))

		var files []string
		if hasDefaultLayout {
			files = append(files, defaultLayout)
		}
		files = append(files, closure...)

		trees = append(trees, scanner.PageTree{
			Label: "pages/" + filepath.ToSlash(rel),
			Files: files,
		})
	}

	return trees
}
