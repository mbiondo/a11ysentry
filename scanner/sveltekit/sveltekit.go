// Package sveltekit provides a ProjectFramework implementation for SvelteKit projects.
//
// Detection: looks for svelte.config.js / svelte.config.ts
//
// Page tree strategy (mirrors Next.js App Router convention):
//   - Pages are +page.svelte files under src/routes/
//   - Layouts are +layout.svelte files — each route segment can have one
//   - Layout chain: walk up from the page collecting +layout.svelte at each
//     directory, then reverse (root-first), then append page + components
package sveltekit

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/scanner"
	"io/fs"
	"path/filepath"
	"strings"
)

// Framework is the SvelteKit project framework.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "SvelteKit" }

// Probe returns true when a svelte.config.* file is present at the project root.
func (f *Framework) Probe(dir string) bool {
	for _, name := range []string{"svelte.config.js", "svelte.config.ts"} {
		if scanner.FileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

// CollectFiles walks dir and returns all .svelte UI files and CSS files.
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

// BuildPageTrees builds one PageTree per +page.svelte in src/routes/:
//
//	[root +layout.svelte, ..., segment +layout.svelte, +page.svelte, components]
func (f *Framework) BuildPageTrees(allFiles []string, importGraph map[string][]string, projectRoot string) []scanner.PageTree {
	routesDir := filepath.Join(projectRoot, "src", "routes")

	// Index all +layout.svelte files by their directory for fast lookup.
	layoutByDir := make(map[string]string)
	for _, file := range allFiles {
		if strings.EqualFold(filepath.Base(file), "+layout.svelte") {
			layoutByDir[filepath.Dir(file)] = file
		}
	}

	var trees []scanner.PageTree

	for _, file := range allFiles {
		if !strings.EqualFold(filepath.Base(file), "+page.svelte") {
			continue
		}
		rel, err := filepath.Rel(routesDir, file)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}

		visited := make(map[string]bool)
		var root *domain.FileNode
		var current *domain.FileNode

		chain := layoutChain(filepath.Dir(file), routesDir, layoutByDir)

		// Stitch layout chain: outermost first.
		for _, layout := range chain {
			layoutNode := scanner.CollectTree(layout, importGraph, visited)
			if layoutNode == nil {
				continue
			}
			if root == nil {
				root = layoutNode
			} else {
				current.Children = append(current.Children, layoutNode)
			}
			current = layoutNode
		}

		// Add page at the end of the layout chain.
		pageNode := scanner.CollectTree(file, importGraph, visited)
		if pageNode != nil {
			if root == nil {
				root = pageNode
			} else {
				current.Children = append(current.Children, pageNode)
			}
		}

		trees = append(trees, scanner.PageTree{
			Label: filepath.ToSlash(rel),
			Root:  root,
		})
	}

	return trees
}

// layoutChain walks up from pageDir to routesDir collecting +layout.svelte at
// each level, then reverses the slice so the root layout comes first.
func layoutChain(pageDir, routesDir string, layoutByDir map[string]string) []string {
	var chain []string
	cur := pageDir
	for {
		if l, ok := layoutByDir[cur]; ok {
			chain = append(chain, l)
		}
		if cur == routesDir {
			break
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	// Reverse: root layout first.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}
