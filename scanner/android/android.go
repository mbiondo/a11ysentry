package android

import (
	"a11ysentry/scanner"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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

var (
	kotlinImportRe      = regexp.MustCompile(`(?m)^import\s+([\w.]+)`)
	androidLayoutRe     = regexp.MustCompile(`R\.layout\.(\w+)`)
	androidXmlIncludeRe = regexp.MustCompile(`<include\s+layout="@layout/(\w+)"`)
	kotlinComposableRe  = regexp.MustCompile(`([A-Z]\w+)\(`)
	kotlinDefRe         = regexp.MustCompile(`(?m)^fun\s+([A-Z]\w+)\s*\(`)
)

// ResolveImports implements Android-specific import resolution.
func (f *Framework) ResolveImports(filePath, projectRoot string, fileSet map[string]bool, aliases *scanner.TSConfigPaths) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	var resolved []string
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Fast lookup map for class names and Composables in the project
	symbolMap := make(map[string]string)
	for f := range fileSet {
		c, _ := os.ReadFile(f)
		defs := kotlinDefRe.FindAllStringSubmatch(string(c), -1)
		for _, d := range defs {
			symbolMap[d[1]] = f
		}
		// Also keep filename as fallback for classes
		base := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		symbolMap[base] = f
	}

	switch ext {
	case ".kt", ".java":
		// 1. Handle Kotlin/Java package imports
		matches := kotlinImportRe.FindAllStringSubmatch(src, -1)
		for _, m := range matches {
			pkgPath := strings.ReplaceAll(m[1], ".", string(filepath.Separator))
			prefixes := []string{
				filepath.Join(projectRoot, "app", "src", "main", "java"),
				filepath.Join(projectRoot, "app", "src", "main", "kotlin"),
				filepath.Join(projectRoot, "src", "main", "java"),
				filepath.Join(projectRoot, "src", "main", "kotlin"),
			}
			for _, prefix := range prefixes {
				candidate := filepath.Join(prefix, pkgPath)
				for _, ext := range []string{".kt", ".java"} {
					if fileSet[candidate+ext] {
						resolved = append(resolved, candidate+ext)
					}
				}
			}
		}

		// 2. Handle R.layout.xxx -> res/layout/xxx.xml
		layoutMatches := androidLayoutRe.FindAllStringSubmatch(src, -1)
		for _, m := range layoutMatches {
			layoutName := m[1]
			prefixes := []string{
				filepath.Join(projectRoot, "app", "src", "main", "res", "layout"),
				filepath.Join(projectRoot, "src", "main", "res", "layout"),
			}
			for _, prefix := range prefixes {
				candidate := filepath.Join(prefix, layoutName+".xml")
				if fileSet[candidate] {
					resolved = append(resolved, candidate)
				}
			}
		}
		
		// 3. Handle Composable/Component calls in same package or via wildcards
		compMatches := kotlinComposableRe.FindAllStringSubmatch(src, -1)
		for _, m := range compMatches {
			name := m[1]
			if realPath, ok := symbolMap[name]; ok {
				if realPath != filePath {
					resolved = append(resolved, realPath)
				}
			}
		}
	case ".xml":
		// 4. Handle XML <include layout="@layout/xxx" />
		includeMatches := androidXmlIncludeRe.FindAllStringSubmatch(src, -1)
		for _, m := range includeMatches {
			layoutName := m[1]
			prefixes := []string{
				filepath.Join(projectRoot, "app", "src", "main", "res", "layout"),
				filepath.Join(projectRoot, "src", "main", "res", "layout"),
			}
			for _, prefix := range prefixes {
				candidate := filepath.Join(prefix, layoutName+".xml")
				if fileSet[candidate] {
					resolved = append(resolved, candidate)
				}
			}
		}
	}

	return resolved
}

// BuildPageTrees identifies Activities/Composables and builds their full import trees.
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
		// Canonical Android entry point detection:
		// 1. Files not imported by anyone.
		// 2. Activities that define the UI root (setContent/setContentView).
		
		content, _ := os.ReadFile(file)
		src := string(content)
		
		isUIEntryPoint := strings.Contains(src, "setContent") || 
			strings.Contains(src, "setContentView") ||
			strings.Contains(file, "MainActivity")
		
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
