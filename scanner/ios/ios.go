package ios

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"a11ysentry/scanner"
)

// Framework implements scanner.ProjectFramework for iOS projects.
type Framework struct{}

// New returns a new iOS framework scanner.
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

// ResolveImports delegates to the shared resolver.
func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	return scanner.ResolveImports(filePath, projectRoot, fileSet)
}

// BuildPageTrees groups all files into a single analysis unit for simplicity.
func (f *Framework) BuildPageTrees(
	allFiles []string,
	importGraph map[string][]string,
	projectRoot string,
) []scanner.PageTree {
	if len(allFiles) == 0 {
		return nil
	}

	return []scanner.PageTree{
		{
			Label: "iOS App",
			Files: allFiles,
		},
	}
}
