package dotnet

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

type dotNetAdapter struct{}

func NewDotNetAdapter() ports.Adapter {
	return &dotNetAdapter{}
}

func (a *dotNetAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		ext := filepath.Ext(file)
		var nodes []domain.USN

		switch ext {
		case ".xaml":
			nodes = a.parseXAML(string(content), file)
		case ".cs":
			nodes = a.parseCSharp(string(content), file)
		}
		
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *dotNetAdapter) parseXAML(content string, filename string) []domain.USN {
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

			// Map common .NET controls to roles
			name := se.Name.Local
			switch name {
			case "Button":
				role = domain.RoleButton
			case "Image":
				role = domain.RoleImage
			case "TextBox", "Entry", "PasswordBox":
				role = domain.RoleInput
			case "TextBlock", "Label":
				// Used as labels for other elements
			}

			if role != "" {
				for _, attr := range se.Attr {
					// AutomationProperties.Name is the key for .NET accessibility
					if strings.Contains(attr.Name.Local, "AutomationProperties.Name") || 
					   attr.Name.Local == "Content" || 
					   attr.Name.Local == "Text" || 
					   attr.Name.Local == "Placeholder" {
						label = attr.Value
					}
				}

				nodes = append(nodes, domain.USN{
					UID:   fmt.Sprintf("%s-xaml-%s-%d", filename, name, decoder.InputOffset()),
					Role:  role,
					Label: label,
					Source: domain.Source{
						Platform: domain.PlatformDotNetXAML,
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

func (a *dotNetAdapter) parseCSharp(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Look for AutomationProperties.SetName(control, "...")
	setRegex := regexp.MustCompile(`AutomationProperties\.SetName\s*\([^,]*,\s*\"([^\"]*)\"`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if matches := setRegex.FindStringSubmatch(trimmed); matches != nil {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-cs-automation-%d", filename, i),
				Role:  domain.RoleButton, // Generic for manual setters
				Label: matches[1],
				Source: domain.Source{
					Platform: domain.PlatformDotNetCSharp,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "AutomationProperties"),
					RawHTML:  trimmed,
				},
			})
		}
	}

	return nodes
}
