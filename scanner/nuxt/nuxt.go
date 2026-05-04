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
	"a11ysentry/engine/core/domain"
	"a11ysentry/scanner"
	"io/fs"
	"path/filepath"
	"strings"
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
	if !scanner.DirExists(pagesDir) {
		pagesDir = filepath.Join(projectRoot, "src", "pages")
	}
	
	globalApp := filepath.Join(projectRoot, "app.vue")
	if !scanner.FileExists(globalApp) {
		globalApp = filepath.Join(projectRoot, "src", "app.vue")
	}
	hasGlobalApp := scanner.FileExists(globalApp)

	defaultLayout := filepath.Join(projectRoot, "layouts", "default.vue")
	if !scanner.FileExists(defaultLayout) {
		defaultLayout = filepath.Join(projectRoot, "src", "layouts", "default.vue")
	}
	hasDefaultLayout := scanner.FileExists(defaultLayout)

	var trees []scanner.PageTree

	for _, file := range allFiles {
		// Only files directly inside pages/ (at any depth) are page roots.
		rel, err := filepath.Rel(pagesDir, file)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}

		var root *domain.FileNode
		var current *domain.FileNode

		// 1. Global app.vue
		if hasGlobalApp {
			root = scanner.CollectTree(globalApp, importGraph, make(map[string]bool))
			current = root
		}

		// 2. Default layout
		if hasDefaultLayout {
			layoutNode := scanner.CollectTree(defaultLayout, importGraph, make(map[string]bool))
			if root == nil {
				root = layoutNode
			} else if layoutNode != nil {
				current.Children = append(current.Children, layoutNode)
			}
			current = layoutNode
		}

		// 3. Page
		pageNode := scanner.CollectTree(file, importGraph, make(map[string]bool))
		if root == nil {
			root = pageNode
		} else if pageNode != nil {
			current.Children = append(current.Children, pageNode)
		}

		label := "pages/" + filepath.ToSlash(rel)
		if strings.Contains(pagesDir, "src") {
			label = "src/" + label
		}

		trees = append(trees, scanner.PageTree{
			Label: label,
			Root:  root,
		})
	}

	return trees
}
