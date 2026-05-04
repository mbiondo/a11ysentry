package angular

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

func (f *Framework) Name() string { return "Angular" }

func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "angular.json"))
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
	angularTemplateRe = regexp.MustCompile(`templateUrl\s*:\s*['"]([^'"]+)['"]`)
	angularStyleRe    = regexp.MustCompile(`styleUrls\s*:\s*\[\s*['"]([^'"]+)['"]`)
)

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	// 1. Standard TS imports
	resolved := scanner.ResolveImports(filePath, projectRoot, fileSet)
	
	base := filepath.Dir(filePath)

	// 2. Angular template resolution
	if m := angularTemplateRe.FindStringSubmatch(src); len(m) > 1 {
		abs := filepath.Clean(filepath.Join(base, m[1]))
		if fileSet[abs] {
			resolved = append(resolved, abs)
		}
	}

	// 3. Angular styles resolution (simplified: first style only for now)
	if m := angularStyleRe.FindAllStringSubmatch(src, -1); len(m) > 0 {
		for _, match := range m {
			abs := filepath.Clean(filepath.Join(base, match[1]))
			if fileSet[abs] {
				resolved = append(resolved, abs)
			}
		}
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
