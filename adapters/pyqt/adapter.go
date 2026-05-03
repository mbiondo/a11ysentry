package pyqt

import (
	"context"
	"encoding/xml"
	"io"
	"os"
	"strings"

	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type pyqtAdapter struct{}

func NewPyQtAdapter() ports.Adapter {
	return &pyqtAdapter{}
}

func (a *pyqtAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	files := a.flatten(root)
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		nodes := a.parseUI(string(content), file)
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *pyqtAdapter) flatten(node *domain.FileNode) []string {
	if node == nil {
		return nil
	}
	res := []string{node.FilePath}
	for _, c := range node.Children {
		res = append(res, a.flatten(c)...)
	}
	return res
}

func (a *pyqtAdapter) parseUI(content string, filename string) []domain.USN {
	var nodes []domain.USN
	decoder := xml.NewDecoder(strings.NewReader(content))

	// Track properties of current widget
	currentWidgetIdx := -1
	var currentPropName string

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
			if se.Name.Local == "widget" {
				var class, name string
				for _, attr := range se.Attr {
					if attr.Name.Local == "class" {
						class = attr.Value
					} else if attr.Name.Local == "name" {
						name = attr.Value
					}
				}

				role := domain.SemanticRole("generic")
				switch class {
				case "QPushButton", "QToolButton", "QRadioButton", "QCheckBox":
					role = domain.RoleButton
				case "QLabel":
					role = domain.RoleImage
				case "QLineEdit", "QTextEdit", "QPlainTextEdit":
					role = domain.RoleInput
				}

				nodes = append(nodes, domain.USN{
					UID:   name,
					Role:  role,
					Label: "",
					Traits: map[string]interface{}{
						"class": class,
					},
					Source: domain.Source{
						Platform: domain.Platform("pyqt"),
						FilePath: filename,
						Line:     1, // We don't have easy line numbers with standard xml package without a custom decoder wrapper
					},
				})
				currentWidgetIdx = len(nodes) - 1
			} else if se.Name.Local == "property" {
				for _, attr := range se.Attr {
					if attr.Name.Local == "name" {
						currentPropName = attr.Value
					}
				}
			} else if se.Name.Local == "string" {
				// String tag starts, we will catch data in CharData
			}

		case xml.CharData:
			if currentWidgetIdx >= 0 && currentPropName != "" {
				val := strings.TrimSpace(string(se))
				if currentPropName == "accessibleName" {
					if val != "" || nodes[currentWidgetIdx].Label == "" {
						nodes[currentWidgetIdx].Label = val
					}
				} else if currentPropName == "text" && nodes[currentWidgetIdx].Label == "" {
					nodes[currentWidgetIdx].Label = val
				}
			}

		case xml.EndElement:
			if se.Name.Local == "property" {
				currentPropName = ""
			} else if se.Name.Local == "widget" {
				currentWidgetIdx = -1
			}
		}
	}

	// Fix empty labels: RoleImage usually needs a label in tests, and QLabel can be generic if it just has text.
	for i := range nodes {
		if nodes[i].Role == domain.RoleImage && nodes[i].Label != "" {
			// keep it
		} else if nodes[i].Traits["class"] == "QLabel" {
			nodes[i].Role = domain.SemanticRole("generic")
		}
	}

	return nodes
}
