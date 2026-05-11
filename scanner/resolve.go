package scanner

import (
	"a11ysentry/engine/core/domain"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// UIExtensions is the set of file extensions that contain UI markup.
var UIExtensions = map[string]bool{
	".html":   true,
	".htm":    true,
	".astro":  true,
	".vue":    true,
	".svelte": true,
	".tsx":    true,
	".jsx":    true,
	".razor":  true, // Blazor
	".kt":     true, // Android Compose
	".xml":    true, // Android Views / Layouts
	".dart":   true, // Flutter
	".swift":  true, // iOS SwiftUI / UIKit
	".xaml":   true, // .NET MAUI / WPF
	".cs":     true, // .NET code-behind
	".fxml":   true, // JavaFX
	".java":   true, // Java Swing / Android
}

// CSSExtensions is the set of file extensions that contain CSS/SCSS.
var CSSExtensions = map[string]bool{
	".css":  true,
	".scss": true,
	".sass": true,
	".less": true,
}

var (
	importRe        = regexp.MustCompile(`(?m)import\s+(?:([\w*\s{},]*)\s+from\s+)?['"]([^'"]+)['"]`)
	dynamicImportRe = regexp.MustCompile(`import\s*\(['"]([^'"]+)['"]\)`)
	requireRe       = regexp.MustCompile(`require\s*\(['"]([^'"]+)['"]\)`)
)

func getUsedUIComponents(identifiers string, src string) []string {
	var used []string
	// Clean identifiers: remove symbols and keywords
	clean := strings.NewReplacer("{", " ", "}", " ", "*", " ", ",", " ", "as", " ").Replace(identifiers)
	parts := strings.Fields(clean)
	for _, p := range parts {
		if len(p) > 0 && p[0] >= 'A' && p[0] <= 'Z' {
			// Intelligent check: is this identifier actually used as a UI component tag?
			// React/Astro/Svelte/Vue/Solid use `<Component` or `</Component>`.
			tagOpen := "<" + p
			tagClose := "</" + p + ">"
			if strings.Contains(src, tagOpen) || strings.Contains(src, tagClose) {
				used = append(used, p)
			}
		}
	}
	return used
}

// ResolveImports returns the absolute paths of project-local files imported
// by filePath. External packages (node_modules) are marked as pkg:// if they
// likely contain UI components.
func ResolveImports(filePath string, projectRoot string, fileSet map[string]bool, aliases *TSConfigPaths) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	src := string(content)

	var resolved []string
	
	// Fast case-insensitive lookup map for Windows
	lowerFileSet := make(map[string]string)
	for f := range fileSet {
		lowerFileSet[strings.ToLower(f)] = f
	}

	add := func(p string) {
		p = filepath.Clean(p)
		if realPath, ok := lowerFileSet[strings.ToLower(p)]; ok {
			resolved = append(resolved, realPath)
		}
	}

	// Default: Generic resolution (Regex-based)
	base := filepath.Dir(filePath)

	tryResolve := func(importPath string, identifiers string) {
		var abs string
		if strings.HasPrefix(importPath, "/") {
			// Web convention: / is project root. 
			// On Windows, filepath.Join with leading slash can be tricky, 
			// so we trim it to ensure it's treated as project-relative.
			abs = filepath.Join(projectRoot, strings.TrimPrefix(importPath, "/"))
		} else if strings.HasPrefix(importPath, ".") {
			abs = filepath.Join(base, importPath)
		} else {
			if p := resolveAlias(importPath, aliases, projectRoot, fileSet); p != "" {
				// resolveAlias should return the correctly cased path from fileSet
				resolved = append(resolved, p)
			} else {
				// If it's not a local path or alias, but looks like a package (e.g. @org/pkg or pkg)
				// we mark it as an opaque package reference ONLY if it looks like it provides components.
				if (!strings.Contains(importPath, "/") || strings.HasPrefix(importPath, "@")) {
					usedComponents := getUsedUIComponents(identifiers, src)
					for _, comp := range usedComponents {
						resolved = append(resolved, "pkg://"+importPath+"/"+comp)
					}
				}
			}
			return
		}

		// Try the path as is
		add(abs)

		// Try with UI extensions if not found
		lowerAbs := strings.ToLower(filepath.Clean(abs))
		if _, ok := lowerFileSet[lowerAbs]; !ok {
			for uiExt := range UIExtensions {
				add(abs + uiExt)
			}
			// Also try with .ts/.js for utility resolution
			for _, ext := range []string{".ts", ".js", ".tsx", ".jsx"} {
				add(abs + ext)
			}
		}
	}

	for _, m := range importRe.FindAllStringSubmatch(src, -1) {
		tryResolve(m[2], m[1])
	}
	for _, m := range dynamicImportRe.FindAllStringSubmatch(src, -1) {
		tryResolve(m[1], "")
	}
	for _, m := range requireRe.FindAllStringSubmatch(src, -1) {
		tryResolve(m[1], "")
	}

	return resolved
}

// CollectTree returns the full transitive closure of rootPath's import graph as a tree.
// It prevents infinite recursion using a path-based cycle detection (visited stack).
func CollectTree(rootPath string, graph map[string][]string, pathStack map[string]bool) *domain.FileNode {
	if strings.HasPrefix(rootPath, "pkg://") {
		return &domain.FileNode{
			FilePath:     rootPath,
			IsOpaque:     true,
			OpaqueSource: strings.TrimPrefix(rootPath, "pkg://"),
		}
	}

	if pathStack[rootPath] {
		// Cycle detected — return a node marked as cycle to stop recursion.
		return &domain.FileNode{FilePath: rootPath, IsCycle: true}
	}

	// Clone the path stack for this branch to allow convergence without cycles.
	newStack := make(map[string]bool, len(pathStack)+1)
	for k, v := range pathStack {
		newStack[k] = v
	}
	newStack[rootPath] = true

	node := &domain.FileNode{FilePath: rootPath}
	for _, dep := range graph[rootPath] {
		child := CollectTree(dep, graph, newStack)
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node
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
	var mu sync.Mutex
	var wg sync.WaitGroup

	aliases := LoadTSConfigPaths(projectRoot)

	for _, f := range allFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			deps := fw.ResolveImports(file, projectRoot, fileSet, aliases)
			mu.Lock()
			graph[file] = deps
			mu.Unlock()
		}(f)
	}

	wg.Wait()
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
			for uiExt := range UIExtensions {
				if fileSet[candidate+uiExt] {
					return candidate + uiExt
				}
			}
			for _, ext := range []string{".ts", ".js", ".tsx", ".jsx"} {
				if fileSet[candidate+ext] {
					return candidate + ext
				}
			}
		}
	}
	return ""
}
