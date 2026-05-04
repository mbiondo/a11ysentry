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

	// Index all special files by their directory for fast lookup.
	specialFilesByDir := make(map[string]map[string]string)
	for _, file := range allFiles {
		base := strings.ToLower(filepath.Base(file))
		if base == "+layout.svelte" || base == "+error.svelte" {
			dir := filepath.Dir(file)
			if _, ok := specialFilesByDir[dir]; !ok {
				specialFilesByDir[dir] = make(map[string]string)
			}
			specialFilesByDir[dir][base] = file
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

		var root *domain.FileNode
		var current *domain.FileNode

		chain := svelteLayoutChain(filepath.Dir(file), routesDir, specialFilesByDir)

		// Stitch layout chain: outermost first.
		for _, component := range chain {
			node := scanner.CollectTree(component, importGraph, make(map[string]bool))
			if node == nil {
				continue
			}
			if root == nil {
				root = node
			} else {
				current.Children = append(current.Children, node)
			}
			current = node
		}

		// Add page at the end of the layout chain.
		pageNode := scanner.CollectTree(file, importGraph, make(map[string]bool))
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

// svelteLayoutChain walks up from pageDir to routesDir collecting +layout.svelte and +error.svelte at
// each level, then reverses the slice so the root layout comes first.
func svelteLayoutChain(pageDir, routesDir string, specialFilesByDir map[string]map[string]string) []string {
	var chain []string
	cur := pageDir
	for {
		if files, ok := specialFilesByDir[cur]; ok {
			// Order within folder: layout then error
			if l, ok := files["+layout.svelte"]; ok {
				chain = append(chain, l)
			}
			if e, ok := files["+error.svelte"]; ok {
				chain = append(chain, e)
			}
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
