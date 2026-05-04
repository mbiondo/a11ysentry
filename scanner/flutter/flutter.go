package flutter

import (
	"a11ysentry/scanner"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Framework implements scanner.ProjectFramework for Flutter projects.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Flutter (Dart)" }

// Probe returns true when dir contains a Flutter project.
func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "pubspec.yaml"))
}

// CollectFiles walks dir and returns UI component files.
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
		if ext == ".dart" {
			uiFiles = append(uiFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

var (
	dartImportRe = regexp.MustCompile(`(?m)import\s+['"]([^'"]+)['"]`)
)

// ResolveImports implements Flutter-specific import resolution.
// It handles both relative imports and 'package:PROJECT_NAME/...' style imports.
func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	var resolved []string
	base := filepath.Dir(filePath)
	
	// Try to detect project name from pubspec.yaml if possible
	projectName := filepath.Base(projectRoot) // fallback
	pubspecPath := filepath.Join(projectRoot, "pubspec.yaml")
	if data, err := os.ReadFile(pubspecPath); err == nil {
		re := regexp.MustCompile(`(?m)^name:\s*(\w+)`)
		if m := re.FindStringSubmatch(string(data)); len(m) > 1 {
			projectName = m[1]
		}
	}

	matches := dartImportRe.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		imp := m[1]
		var abs string
		
		if strings.HasPrefix(imp, "package:" + projectName + "/") {
			// Resolve project-local package import
			relPath := strings.TrimPrefix(imp, "package:" + projectName + "/")
			abs = filepath.Join(projectRoot, "lib", relPath)
		} else if !strings.HasPrefix(imp, "package:") && !strings.HasPrefix(imp, "dart:") {
			// Resolve relative import
			abs = filepath.Join(base, imp)
		} else {
			continue
		}

		abs = filepath.Clean(abs)
		if fileSet[abs] {
			resolved = append(resolved, abs)
		}
	}

	return resolved
}

// BuildPageTrees identifies top-level Widgets and builds their full import trees.
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
		// Canonical Flutter entry point detection:
		// 1. Files not imported.
		// 2. OR lib/main.dart (the standard starting point)
		isMain := strings.HasSuffix(filepath.ToSlash(file), "lib/main.dart")
		
		if importedByAnyone[file] && !isMain {
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
