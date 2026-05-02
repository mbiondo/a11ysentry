package unity

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type unityAdapter struct{}

func NewUnityAdapter() ports.Adapter {
	return &unityAdapter{}
}

func (a *unityAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	files := a.flatten(root)
	var allNodes []domain.USN
	nodeChan := make(chan []domain.USN, len(files))
	errChan := make(chan error, len(files))

	for _, file := range files {
		go func(f string) {
			content, err := os.ReadFile(f)
			if err != nil {
				errChan <- err
				return
			}

			nodes := a.parseUnityYAML(string(content), f)
			nodeChan <- nodes
		}(file)
	}

	for i := 0; i < len(files); i++ {
		select {
		case nodes := <-nodeChan:
			allNodes = append(allNodes, nodes...)
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return allNodes, nil
}

func (a *unityAdapter) flatten(node *domain.FileNode) []string {
	if node == nil {
		return nil
	}
	res := []string{node.FilePath}
	for _, c := range node.Children {
		res = append(res, a.flatten(c)...)
	}
	return res
}

func (a *unityAdapter) parseUnityYAML(content string, filename string) []domain.USN {
	var nodes []domain.USN
	
	// Split by YAML document separator
	documents := strings.Split(content, "---")

	// Regex for m_Text: "..."
	textRegex := regexp.MustCompile(`m_Text:\s*\"?([^\n\"]*)\"?`)

	for i, doc := range documents {
		if strings.Contains(doc, "GameObject") || strings.Contains(doc, "Button") || strings.Contains(doc, "Text") || strings.Contains(doc, "Image") {
			label := ""
			role := domain.RoleLiveRegion // Default for generic objects

			if matches := textRegex.FindStringSubmatch(doc); matches != nil {
				label = matches[1]
			}

			if strings.Contains(doc, "Button") {
				role = domain.RoleButton
			} else if strings.Contains(doc, "Image") {
				role = domain.RoleImage
			} else if strings.Contains(doc, "Text") {
				role = domain.RoleHeading // Often texts are headings or labels
			}

			// If we have a label or a clear role, add it
			if label != "" || role != domain.RoleLiveRegion {
				nodes = append(nodes, domain.USN{
					UID:   fmt.Sprintf("%s-unity-doc-%d", filename, i),
					Role:  role,
					Label: label,
					Source: domain.Source{
						Platform: domain.PlatformUnity,
						FilePath: filename,
						Line:     i + 1, // Document index as a proxy for line
						RawHTML:  "Unity YAML Document",
					},
				})
			}
		}
	}

	return nodes
}
