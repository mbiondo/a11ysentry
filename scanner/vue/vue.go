package vue

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Vue" }

func (f *Framework) Probe(dir string) bool {
	// Don't claim Nuxt projects
	if scanner.FileExists(filepath.Join(dir, "nuxt.config.ts")) || 
	   scanner.FileExists(filepath.Join(dir, "nuxt.config.js")) {
		return false
	}

	// Check package.json for Vue
	pkgPath := filepath.Join(dir, "package.json")
	if scanner.FileExists(pkgPath) {
		data, err := os.ReadFile(pkgPath)
		if err == nil {
			content := string(data)
			if strings.Contains(content, `"nuxt"`) {
				return false // Definitely Nuxt
			}
			if strings.Contains(content, `"vue"`) {
				return true
			}
		}
	}
	
	// Fallback to checking for .vue files if no package.json
	hasVueFiles := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".vue") {
			hasVueFiles = true
			return filepath.SkipDir // Stop early
		}
		return nil
	})
	
	return hasVueFiles
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
		if ext == ".xml" {
			return nil
		}

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

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	return scanner.ResolveImports(filePath, projectRoot, fileSet)
}

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

	var trees []scanner.PageTree
	for _, file := range allFiles {
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
