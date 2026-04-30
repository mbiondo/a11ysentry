// Package scanner provides framework-aware project scanning for A11ySentry.
// It detects the frontend framework used in a directory and decomposes the
// project into PageTrees — self-contained analysis units (e.g. layout chain +
// page + imported components) that the engine can audit as a whole.
package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// ProjectMarkers is the set of filenames that identify a project root.
var ProjectMarkers = []string{
	"package.json",      // JS/TS
	"build.gradle",      // Android (Groovy)
	"build.gradle.kts",  // Android (Kotlin)
	"settings.gradle",   // Android settings
	"settings.gradle.kts",
	"Package.swift",     // Swift Package Manager
}

// ProjectDirExtensions is the set of directory extensions that identify a project root.
var ProjectDirExtensions = []string{
	".xcodeproj",    // iOS
	".xcworkspace",  // iOS
}

// isProjectRoot returns true if dir contains any of the known project markers.
func isProjectRoot(dir string) bool {
	for _, marker := range ProjectMarkers {
		if FileExists(filepath.Join(dir, marker)) {
			return true
		}
	}
	for _, ext := range ProjectDirExtensions {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), ext) {
				return true
			}
		}
	}
	return false
}

// FindProjectRoots walks dir recursively and returns the absolute paths of
// directories that look like project roots (JS, Android, or iOS).
//
// Once a project root is found, the walk does NOT descend further into it —
// nested build artifacts or sub-projects (in some cases) are skipped.
//
// If dir itself is a project root it is returned immediately without further
// traversal (single-project fast path).
func FindProjectRoots(dir string) []string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	// Fast path: the given dir is itself a project root.
	if isProjectRoot(abs) {
		return []string{abs}
	}
	return collectRoots(abs)
}

// collectRoots is the recursive implementation of FindProjectRoots.
func collectRoots(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var roots []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if SkipDirs[entry.Name()] {
			continue
		}
		sub := filepath.Join(dir, entry.Name())
		if isProjectRoot(sub) {
			roots = append(roots, sub)
			// Don't descend — projects don't nest.
		} else {
			roots = append(roots, collectRoots(sub)...)
		}
	}
	return roots
}

// PageTree is a single analysis unit: an ordered set of files that form a
// complete rendering context for one route or view.
//
// Files are ordered so that wrapping contexts (layouts, providers) come before
// the page and its components. This ensures the analyzer sees the full document
// tree, including attributes like html[lang] that live in a root layout.
type PageTree struct {
	// Label is a human-readable identifier (e.g. "app/(app)/admin/page.tsx").
	Label string
	// Files holds the ordered list of absolute paths to analyze together.
	Files []string
}

// ProjectFramework knows how to decompose a specific frontend framework's
// project into PageTrees ready for accessibility analysis.
type ProjectFramework interface {
	// Name returns a display name for the framework (e.g. "Next.js App Router").
	Name() string

	// CollectFiles returns all UI component files and CSS/config files found
	// under dir. uiFiles are fed to the import graph builder; cssFiles are
	// pre-loaded for color resolution.
	CollectFiles(dir string) (uiFiles []string, cssFiles []string, err error)

	// ResolveImports returns the absolute paths of project-local files imported
	// by filePath. External packages (node_modules) are ignored.
	ResolveImports(filePath, projectRoot string, fileSet map[string]bool) []string

	// BuildPageTrees groups files into analysis units given the full import
	// graph (file → []imported abs paths).
	BuildPageTrees(allFiles []string, importGraph map[string][]string, projectRoot string) []PageTree
}

// Detectable extends ProjectFramework with framework detection capability.
type Detectable interface {
	ProjectFramework
	// Probe returns true if this framework owns the given directory.
	Probe(dir string) bool
}

// Detect inspects dir and returns the first framework in candidates whose
// Probe() method returns true. If none match, the last candidate is returned
// as the designated fallback (conventionally generic.New()).
//
// Usage (from cmd):
//
//	fw := scanner.Detect(dir,
//	    nextjs.New(),
//	    astro.New(),
//	    generic.New(), // fallback — returned when nothing else matches
//	)
func Detect(dir string, candidates ...Detectable) ProjectFramework {
	for _, fw := range candidates {
		if fw.Probe(dir) {
			return fw
		}
	}
	// Return the last candidate as the designated fallback.
	if len(candidates) > 0 {
		return candidates[len(candidates)-1]
	}
	return nil
}

// FileExists is a small helper used by framework detectors.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists returns true if path is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// AbsDir returns the absolute version of dir, or dir itself on error.
func AbsDir(dir string) string {
	a, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return a
}
