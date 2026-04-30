package reactnative

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type reactNativeAdapter struct{}

func NewReactNativeAdapter() ports.Adapter {
	return &reactNativeAdapter{}
}

func (a *reactNativeAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		nodes := a.parseJS(string(content), file)
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *reactNativeAdapter) parseJS(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Regex for accessibilityLabel="..." or accessibilityLabel={'...'}
	labelRegex := regexp.MustCompile(`accessibilityLabel\s*=\s*[\"{]([^\"]*)[\"}]`)
	// Regex for accessibilityRole="..."
	roleRegex := regexp.MustCompile(`accessibilityRole\s*=\s*[\"{]([^\"]*)[\"}]`)
	// Regex for alt="..." (sometimes used in newer RN or web-compat)
	altRegex := regexp.MustCompile(`alt\s*=\s*[\"{]([^\"]*)[\"}]`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip import statements, comment lines, and closing JSX tags
		if strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "{/*") ||
			strings.HasPrefix(trimmed, "</") {
			continue
		}
		isButton := strings.Contains(trimmed, "TouchableOpacity") ||
			strings.Contains(trimmed, "TouchableHighlight") ||
			strings.Contains(trimmed, "Pressable") ||
			strings.Contains(trimmed, "Button")

		isImage := strings.Contains(trimmed, "Image") && !strings.Contains(trimmed, "ImageSource")

		if isButton || isImage {
			label := ""
			role := domain.RoleButton
			if isImage {
				role = domain.RoleImage
			}

			// Look for labels in current or nearby lines (React components often span multiple lines)
			lookAhead := 5
			contextBlock := ""
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				contextBlock += lines[j] + " "
			}

			if matches := labelRegex.FindStringSubmatch(contextBlock); matches != nil {
				label = matches[1]
			} else if matches := altRegex.FindStringSubmatch(contextBlock); matches != nil {
				label = matches[1]
			}

			if matches := roleRegex.FindStringSubmatch(contextBlock); matches != nil {
				// Map RN roles to domain roles if possible
				rnRole := matches[1]
				switch rnRole {
				case "button":
					role = domain.RoleButton
				case "image":
					role = domain.RoleImage
				case "link":
					role = domain.RoleLink
				case "header":
					role = domain.RoleHeading
				}
			}

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-rn-%d", filename, i),
				Role:  role,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformReactNative,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, trimmed),
					RawHTML:  trimmed,
				},
			})
		}
	}

	return nodes
}
