package ios

import (
	"a11ysentry/scanner"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Framework implements scanner.ProjectFramework for iOS projects.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "iOS (Swift/SwiftUI)" }

// Probe returns true when dir contains an iOS project.
func (f *Framework) Probe(dir string) bool {
	if scanner.FileExists(filepath.Join(dir, "Package.swift")) {
		return true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && (strings.HasSuffix(entry.Name(), ".xcodeproj") || strings.HasSuffix(entry.Name(), ".xcworkspace")) {
			return true
		}
	}
	return false
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
		if scanner.UIExtensions[ext] {
			uiFiles = append(uiFiles, path)
		}
		return nil
	})
	return uiFiles, cssFiles, err
}

var (
	swiftViewRe = regexp.MustCompile(`(\w+)\(\)`)
)

// ResolveImports implements iOS-specific import resolution.
// Since Swift doesn't require explicit file imports within the same target,
// we look for View() instantiation patterns as a heuristic.
func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	var resolved []string
	
	// Fast lookup map
	baseNames := make(map[string]string)
	for f := range fileSet {
		base := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		baseNames[base] = f
	}

	matches := swiftViewRe.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		viewName := m[1]
		if realPath, ok := baseNames[viewName]; ok {
			if realPath != filePath {
				resolved = append(resolved, realPath)
			}
		}
	}

	return resolved
}

// BuildPageTrees identifies Views/Controllers and builds their full import trees.
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
		// Canonical iOS entry point detection:
		// 1. Files not imported/referenced.
		// 2. OR files containing @main or App struct.
		
		content, _ := os.ReadFile(file)
		src := string(content)
		
		isUIEntryPoint := strings.Contains(src, "@main") || 
			strings.Contains(src, ": App") ||
			strings.HasSuffix(file, "AppDelegate.swift")
		
		if importedByAnyone[file] && !isUIEntryPoint {
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
