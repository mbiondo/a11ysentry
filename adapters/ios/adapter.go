package ios

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
)

type iosAdapter struct{}

func NewIOSAdapter() ports.Adapter {
	return &iosAdapter{}
}

func (a *iosAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		src := string(content)
		// Try SwiftUI first, then UIKit (simple heuristic)
		nodes := a.parseSwiftUI(src, file)
		nodes = append(nodes, a.parseUIKit(src, file)...)
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

// parseSwiftUI identifies SwiftUI components and accessibility modifiers.
func (a *iosAdapter) parseSwiftUI(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	imageRegex := regexp.MustCompile(`Image\s*\(\s*\"([^\"]*)\"`)
	labelRegex := regexp.MustCompile(`\.accessibilityLabel\s*\(\s*\"([^\"]*)\"`)
	buttonRegex := regexp.MustCompile(`Button\s*\(`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := imageRegex.FindStringSubmatch(trimmed); matches != nil {
// ... (omitted for brevity in instruction, but keep logic)
			imageName := matches[1]
			label := "" 
			lookAhead := 3
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				if m := labelRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
			}
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-swiftui-img-%d", filename, i),
				Role:  domain.RoleImage,
				Label: label,
				Traits: map[string]any{"imageName": imageName},
				Source: domain.Source{
					Platform: domain.PlatformIOSSwiftUI,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Image"),
					RawHTML:  trimmed,
				},
			})
		}

		if buttonRegex.MatchString(trimmed) {
			label := ""
			// Look ahead for accessibilityLabel override
			lookAhead := 5
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				if m := labelRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
			}
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-swiftui-btn-%d", filename, i),
				Role:  domain.RoleButton,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformIOSSwiftUI,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Button"),
					RawHTML:  trimmed,
				},
			})
		}
	}
	return nodes
}

func (a *iosAdapter) parseUIKit(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	uiImageRegex := regexp.MustCompile(`UIImageView\s*\(\s*image:\s*`)
	uiButtonRegex := regexp.MustCompile(`UIButton\s*\(\s*type:\s*`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if uiImageRegex.MatchString(trimmed) {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-uikit-img-%d", filename, i),
				Role:  domain.RoleImage,
				Label: "", // UIKit requires explicit accessibilityLabel
				Source: domain.Source{
					Platform: domain.PlatformIOSSwiftUI, // Using SwiftUI platform for now as a catch-all
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "UIImageView"),
					RawHTML:  trimmed,
				},
			})
		}
		if uiButtonRegex.MatchString(trimmed) {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-uikit-btn-%d", filename, i),
				Role:  domain.RoleButton,
				Label: "", 
				Source: domain.Source{
					Platform: domain.PlatformIOSSwiftUI,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "UIButton"),
					RawHTML:  trimmed,
				},
			})
		}
	}
	return nodes
}
