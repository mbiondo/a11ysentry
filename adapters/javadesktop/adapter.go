package javadesktop

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type javaDesktopAdapter struct{}

func NewJavaDesktopAdapter() ports.Adapter {
	return &javaDesktopAdapter{}
}

func (a *javaDesktopAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
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

			ext := filepath.Ext(f)
			var nodes []domain.USN

			switch ext {
			case ".fxml":
				nodes = a.parseFXML(string(content), f)
			case ".java":
				nodes = a.parseJavaSwing(string(content), f)
			}

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

func (a *javaDesktopAdapter) flatten(node *domain.FileNode) []string {
	if node == nil {
		return nil
	}
	res := []string{node.FilePath}
	for _, c := range node.Children {
		res = append(res, a.flatten(c)...)
	}
	return res
}

func (a *javaDesktopAdapter) parseFXML(content string, filename string) []domain.USN {
	var nodes []domain.USN
	decoder := xml.NewDecoder(strings.NewReader(content))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			role := domain.SemanticRole("")
			label := ""

			name := se.Name.Local
			switch name {
			case "Button":
				role = domain.RoleButton
			case "ImageView":
				role = domain.RoleImage
			case "TextField", "PasswordField":
				role = domain.RoleInput
			}

			if role != "" {
				for _, attr := range se.Attr {
					// accessibleText is the key for JavaFX accessibility
					if attr.Name.Local == "accessibleText" || 
					   attr.Name.Local == "text" || 
					   attr.Name.Local == "promptText" {
						label = attr.Value
					}
				}

				nodes = append(nodes, domain.USN{
					UID:   fmt.Sprintf("%s-fxml-%s-%d", filename, name, decoder.InputOffset()),
					Role:  role,
					Label: label,
					Source: domain.Source{
						Platform: domain.PlatformJavaFX,
						FilePath: filename,
						Line:     1,
						RawHTML:  fmt.Sprintf("<%s ...>", name),
					},
				})
			}
		}
	}

	return nodes
}

func (a *javaDesktopAdapter) parseJavaSwing(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Look for .getAccessibleContext().setAccessibleName("...")
	swingRegex := regexp.MustCompile(`\.getAccessibleContext\(\)\.setAccessibleName\s*\(\s*\"([^\"]*)\"`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if matches := swingRegex.FindStringSubmatch(trimmed); matches != nil {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-swing-desc-%d", filename, i),
				Role:  domain.RoleImage, // Default for manual accessibility context
				Label: matches[1],
				Source: domain.Source{
					Platform: domain.PlatformJavaSwing,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "getAccessibleContext"),
					RawHTML:  trimmed,
				},
			})
		}
	}

	return nodes
}
