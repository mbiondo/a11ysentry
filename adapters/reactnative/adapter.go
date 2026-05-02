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

func (a *reactNativeAdapter) flatten(node *domain.FileNode) []string {
	if node == nil {
		return nil
	}
	var res []string
	if node.FilePath != "" {
		info, err := os.Stat(node.FilePath)
		if err == nil && !info.IsDir() {
			res = append(res, node.FilePath)
		}
	}
	for _, c := range node.Children {
		res = append(res, a.flatten(c)...)
	}
	return res
}

func (a *reactNativeAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	files := a.flatten(root)
	var allNodes []domain.USN
	if len(files) == 0 {
		return nil, nil
	}
	nodeChan := make(chan []domain.USN, len(files))
	errChan := make(chan error, len(files))

	for _, file := range files {
		go func(f string) {
			content, err := os.ReadFile(f)
			if err != nil {
				errChan <- err
				return
			}

			nodes := a.parseReactNative(string(content), f)
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

func (a *reactNativeAdapter) parseReactNative(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Regex for accessibilityLabel="..." or accessibilityLabel={'...'}
	labelRegex := regexp.MustCompile(`accessibilityLabel\s*=\s*[\"{]([^\"]*)[\"}]`)
	// Regex for accessibilityRole="..."
	roleRegex := regexp.MustCompile(`accessibilityRole\s*=\s*[\"{]([^\"]*)[\"}]`)
	// Regex for alt="..." (sometimes used in newer RN or web-compat)
	altRegex := regexp.MustCompile(`alt\s*=\s*[\"{]([^\"]*)[\"}]`)
	// Hidden traits
	hiddenAndroidRegex := regexp.MustCompile(`importantForAccessibility\s*=\s*[\"{]no-hide-descendants[\"}]`)
	hiddenIOSRegex := regexp.MustCompile(`accessibilityElementsHidden\s*=\s*[\"{]true[\"}]`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip import statements and other noise
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
		
		// Container with potential hidden trait
		isView := strings.Contains(trimmed, "<View")

		if isButton || isImage || isView {
			label := ""
			role := domain.SemanticRole("generic")
			if isButton {
				role = domain.RoleButton
			} else if isImage {
				role = domain.RoleImage
			}

			// Look for labels and traits in current or nearby lines
			lookAhead := 10
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

			traits := make(map[string]any)
			if hiddenAndroidRegex.MatchString(contextBlock) || hiddenIOSRegex.MatchString(contextBlock) {
				traits["aria-hidden"] = "true"
			}

			// Only add if it's a component we care about or it has a hidden trait
			if isButton || isImage || (isView && traits["aria-hidden"] == "true") {
				nodes = append(nodes, domain.USN{
					UID:    fmt.Sprintf("%s-rn-%d", filename, i),
					Role:   role,
					Label:  label,
					Traits: traits,
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
	}

	return nodes
}
