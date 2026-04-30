package scanner

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UIExtensions is the set of file extensions that contain UI markup.
var UIExtensions = map[string]bool{
	".html": true, ".htm": true,
	".astro":  true,
	".vue":    true,
	".svelte": true,
	".tsx":    true,
	".jsx":    true,
	// Android
	".xml":  true,
	".kt":   true,
	".java": true,
	// iOS
	".swift":      true,
	".storyboard": true,
	".xib":        true,
}

// CSSExtensions is the set of file extensions treated as stylesheets.
var CSSExtensions = map[string]bool{
	".css": true, ".scss": true,
}

// SkipDirs is the set of directory names to skip during filesystem walks.
var SkipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	"out":          true,
	".astro":       true,
	".next":        true,
	".nuxt":        true,
	".svelte-kit":  true,
	// Android / Mobile
	".gradle":     true,
	".idea":       true,
	"DerivedData": true,
	".build":      true,
	// Demo / Internal
	"examples": true,
	"landing":  true,
}

// importRe matches ES module static imports and re-exports.
var importRe = regexp.MustCompile(`(?m)(?:import|export)\s+(?:[^'"` + "`" + `\n]+\s+from\s+)?['"]([^'"` + "`" + `]+)['"]`)

// dynamicImportRe matches dynamic imports: import('...')
var dynamicImportRe = regexp.MustCompile(`(?m)import\(\s*['"]([^'"]+)['"]\s*\)`)

// kotlinImportRe matches Kotlin/Java imports: import com.example.MyComponent
var kotlinImportRe = regexp.MustCompile(`(?m)^import\s+([a-zA-Z0-9._]+)`)

// swiftImportRe matches Swift imports: import MyModule
var swiftImportRe = regexp.MustCompile(`(?m)^import\s+([a-zA-Z0-9_]+)`)

// ResolveImports returns the absolute paths of project-local files imported by
// filePath. It automatically detects the language based on file extension.
func ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	src := string(data)
	var resolved []string
	seen := make(map[string]bool)

	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			resolved = append(resolved, p)
		}
	}

	switch ext {
	case ".kt", ".java":
		// Kotlin/Java imports are tricky because they use package names.
		// For now, we do a simple heuristic: if a file in the fileSet
		// matches the end of the import path, we include it.
		// In a real Android project, we'd need to map package names to directories.
		for _, m := range kotlinImportRe.FindAllStringSubmatch(src, -1) {
			importPath := m[1]
			parts := strings.Split(importPath, ".")
			className := parts[len(parts)-1]

			// Look for files that end with /ClassName.kt or /ClassName.java
			for f := range fileSet {
				base := filepath.Base(f)
				if strings.HasPrefix(base, className+".") {
					add(f)
				}
			}
		}

	case ".swift":
		for _, m := range swiftImportRe.FindAllStringSubmatch(src, -1) {
			moduleName := m[1]
			// Similar heuristic for Swift.
			for f := range fileSet {
				if strings.Contains(f, "/"+moduleName+"/") || filepath.Base(f) == moduleName+".swift" {
					add(f)
				}
			}
		}

	default:
		// JS/TS/Web fallback
		aliases := LoadTSConfigPaths(projectRoot)
		base := filepath.Dir(filePath)

		tryResolve := func(importPath string) {
			if strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/") {
				abs := filepath.Clean(filepath.Join(base, importPath))
				if fileSet[abs] {
					add(abs)
					return
				}
				for uiExt := range UIExtensions {
					if fileSet[abs+uiExt] {
						add(abs + uiExt)
						return
					}
				}
			} else {
				if p := resolveAlias(importPath, aliases, projectRoot, fileSet); p != "" {
					add(p)
				}
			}
		}

		for _, m := range importRe.FindAllStringSubmatch(src, -1) {
			tryResolve(m[1])
		}
		for _, m := range dynamicImportRe.FindAllStringSubmatch(src, -1) {
			tryResolve(m[1])
		}
	}

	return resolved
}

// CollectTree returns the full transitive closure of root's import graph.
func CollectTree(root string, graph map[string][]string, visited map[string]bool) []string {
	if visited[root] {
		return nil
	}
	visited[root] = true
	result := []string{root}
	for _, dep := range graph[root] {
		result = append(result, CollectTree(dep, graph, visited)...)
	}
	return result
}

// BuildFileSet builds a fast-lookup set from a list of absolute paths.
func BuildFileSet(files []string) map[string]bool {
	s := make(map[string]bool, len(files))
	for _, f := range files {
		s[f] = true
	}
	return s
}

// BuildImportGraph constructs the full import graph for allFiles.
func BuildImportGraph(allFiles []string, fw ProjectFramework, projectRoot string) map[string][]string {
	fileSet := BuildFileSet(allFiles)
	graph := make(map[string][]string, len(allFiles))
	for _, f := range allFiles {
		graph[f] = fw.ResolveImports(f, projectRoot, fileSet)
	}
	return graph
}

// ─────────────────────────────────────────────────────────────────────────────
// tsconfig / jsconfig path alias resolution
// ─────────────────────────────────────────────────────────────────────────────

// TSConfigPaths holds the parsed path aliases from tsconfig.json / jsconfig.json.
type TSConfigPaths struct {
	BaseURL string
	Paths   map[string]string // alias prefix → target dir (relative to BaseURL)
}

// LoadTSConfigPaths reads tsconfig.json or jsconfig.json from projectRoot and
// extracts compilerOptions.baseUrl and compilerOptions.paths. Returns nil if no
// config is found or if it contains no alias information.
func LoadTSConfigPaths(projectRoot string) *TSConfigPaths {
	for _, name := range []string{"tsconfig.json", "jsconfig.json"} {
		data, err := os.ReadFile(filepath.Join(projectRoot, name))
		if err != nil {
			continue
		}
		result := &TSConfigPaths{Paths: make(map[string]string)}

		baseRe := regexp.MustCompile(`"baseUrl"\s*:\s*"([^"]+)"`)
		if m := baseRe.FindStringSubmatch(string(data)); len(m) > 1 {
			result.BaseURL = m[1]
		}

		// Matches entries like: "@/*": ["./src/*"] or "@/components": ["components"]
		pathRe := regexp.MustCompile(`"([@~][^"]*)"\s*:\s*\[\s*"([^"]+)"`)
		for _, m := range pathRe.FindAllStringSubmatch(string(data), -1) {
			alias := m[1]
			target := strings.TrimSuffix(m[2], "/*")
			target = strings.TrimPrefix(target, "./")
			result.Paths[alias] = target
		}

		if len(result.Paths) > 0 || result.BaseURL != "" {
			return result
		}
	}
	return nil
}

func resolveAlias(importPath string, aliases *TSConfigPaths, projectRoot string, fileSet map[string]bool) string {
	if aliases == nil {
		return ""
	}
	for alias, target := range aliases.Paths {
		prefix := strings.TrimSuffix(alias, "/*")
		if strings.HasPrefix(importPath, prefix+"/") || importPath == prefix {
			suffix := strings.TrimPrefix(importPath, prefix+"/")
			candidate := filepath.Clean(filepath.Join(projectRoot, aliases.BaseURL, target, suffix))
			if fileSet[candidate] {
				return candidate
			}
			for ext := range UIExtensions {
				if fileSet[candidate+ext] {
					return candidate + ext
				}
			}
		}
	}
	return ""
}
