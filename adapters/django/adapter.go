package django

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type djangoAdapter struct {
	webAdapter ports.Adapter
}

func NewDjangoAdapter() ports.Adapter {
	return &djangoAdapter{
		webAdapter: web.NewHTMLAdapter(),
	}
}

func (a *djangoAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	if root == nil {
		return nil, nil
	}

	// 1. Build a map of all files in the tree for template resolution
	fileMap := make(map[string]string)
	err := buildFileMap(root, fileMap)
	if err != nil {
		return nil, err
	}

	// 2. Resolve templates starting from the root file
	resolvedHTML, err := resolveTemplate(root.FilePath, fileMap)
	if err != nil {
		return nil, err
	}

	// 3. Create a temporary file or use memory. 
	// The webAdapter requires a FileNode with a real file path because it uses os.ReadFile.
	// Since we expanded the template, we'll write the expanded HTML to a temp file
	// and feed that to the web adapter.
	tmpFile, err := os.CreateTemp("", "expanded_django_*.html")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name()) //nolint:errcheck

	if _, err := tmpFile.WriteString(resolvedHTML); err != nil {
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	// Create a dummy root node pointing to the expanded HTML
	expandedNode := &domain.FileNode{
		FilePath: tmpFile.Name(),
	}

	// Run the standard web adapter on the expanded HTML
	nodes, err := a.webAdapter.Ingest(ctx, expandedNode)
	if err != nil {
		return nil, err
	}

	// Fix the source FilePath to point back to the original template file instead of the temp file
	for i := range nodes {
		nodes[i].Source.FilePath = root.FilePath
		nodes[i].Source.Platform = domain.Platform("django") // Overwrite platform
	}

	return nodes, nil
}

func buildFileMap(node *domain.FileNode, fileMap map[string]string) error {
	data, err := os.ReadFile(node.FilePath)
	if err != nil {
		return err
	}
	fileMap[filepath.Base(node.FilePath)] = string(data)
	fileMap[node.FilePath] = string(data) // store by absolute path too

	for _, child := range node.Children {
		if err := buildFileMap(child, fileMap); err != nil {
			return err
		}
	}
	return nil
}

func resolveTemplate(filePath string, fileMap map[string]string) (string, error) {
	content, ok := fileMap[filePath]
	if !ok {
		content, ok = fileMap[filepath.Base(filePath)]
		if !ok {
			return "", nil // Not found, ignore
		}
	}

	// Simple Template Expansion
	return expandBlocks(content, fileMap)
}

func expandBlocks(content string, fileMap map[string]string) (string, error) {
	extendsRe := regexp.MustCompile(`(?i)\{%\s*extends\s+['"]([^'"]+)['"]\s*%\}`)
	blockRe := regexp.MustCompile(`(?s)\{%\s*block\s+([a-zA-Z0-9_]+)\s*%\}(.*?)\{%\s*endblock\s*(?:[a-zA-Z0-9_]+\s*)?%\}`)
	includeRe := regexp.MustCompile(`(?i)\{%\s*include\s+['"]([^'"]+)['"]\s*%\}`)

	// Extract blocks from current file
	blocks := make(map[string]string)
	for _, m := range blockRe.FindAllStringSubmatch(content, -1) {
		blocks[m[1]] = m[2]
	}

	// Handle extends
	if extendsMatch := extendsRe.FindStringSubmatch(content); extendsMatch != nil {
		parentName := extendsMatch[1]
		parentContent, ok := fileMap[parentName]
		if ok {
			// Resolve includes in parent first (or later, order might matter but simple is ok)
			// Replace blocks in parent with child's blocks
			parentContent = blockRe.ReplaceAllStringFunc(parentContent, func(match string) string {
				sm := blockRe.FindStringSubmatch(match)
				blockName := sm[1]
				if childContent, exists := blocks[blockName]; exists {
					return childContent
				}
				return sm[2] // default block content
			})
			content = parentContent
		}
	}

	// Handle includes
	content = includeRe.ReplaceAllStringFunc(content, func(match string) string {
		sm := includeRe.FindStringSubmatch(match)
		includeName := sm[1]
		if includeContent, ok := fileMap[includeName]; ok {
			// Expand includes recursively
			expanded, _ := expandBlocks(includeContent, fileMap)
			return expanded
		}
		return ""
	})

	// Strip remaining template tags that might confuse the HTML parser
	tagRe := regexp.MustCompile(`\{%[^%]+%\}`)
	content = tagRe.ReplaceAllString(content, "")

	varRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	// Replace variables with a safe placeholder to avoid breaking HTML structure
	content = varRe.ReplaceAllString(content, "var_placeholder")

	return content, nil
}

// Make sure the htmlAdapter allows overriding platform. Wait, htmlAdapter's struct fields are private.
// But we just overwrite usn.Source.Platform in the returned USNs, which is fine since USN is public.

