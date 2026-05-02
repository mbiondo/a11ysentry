package flutter

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type flutterAdapter struct{}

func NewFlutterAdapter() ports.Adapter {
	return &flutterAdapter{}
}

func (a *flutterAdapter) flatten(node *domain.FileNode) []string {
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

func (a *flutterAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
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

			nodes := a.parseDart(string(content), f)
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

func (a *flutterAdapter) parseDart(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	semanticsStartRegex := regexp.MustCompile(`Semantics\s*\(`)
	labelRegex := regexp.MustCompile(`label\s*:\s*\"([^\"]*)\"`)
	textRegex := regexp.MustCompile(`Text\s*\(\s*\"([^\"]*)\"`)
	imageStartRegex := regexp.MustCompile(`Image\.(asset|network|file|memory)\s*\(`)
	semanticLabelRegex := regexp.MustCompile(`semanticLabel\s*:\s*\"([^\"]*)\"`)
	excludeSemanticsRegex := regexp.MustCompile(`excludeSemantics\s*:\s*true`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Match Semantics
		if semanticsStartRegex.MatchString(trimmed) {
			label := ""
			isExclude := false
			lookAhead := 10
			block := ""
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				block += lines[j] + " "
			}
			
			if m := labelRegex.FindStringSubmatch(block); m != nil {
				label = m[1]
			}
			if excludeSemanticsRegex.MatchString(block) {
				isExclude = true
			}

			traits := make(map[string]any)
			if isExclude {
				traits["aria-hidden"] = "true"
			}

			nodes = append(nodes, domain.USN{
				UID:    fmt.Sprintf("%s-flutter-semantics-%d", filename, i),
				Role:   domain.RoleLiveRegion,
				Label:  label,
				Traits: traits,
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Semantics"),
					RawHTML:  trimmed,
				},
			})
		}

		// Match Images
		if imageStartRegex.MatchString(trimmed) {
			label := ""
			lookAhead := 10
			block := ""
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				block += lines[j] + " "
			}

			if m := semanticLabelRegex.FindStringSubmatch(block); m != nil {
				label = m[1]
			}

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-flutter-img-%d", filename, i),
				Role:  domain.RoleImage,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Image"),
					RawHTML:  trimmed,
				},
			})
		}

		// Match Buttons
		if (strings.Contains(trimmed, "Button") || strings.Contains(trimmed, "IconButton")) && 
			!strings.Contains(trimmed, "ButtonStyle") && !strings.Contains(trimmed, "import") {
			label := ""
			lookAhead := 10
			block := ""
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				block += lines[j] + " "
			}

			if m := textRegex.FindStringSubmatch(block); m != nil {
				label = m[1]
			} else if m := labelRegex.FindStringSubmatch(block); m != nil {
				label = m[1]
			}

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-flutter-btn-%d", filename, i),
				Role:  domain.RoleButton,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
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
