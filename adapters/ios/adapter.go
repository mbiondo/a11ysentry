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

func (a *iosAdapter) flatten(node *domain.FileNode) []string {
	if node == nil {
		return nil
	}
	var res []string
	info, err := os.Stat(node.FilePath)
	if err == nil && !info.IsDir() {
		res = append(res, node.FilePath)
	}
	for _, c := range node.Children {
		res = append(res, a.flatten(c)...)
	}
	return res
}

func (a *iosAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
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

			src := string(content)
			// Try SwiftUI first, then UIKit (simple heuristic)
			nodes := a.parseSwiftUI(src, f)
			nodes = append(nodes, a.parseUIKit(src, f)...)
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

// parseSwiftUI identifies SwiftUI components and accessibility modifiers.
func (a *iosAdapter) parseSwiftUI(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	imageRegex := regexp.MustCompile(`Image\s*\(\s*\"([^\"]*)\"`)
	labelRegex := regexp.MustCompile(`\.accessibilityLabel\s*\(\s*\"([^\"]*)\"`)
	buttonRegex := regexp.MustCompile(`Button\s*\(\s*\"([^\"]*)\"`)
	buttonGenericRegex := regexp.MustCompile(`Button\s*\(`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := imageRegex.FindStringSubmatch(trimmed); matches != nil {
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

		if matches := buttonRegex.FindStringSubmatch(trimmed); matches != nil {
			label := matches[1]
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
		} else if buttonGenericRegex.MatchString(trimmed) {
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
