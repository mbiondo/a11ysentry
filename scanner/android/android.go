package android

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/scanner"
	"io/fs"
	"path/filepath"
	"strings"
)

// Framework implements scanner.ProjectFramework for Android projects.
type Framework struct{}

func New() *Framework { return &Framework{} }

func (f *Framework) Name() string { return "Android (Kotlin/Java)" }

// Probe returns true when dir contains an Android project.
func (f *Framework) Probe(dir string) bool {
	return scanner.FileExists(filepath.Join(dir, "build.gradle")) ||
		scanner.FileExists(filepath.Join(dir, "build.gradle.kts")) ||
		scanner.FileExists(filepath.Join(dir, "settings.gradle")) ||
		scanner.FileExists(filepath.Join(dir, "settings.gradle.kts"))
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
		// Android doesn't typically use CSS, but we might have color XMLs.
		// For now, we focus on UI files.
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

	root := &domain.FileNode{FilePath: projectRoot}
	for _, f := range allFiles {
		root.Children = append(root.Children, &domain.FileNode{FilePath: f})
	}

	return []scanner.PageTree{
		{
			Label: "Android App",
			Root:  root,
		},
	}
}
