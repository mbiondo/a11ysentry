package django

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"a11ysentry/scanner"
)

type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Django" }

func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "manage.py")) ||
		scanner.FileExists(filepath.Join(dir, "settings.py"))
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

var (
	djangoExtendsRe = regexp.MustCompile(`{%\s*extends\s*['"]([^'"]+)['"]\s*%}`)
	djangoIncludeRe = regexp.MustCompile(`{%\s*include\s*['"]([^'"]+)['"]\s*%}`)
)

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	var resolved []string
	base := filepath.Dir(filePath)

	tryResolve := func(path string) {
		// 1. Try relative to current file
		abs := filepath.Clean(filepath.Join(base, path))
		if fileSet[abs] {
			resolved = append(resolved, abs)
			return
		}

		// 2. Try relative to template dirs (templates/ or src/templates/)
		prefixes := []string{
			filepath.Join(projectRoot, "templates"),
			filepath.Join(projectRoot, "src", "templates"),
		}
		for _, prefix := range prefixes {
			abs = filepath.Clean(filepath.Join(prefix, path))
			if fileSet[abs] {
				resolved = append(resolved, abs)
				return
			}
		}
	}

	for _, m := range djangoExtendsRe.FindAllStringSubmatch(src, -1) {
		tryResolve(m[1])
	}
	for _, m := range djangoIncludeRe.FindAllStringSubmatch(src, -1) {
		tryResolve(m[1])
	}

	return resolved
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
