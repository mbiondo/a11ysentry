package godot

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type godotAdapter struct{}

func NewGodotAdapter() ports.Adapter {
	return &godotAdapter{}
}

func (a *godotAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		nodes := a.parseTSCN(string(content), file)
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *godotAdapter) parseTSCN(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Regex for [node name="..." type="..."]
	nodeRegex := regexp.MustCompile(`\[node name=\"([^\"]*)\" type=\"([^\"]*)\"`)
	// Regex for text = "..."
	textRegex := regexp.MustCompile(`text\s*=\s*\"([^\"]*)\"`)
	// Regex for tooltip_text = "..."
	tooltipRegex := regexp.MustCompile(`tooltip_text\s*=\s*\"([^\"]*)\"`)

	var currentRole domain.SemanticRole
	var currentLabel string
	var currentLine int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := nodeRegex.FindStringSubmatch(trimmed); matches != nil {
			// Save previous node if it had data
			if currentRole != "" || currentLabel != "" {
				nodes = append(nodes, domain.USN{
					UID:   fmt.Sprintf("%s-godot-%d", filename, currentLine),
					Role:  currentRole,
					Label: currentLabel,
					Source: domain.Source{
						Platform: domain.PlatformGodot,
						FilePath: filename,
						Line:     currentLine + 1,
						RawHTML:  fmt.Sprintf("Godot Node: %s", currentRole),
					},
				})
			}

			// New node
			nodeType := matches[2]
			currentRole = a.mapGodotRole(nodeType)
			currentLabel = ""
			currentLine = i
		}

		if matches := textRegex.FindStringSubmatch(trimmed); matches != nil {
			currentLabel = matches[1]
		}
		if matches := tooltipRegex.FindStringSubmatch(trimmed); matches != nil && currentLabel == "" {
			currentLabel = matches[1]
		}
	}

	// Add last node
	if currentRole != "" || currentLabel != "" {
		nodes = append(nodes, domain.USN{
			UID:   fmt.Sprintf("%s-godot-%d", filename, currentLine),
			Role:  currentRole,
			Label: currentLabel,
			Source: domain.Source{
				Platform: domain.PlatformGodot,
				FilePath: filename,
				Line:     currentLine + 1,
				RawHTML:  "Godot Node",
			},
		})
	}

	return nodes
}

func (a *godotAdapter) mapGodotRole(nodeType string) domain.SemanticRole {
	switch nodeType {
	case "Button", "TextureButton", "CheckButton":
		return domain.RoleButton
	case "LinkButton":
		return domain.RoleLink
	case "Label":
		return domain.RoleHeading
	case "LineEdit", "TextEdit":
		return domain.RoleInput
	case "Sprite2D", "Sprite3D", "TextureRect":
		return domain.RoleImage
	default:
		return ""
	}
}
