package blazor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"

	"golang.org/x/net/html"
)

type blazorAdapter struct{}

func NewBlazorAdapter() ports.Adapter {
	return &blazorAdapter{}
}

func (a *blazorAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		doc, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			// .razor files might have @code blocks that break standard HTML parsing
			// but golang.org/x/net/html is quite robust.
			// If it fails, we might need a more custom parser.
		}

		nodes := a.traverse(doc, file, string(content))
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *blazorAdapter) traverse(n *html.Node, filename string, fullContent string) []domain.USN {
	var nodes []domain.USN

	if n.Type == html.ElementNode {
		role := a.mapRole(n.Data)
		label := a.getLabel(n)

		if role != "" || label != "" {
			raw := a.renderNode(n)
			line, col := a.findPosition(raw, fullContent)

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-blazor-%s-%d", filename, n.Data, line),
				Role:  role,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformBlazor,
					FilePath: filename,
					Line:     line,
					Column:   col,
					RawHTML:  raw,
				},
			})
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, a.traverse(c, filename, fullContent)...)
	}

	return nodes
}

func (a *blazorAdapter) mapRole(tag string) domain.SemanticRole {
	// Standard HTML tags
	switch tag {
	case "button":
		return domain.RoleButton
	case "a":
		return domain.RoleLink
	case "img":
		return domain.RoleImage
	case "input":
		return domain.RoleInput
	}

	// Blazor Components
	if strings.HasPrefix(tag, "Input") {
		return domain.RoleInput
	}
	if strings.Contains(tag, "Button") {
		return domain.RoleButton
	}

	return ""
}

func (a *blazorAdapter) getLabel(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "aria-label" || attr.Key == "alt" || attr.Key == "Title" || attr.Key == "Placeholder" {
			return attr.Val
		}
	}
	return ""
}

func (a *blazorAdapter) renderNode(n *html.Node) string {
	var buf bytes.Buffer
	buf.WriteString("<" + n.Data)
	for _, attr := range n.Attr {
		buf.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Key, attr.Val))
	}
	buf.WriteString(">")
	return buf.String()
}

func (a *blazorAdapter) findPosition(raw string, fullContent string) (int, int) {
	idx := strings.Index(fullContent, raw)
	if idx == -1 {
		return 1, 1
	}

	line := 1
	col := 1
	for i := 0; i < idx; i++ {
		if fullContent[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
