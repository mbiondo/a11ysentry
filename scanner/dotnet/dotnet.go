package dotnet

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

func (f *Framework) Name() string { return "DotNet" }

func (f *Framework) Probe(dir string) bool {
	hasDotNetFiles := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && scanner.SkipDirs[d.Name()] {
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".csproj" || ext == ".sln" {
			hasDotNetFiles = true
			return filepath.SkipDir // Found it, stop scanning
		}
		return nil
	})
	return hasDotNetFiles
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
		if ext == ".xaml" || ext == ".cs" || ext == ".razor" {
			uiFiles = append(uiFiles, path)
		}
		if scanner.CSSExtensions[ext] {
			cssFiles = append(cssFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

var (
	razorImportRe = regexp.MustCompile(`(?m)@using\s+([\w.]+)`)
)

func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	ext := strings.ToLower(filepath.Ext(filePath))
	var resolved []string

	// 1. Link XAML to Code-Behind
	if ext == ".xaml" {
		codeBehind := filePath + ".cs"
		if fileSet[codeBehind] {
			resolved = append(resolved, codeBehind)
		}
	}

	// 2. Handle Razor @using imports (best effort)
	if ext == ".razor" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			src := string(content)
			matches := razorImportRe.FindAllStringSubmatch(src, -1)
			for _, m := range matches {
				pkgPath := strings.ReplaceAll(m[1], ".", string(filepath.Separator))
				// Search in project root for matching folder/file
				candidate := filepath.Join(projectRoot, pkgPath)
				if fileSet[candidate+".razor"] {
					resolved = append(resolved, candidate+".razor")
				}
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
		ext := strings.ToLower(filepath.Ext(file))
		// Only XAML or top-level Razor files are likely page roots
		if importedByAnyone[file] && ext != ".xaml" {
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
