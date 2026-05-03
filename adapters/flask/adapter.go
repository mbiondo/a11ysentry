package flask

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type flaskAdapter struct {
	webAdapter ports.Adapter
}

func NewFlaskAdapter() ports.Adapter {
	return &flaskAdapter{
		webAdapter: web.NewHTMLAdapter(),
	}
}

func (a *flaskAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	if root == nil {
		return nil, nil
	}

	fileMap := make(map[string]string)
	err := buildFileMap(root, fileMap)
	if err != nil {
		return nil, err
	}

	resolvedHTML, err := resolveTemplate(root.FilePath, fileMap)
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "expanded_flask_*.html")
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

	expandedNode := &domain.FileNode{
		FilePath: tmpFile.Name(),
	}

	nodes, err := a.webAdapter.Ingest(ctx, expandedNode)
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		nodes[i].Source.FilePath = root.FilePath
		nodes[i].Source.Platform = domain.Platform("flask")
	}

	return nodes, nil
}

func buildFileMap(node *domain.FileNode, fileMap map[string]string) error {
	data, err := os.ReadFile(node.FilePath)
	if err != nil {
		return err
	}
	fileMap[filepath.Base(node.FilePath)] = string(data)
	fileMap[node.FilePath] = string(data)

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
			return "", nil
		}
	}
	return expandBlocks(content, fileMap)
}

func expandBlocks(content string, fileMap map[string]string) (string, error) {
	extendsRe := regexp.MustCompile(`(?i)\{%\s*extends\s+['"]([^'"]+)['"]\s*%\}`)
	blockRe := regexp.MustCompile(`(?s)\{%\s*block\s+([a-zA-Z0-9_]+)\s*%\}(.*?)\{%\s*endblock\s*(?:[a-zA-Z0-9_]+\s*)?%\}`)
	includeRe := regexp.MustCompile(`(?i)\{%\s*include\s+['"]([^'"]+)['"]\s*%\}`)

	blocks := make(map[string]string)
	for _, m := range blockRe.FindAllStringSubmatch(content, -1) {
		blocks[m[1]] = m[2]
	}

	if extendsMatch := extendsRe.FindStringSubmatch(content); extendsMatch != nil {
		parentName := extendsMatch[1]
		parentContent, ok := fileMap[parentName]
		if ok {
			parentContent = blockRe.ReplaceAllStringFunc(parentContent, func(match string) string {
				sm := blockRe.FindStringSubmatch(match)
				blockName := sm[1]
				if childContent, exists := blocks[blockName]; exists {
					return childContent
				}
				return sm[2]
			})
			content = parentContent
		}
	}

	content = includeRe.ReplaceAllStringFunc(content, func(match string) string {
		sm := includeRe.FindStringSubmatch(match)
		includeName := sm[1]
		if includeContent, ok := fileMap[includeName]; ok {
			expanded, _ := expandBlocks(includeContent, fileMap)
			return expanded
		}
		return ""
	})

	tagRe := regexp.MustCompile(`\{%[^%]+%\}`)
	content = tagRe.ReplaceAllString(content, "")

	varRe := regexp.MustCompile(`\{\{[^}]+\}\}`)
	content = varRe.ReplaceAllString(content, "var_placeholder")

	return content, nil
}
